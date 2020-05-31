package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/stdlib"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func FQDN() (string, error) {
	stdout := bytes.NewBuffer(nil)
	cmd := exec.Command("hostname", "-f")
	cmd.Stdout = stdout

	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "could not look up FQDN")
	}

	return strings.TrimSpace(stdout.String()), nil
}

func ClientSet() (kubernetes.Interface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "could not get in cluster config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrap(err, "could not construct new client")
	}

	return clientset, nil
}

func SQLConn() (*sql.DB, error) {
	dsn := "postgres://root@localhost:26257/system"
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	return stdlib.OpenDB(*config), nil
}

func CurrentNodeID(ctx context.Context, conn *sql.DB, fqdn string, sqlPort uint) (string, error) {
	query := `SELECT node_id FROM crdb_internal.gossip_nodes WHERE advertise_address = $1`

	row := conn.QueryRowContext(ctx, query, fmt.Sprintf("%s:%d", fqdn, sqlPort))

	var nodeID string
	if err := row.Scan(&nodeID); err != nil {
		return "", err
	}

	return nodeID, nil
}

func IsBeingRemoved(sts *appsv1.StatefulSet, podName string) bool {
	index, err := strconv.Atoi(podName[len(sts.Name)+1:])
	if err != nil {
		panic(err)
	}

	return *sts.Spec.Replicas < int32(index)
}

type PostStopParams struct {
	FQDN            string
	PodName         string
	StatefulSetName string
	SQLPort         uint
	Namespace       string
}

func PostStop(ctx context.Context, conn *sql.DB, clientset kubernetes.Interface, params PostStopParams) error {
	sts, err := clientset.AppsV1().StatefulSets(params.Namespace).Get(params.StatefulSetName, metav1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "could not get statefulset %s", params.StatefulSetName)
	}

	// This pod is just being restarted, no need to drain
	if !IsBeingRemoved(sts, params.PodName) {
		return nil
	}

	nodeID, err := CurrentNodeID(ctx, conn, params.FQDN, params.SQLPort)
	if err != nil {
		return errors.Wrap(err, "could not get current node ID")
	}

	cmd := exec.Command("cockroach", "node", "decommission", nodeID, "--wait=all")
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	ctx := context.Background()

	if len(os.Args) != 2 {
		log.Fatalf("invalid invocation: %v", os.Args[1:])
	}

	conn, err := SQLConn()
	if err != nil {
		log.Fatalf("could not connect to DB: %v", err)
	}

	clientset, err := ClientSet()
	if err != nil {
		log.Fatalf("could not connect to kubernetes: %v", err)
	}

	switch os.Args[1] {
	case "post-stop":
		fqdn, err := FQDN()
		if err != nil {
			panic(err)
		}

		namespace, _ := os.LookupEnv("KUBERNETES_NAMESPACE")
		podName, _ := os.LookupEnv("KUBERNETES_POD")
		statefulsetName, _ := os.LookupEnv("KUBERNETES_STATEFULSET")

		params := PostStopParams{
			FQDN:            fqdn,
			PodName:         podName,
			StatefulSetName: statefulsetName,
			Namespace:       namespace,
			SQLPort:         26257,
		}

		if err := PostStop(ctx, conn, clientset, params); err != nil {
			log.Fatalf("failed to execute post-stop hook: %v", err)
		}
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}

	nodeName := os.Getenv("KUBERNETES_POD")
	if nodeName == "" {
		log.Fatal("KUBERNETES_NODE must be set")
	}
}

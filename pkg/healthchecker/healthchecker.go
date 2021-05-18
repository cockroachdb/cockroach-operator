/*
Copyright 2021 The Cockroach Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package healthchecker

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	underreplicatedmetric = "ranges_underreplicated{store=\"%v\"}"
	cmdunderreplicted     = "curl -Lk localhost:%s/_status/vars --silent | grep -i '%s'"
)

type HealthChecker interface { // for testing
	Probe(ctx context.Context, l logr.Logger, logSuffix string, partition int) error
}

type HealthCheckerImpl struct {
	clientset *kubernetes.Clientset
	scheme    *runtime.Scheme
	cluster   *resource.Cluster
	config    *rest.Config
}

func NewHealthChecker(cluster *resource.Cluster, clientset *kubernetes.Clientset, scheme *runtime.Scheme, config *rest.Config) *HealthCheckerImpl {
	return &HealthCheckerImpl{
		clientset: clientset,
		scheme:    scheme,
		cluster:   cluster,
		config:    config,
	}
}

func (s *HealthCheckerImpl) Probe(ctx context.Context, l logr.Logger, logSuffix string, nodeID int) error {
	l.V(int(zapcore.DebugLevel)).Info("Health check probe", "label", logSuffix, "nodeID", nodeID)
	stsname := s.cluster.StatefulSetName()
	stsnamespace := s.cluster.Namespace()
	podname := fmt.Sprintf("%s-%v", stsname, nodeID)

	//we sleep 1 minunte because the prestop hook is running and we need to wait
	//to finish the drain of the node
	time.Sleep(1 * time.Minute)

	//WaitUntilAllStsPodsAreReady waits for all pods from statefulset to be ready
	err := kube.WaitUntilAllStsPodsAreReady(ctx, s.clientset, l, stsname, stsnamespace, 30*time.Minute, 10*time.Minute)
	if err != nil {
		return err
	}

	sts, err := s.clientset.AppsV1().StatefulSets(stsnamespace).Get(ctx, stsname, metav1.GetOptions{})
	if err != nil {
		return kube.HandleStsError(err, l, stsname, stsnamespace)
	}

	//TODO: add goroutine for each partition
	for partition := *sts.Spec.Replicas - 1; partition >= 0; partition-- {
		//we check the underreplicated  metric for each pod from statefullset
		err = s.waitUntilUnderReplicatedMetricIsZero(ctx, l, logSuffix, podname, partition)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *HealthCheckerImpl) waitUntilUnderReplicatedMetricIsZero(ctx context.Context, l logr.Logger, logSuffix, podname string, partition int32) error {
	f := func() error {
		return s.checkUnderReplicatedMetric(ctx, l, logSuffix, podname, partition)
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 3 * time.Minute
	b.MaxInterval = 5 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "replicas check probe failed for cluster %s", logSuffix)
	}
	return nil
}

//checkUnderReplicatedMetric uses exec on the pod for now
func (s *HealthCheckerImpl) checkUnderReplicatedMetric(ctx context.Context, l logr.Logger, logSuffix, podname string, partition int32) error {
	l.V(int(zapcore.DebugLevel)).Info("checkUnderReplicatedMetric", "label", logSuffix, "podname", podname, "partition", partition)
	port := strconv.FormatInt(int64(*s.cluster.Spec().HTTPPort), 10)
	underrepmetric := fmt.Sprintf(underreplicatedmetric, partition)
	cmd := []string{
		fmt.Sprintf(cmdunderreplicted, port, underrepmetric),
	}
	l.V(int(zapcore.DebugLevel)).Info("get ranges_underreplicated metric", "node", podname)
	output, stderr, err := kube.ExecInPod(s.scheme, s.config, s.cluster.Namespace(),
		podname, resource.DbContainerName, cmd)
	if stderr != "" {
		return errors.Errorf("exec in pod %s failed with stderror: %s ", stderr)
	}
	if err != nil {
		return errors.Wrapf(err, "health check probe for pod %s failed", podname)
	}
	_, err = extractMetric(output, underrepmetric, partition)
	return err
}

func extractMetric(output, underepmetric string, partition int32) (int, error) {
	if output == "" {
		return -1, errors.Errorf("non existing ranges_underreplicated metric for partition %v", partition)
	}
	if strings.HasPrefix(output, underepmetric) {
		return -1, errors.Errorf("incorrect format of the output: actual='%s' expected to start with=%s", output, underepmetric)
	}
	out := strings.Split(output, " ")
	if out != nil && len(out) != 2 {
		return -1, errors.Errorf("incorrect format of the output: actual='%s' expected to start with=%s", output, underepmetric)
	}
	//the value of the metric should be 0 to return nil
	if i, err := strconv.Atoi(out[1]); err != nil {
		return -1, err
	} else if i > 0 {
		return -1, errors.Errorf("under replica is not zero for partition %v", partition)
	}
	return 0, nil
}

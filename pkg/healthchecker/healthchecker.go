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
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/scale"
	"github.com/cockroachdb/errors"
	"github.com/go-logr/logr"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const underreplicatedmetric = "ranges_underreplicated{store="

//HealthChecker interface
type HealthChecker interface { // for testing
	Probe(ctx context.Context, l logr.Logger, logSuffix string, partition int) error
}

//HealthCheckerImpl struct
type HealthCheckerImpl struct {
	clientset kubernetes.Interface
	scheme    *runtime.Scheme
	cluster   *resource.Cluster
	config    *rest.Config
}

//NewHealthChecker ctor
func NewHealthChecker(cluster *resource.Cluster, clientset kubernetes.Interface, scheme *runtime.Scheme, config *rest.Config) *HealthCheckerImpl {
	return &HealthCheckerImpl{
		clientset: clientset,
		scheme:    scheme,
		cluster:   cluster,
		config:    config,
	}
}

// Probe will check the ranges_underreplicated metric  for value 0 on all pods after the resart of a
// pod, before continue the rolling update of the next pod
func (hc *HealthCheckerImpl) Probe(ctx context.Context, l logr.Logger, logSuffix string, nodeID int) error {
	l.V(int(zapcore.DebugLevel)).Info("Health check probe", "label", logSuffix, "nodeID", nodeID)
	stsname := hc.cluster.StatefulSetName()
	stsnamespace := hc.cluster.Namespace()

	sts, err := hc.clientset.AppsV1().StatefulSets(stsnamespace).Get(ctx, stsname, metav1.GetOptions{})
	if err != nil {
		return kube.HandleStsError(err, l, stsname, stsnamespace)
	}

	if err := scale.WaitUntilStatefulSetIsReadyToServe(ctx, hc.clientset, stsnamespace, stsname, *sts.Spec.Replicas); err != nil {
		return errors.Wrapf(err, "error rolling update stategy on pod %d", nodeID)
	}

	// we check _status/vars on all cockroachdb pods looking for pairs like
	// ranges_underreplicated{store="1"} 0 and wait if any are non-zero until all are 0.
	// We can recheck every 10 seconds. We are waiting for this maximum 3 minutes
	err = hc.waitUntilUnderReplicatedMetricIsZero(ctx, l, logSuffix, stsname, stsnamespace, *sts.Spec.Replicas)
	if err != nil {
		return err
	}

	// we will wait 22 seconds and check again  _status/vars on all cockroachdb pods looking for pairs like
	// ranges_underreplicated{store="1"} 0. This time we do not wait anymore. This suplimentary check
	// is due to the fact that a node can be evicted in some cases
	time.Sleep(22 * time.Second)
	l.V(int(zapcore.DebugLevel)).Info("second wait loop for range_underreplicated metric", "label", logSuffix, "nodeID", nodeID)
	err = hc.waitUntilUnderReplicatedMetricIsZero(ctx, l, logSuffix, stsname, stsnamespace, *sts.Spec.Replicas)
	if err != nil {
		return err
	}
	return nil
}

//waitUntilUnderReplicatedMetricIsZero will check _status/vars on all cockroachdb pods looking for pairs like
//ranges_underreplicated{store="1"} 0 and wait if any are non-zero until all are 0.
func (hc *HealthCheckerImpl) waitUntilUnderReplicatedMetricIsZero(ctx context.Context, l logr.Logger, logSuffix, stsname, stsnamespace string, replicas int32) error {
	f := func() error {
		return hc.checkUnderReplicatedMetricAllPods(ctx, l, logSuffix, stsname, stsnamespace, replicas)
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 3 * time.Minute
	b.MaxInterval = 10 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "replicas check probe failed for cluster %s", logSuffix)
	}
	return nil
}

//checkUnderReplicatedMetric will make an http get call to _status/vars on a specific pod looking for pairs like
//ranges_underreplicated{store="1"} 0
func (hc *HealthCheckerImpl) checkUnderReplicatedMetric(ctx context.Context, l logr.Logger, logSuffix, podname, stsname, stsnamespace string, partition int32) error {
	l.V(int(zapcore.DebugLevel)).Info("checkUnderReplicatedMetric", "label", logSuffix, "podname", podname, "partition", partition)
	port := strconv.FormatInt(int64(*hc.cluster.Spec().HTTPPort), 10)
	url := fmt.Sprintf("https://%s.%s.%s:%s/_status/vars", podname, stsname, stsnamespace, port)

	runningInsideK8s := inK8s("/var/run/secrets/kubernetes.io/serviceaccount/token")

	var resp *http.Response
	var err error
	// Not running inside of Kubernetes so we need to use
	// the pod dialer
	if !runningInsideK8s {
		podDialer, err := kube.NewPodDialer(hc.config, stsnamespace)

		if err != nil {
			msg := "creating dialer failed"
			l.Error(err, msg)
			return errors.Wrap(err, msg)
		}
		tr := &http.Transport{
			Dial: podDialer.Dial,
			// When we generate the certs there is no CA we can
			// validate against
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client := http.Client{Transport: tr}
		resp, err = client.Get(url)
		if err != nil {
			msg := "health check failed, http get failed"
			l.Error(err, msg)
			return errors.Wrapf(err, msg)
		}
	} else {

		// When we generate the certs there is no CA we can
		// validate against
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		resp, err = http.Get(url)
		if err != nil {
			msg := "health check failed, http get failed"
			l.Error(err, msg)
			return errors.Wrapf(err, msg)
		}
	}

	defer resp.Body.Close()
	line, err := findLine(resp.Body)
	if err != nil {
		msg := "health check failed, error finding line in Body"
		l.Error(err, msg)
		return errors.Wrapf(err, msg)
	}

	if line == "" {
		msg := "health check failed, failed unable to find metric line in response body"
		l.Error(err, msg)
		return errors.Wrapf(err, msg)
	}

	metric, err := extractMetric(l, line, underreplicatedmetric, partition)
	l.V(int(zapcore.DebugLevel)).Info("after get ranges_underreplicated metric", "node", podname, "line", line, "metric", metric)
	return err
}

// findLine finds the line with the phrase "ranges_underreplicated{" in it
func findLine(r io.Reader) (string, error) {

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.Contains(text, "ranges_underreplicated{") {
			return text, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", nil
}

//checkUnderReplicatedMetric will check _status/vars on all cockroachdb pods looking for pairs like
//ranges_underreplicated{store="1"} 0
func (hc *HealthCheckerImpl) checkUnderReplicatedMetricAllPods(ctx context.Context, l logr.Logger, logSuffix, stsname, stsnamespace string, replicas int32) error {
	l.V(int(zapcore.DebugLevel)).Info("checkUnderReplicatedMetric", "label", logSuffix, "replicas", replicas)
	for partition := replicas - 1; partition >= 0; partition-- {
		podName := fmt.Sprintf("%s-%v", stsname, partition)
		if err := hc.checkUnderReplicatedMetric(ctx, l, logSuffix, podName, stsname, stsnamespace, partition); err != nil {
			return err
		}
	}

	return nil
}

//extractMetric gets the value of the ranges_underreplicated metric for the specific store
func extractMetric(l logr.Logger, output, underepmetric string, partition int32) (int, error) {
	l.V(int(zapcore.DebugLevel)).Info("extractMetric", "output", output, "underepmetric", underepmetric, "partition", partition)
	if output == "" {
		l.V(int(zapcore.DebugLevel)).Info("output is empty")
		return -1, errors.Errorf("non existing ranges_underreplicated metric for partition %v", partition)
	}
	if !strings.HasPrefix(output, underepmetric) {
		msg := fmt.Sprintf("incorrect format of the output: actual='%s' expected to start with=%s", output, underepmetric)
		l.V(int(zapcore.DebugLevel)).Info(msg)
		return -1, errors.New(msg)
	}
	out := strings.Split(output, " ")
	if out != nil && len(out) <= 1 {
		return -1, errors.Errorf("incorrect format of the output: actual='%s' expected to start with=%s", output, underepmetric)
	}
	metric := strings.TrimSuffix(out[1], "\n")
	//the value of the metric should be 0 to return nil
	if i, err := strconv.ParseFloat(metric, 1); err != nil {
		l.V(int(zapcore.DebugLevel)).Info(err.Error())
		return -1, err
	} else if i > 0 {
		l.V(int(zapcore.DebugLevel)).Info("Metric is greater than 0", "under_replicated", i)
		return -1, errors.Errorf("under replica is not zero for partition %v", partition)
	}
	return 0, nil
}

// inK8s checks to see if the a file exists
func inK8s(file string) bool {
	_, err := os.Stat(file)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

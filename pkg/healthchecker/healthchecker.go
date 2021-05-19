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
	"github.com/cockroachdb/cockroach-operator/pkg/scale"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	underreplicatedmetric = "ranges_underreplicated{store=\"%v\"}"
	//TODO: remove the svc.cluster.local
	cmdunderreplicted = "curl -ks https://%s.%s.%s.svc.cluster.local:%s/_status/vars | grep 'ranges_underreplicated{'"
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

	sts, err := s.clientset.AppsV1().StatefulSets(stsnamespace).Get(ctx, stsname, metav1.GetOptions{})
	if err != nil {
		return kube.HandleStsError(err, l, stsname, stsnamespace)
	}

	if err := scale.WaitUntilStatefulSetIsReadyToServe(ctx, s.clientset, stsnamespace, stsname, *sts.Spec.Replicas); err != nil {
		return errors.Wrapf(err, "error rolling update stategy on pod %d", nodeID)
	}

	// we check _status/vars on all cockroachdb pods looking for pairs like
	// ranges_underreplicated{store="1"} 0 and wait if any are non-zero until all are 0.
	// We can recheck every 10 seconds. We are waiting for this maximum 3 minutes
	err = s.waitUntilUnderReplicatedMetricIsZero(ctx, l, logSuffix, stsname, stsnamespace, *sts.Spec.Replicas)
	if err != nil {
		return err
	}

	// we will wait 22 seconds and check again  _status/vars on all cockroachdb pods looking for pairs like
	// ranges_underreplicated{store="1"} 0. This time we do not wait anymore. This suplimentary check
	// is due to the fact that a node can be evicted in some cases
	time.Sleep(22 * time.Second)

	err = s.waitUntilUnderReplicatedMetricIsZero(ctx, l, logSuffix, stsname, stsnamespace, *sts.Spec.Replicas)
	if err != nil {
		return err
	}
	return nil
}

//waitUntilUnderReplicatedMetricIsZero will check _status/vars on all cockroachdb pods looking for pairs like
//ranges_underreplicated{store="1"} 0 and wait if any are non-zero until all are 0.
func (s *HealthCheckerImpl) waitUntilUnderReplicatedMetricIsZero(ctx context.Context, l logr.Logger, logSuffix, stsname, stsnamespace string, replicas int32) error {
	f := func() error {
		return s.checkUnderReplicatedMetricAllPods(ctx, l, logSuffix, stsname, stsnamespace, replicas)
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 3 * time.Minute
	b.MaxInterval = 10 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "replicas check probe failed for cluster %s", logSuffix)
	}
	return nil
}

//checkUnderReplicatedMetric will check _status/vars on a specific pod looking for pairs like
//ranges_underreplicated{store="1"} 0
func (s *HealthCheckerImpl) checkUnderReplicatedMetric(ctx context.Context, l logr.Logger, logSuffix, podname, stsname, stsnamespace string, partition int32) error {
	l.V(int(zapcore.DebugLevel)).Info("checkUnderReplicatedMetric", "label", logSuffix, "podname", podname, "partition", partition)
	port := strconv.FormatInt(int64(*s.cluster.Spec().HTTPPort), 10)
	store := partition + 1
	underrepmetric := fmt.Sprintf(underreplicatedmetric, int(store))
	cmd := []string{
		"/bin/bash",
		"-c",
		fmt.Sprintf(cmdunderreplicted, podname, stsname, stsnamespace, port),
	}
	l.V(int(zapcore.DebugLevel)).Info("get ranges_underreplicated metric", "node", podname, "underrepmetric", underrepmetric, "cmd", cmd)
	output, stderr, err := kube.ExecInPod(s.scheme, s.config, s.cluster.Namespace(),
		podname, resource.DbContainerName, cmd)
	if stderr != "" {
		return errors.Errorf("exec in pod %s failed with stderror: %s ", stderr)
	}
	if err != nil {
		return errors.Wrapf(err, "health check probe for pod %s failed", podname)
	}
	metric, err := extractMetric(l, output, underrepmetric, partition)
	l.V(int(zapcore.DebugLevel)).Info("after get ranges_underreplicated metric", "node", podname, "output", output, "metric", metric)
	return err
}

//checkUnderReplicatedMetric will check _status/vars on all cockroachdb pods looking for pairs like
//ranges_underreplicated{store="1"} 0
func (s *HealthCheckerImpl) checkUnderReplicatedMetricAllPods(ctx context.Context, l logr.Logger, logSuffix, stsname, stsnamespace string, replicas int32) error {
	l.V(int(zapcore.DebugLevel)).Info("checkUnderReplicatedMetric", "label", logSuffix, "replicas", replicas)
	for partition := replicas - 1; partition >= 0; partition-- {
		podName := fmt.Sprintf("%s-%v", stsname, partition)
		if err := s.checkUnderReplicatedMetric(ctx, l, logSuffix, podName, stsname, stsnamespace, partition); err != nil {
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
	if i, err := strconv.Atoi(metric); err != nil {
		l.V(int(zapcore.DebugLevel)).Info(err.Error())
		return -1, err
	} else if i > 0 {
		l.V(int(zapcore.DebugLevel)).Info("Metric is greater than 0", "under_replicated", i)
		return -1, errors.Errorf("under replica is not zero for partition %v", partition)
	}
	return 0, nil
}

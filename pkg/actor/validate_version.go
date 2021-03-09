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

package actor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	api "github.com/cockroachdb/cockroach-operator/api/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/ptr"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	kubetypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func newVersionChecker(scheme *runtime.Scheme, cl client.Client, config *rest.Config) Actor {
	return &versionChecker{
		action: newAction("Crdb Version Validator", scheme, cl),
		config: config,
	}
}

// versionChecker performs the validation of the crdb image for the new cluster
type versionChecker struct {
	action

	config *rest.Config
}

//GetActionType returns api.VersionCheckerAction action used to set the cluster status errors
func (v *versionChecker) GetActionType() api.ActionType {
	return api.VersionCheckerAction
}

func (v *versionChecker) Handles(conds []api.ClusterCondition) bool {
	return utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator) && (condition.False(api.NotInitializedCondition, conds) || condition.True(api.NotInitializedCondition, conds)) && condition.True(api.CrdbVersionNotChecked, conds)
}

func (v *versionChecker) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := v.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.Info("version checker")

	r := resource.NewManagedKubeResource(ctx, v.client, cluster, kube.AnnotatingPersister)
	owner := cluster.Unwrap()
	// we check if the image tag version is supported by the operator,
	// this can return false only for api fieled CockroachDBVersion
	if !cluster.IsSupportedImage() {
		return InvalidContainerVersionError{Err: errors.New(fmt.Sprintf("crdb image %s not supported", cluster.Spec().CockroachDBVersion))}
	}
	var calVersion, containerImage string
	//reset the values of the annotation and make sure we will have the correct one
	cluster.SetClusterVersion(calVersion)
	cluster.SetAnnotationVersion(calVersion)
	cluster.SetCrdbContainerImage(containerImage)
	cluster.SetAnnotationContainerImage(containerImage)
	changed, err := (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.JobBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
		},
		Owner:  owner,
		Scheme: v.scheme,
	}).Reconcile()
	if err != nil && kube.IgnoreNotFound(err) == nil {
		return errors.Wrap(err, "failed to reconcile job")
	}

	if changed {
		log.Info("created/updated job, stopping request processing")
		CancelLoop(ctx)
		return nil
	}
	jobName := cluster.JobName()
	log.Info("version checker", "job", jobName)
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      jobName,
	}
	job := &kbatch.Job{}
	if err := v.client.Get(ctx, key, job); err != nil {
		return kube.IgnoreNotFound(err)
	}
	clientset, err := kubernetes.NewForConfig(v.config)
	if err != nil {
		return errors.Wrapf(err, "check version failed to create kubernetes clientset")
	}
	// check if the job is completed or failed before EXEC
	if finished, _ := isJobCompletedOrFailed(job); !finished {
		if err := WaitUntilJobPodIsRunning(ctx, clientset, job, v.log); err != nil {
			return errors.Wrapf(err, "failed to check the version of the cluster")
		}
		//get pod for the job we created
		pods, err := clientset.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set(job.Spec.Selector.MatchLabels).AsSelector().String(),
		})

		if err != nil {
			return errors.Wrapf(err, "failed to list running pod for job")
		}
		if len(pods.Items) == 0 {
			log.Info("No running pods yet for version checker... we will retry later")
			return nil
		}
		podName := pods.Items[0].Name
		log.Info("versionchecker", "jobPodName", podName)
		cmd := []string{
			"/bin/bash",
			"-c",
			resource.GetTagVersionCommand,
		}
		// exec on the pod of the job to obtain the version of the cockroach DB
		output, stderr, err := kube.ExecInPod(v.scheme, v.config, cluster.Namespace(),
			podName, resource.JobContainerName, cmd)
		log.Info("version checker result after exec in pod: ", "output", output)
		// if the container is running but the exec retrieved a stderr on our cmd
		if stderr != "" {
			// PermanentErr will requeue after 5 min
			log.Error(errors.New(stderr), "failure after exec in pod")
			return InvalidContainerVersionError{Err: errors.New(stderr)}
		}
		if err != nil {
			// can happen if container has not finished its startup
			if strings.Contains(err.Error(), "unable to upgrade connection: container not found") ||
				strings.Contains(err.Error(), "does not have a host assigned") {
				return NotReadyErr{Err: errors.New("pod has not completely started")}
			}
			return errors.Wrapf(err, "failed to check the version of the cluster")
		}

		// This is the value from Build Tag taken from the container
		calVersion = strings.Replace(output, "\n", "", -1)
		// if no image is retrieved we exit
		if calVersion == "" {
			return PermanentErr{Err: errors.New("failed to check the version of the cluster")}
		}

		// we check if the image tag version is supported by the operator
		if _, ok := cluster.LookupSupportedVersion(calVersion); !ok {
			return InvalidContainerVersionError{Err: errors.New(fmt.Sprintf("crdb version %s not supported ", calVersion))}
		}
		dbContainer, err := kube.FindContainer(resource.JobContainerName, &job.Spec.Template.Spec)
		if err != nil {
			return err
		}
		containerImage = dbContainer.Image
		if strings.EqualFold(cluster.GetVersionAnnotation(), calVersion) {
			log.Info("No update on version annotation -> nothing changed")
			return nil
		}
		if strings.EqualFold(cluster.GetAnnotationContainerImage(), containerImage) {
			log.Info("No update on container image annotation -> nothing changed")
			return nil
		}
		cluster.SetClusterVersion(calVersion)
		cluster.SetAnnotationVersion(calVersion)
		cluster.SetCrdbContainerImage(containerImage)
		cluster.SetAnnotationContainerImage(containerImage)
		if err := v.client.Update(ctx, cluster.Unwrap()); err != nil {
			log.Error(err, "failed saving the annotations on version checker")
		}
	} else {
		// after 2 minutes the pod enters in the completed state
		// if the container it is running and the version was not retrieved
		// for instance if the image is nginx and we want to get the crdb version from it case
		return PermanentErr{Err: errors.New("job completed with version empty-container running but no version was retrieved")}
	}
	dp := metav1.DeletePropagationForeground

	//delete the job only if we have managed to get the version and we do not have any errors
	err = clientset.BatchV1().Jobs(cluster.Namespace()).Delete(ctx, job.Name, metav1.DeleteOptions{
		GracePeriodSeconds: ptr.Int64(5),
		PropagationPolicy:  &dp,
	})
	if err != nil {
		log.Error(err, "failed to delete the job")
	}

	// we force the saving of the status on the cluster and cancel the loop
	fetcher := resource.NewKubeFetcher(ctx, cluster.Namespace(), v.client)

	cr := resource.ClusterPlaceholder(cluster.Name())
	if err := fetcher.Fetch(cr); err != nil {
		log.Error(err, "failed to retrieve CrdbCluster resource")
		return err
	}

	refreshedCluster := resource.NewCluster(cr)
	// save the status of the cluster
	refreshedCluster.SetFalse(api.CrdbVersionNotChecked)
	refreshedCluster.SetClusterVersion(calVersion)
	refreshedCluster.SetCrdbContainerImage(containerImage)
	if err := v.client.Status().Update(ctx, refreshedCluster.Unwrap()); err != nil {
		log.Error(err, "failed saving cluster status on version checker")
		return err
	}
	log.Info("completed version checker")
	CancelLoop(ctx)
	return nil
}

func isJobCompletedOrFailed(job *kbatch.Job) (bool, kbatch.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

func IsJobPodRunning(
	ctx context.Context,
	clientset kubernetes.Interface,
	job *kbatch.Job,
	l logr.Logger,
) error {
	//get pod for the job we created
	pods, err := clientset.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(job.Spec.Selector.MatchLabels).AsSelector().String(),
	})

	if err != nil {
		return err
	}
	if len(pods.Items) == 0 {
		l.Info("No running pods yet for version checker... we will retry later")
		return nil
	}
	l.Info("job pod is ready")
	return nil
}

func WaitUntilJobPodIsRunning(ctx context.Context, clientset kubernetes.Interface, job *kbatch.Job, l logr.Logger) error {
	f := func() error {
		return IsJobPodRunning(ctx, clientset, job, l)
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 120 * time.Second
	b.MaxInterval = 10 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "pod is not running for job: %s", job.Name)
	}
	return nil
}

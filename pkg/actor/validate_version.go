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
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cenkalti/backoff"
	api "github.com/cockroachdb/cockroach-operator/apis/v1alpha1"
	"github.com/cockroachdb/cockroach-operator/pkg/condition"
	"github.com/cockroachdb/cockroach-operator/pkg/features"
	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/ptr"
	"github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/cockroach-operator/pkg/utilfeature"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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

//Handles will return true if the conditions to run this action are satisfied
func (v *versionChecker) Handles(conds []api.ClusterCondition) bool {
	return utilfeature.DefaultMutableFeatureGate.Enabled(features.CrdbVersionValidator) && (condition.True(api.InitializedCondition, conds) || condition.False(api.InitializedCondition, conds)) && condition.False(api.CrdbVersionChecked, conds)
}

// Act will create a job that will try to extract the crdb version using the image field or the CockroachDBVersion field
// Initially version checker job was created to have a unique name, the same name as the CR
// This was the mechanism that assured a single job was created for each CR
// To run the tests in parallel we change the name of the job to use a timestamp
// We want to run different test scenarios in parallel with the same CR.
// But for the crdb container to run in our version checker we need also requestcerts action.
// The timestamp from the nameing of the job will ensure for a short period that the name of the job will be the same,
// We need a better mechanism to avoid creation of duplicate version checkers when the requestcerts finish
// and version checker needs to complete the work.
func (v *versionChecker) Act(ctx context.Context, cluster *resource.Cluster) error {
	log := v.log.WithValues("CrdbCluster", cluster.ObjectKey())
	log.V(DEBUGLEVEL).Info("starting to check the crdb version of the container provided")

	r := resource.NewManagedKubeResource(ctx, v.client, cluster, kube.AnnotatingPersister)
	owner := cluster.Unwrap()

	// If the image.name is set use that value and do not check that the
	// version is set in the supported versions.
	// If it is not set then pass through the if statement and check that
	if cluster.Spec().Image.Name == "" {
		if cluster.Spec().CockroachDBVersion == "" {
			err := ValidationError{Err: errors.New("Cockroach image name and cockroachDBVersion api fields are not set, you must set one of them")}
			log.Error(err, "invalid custom resources")
			return err
		}
		log.V(DEBUGLEVEL).Info("User set cockroachDBVersion")
		// we check if the cockroachDBVersion is supported by the operator,
		// this can return false only for api field CockroachDBVersion
		// The supported versions are set as enviroment variables in the operator manifest.
		if !cluster.IsSupportedImage() {
			err := ValidationError{Err: errors.New(fmt.Sprintf("crdb version %s not supported", cluster.Spec().CockroachDBVersion))}
			log.Error(err, "The cockroachDBVersion API value is set to a value that is not supported by the operator. Supported versions are set via the RELATED_IMAGE env variables in the operator manifest.")
			return err
		}
		log.V(DEBUGLEVEL).Info(fmt.Sprintf("supported CockroachDBVersion %s", cluster.Spec().CockroachDBVersion))
	} else {
		log.V(DEBUGLEVEL).Info("User set image.name, using that field instead of cockroachDBVersion")
	}

	var calVersion, containerImage string
	//reset the values of the annotation and make sure we will have the correct one
	cluster.SetClusterVersion(calVersion)
	cluster.SetAnnotationVersion(calVersion)
	cluster.SetCrdbContainerImage(containerImage)
	cluster.SetAnnotationContainerImage(containerImage)
	jobName := cluster.JobName()
	log.V(DEBUGLEVEL).Info(fmt.Sprintf("Reconcile jobName= %s", jobName))
	_, err := (resource.Reconciler{
		ManagedResource: r,
		Builder: resource.JobBuilder{
			Cluster:  cluster,
			Selector: r.Labels.Selector(),
			JobName:  jobName,
		},
		Owner:  owner,
		Scheme: v.scheme,
	}).Reconcile()

	if err != nil && kube.IgnoreNotFound(err) == nil {
		err := errors.Wrap(err, "failed to reconcile job not found")
		log.Error(err, "failed to reconcile job")
		return err
	} else if err != nil {
		log.Error(err, "failed to reconcile job only err")
	}
	// we comment this because after reconcile the object it is modified in k8s and
	// this will trigger a restart.
	// TODO: we need to delete this lines
	// if changed {
	// 	log.V(DEBUGLEVEL).Info("created/updated job, stopping request processing")
	// 	CancelLoop(ctx)
	// 	return nil
	// }
	log.V(DEBUGLEVEL).Info("version checker", "job", jobName)
	job := &kbatch.Job{}
	key := kubetypes.NamespacedName{
		Namespace: cluster.Namespace(),
		Name:      jobName,
	}

	clientset, err := kubernetes.NewForConfig(v.config)
	if err != nil {
		log.Error(err, "cannot create k8s client")
		return errors.Wrapf(err, "check version failed to create kubernetes clientset")
	}
	// sometimes the Get on the job will return an error and an empty job struct.
	// this is especially valid for kind cluster in our e2e
	// To fix this we wait for the pod of the job to run and than get the job.
	if err := waitUntilJobPodIsRunning(ctx, clientset, jobName, job.Namespace, r.Labels.Selector(), log); err != nil {
		log.Error(err, "job pod not found")
		return err
	}

	//we get the job after we checked that the pod is running
	if err := v.client.Get(ctx, key, job); err != nil {
		msg := fmt.Sprintf("failure: retrieved job%+v", *job)
		log.Error(err, msg)
		return err
	}

	// check if the job is completed or failed before getting the logs from the pod
	if finished, _ := isJobCompletedOrFailed(job); !finished {
		//we check first to see that the job pod it is running
		if err := waitUntilJobPodIsRunning(ctx, clientset, jobName, job.Namespace, r.Labels.Selector(), v.log); err != nil {
			// if after 2 minutes the job pod is not ready and container status is ImagePullBackoff
			// We need to stop requeueing until further changes on the CR
			image := cluster.GetCockroachDBImageName()
			if errBackoff := isContainerStatusImagePullBackoff(ctx, clientset, job, log, image); errBackoff != nil {
				err := InvalidContainerVersionError{Err: errBackoff}
				msg := "job image incorrect"
				log.V(DEBUGLEVEL).Info(msg)
				return errors.Wrapf(err, msg)
			}
			return errors.Wrapf(err, "failed to check the version of the crdb")
		}
		podLogOpts := corev1.PodLogOptions{}
		//get pod for the job we created
		pods, err := clientset.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
			LabelSelector: labels.Set(job.Spec.Selector.MatchLabels).AsSelector().String(),
		})

		if err != nil {
			log.Error(err, "failed to find running pod for job")
			return errors.Wrapf(err, "failed to list running pod for job")
		}
		if len(pods.Items) == 0 {
			log.V(DEBUGLEVEL).Info("No running pods yet for version checker... we will retry later")
			return nil
		}
		tmpPod := &pods.Items[0]
		// when we have more jobs take the latest in consideration
		if len(pods.Items) > 1 {
			for _, po := range pods.Items {
				if !po.CreationTimestamp.Before(&tmpPod.CreationTimestamp) {
					tmpPod = &po
				}
			}
		}
		podName := tmpPod.Name

		req := clientset.CoreV1().Pods(job.Namespace).GetLogs(podName, &podLogOpts)
		podLogs, err := req.Stream(ctx)
		if err != nil {
			msg := "error in opening stream"
			log.Error(err, msg)
			return errors.Wrapf(err, msg)
		}
		defer podLogs.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, podLogs)
		if err != nil {
			msg := "error in copy information from podLogs to buf"
			log.Error(err, msg)
			return errors.Wrapf(err, msg)
		}
		output := buf.String()

		// This is the value from Build Tag taken from the container
		calVersion = strings.Replace(output, "\n", "", -1)
		// if no image is retrieved we exit
		if calVersion == "" {
			err := PermanentErr{Err: errors.New("failed to check the version of the cluster")}
			log.Error(err, "crdb version not found")
			return err
		}

		// If the user has not set image.name then check if the calVersion is supported
		// We already check above that if image.name is not set then cockroachDBVersion is set.
		if cluster.Spec().Image.Name == "" {
			// we check if the image tag version is supported by the operator
			if _, ok := cluster.LookupSupportedVersion(calVersion); !ok {
				err := ValidationError{Err: errors.New(fmt.Sprintf("crdb version %s not supported ", calVersion))}
				log.Error(err, "crdb version not supported")
				return err
			}
		}

		dbContainer, err := kube.FindContainer(resource.JobContainerName, &job.Spec.Template.Spec)
		if err != nil {
			log.Error(err, "unable to find container version")
			return err
		}
		containerImage = dbContainer.Image
		if strings.EqualFold(cluster.GetVersionAnnotation(), calVersion) {
			log.V(DEBUGLEVEL).Info("No update on version annotation -> nothing changed")
			return nil
		}
		if strings.EqualFold(cluster.GetAnnotationContainerImage(), containerImage) {
			log.V(DEBUGLEVEL).Info("No update on container image annotation -> nothing changed")
			return nil
		}
		//we refresh the resource to make sure we use the latest version
		fetcher := resource.NewKubeFetcher(ctx, cluster.Namespace(), v.client)

		cr := resource.ClusterPlaceholder(cluster.Name())
		if err := fetcher.Fetch(cr); err != nil {
			log.Error(err, "failed to retrieve CrdbCluster resource")
			return err
		}

		refreshedCluster := resource.NewCluster(cr)
		refreshedCluster.SetClusterVersion(calVersion)
		refreshedCluster.SetAnnotationVersion(calVersion)
		refreshedCluster.SetCrdbContainerImage(containerImage)
		refreshedCluster.SetAnnotationContainerImage(containerImage)
		if err := v.client.Update(ctx, refreshedCluster.Unwrap()); err != nil {
			log.Error(err, "failed saving the annotations on version checker")
			// TODO should we fail here?
		}
	} else {
		// after 2 minutes the pod enters in the completed state
		// if the container it is running and the version was not retrieved
		// for instance if the image is nginx and we want to get the crdb version from it case
		err := PermanentErr{Err: errors.New("job completed with version empty-container running but no version was retrieved")}
		log.Error(err, "job completed and we cannot find crdb version")
		return err
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
	refreshedCluster.SetTrue(api.CrdbVersionChecked)
	refreshedCluster.SetClusterVersion(calVersion)
	refreshedCluster.SetCrdbContainerImage(containerImage)
	if err := v.client.Status().Update(ctx, refreshedCluster.Unwrap()); err != nil {
		log.Error(err, "failed saving cluster status on version checker")
		return err
	}
	log.V(DEBUGLEVEL).Info("completed version checker", "calVersion", calVersion, "containerImage", containerImage)
	CancelLoop(ctx)
	return nil
}

//isJobCompletedOrFailed checks if a job is in state completed of failed
func isJobCompletedOrFailed(job *kbatch.Job) (bool, kbatch.JobConditionType) {
	for _, c := range job.Status.Conditions {
		if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
			return true, c.Type
		}
	}
	return false, ""
}

//isJobPodRunning checks that the version checker pod it is in state running
func isJobPodRunning(
	ctx context.Context,
	clientset kubernetes.Interface,
	jobName, jobNamespace string,
	labelsSel map[string]string,
	l logr.Logger,
) error {
	labelSelector := metav1.LabelSelector{
		MatchLabels: labelsSel,
	}

	//get pod for the job we created
	pods, err := clientset.CoreV1().Pods(jobNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).AsSelector().String(),
	})
	if k8sErrors.IsNotFound(err) { // this is not an error
		l.V(DEBUGLEVEL).Info("cannot find pods for vcheck job", "jobName", jobName, "namespace", jobNamespace)
		return err
	} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus { // this is an error
		l.Error(statusError, fmt.Sprintf("status error getting pod %v", statusError.ErrStatus.Message))
		return err
	} else if err != nil {
		l.V(int(zapcore.ErrorLevel)).Info("error finding pods for vcheck job", "jobName", jobName, "namespace", jobNamespace)
		return err
	}

	if len(pods.Items) == 0 {
		l.V(DEBUGLEVEL).Info("job pods are not running yet waiting longer")
		return err
	}
	pod := pods.Items[0]
	if !kube.IsPodReady(&pod) {
		msg := "job pod is not ready yet waiting longer"
		l.V(DEBUGLEVEL).Info(msg)
		return errors.New(msg)
	}
	l.V(DEBUGLEVEL).Info("job pod is ready")
	return nil
}

//isContainerStatusImagePullBackoff checks that the container status is ImagePullBackOff
func isContainerStatusImagePullBackoff(
	ctx context.Context,
	clientset kubernetes.Interface,
	job *kbatch.Job,
	l logr.Logger,
	image string,
) error {
	//get pod for the job we created
	pods, err := clientset.CoreV1().Pods(job.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labels.Set(job.Spec.Selector.MatchLabels).AsSelector().String(),
	})

	if k8sErrors.IsNotFound(err) { // this is not an error
		l.V(DEBUGLEVEL).Info("cannot find pods for vcheck job", "jobName", job.ObjectMeta.Name, "namespace", job.Namespace)
		return err
	} else if statusError, isStatus := err.(*k8sErrors.StatusError); isStatus { // this is an error
		l.Error(statusError, fmt.Sprintf("status error getting pod %v", statusError.ErrStatus.Message))
		return err
	} else if err != nil {
		l.V(int(zapcore.ErrorLevel)).Info("error finding pods for vcheck job", "jobName", job.ObjectMeta.Name, "namespace", job.Namespace)
		return err
	}

	if len(pods.Items) == 0 {
		l.V(DEBUGLEVEL).Info("job pods are not running.")
		return nil
	}
	pod := pods.Items[0]
	if !kube.IsPodReady(&pod) && kube.IsImagePullBackOff(&pod, image) {
		l.V(DEBUGLEVEL).Info(fmt.Sprintf("Back-off pulling image %s", image))
		return nil
	}
	l.V(int(DEBUGLEVEL)).Info("job pod is ready")
	return nil
}

// waitUntilJobPodIsRunning will retry until the versionchecker pod it is running
func waitUntilJobPodIsRunning(ctx context.Context, clientset kubernetes.Interface, jobName, jobNamespace string, labelSelector map[string]string, l logr.Logger) error {
	if labelSelector == nil {
		return errors.New("selector cannot be nil")
	}
	f := func() error {
		return isJobPodRunning(ctx, clientset, jobName, jobNamespace, labelSelector, l)
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = 120 * time.Second
	b.MaxInterval = 10 * time.Second
	if err := backoff.Retry(f, b); err != nil {
		return errors.Wrapf(err, "pod is not running for job: %s", jobName)
	}
	return nil
}

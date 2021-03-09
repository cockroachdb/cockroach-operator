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

package resource

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/cockroach-operator/pkg/kube"
	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/ptr"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// JobContainerName used on spec for teh container
	JobContainerName     = "crdb"
	GetTagVersionCommand = "/cockroach/cockroach version | grep 'Build Tag:'| awk '{print $3}'"
)

type JobBuilder struct {
	*Cluster

	Selector labels.Labels
}

func (b JobBuilder) Build(obj runtime.Object) error {
	job, ok := obj.(*kbatch.Job)
	if !ok {
		return errors.New("failed to cast to Job object")
	}

	if job.ObjectMeta.Name == "" {
		job.ObjectMeta.Name = b.JobName()
	}

	// we recreate spec from ground only if we do not find the container job
	if dbContainer, err := kube.FindContainer(JobContainerName, &job.Spec.Template.Spec); err != nil {
		job.Spec = kbatch.JobSpec{
			// This field is alpha-level and is only honored by servers that enable the TTLAfterFinished feature.
			// see https://v1-18.docs.kubernetes.io/docs/reference/generated/kubernetes-api/v1.18/#job-v1-batch
			TTLSecondsAfterFinished: ptr.Int32(300),
			Template:                b.buildPodTemplate(),
		}
	} else {
		//if job with the container already exists we update the image only
		dbContainer.Image = b.GetCockroachDBImageName()
	}

	return nil
}
func (b JobBuilder) buildPodTemplate() corev1.PodTemplateSpec {
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.Selector,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: ptr.Int64(60),
			Containers:                    b.MakeContainers(),
			AutomountServiceAccountToken:  ptr.Bool(false),
			ServiceAccountName:            "cockroach-database-sa",
			RestartPolicy:                 corev1.RestartPolicyNever,
		},
	}

	secret := b.Spec().Image.PullSecret
	if secret != nil {
		local := corev1.LocalObjectReference{
			Name: *secret,
		}
		pod.Spec.ImagePullSecrets = []corev1.LocalObjectReference{local}
	}

	return pod
}

// MakeContainers creates a slice of corev1.Containers which includes a single
// corev1.Container that is based on the CR.
func (b JobBuilder) MakeContainers() []corev1.Container {
	return []corev1.Container{
		{
			Name:            JobContainerName,
			Image:           b.GetCockroachDBImageName(),
			ImagePullPolicy: *b.Spec().Image.PullPolicyName,
			Resources:       b.Spec().Resources,
			Command:         []string{"/bin/bash"},
			Args:            []string{"-c", fmt.Sprintf("%s; sleep 120", GetTagVersionCommand)},
		},
	}
}
func (b JobBuilder) Placeholder() runtime.Object {
	return &kbatch.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.JobName(),
		},
	}
}

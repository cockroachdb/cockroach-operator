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
	"os"
	"strings"

	"github.com/cockroachdb/cockroach-operator/pkg/labels"
	"github.com/cockroachdb/cockroach-operator/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	httpPortName = "http"
	grpcPortName = "grpc"

	dataDirName      = "datadir"
	dataDirMountPath = "/cockroach/cockroach-data/"

	certsDirName = "certs"

	DbContainerName = "db"
	RHEnvVar        = "RELATED_IMAGE_COCKROACH"
)

type StatefulSetBuilder struct {
	*Cluster

	Selector labels.Labels
}

func (b StatefulSetBuilder) Build(obj runtime.Object) error {
	ss, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return errors.New("failed to cast to StatefulSet object")
	}

	if ss.ObjectMeta.Name == "" {
		ss.ObjectMeta.Name = b.StatefulSetName()
	}

	ss.Spec = appsv1.StatefulSetSpec{
		ServiceName: b.Cluster.DiscoveryServiceName(),
		Replicas:    ptr.Int32(b.Spec().Nodes),
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{},
		},
		PodManagementPolicy: appsv1.OrderedReadyPodManagement,
		Selector: &metav1.LabelSelector{
			MatchLabels: b.Selector,
		},
		Template: b.makePodTemplate(),
	}

	if err := b.Spec().DataStore.Apply(dataDirName, DbContainerName, dataDirMountPath, &ss.Spec,
		func(name string) metav1.ObjectMeta {
			return metav1.ObjectMeta{
				Name: dataDirName,
			}
		}); err != nil {
		return err
	}

	if b.Spec().TLSEnabled {
		if err := addCertsVolumeMount(DbContainerName, &ss.Spec.Template.Spec); err != nil {
			return err
		}

		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: certsDirName,
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					DefaultMode: ptr.Int32(0400),
					Sources: []corev1.VolumeProjection{
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: b.nodeTLSSecretName(),
								},
								Items: []corev1.KeyToPath{
									{
										Key:  "ca.crt",
										Path: "ca.crt",
									},
									{
										Key:  corev1.TLSCertKey,
										Path: "node.crt",
									},
									{
										Key:  corev1.TLSPrivateKeyKey,
										Path: "node.key",
									},
								},
							},
						},
						{
							Secret: &corev1.SecretProjection{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: b.clientTLSSecretName(),
								},
								Items: []corev1.KeyToPath{
									{
										Key:  corev1.TLSCertKey,
										Path: "client.root.crt",
									},
									{
										Key:  corev1.TLSPrivateKeyKey,
										Path: "client.root.key",
									},
								},
							},
						},
					},
				},
			},
		})
	}

	return nil
}

func (b StatefulSetBuilder) Placeholder() runtime.Object {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.StatefulSetName(),
		},
	}
}

func (b StatefulSetBuilder) makePodTemplate() corev1.PodTemplateSpec {
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.Selector,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: ptr.Int64(60),
			Containers:                    b.MakeContainers(),
			AutomountServiceAccountToken:  ptr.Bool(false),
			ServiceAccountName:            "cockroach-database-sa",
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
func (b StatefulSetBuilder) MakeContainers() []corev1.Container {

	//
	// This code block allows for RedHat to override the coachroach image name during
	// openshift testing.  They need to set the image name dynamically using a environment
	// variable to allow the testing of a specific image.
	//
	image := os.Getenv(RHEnvVar)
	if image == "" {
		image = b.Spec().Image.Name
	}

	return []corev1.Container{
		{
			Name:            DbContainerName,
			Image:           image,
			ImagePullPolicy: *b.Spec().Image.PullPolicyName,
			Resources:       b.Spec().Resources,
			Command:         []string{"/cockroach/cockroach"},
			Args:            b.dbArgs(),
			Env: []corev1.EnvVar{
				{
					Name:  "COCKROACH_CHANNEL",
					Value: "kubernetes-operator",
				},
				{
					Name: "POD_NAME",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.name",
						},
					},
				},
			},
			Ports: []corev1.ContainerPort{
				{
					Name:          grpcPortName,
					ContainerPort: *b.Spec().GRPCPort,
					Protocol:      corev1.ProtocolTCP,
				},
				{
					Name:          httpPortName,
					ContainerPort: *b.Spec().HTTPPort,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			LivenessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/health",
						Port:   intstr.FromString(httpPortName),
						Scheme: b.probeScheme(),
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       5,
			},
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/health?ready=1",
						Port:   intstr.FromString(httpPortName),
						Scheme: b.probeScheme(),
					},
				},
				InitialDelaySeconds: 10,
				PeriodSeconds:       5,
				FailureThreshold:    2,
			},
		},
	}
}

func (b StatefulSetBuilder) secureMode() string {
	if b.Spec().TLSEnabled {
		return " --certs-dir=/cockroach/cockroach-certs/"
	}

	return " --insecure"
}

func (b StatefulSetBuilder) probeScheme() corev1.URIScheme {
	if b.Spec().TLSEnabled {
		return corev1.URISchemeHTTPS
	}

	return corev1.URISchemeHTTP
}

func (b StatefulSetBuilder) nodeTLSSecretName() string {
	if b.Spec().NodeTLSSecret == "" {
		return b.Cluster.NodeTLSSecretName()
	}

	return b.Spec().NodeTLSSecret
}

func (b StatefulSetBuilder) clientTLSSecretName() string {
	if b.Spec().NodeTLSSecret == "" {
		return b.Cluster.ClientTLSSecretName()
	}

	return b.Spec().ClientTLSSecret
}

func (b StatefulSetBuilder) dbArgs() []string {
	aa := []string{
		"start",
		"--join=" + b.joinStr(),
		fmt.Sprintf("--advertise-host=$(POD_NAME).%s.%s",
			b.Cluster.DiscoveryServiceName(), b.Cluster.Namespace()),
		"--logtostderr=INFO",
		b.Cluster.SecureMode(),
		"--http-port=" + fmt.Sprint(*b.Spec().HTTPPort),
		"--port=" + fmt.Sprint(*b.Spec().GRPCPort),
		"--cache=" + b.Spec().Cache,
		"--max-sql-memory=" + b.Spec().MaxSQLMemory,
	}

	return append(aa, b.Spec().AdditionalArgs...)
}

func (b StatefulSetBuilder) joinStr() string {
	var seeds []string

	for i := 0; i < int(b.Spec().Nodes) && i < 3; i++ {
		seeds = append(seeds, fmt.Sprintf("%s-%d.%s.%s:%d", b.Cluster.StatefulSetName(), i,
			b.Cluster.DiscoveryServiceName(), b.Cluster.Namespace(), *b.Cluster.Spec().GRPCPort))
	}

	return strings.Join(seeds, ",")
}

func addCertsVolumeMount(container string, spec *corev1.PodSpec) error {
	found := false
	for i := range spec.Containers {
		c := &spec.Containers[i]
		if c.Name == container {
			found = true

			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      certsDirName,
				MountPath: "/cockroach/cockroach-certs/",
			})
			break
		}
	}

	if !found {
		return fmt.Errorf("failed to find container %s to attach volume", container)
	}

	return nil
}

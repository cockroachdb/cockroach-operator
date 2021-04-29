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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	httpPortName = "http"
	grpcPortName = "grpc"
	sqlPortName  = "sql"

	DataDirName      = "datadir"
	DataDirMountPath = "/cockroach/cockroach-data/"

	certsDirName = "certs"
	emptyDirName = "emptydir"

	DbContainerName = "db"
	certCpCmd       = ">- cp -p /cockroach/cockroach-certs-prestage/..data/* /cockroach/cockroach-certs/ && chmod 700 /cockroach/cockroach-certs/*.key && chown 10001:10001 /cockroach/cockroach-certs/*.key"
)

type StatefulSetBuilder struct {
	*Cluster

	Selector  labels.Labels
	Telemetry string
}

func (b StatefulSetBuilder) Build(obj client.Object) error {
	ss, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return errors.New("failed to cast to StatefulSet object")
	}
	if ss.ObjectMeta.Name == "" {
		ss.ObjectMeta.Name = b.StatefulSetName()
	}
	if ss.Annotations == nil {
		ss.Annotations = make(map[string]string)
	}
	ss.Annotations[CrdbVersionAnnotation] = b.Cluster.GetVersionAnnotation()
	ss.Annotations[CrdbContainerImageAnnotation] = b.Cluster.GetAnnotationContainerImage()
	ss.Spec = appsv1.StatefulSetSpec{
		ServiceName: b.Cluster.DiscoveryServiceName(),
		Replicas:    ptr.Int32(b.Spec().Nodes),
		UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{},
		},
		PodManagementPolicy: appsv1.ParallelPodManagement,
		Selector: &metav1.LabelSelector{
			MatchLabels: b.Selector,
		},
		Template: b.makePodTemplate(),
	}

	if err := b.Spec().DataStore.Apply(DataDirName, DbContainerName, DataDirMountPath, &ss.Spec,
		func(name string) metav1.ObjectMeta {
			return metav1.ObjectMeta{
				Name: DataDirName,
			}
		}); err != nil {
		return err
	}

	if b.Spec().TLSEnabled {
		if err := addCertsVolumeMountOnInitContiners(DbContainerName, &ss.Spec.Template.Spec); err != nil {
			return err
		}
		if err := addCertsVolumeMount(DbContainerName, &ss.Spec.Template.Spec); err != nil {
			return err
		}

		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: emptyDirName,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			}})

		ss.Spec.Template.Spec.Volumes = append(ss.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: certsDirName,
			VolumeSource: corev1.VolumeSource{
				Projected: &corev1.ProjectedVolumeSource{
					DefaultMode: ptr.Int32(400),
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
										Mode: ptr.Int32(504),
									},
									{
										Key:  corev1.TLSCertKey,
										Path: "node.crt",
										Mode: ptr.Int32(504),
									},
									{
										Key:  corev1.TLSPrivateKeyKey,
										Path: "node.key",
										Mode: ptr.Int32(400),
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
										Mode: ptr.Int32(504),
									},
									{
										Key:  corev1.TLSPrivateKeyKey,
										Path: "client.root.key",
										Mode: ptr.Int32(400),
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

func (b StatefulSetBuilder) Placeholder() client.Object {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.StatefulSetName(),
		},
	}
}

func (b StatefulSetBuilder) SetAnnotations(obj client.Object) error {
	ss, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return errors.New("failed to cast to StatefulSet object")
	}
	ss.Annotations[CrdbVersionAnnotation] = b.Cluster.Status().Version
	ss.Annotations[CrdbContainerImageAnnotation] = b.Cluster.Status().CrdbContainerImage
	timeNow := metav1.Now()
	if val, ok := ss.Annotations[CrdbHistoryAnnotation]; !ok {
		ss.Annotations[CrdbHistoryAnnotation] = fmt.Sprintf("%s:%s", timeNow.String(), b.Cluster.Status().Version)
	} else {
		ss.Annotations[CrdbHistoryAnnotation] = fmt.Sprintf("%s %s:%s", val, timeNow.String(), b.Cluster.Status().Version)
	}
	return nil
}

func (b StatefulSetBuilder) makePodTemplate() corev1.PodTemplateSpec {
	pod := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.Selector,
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser: ptr.Int64(10001),
				FSGroup:   ptr.Int64(10001),
			},
			TerminationGracePeriodSeconds: ptr.Int64(60),
			InitContainers:                b.MakeInitContainers(),
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

// MakeInitContainers creates a slice of corev1.Containers which includes a single
// corev1.Container that is based on the CR.
func (b StatefulSetBuilder) MakeInitContainers() []corev1.Container {
	image := b.GetCockroachDBImageName()
	initContainer := fmt.Sprintf("%s-init", DbContainerName)
	return []corev1.Container{
		{
			Name:            initContainer,
			Image:           image,
			Command:         []string{"/bin/sh", "-c", certCpCmd},
			ImagePullPolicy: *b.Spec().Image.PullPolicyName,
			SecurityContext: &corev1.SecurityContext{
				RunAsUser:                ptr.Int64(0),
				AllowPrivilegeEscalation: ptr.Bool(false),
			},
		},
	}
}

// MakeContainers creates a slice of corev1.Containers which includes a single
// corev1.Container that is based on the CR.
func (b StatefulSetBuilder) MakeContainers() []corev1.Container {
	image := b.GetCockroachDBImageName()

	return []corev1.Container{
		{
			Name:            DbContainerName,
			Image:           image,
			ImagePullPolicy: *b.Spec().Image.PullPolicyName,
			Resources:       b.Spec().Resources,
			Command:         b.commandArgs(),
			Env:             b.envVars(),
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
				{
					Name:          sqlPortName,
					ContainerPort: *b.Spec().SQLPort,
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

func (b StatefulSetBuilder) commandArgs() []string {
	exec := "exec " + strings.Join(b.dbArgs(), " ")
	return []string{"/bin/bash", "-ecx", exec}
}

func (b StatefulSetBuilder) dbArgs() []string {
	aa := []string{
		"/cockroach/cockroach.sh",
		"start",
		"--join=" + b.joinStr(),
		fmt.Sprintf("--advertise-host=$(POD_NAME).%s.%s",
			b.Cluster.DiscoveryServiceName(), b.Cluster.Namespace()),
		"--logtostderr=INFO",
		b.Cluster.SecureMode(),
		"--http-port=" + fmt.Sprint(*b.Spec().HTTPPort),
		"--sql-addr=:" + fmt.Sprint(*b.Spec().SQLPort),
		"--listen-addr=:" + fmt.Sprint(*b.Spec().GRPCPort),
	}

	if b.Spec().Cache != "" {
		aa = append(aa, "--cache="+b.Spec().Cache)
	} else {
		aa = append(aa, "--cache $(expr $MEMORY_LIMIT_MIB / 4)MiB")
	}

	if b.Spec().MaxSQLMemory != "" {
		aa = append(aa, "--max-sql-memory="+b.Spec().MaxSQLMemory)
	} else {
		aa = append(aa, "--max-sql-memory $(expr $MEMORY_LIMIT_MIB / 4)MiB")
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
func addCertsVolumeMountOnInitContiners(container string, spec *corev1.PodSpec) error {
	found := false
	initContainer := fmt.Sprintf("%s-init", container)
	for i := range spec.InitContainers {
		c := &spec.InitContainers[i]
		if c.Name == initContainer {
			found = true

			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      certsDirName,
				MountPath: "/cockroach/cockroach-certs-prestage/",
			})
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      emptyDirName,
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

func addCertsVolumeMount(container string, spec *corev1.PodSpec) error {
	found := false
	for i := range spec.Containers {
		c := &spec.Containers[i]
		if c.Name == container {
			found = true

			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      emptyDirName,
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

var CRDB_PREFIX string = "CRDB_"

func (b StatefulSetBuilder) envVars() []corev1.EnvVar {
	values := make([]corev1.EnvVar, 0)

	one := resource.MustParse("1")
	oneMi := resource.MustParse("1Mi")

	// append the POD_NAME and the COCKROACH_CHANNEL values
	values = append(values,
		// set the telemetry
		// You can disable the telemetry by setting
		// CRDB_COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING=true
		// in the operator manifest.
		// Or set COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING=true
		// using the podEnvVariables stanza in the CRD
		corev1.EnvVar{
			Name:  "COCKROACH_CHANNEL",
			Value: b.Telemetry,
		},
		corev1.EnvVar{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		// values for used to calc --cache and --max-sql-memory
		// these values do exist in the CRD and the user can
		// override them.
		corev1.EnvVar{
			Name: "GOMAXPROCS",
			ValueFrom: &corev1.EnvVarSource{
				ResourceFieldRef: &corev1.ResourceFieldSelector{
					Resource: "limits.cpu",
					Divisor:  one,
				},
			},
		},
		corev1.EnvVar{
			Name: "MEMORY_LIMIT_MIB",
			ValueFrom: &corev1.EnvVarSource{
				ResourceFieldRef: &corev1.ResourceFieldSelector{
					Resource: "limits.memory",
					Divisor:  oneMi,
				},
			},
		},
	)

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.HasPrefix(pair[0], CRDB_PREFIX) {
			key := strings.ReplaceAll(pair[0], CRDB_PREFIX, "")
			env := corev1.EnvVar{
				Name:  key,
				Value: pair[1],
			}
			values = append(values, env)
		}
	}

	if len(b.Cluster.Spec().PodEnvVariables) != 0 {
		values = append(values, b.Cluster.Spec().PodEnvVariables...)
	}

	return values
}

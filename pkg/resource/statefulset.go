package resource

import (
	"errors"
	"fmt"
	"github.com/cockroachlabs/crdb-operator/pkg/labels"
	"github.com/cockroachlabs/crdb-operator/pkg/ptr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"
)

const (
	httpPortName = "http"
	grpcPortName = "grpc"

	dataDirName      = "datadir"
	dataDirMountPath = "/cockroach/cockroach-data/"

	certsDirName = "certs"

	DbContainerName = "db"
)

type StatefulSetBuilder struct {
	*Cluster

	Selector labels.Labels
}

func (b StatefulSetBuilder) Build(obj runtime.Object) error {
	ss, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return errors.New("failed to access StatefulSet object")
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
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: b.Selector,
		},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: ptr.Int64(60),
			Containers:                    b.makeContainers(),
			ServiceAccountName:            b.ServiceAccountName(),
		},
	}
}

func (b StatefulSetBuilder) makeContainers() []corev1.Container {
	return []corev1.Container{
		{
			Name:            DbContainerName,
			Image:           b.Spec().Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Resources:       b.Spec().Resources,
			Args:            b.dbArgs(),
			Env:             b.env(),
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
			Lifecycle: &corev1.Lifecycle{
				// PostStart: &corev1.Handler{
				// 	Exec: &corev1.ExecAction{
				// 		Command: []string{
				// 			"shell",
				// 			"-ecx",
				// 			">- exec /cockroach/cockroach-util post-stop",
				// 		},
				// 	},
				// },
				PreStop: &corev1.Handler{
					Exec: &corev1.ExecAction{
						Command: []string{
							"shell",
							"-ecx",
							">- exec /cockroach/cockroach-util post-stop",
						},
					},
				},
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
		"shell",
		"-ecx",
		">- exec /cockroach/cockroach start" +
			" --join=" + b.joinStr() +
			" --advertise-host=$(hostname -f)" +
			" --logtostderr=INFO" +
			b.Cluster.SecureMode() +
			" --http-port=" + fmt.Sprint(*b.Spec().HTTPPort) +
			" --port=" + fmt.Sprint(*b.Spec().GRPCPort) +
			" --cache=" + b.Spec().Cache +
			" --max-sql-memory=" + b.Spec().MaxSQLMemory,
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

func (b StatefulSetBuilder) env() []corev1.EnvVar {
	env := []corev1.EnvVar{
		{
			Name:  "COCKROACH_CHANNEL",
			Value: "kubernetes-operator",
		},
		{
			Name:  "KUBERNETES_STATEFULSET",
			Value: b.StatefulSetName(),
		},
		{
			Name: "KUBERNETES_POD",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "KUBERNETES_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
	}

	if b.Spec().TLSEnabled {
		env = append(env, corev1.EnvVar{
			Name:  "COCKROACH_CERTS_DIR",
			Value: "/cockroach/cockroach-certs/",
		})
	} else {
		env = append(env, corev1.EnvVar{
			Name:  "COCKROACH_INSECURE",
			Value: "TRUE",
		})
	}

	return env
}

func addCertsVolumeMount(container string, spec *corev1.PodSpec) error {
	found := false
	for i, _ := range spec.Containers {
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

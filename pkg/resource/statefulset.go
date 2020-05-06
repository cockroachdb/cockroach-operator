package resource

import (
	"fmt"
	api "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/labels"
	"github.com/cockroachlabs/crdb-operator/pkg/ptr"
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
)

func NewStatefulSetBuilder(cluster *Cluster, name string, nodes int32, join string, locality string, nodeSelector map[string]string) StatefulSetBuilder {
	return StatefulSetBuilder{
		Cluster:         cluster,
		StatefulSetName: name,
		Nodes:           nodes,
		Selector:        labels.Common(cluster.Unwrap()).Selector(),
		NodeSelector:    nodeSelector,
		JoinStr:         join,
		Locality:        locality,
	}
}

type StatefulSetBuilder struct {
	*Cluster

	StatefulSetName string
	Nodes           int32
	NodeSelector    map[string]string
	Selector        labels.Labels
	JoinStr         string
	Locality        string
}

func (b StatefulSetBuilder) Build() (runtime.Object, error) {
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.StatefulSetName,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: b.Cluster.DiscoveryServiceName(),
			Replicas:    ptr.Int32(b.Nodes),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{},
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: b.Selector,
			},
			Template: b.makePodTemplate(),
		},
	}

	if err := b.Spec().DataStore.Apply(dataDirName, DbContainerName, dataDirMountPath, &ss.Spec,
		func(name string) metav1.ObjectMeta {
			return metav1.ObjectMeta{
				Name: dataDirName,
			}
		}); err != nil {
		return nil, err
	}

	if b.Spec().TLSEnabled {
		if err := addCertsVolumeMount(DbContainerName, &ss.Spec.Template.Spec); err != nil {
			return nil, err
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

	return ss, nil
}

func (b StatefulSetBuilder) Placeholder() runtime.Object {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: b.StatefulSetName,
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
			NodeSelector:                  b.NodeSelector,
			Containers:                    b.makeContainers(),
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
			Env: []corev1.EnvVar{
				{
					Name: "COCKROACH_CHANNEL",
					// TODO(vladdy): should be custom
					// ༼∵༽ ༼⍨༽ ༼⍢༽ ༼⍤༽
					Value: "kubernetes-helm",
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

func (b StatefulSetBuilder) localityOrNothing() string {
	if b.Locality == "" {
		return ""
	}

	return " --locality=" + b.Locality
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
	if b.Spec().NodeTLSSecret == api.NodeTLSSecretKeyword {
		return b.Cluster.NodeTLSSecretName()
	}

	return b.Spec().NodeTLSSecret
}

func (b StatefulSetBuilder) clientTLSSecretName() string {
	if b.Spec().NodeTLSSecret == api.NodeTLSSecretKeyword {
		return b.Cluster.ClientTLSSecretName()
	}

	return b.Spec().ClientTLSSecret
}

func (b StatefulSetBuilder) dbArgs() []string {
	aa := []string{
		"shell",
		"-ecx",
		">- exec /cockroach/cockroach start" +
			b.localityOrNothing() +
			" --join=" + b.JoinStr +
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

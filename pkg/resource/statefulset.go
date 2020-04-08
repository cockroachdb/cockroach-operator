package resource

import (
	"fmt"
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

	DbContainerName = "db"
)

func NewStatefulSetBuilder(cluster *Cluster, name string, nodes int32, join string, locality string, nodeSelector map[string]string) StatefulSetBuilder {
	return StatefulSetBuilder{
		Cluster:      cluster,
		Name:         name,
		Nodes:        nodes,
		Selector:     labels.Common(cluster.Unwrap()).Selector(),
		NodeSelector: nodeSelector,
		JoinStr:      join,
		Locality:     locality,
	}
}

type StatefulSetBuilder struct {
	*Cluster

	Name         string
	Nodes        int32
	NodeSelector map[string]string
	Selector     labels.Labels
	JoinStr      string
	Locality     string
}

func (b StatefulSetBuilder) Build() (runtime.Object, error) {
	ss := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   b.Name,
			Labels: map[string]string{},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: b.Name,
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

	return ss, nil
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
			NodeSelector:                  b.nodeSelectorOrNil(),
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
			Args: []string{
				"shell",
				"-ecx",
				">- exec /cockroach/cockroach start" +
					b.localityOrNothing() +
					" --join=" + b.JoinStr +
					" --advertise-host=$(hostname -f)" +
					" --logtostderr=INFO" +
					" --insecure" +
					" --http-port=" + fmt.Sprint(*b.Spec().HTTPPort) +
					" --port=" + fmt.Sprint(*b.Spec().GRPCPort) +
					" --cache=25%" +
					" --max-sql-memory=25%",
			},
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
						Path: "/health",
						Port: intstr.FromString(httpPortName),
					},
				},
				InitialDelaySeconds: 30,
				PeriodSeconds:       5,
			},
			ReadinessProbe: &corev1.Probe{
				Handler: corev1.Handler{
					HTTPGet: &corev1.HTTPGetAction{
						Path: "/health?ready=1",
						Port: intstr.FromString(httpPortName),
					},
				},
				InitialDelaySeconds: 10,
				PeriodSeconds:       5,
				FailureThreshold:    2,
			},
		},
	}
}

func (b StatefulSetBuilder) nodeSelectorOrNil() map[string]string {
	if b.NodeSelector == nil {
		return nil
	}

	selector := b.Selector.Copy()
	selector.Merge(b.NodeSelector)

	return selector.AsMap()
}

func (b StatefulSetBuilder) localityOrNothing() string {
	if b.Locality == "" {
		return ""
	}

	return " --locality=" + b.Locality
}

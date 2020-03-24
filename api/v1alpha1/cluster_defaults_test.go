package v1alpha1

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestSetClusterSpecDefaults(t *testing.T) {
	s := &CrdbClusterSpec{}

	expected := &CrdbClusterSpec{
		GrpcPort: &DefaultGrpcPort,
		HttpPort: &DefaultHttpPort,
		NodesSpec: &appsv1.StatefulSetSpec{
			ServiceName: "cockroachdb",
			Replicas:    &DefaultReplicasNum,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{},
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{},
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &DefaultTerminationGracePeriodSeconds,
					Containers: []corev1.Container{
						{
							Name:            "db",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"shell",
								"-ecx",
								">- exec /cockroach/cockroach " +
									"start --join ${STATEFULSET_NAME}-0.${STATEFULSET_FQDN}:26257 " +
									"--advertise-host=$(hostname).${STATEFULSET_FQDN} " +
									"--disable-cluster-name-verification " +
									"--logtostderr=INFO " +
									"--insecure " +
									"--http-port=8080 " +
									"--port=26257 " +
									"--cache=25%" +
									"--max-disk-temp-storage=0" +
									"--max-offset=500ms" +
									"--max-sql-memory=25%",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "STATEFULSET_NAME",
									Value: "cockroachdb",
								},
								{
									Name:  "STATEFULSET_FQDN",
									Value: "cockroachdb.default.svc.cluster.local",
								},
								{
									Name: "COCKROACH_CHANNEL",
									Value: "kubernetes-helm",
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "grpc",
									ContainerPort: 26257,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "http",
									ContainerPort: 8080,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/cockroach/cockroach-data/",
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromString("http"),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       5,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health?ready=1",
										Port: intstr.FromString("http"),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								FailureThreshold:    2,
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "datadir",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}

	SetClusterSpecDefaults(s)

	diff := cmp.Diff(expected, s)
	if diff != "" {
		assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
	}
}

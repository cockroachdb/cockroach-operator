package resource_test

import (
	"fmt"
	"github.com/cockroachlabs/crdb-operator/pkg/labels"
	"github.com/cockroachlabs/crdb-operator/pkg/ptr"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestStatefulSetBuilder(t *testing.T) {
	cluster := testutil.NewBuilder("test-cluster").Namespaced("test-ns").
		WithEmptyDirDataStore().WithNodeCount(1)
	commonLabels := labels.Common(cluster.Cr())

	tests := []struct {
		name     string
		cluster  *resource.Cluster
		selector map[string]string
		expected *appsv1.StatefulSet
	}{
		{
			name:     "builds default insecure statefulset",
			cluster:  cluster.Cluster(),
			selector: commonLabels.Selector(),
			expected: insecureOneNode(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := resource.StatefulSetBuilder{
				Cluster:  tt.cluster,
				Name:     "test-cluster",
				Nodes:    1,
				Selector: tt.selector,
				NodeSelector: map[string]string{
					"failure-domain.beta.kubernetes.io/zone": "zone-a",
				},
				JoinStr:  "test-cluster-0.test-cluster.test-ns:26257",
				Locality: "",
			}.Build()
			require.NoError(t, err)

			diff := cmp.Diff(tt.expected, actual, testutil.RuntimeObjCmpOpts...)
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}

func insecureOneNode() *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "test-cluster",
			Labels: map[string]string{},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "test-cluster",
			Replicas:    ptr.Int32(1),
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{},
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "cockroachdb",
					"app.kubernetes.io/instance":  "test-cluster",
					"app.kubernetes.io/component": "database",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":      "cockroachdb",
						"app.kubernetes.io/instance":  "test-cluster",
						"app.kubernetes.io/component": "database",
					},
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: ptr.Int64(60),
					NodeSelector: map[string]string{
						"app.kubernetes.io/name":                 "cockroachdb",
						"app.kubernetes.io/instance":             "test-cluster",
						"app.kubernetes.io/component":            "database",
						"failure-domain.beta.kubernetes.io/zone": "zone-a",
					},
					Containers: []corev1.Container{
						{
							Name:            "db",
							Image:           "cockroachdb/cockroach:v19.2.6",
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"shell",
								"-ecx",
								">- exec /cockroach/cockroach start" +
									" --join=test-cluster-0.test-cluster.test-ns:26257" +
									" --advertise-host=$(hostname -f)" +
									" --logtostderr=INFO" +
									" --insecure" +
									" --http-port=8080" +
									" --port=26257" +
									" --cache=25%" +
									" --max-sql-memory=25%",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "COCKROACH_CHANNEL",
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
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/cockroach/cockroach-data/",
								},
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
}

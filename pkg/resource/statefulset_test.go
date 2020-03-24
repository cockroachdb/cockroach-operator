package resource_test

import (
	"context"
	"errors"
	"fmt"
	crdbv1alpha1 "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	optesting "github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

func TestStatefulSet_Reconcile(t *testing.T) {
	ctx := context.TODO()
	cluster := optesting.MakeStandAloneCluster(t)
	scheme := optesting.InitScheme(t)

	var (
		expectedReplicasNum            int32 = 1
		expectedTerminationGracePeriod int64 = 60
	)

	expected := &appsv1.StatefulSet{
		ObjectMeta: v1.ObjectMeta{
			Name:      "cockroachdb",
			Namespace: cluster.Namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "cockroachdb",
				"app.kubernetes.io/instance":   "test-ns/test-cluster",
				"app.kubernetes.io/version":    "v19.2",
				"app.kubernetes.io/component":  "database",
				"app.kubernetes.io/part-of":    "cockroachdb",
				"app.kubernetes.io/managed-by": "crdb-operator",
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: "cockroachdb",
			Replicas:    &expectedReplicasNum,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{},
			},
			PodManagementPolicy: appsv1.OrderedReadyPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name":      "cockroachdb",
					"app.kubernetes.io/instance":  "test-ns/test-cluster",
					"app.kubernetes.io/component": "database",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name":      "cockroachdb",
						"app.kubernetes.io/instance":  "test-ns/test-cluster",
						"app.kubernetes.io/component": "database",
					},
				},
				Spec: corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									Weight: 100,
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"app.kubernetes.io/name":      "cockroachdb",
												"app.kubernetes.io/instance":  "test-ns/test-cluster",
												"app.kubernetes.io/component": "database",
											},
										},
										TopologyKey: "kubernetes.io/hostname",
									},
								},
							},
						},
					},
					TerminationGracePeriodSeconds: &expectedTerminationGracePeriod,
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

	tests := []struct {
		name             string
		clusterSpec      *crdbv1alpha1.CrdbCluster
		preexistingObjs  []runtime.Object
		expected         *appsv1.StatefulSet
		wantErr          string
		reactionInjector func(*optesting.FakeClient)
	}{
		{
			name:        "statefulset definition can't be retrieved",
			clusterSpec: cluster,
			expected:    nil,
			wantErr:     "failed to fetch statefulset: test-ns/cockroachdb: internal error",
			reactionInjector: func(c *optesting.FakeClient) {
				c.AddReactor(
					"get",
					"statefulsets",
					func(action optesting.Action) (bool, error) {
						return true, errors.New("internal error")
					})
			},
		},
		{
			name:        "service resource does not exist",
			clusterSpec: cluster,
			expected:    expected,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := optesting.NewFakeClient(scheme, tt.preexistingObjs...)
			if tt.reactionInjector != nil {
				tt.reactionInjector(client)
			}

			actual, err := (&resource.StatefulSet{Cluster: tt.clusterSpec}).Reconcile(ctx, client)
			if tt.wantErr != "" {
				assert.Nil(t, actual)
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err, "failed to reconcile due: %v", err)

			diff := cmp.Diff(tt.expected, actual, cmpopts.IgnoreTypes(metav1.TypeMeta{}))
			if diff != "" {
				assert.Fail(t, fmt.Sprintf("unexpected result (-want +got):\n%v", diff))
			}
		})
	}
}

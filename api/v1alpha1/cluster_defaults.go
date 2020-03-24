package v1alpha1

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	DefaultReplicasNum                   int32 = 1
	DefaultTerminationGracePeriodSeconds int64 = 60
	DefaultGrpcPort                      int32 = 26257
	DefaultHttpPort                      int32 = 8080
)

func SetClusterSpecDefaults(cs *CrdbClusterSpec) {
	if cs.GrpcPort == nil {
		cs.GrpcPort = &DefaultGrpcPort
	}

	if cs.HttpPort == nil {
		cs.HttpPort = &DefaultHttpPort
	}

	setNodesSpecDefaults(cs)
}

func setNodesSpecDefaults(cs *CrdbClusterSpec) {
	if cs.NodesSpec == nil {
		cs.NodesSpec = &appsv1.StatefulSetSpec{}
	}

	ns := cs.NodesSpec

	if ns.ServiceName == "" {
		ns.ServiceName = "cockroachdb"
	}

	if ns.Replicas == nil {
		ns.Replicas = &DefaultReplicasNum
	}

	if ns.UpdateStrategy.Type == appsv1.OnDeleteStatefulSetStrategyType {
		ns.UpdateStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
	}

	if ns.UpdateStrategy.RollingUpdate == nil {
		ns.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{}
	}

	if ns.PodManagementPolicy == "" {
		ns.PodManagementPolicy = appsv1.OrderedReadyPodManagement
	}

	if ns.Selector == nil {
		ns.Selector = &metav1.LabelSelector{
			MatchLabels: map[string]string{},
		}
	}

	if ns.Template.ObjectMeta.Labels == nil {
		ns.Template.ObjectMeta.Labels = map[string]string{}
	}

	podSpec := &ns.Template.Spec

	if podSpec.TerminationGracePeriodSeconds == nil {
		podSpec.TerminationGracePeriodSeconds = &DefaultTerminationGracePeriodSeconds
	}

	if podSpec.Affinity == nil {
		podSpec.Affinity = &corev1.Affinity{
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
		}
	}

	if len(podSpec.Containers) == 0 {
		podSpec.Containers = append(podSpec.Containers, corev1.Container{
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
					// TODO: should be customizable
					// ༼∵༽ ༼⍨༽ ༼⍢༽ ༼⍤༽
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
		})

		podSpec.Volumes = []corev1.Volume{
			{
				Name: "datadir",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			},
		}
	}
}

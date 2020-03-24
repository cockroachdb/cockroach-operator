package resource

import (
	"context"
	"fmt"
	crdbv1alpha1 "github.com/cockroachlabs/crdb-operator/api/v1alpha1"
	"github.com/cockroachlabs/crdb-operator/pkg/label"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DiscoveryService struct {
	Cluster *crdbv1alpha1.CrdbCluster
}

func (s *DiscoveryService) Reconcile(ctx context.Context, cl client.Client) (*corev1.Service, error) {
	existing := &corev1.Service{}

	key := s.makeNamespacedName()
	if err := cl.Get(ctx, key, existing); client.IgnoreNotFound(err) != nil {
		return nil, errors.Wrapf(err, "failed to fetch discovery service: %s", key)
	}

	desired := s.makeDesired(key)

	if equalService(existing, desired) {
		return nil, nil
	}

	return desired, nil
}

func (s *DiscoveryService) makeDesired(nn types.NamespacedName) *corev1.Service {
	meta := metav1.ObjectMeta{
		Name:        nn.Name,
		Namespace:   nn.Namespace,
		Labels:      label.MakeCommonLabels(s.Cluster),
		Annotations: s.makeMonitoringAnnotations(),
	}

	service := &corev1.Service{
		ObjectMeta: meta,
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			PublishNotReadyAddresses: true,
			Ports: []corev1.ServicePort{
				{Name: "grpc", Port: *s.Cluster.Spec.GrpcPort},
				{Name: "http", Port: *s.Cluster.Spec.HttpPort},
			},
			Selector: map[string]string{
				label.ComponentLabelKey: meta.Labels[label.ComponentLabelKey],
				label.InstanceLabelKey:  meta.Labels[label.InstanceLabelKey],
				label.NameLabelKey:      meta.Labels[label.NameLabelKey],
			},
		},
	}

	// Use this annotation in addition to the actual field below because the
	// annotation will stop being respected soon, but the field is broken in
	// some versions of Kubernetes:
	// https://github.com/kubernetes/kubernetes/issues/58662
	service.ObjectMeta.Labels["service.alpha.kubernetes.io/tolerate-unready-endpoints"] = "true"

	return service
}

func (s *DiscoveryService) makeNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      "cockroachdb",
		Namespace: s.Cluster.GetNamespace(),
	}
}

func (s *DiscoveryService) makeMonitoringAnnotations() map[string]string {
	return map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/path":   "_status/vars",
		"prometheus.io/port":   fmt.Sprint(*(s.Cluster.Spec.HttpPort)),
	}
}

func equalService(l *corev1.Service, r *corev1.Service) bool {
	portsCmpOpts := []cmp.Option{
		cmpopts.SortSlices(func(a, b corev1.ServicePort) bool { return a.Port < b.Port }),
	}

	return cmp.Equal(l.ObjectMeta.Labels, r.ObjectMeta.Labels) &&
		cmp.Equal(l.ObjectMeta.Annotations, r.ObjectMeta.Annotations) &&
		cmp.Equal(l.Spec.Ports, r.Spec.Ports, portsCmpOpts...) &&
		cmp.Equal(l.Spec, r.Spec, cmpopts.IgnoreTypes(corev1.ServicePort{}))
}

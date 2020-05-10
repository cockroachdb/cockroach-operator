package resource_test

import (
	"context"
	"github.com/cockroachlabs/crdb-operator/pkg/kube"
	"github.com/cockroachlabs/crdb-operator/pkg/resource"
	"github.com/cockroachlabs/crdb-operator/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestLoadTLSSecret(t *testing.T) {
	ctx := context.TODO()
	scheme := testutil.InitScheme(t)
	fakeClient := testutil.NewFakeClient(scheme)
	r := resource.NewKubeResource(ctx, fakeClient, "test-namespace", kube.DefaultPersister)

	_, err := resource.LoadTLSSecret("non-existing", r)
	assert.True(t, apierrors.IsNotFound(err))
}

func TestTLSSecretReady(t *testing.T) {
	ctx := context.TODO()
	scheme := testutil.InitScheme(t)
	name := "test-secret"
	namespace := "test-namespace"

	tests := []struct {
		name     string
		secret   runtime.Object
		expected bool
	}{
		{
			name: "secret missing required fields",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"someKey": {},
				},
			},
			expected: false,
		},
		{
			name: "secret has all required fields",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
				},
				Data: map[string][]byte{
					"ca.crt":  {},
					"tls.crt": {},
					"tls.key": {},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := testutil.NewFakeClient(scheme, tt.secret)
			r := resource.NewKubeResource(ctx, fakeClient, namespace, kube.DefaultPersister)

			actual, err := resource.LoadTLSSecret(name, r)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, actual.Ready())

		})
	}
}

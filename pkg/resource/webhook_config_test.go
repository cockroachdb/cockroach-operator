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
package resource_test

import (
	"context"
	"testing"

	. "github.com/cockroachdb/cockroach-operator/pkg/resource"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/admissionregistration/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestConfigureMutatingWebhook(t *testing.T) {
	name := "mutating-webhook-configuration"

	tests := []struct {
		name   string
		config v1.MutatingWebhookConfiguration
		err    error
	}{
		{
			name: "valid definition found",
			config: v1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Webhooks: []v1.MutatingWebhook{
					{Name: "mcrdbcluster.kb.io"},
				},
			},
		},
		{
			name: "config not found",
			err:  &apiErrors.StatusError{},
		},
		{
			name: "webhook not defined",
			config: v1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
			},
			err: &ErrWebhookNotFound{},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		api := fake.NewSimpleClientset().AdmissionregistrationV1().MutatingWebhookConfigurations()

		if tt.config.Name != "" {
			_, err := api.Create(ctx, &tt.config, metav1.CreateOptions{})
			require.NoError(t, err, tt.name)
		}

		err := ConfigureMutatingWebhook(ctx, api, []byte("PEM ENCODED CERT"))
		if tt.err != nil {
			require.IsType(t, tt.err, errors.Cause(err), tt.name)
			continue
		}

		require.NoError(t, err, tt.name)

		cfg, err := api.Get(ctx, name, metav1.GetOptions{})
		require.NoError(t, err, tt.name)
		require.Equal(t, []byte("PEM ENCODED CERT"), cfg.Webhooks[0].ClientConfig.CABundle)
	}
}

func TestConfigureValidatingWebhook(t *testing.T) {
	name := "validating-webhook-configuration"

	tests := []struct {
		name   string
		config v1.ValidatingWebhookConfiguration
		err    error
	}{
		{
			name: "valid definition found",
			config: v1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
				Webhooks: []v1.ValidatingWebhook{
					{Name: "vcrdbcluster.kb.io"},
				},
			},
		},
		{
			name: "config not found",
			err:  &apiErrors.StatusError{},
		},
		{
			name: "webhook not defined",
			config: v1.ValidatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{Name: name},
			},
			err: &ErrWebhookNotFound{},
		},
	}

	for _, tt := range tests {
		ctx := context.Background()
		api := fake.NewSimpleClientset().AdmissionregistrationV1().ValidatingWebhookConfigurations()

		if tt.config.Name != "" {
			_, err := api.Create(ctx, &tt.config, metav1.CreateOptions{})
			require.NoError(t, err, tt.name)
		}

		err := ConfigureValidatingWebhook(ctx, api, []byte("PEM ENCODED CERT"))
		if tt.err != nil {
			require.IsType(t, tt.err, errors.Cause(err), tt.name)
			continue
		}

		require.NoError(t, err, tt.name)

		cfg, err := api.Get(ctx, name, metav1.GetOptions{})
		require.NoError(t, err, tt.name)
		require.Equal(t, []byte("PEM ENCODED CERT"), cfg.Webhooks[0].ClientConfig.CABundle)
	}
}

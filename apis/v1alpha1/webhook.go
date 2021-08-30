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

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// DefaultGRPCPort is the default port used for GRPC communication
	DefaultGRPCPort int32 = 26258
	// DefaultSQLPort is the default port used for SQL connections
	DefaultSQLPort int32 = 26257
	// DefaultHTTPPort is the default port for the Web UI
	DefaultHTTPPort int32 = 8080
	// DefaultMaxUnavailable is the default max unavailable nodes during a rollout
	DefaultMaxUnavailable int32 = 1
)

var (
	// log is for logging in this package.
	webhookLog = logf.Log.WithName("webhooks")

	// this just ensures that we've implemented the interface
	_ webhook.Defaulter = &CrdbCluster{}
	_ webhook.Validator = &CrdbCluster{}
)

// SetupWebhookWithManager ensures webhooks are enabled for the CrdbCluster resource.
func (r *CrdbCluster) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(r).Complete()
}

//+kubebuilder:webhook:path=/mutate-crdb-cockroachlabs-com-v1alpha1-crdbcluster,mutating=true,failurePolicy=fail,groups=crdb.cockroachlabs.com,resources=crdbclusters,verbs=create;update,versions=v1alpha1,name=mcrdbcluster.kb.io,sideEffects=None,admissionReviewVersions={v1,v1beta1}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *CrdbCluster) Default() {
	webhookLog.Info("default", "name", r.Name)

	if r.Spec.GRPCPort == nil {
		r.Spec.GRPCPort = &DefaultGRPCPort
	}

	if r.Spec.SQLPort == nil {
		r.Spec.SQLPort = &DefaultSQLPort
	}

	if r.Spec.HTTPPort == nil {
		r.Spec.HTTPPort = &DefaultHTTPPort
	}

	if r.Spec.MaxUnavailable == nil && r.Spec.MinAvailable == nil {
		r.Spec.MaxUnavailable = &DefaultMaxUnavailable
	}

	if r.Spec.Image.PullPolicyName == nil {
		policy := v1.PullIfNotPresent
		r.Spec.Image.PullPolicyName = &policy
	}
}

//+kubebuilder:webhook:path=/validate-crdb-cockroachlabs-com-v1alpha1-crdbcluster,mutating=false,failurePolicy=fail,groups=crdb.cockroachlabs.com,resources=crdbclusters,verbs=create;update,versions=v1alpha1,name=vcrdbcluster.kb.io,sideEffects=None,admissionReviewVersions={v1,v1beta1}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *CrdbCluster) ValidateCreate() error {
	webhookLog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *CrdbCluster) ValidateUpdate(old runtime.Object) error {
	webhookLog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *CrdbCluster) ValidateDelete() error {
	webhookLog.Info("validate delete", "name", r.Name)

	// we're not validating anything on delete. This is just a placeholder for now to satisfy the Validator interface
	return nil
}

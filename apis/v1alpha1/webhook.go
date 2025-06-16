/*
Copyright 2025 The Cockroach Authors

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
	"fmt"
	"reflect"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var (
	// DefaultGRPCAddr is the default grpc address used for GRPC communication
	DefaultGRPCAddr string = ":26258"
	// DefaultSQLAddr is the default sql address used for SQL connections
	DefaultSQLAddr string = ":26257"
	// DefaultHTTPAddr is the default http address for the Web UI
	DefaultHTTPAddr string = ":8080"
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

// +kubebuilder:webhook:path=/mutate-crdb-cockroachlabs-com-v1alpha1-crdbcluster,mutating=true,failurePolicy=fail,groups=crdb.cockroachlabs.com,resources=crdbclusters,verbs=create;update,versions=v1alpha1,name=mcrdbcluster.kb.io,sideEffects=None,admissionReviewVersions=v1

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *CrdbCluster) Default() {
	webhookLog.Info("default", "name", r.Name)

	if r.Spec.GRPCPort == nil && r.Spec.ListenAddr == nil {
		r.Spec.ListenAddr = &DefaultGRPCAddr
	} else if r.Spec.GRPCPort != nil && r.Spec.ListenAddr == nil {
		listenAddr := fmt.Sprintf(":%d", *r.Spec.GRPCPort)
		r.Spec.ListenAddr = &listenAddr
		r.Spec.GRPCPort = nil
	}

	if r.Spec.SQLPort == nil && r.Spec.SQLAddr == nil {
		r.Spec.SQLAddr = &DefaultSQLAddr
	} else if r.Spec.SQLPort != nil && r.Spec.SQLAddr == nil {
		sqlAddr := fmt.Sprintf(":%d", *r.Spec.SQLPort)
		r.Spec.SQLAddr = &sqlAddr
		r.Spec.SQLPort = nil
	}

	if r.Spec.HTTPPort == nil && r.Spec.HTTPAddr == nil {
		r.Spec.HTTPAddr = &DefaultHTTPAddr
	} else if r.Spec.HTTPPort != nil && r.Spec.HTTPAddr == nil {
		httpAddr := fmt.Sprintf(":%d", *r.Spec.HTTPPort)
		r.Spec.HTTPAddr = &httpAddr
		r.Spec.HTTPPort = nil
	}

	if r.Spec.MaxUnavailable == nil && r.Spec.MinAvailable == nil {
		r.Spec.MaxUnavailable = &DefaultMaxUnavailable
	}

	if r.Spec.Image != nil && r.Spec.Image.PullPolicyName == nil {
		policy := v1.PullIfNotPresent
		r.Spec.Image.PullPolicyName = &policy
	}
}

// +kubebuilder:webhook:path=/validate-crdb-cockroachlabs-com-v1alpha1-crdbcluster,mutating=false,failurePolicy=fail,groups=crdb.cockroachlabs.com,resources=crdbclusters,verbs=create;update,versions=v1alpha1,name=vcrdbcluster.kb.io,sideEffects=None,admissionReviewVersions=v1

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *CrdbCluster) ValidateCreate() error {
	webhookLog.Info("validate create", "name", r.Name)
	var errors []error
	if r.Spec.Ingress != nil {
		if err := r.ValidateIngress(); err != nil {
			errors = append(errors, err...)
		}
	}

	if err := r.ValidateCockroachVersion(); err != nil {
		errors = append(errors, err)
	}

	if err := r.ValidateVolumeMode(); err != nil {
		errors = append(errors, err)
	}

	if len(errors) != 0 {
		return kerrors.NewAggregate(errors)
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *CrdbCluster) ValidateUpdate(old runtime.Object) error {
	webhookLog.Info("validate update", "name", r.Name)
	var errors []error

	oldCluster, ok := old.(*CrdbCluster)
	if !ok {
		webhookLog.Info(fmt.Sprintf("unexpected old cluster type %T", old))
	} else {
		// Validate if labels changed.
		// k8s does not support changing selector/labels on sts:
		//  https://github.com/kubernetes/kubernetes/issues/90519.
		if !reflect.DeepEqual(oldCluster.Spec.AdditionalLabels, r.Spec.AdditionalLabels) {
			errors = append(errors, fmt.Errorf("mutating additionalLabels field is not supported"))
		}
	}

	if r.Spec.Ingress != nil {
		if err := r.ValidateIngress(); err != nil {
			errors = append(errors, err...)
		}
	}

	if err := r.ValidateCockroachVersion(); err != nil {
		errors = append(errors, err)
	}

	if len(errors) != 0 {
		return kerrors.NewAggregate(errors)
	}

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *CrdbCluster) ValidateDelete() error {
	webhookLog.Info("validate delete", "name", r.Name)

	// we're not validating anything on delete. This is just a placeholder for now to satisfy the Validator interface
	return nil
}

// ValidateIngress validates the ingress configuration used to create ingress resource
func (r *CrdbCluster) ValidateIngress() (errors []error) {
	webhookLog.Info("validate ingress", "name", r.Name)

	if r.Spec.Ingress.UI == nil && r.Spec.Ingress.SQL == nil {
		errors = append(errors, fmt.Errorf("at least one of UI or SQL ingresses must be present"))
	}

	if r.Spec.Ingress.UI != nil && r.Spec.Ingress.UI.Host == "" {
		errors = append(errors, fmt.Errorf("host required for UI"))
	}

	if r.Spec.Ingress.SQL != nil && r.Spec.Ingress.SQL.Host == "" {
		errors = append(errors, fmt.Errorf("host required for SQL"))
	}

	return
}

// ValidateCockroachVersion validates the cockroachdb version or image provided
func (r *CrdbCluster) ValidateCockroachVersion() error {
	if r.Spec.CockroachDBVersion == "" && (r.Spec.Image == nil || r.Spec.Image.Name == "") {
		return fmt.Errorf("you have to provide the cockroachDBVersion or cockroach image")
	} else if r.Spec.CockroachDBVersion != "" && (r.Spec.Image != nil && r.Spec.Image.Name != "") {
		return fmt.Errorf("you have provided both cockroachDBVersion and cockroach image, please provide only one")
	}

	return nil
}

// ValidateVolumeMode validates that the volumeMode of the pvc is set to filesystem.
func (r *CrdbCluster) ValidateVolumeMode() error {
	claim := r.Spec.DataStore.VolumeClaim
	if claim == nil {
		return nil
	}
	if claim.PersistentVolumeClaimSpec.VolumeMode == nil {
		return fmt.Errorf("you have not provided pvc.volumeMode value.")
	}
	if *claim.PersistentVolumeClaimSpec.VolumeMode != v1.PersistentVolumeFilesystem {
		return fmt.Errorf(
			"you have provided unsupported pvc.volumeMode, currently only Filesystem is supported.")
	}
	return nil
}

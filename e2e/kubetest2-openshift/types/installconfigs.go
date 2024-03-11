/*
Copyright 2024 The Cockroach Authors

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

package types

import (
	"github.com/cockroachdb/cockroach-operator/e2e/kubetest2-openshift/types/gcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// These structs are based on the struct in the openshift installer code
// https://github.com/openshift/installer/blob/3cee5f9d844e773fe102c53d7f50186566a93165/pkg/types/installconfig.go#L1
// Importing that huge code base is not a great idea, but this
// struct will require maintenance if there code changes.
// This struct is used to create the YAML file that the openshift-installer
// uses to create a cluster, and if the format of that YAML file changes
// the installer will fail.

// InstallConfig is the configuration for an OpenShift install.
type InstallConfig struct {

	// APIVersion of the YAML file
	APIVersion string `yaml:"apiVersion"`

	metav1.ObjectMeta `yaml:"metadata"`

	// BaseDomain is the base domain to which the cluster should belong.
	BaseDomain string `yaml:"baseDomain"`

	// PullSecret is the secret to use when pulling images.
	PullSecret string `yaml:"pullSecret"`

	// Platform is the configuration for the specific platform upon which to
	// perform the installation.
	Platform `yaml:"platform"`
}

// We are only supporting GCP at this time, but the installer supports
// multiple platforms, so I have left the commented code here
// as a starting place for the other platforms if we decide to
// support them.

// Platform is the configuration for the specific platform upon which to perform
// the installation. Only one of the platform configuration should be set.
type Platform struct {
	// AWS is the configuration used when installing on AWS.
	//AWS *aws.Platform `yaml:"aws,omitempty"`

	// Azure is the configuration used when installing on Azure.
	//Azure *azure.Platform `yaml:"azure,omitempty"`

	// BareMetal is the configuration used when installing on bare metal.
	//BareMetal *baremetal.Platform `yaml:"baremetal,omitempty"`

	// GCP is the configuration used when installing on Google Cloud Platform.
	// +optional
	GCP *gcp.Platform `yaml:"gcp,omitempty"`

	// Libvirt is the configuration used when installing on libvirt.
	//Libvirt *libvirt.Platform `yaml:"libvirt,omitempty"`

	// None is the empty configuration used when installing on an unsupported
	// platform.
	//None *none.Platform `yaml:"none,omitempty"`

	// OpenStack is the configuration used when installing on OpenStack.
	//OpenStack *openstack.Platform `yaml:"openstack,omitempty"`

	// VSphere is the configuration used when installing on vSphere.
	//VSphere *vsphere.Platform `yaml:"vsphere,omitempty"`

	// Ovirt is the configuration used when installing on oVirt.
	//Ovirt *ovirt.Platform `yaml:"ovirt,omitempty"`

	// Kubevirt is the configuration used when installing on kubevirt.
	//Kubevirt *kubevirt.Platform `yaml:"kubevirt,omitempty"`
}

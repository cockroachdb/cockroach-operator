/*
Copyright 2026 The Cockroach Authors

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

package deployer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach-operator/e2e/kubetest2-openshift/types"
	"github.com/cockroachdb/cockroach-operator/e2e/kubetest2-openshift/types/gcp"
	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/process"
)

var home = os.Getenv("HOME")

// Up creates a OpenShift cluster. Optionally this provider
// creates a GCP token.
func (d *deployer) Up() error {

	if d.ClusterName == "" {
		return errors.New("flag cluster-name is not set")
	}
	if d.GCPProjectId == "" {
		return errors.New("flag gcp-project-id is not set")
	}
	if d.GCPRegion == "" {
		return errors.New("flag gcp-region is not set")
	}
	if d.BaseDomain == "" {
		return errors.New("flag base-domain is not set")
	}
	if d.PullSecret == "" {
		return errors.New("flag pull-secret-file is not set")
	}

	if !fileExists(d.PullSecret) {
		return errors.New("the pull-secret-file flag is set incorrectly and does not point to a file")
	}

	if d.ServiceAccountToken != "" && !fileExists(d.ServiceAccountToken) {
		return errors.New("the sa-token flag is set incorrectly and does not point to a file")
	}

	baseDir := d.getBaseDir()

	if err := os.Mkdir(baseDir, 0755); err != nil {
		klog.Fatalf("unable to create directory '%s': %v", baseDir, err)
		return err
	}

	if !d.DoNotEnableServices {
		if err := d.enableServices(); err != nil {
			klog.Fatalf("unable to enable GCP services: %v", err)
			return err
		}
	}

	// If we do not have a service account token then create the SA
	if d.ServiceAccountToken == "" {
		if err := d.createServiceAccount(); err != nil {
			klog.Fatalf("unable to create GCP service account: %v", err)
			return err
		}
	}

	pullSecretContent, err := os.ReadFile(d.PullSecret)
	if err != nil {
		klog.Fatalf("unable to read pull secret file '%s': %v", d.PullSecret, err)
		return err
	}

	installConfig := &types.InstallConfig{
		BaseDomain: d.BaseDomain,
		Platform: types.Platform{
			GCP: &gcp.Platform{
				ProjectID: d.GCPProjectId,
				Region:    d.GCPRegion,
			},
		},
		PullSecret: string(pullSecretContent),
	}

	installConfig.Name = d.ClusterName
	installConfig.APIVersion = "v1"

	b, err := yaml.Marshal(&installConfig)
	if err != nil {
		klog.Fatalf("unable to marshal openshift install file %v", err)
		return err
	}

	installConfigFile := filepath.Join(baseDir, "install-config.yaml")

	if err := os.WriteFile(installConfigFile, b, 0644); err != nil {
		klog.Fatalf("unable to write openshift install file %v", err)
		return err
	}

	args := []string{
		"create", "cluster",
		"--dir", baseDir,
		// TODO allow debugging
		// "--log-level=debug",
	}

	// use the service account to athenticate when executing openshift-install binary
	os.Setenv("GOOGLE_CREDENTIALS", d.getServiceAccountToken())
	// we want to see the output so use process.ExecJUnit
	if err := process.ExecJUnit("openshift-install", args, os.Environ()); err != nil {
		klog.Fatalf("unable to create open shift cluster %v", err)
		return err
	}

	return nil
}

// serviceName that are required to run an OpenShift cluster
var serviceNames = []string{
	"compute.googleapis.com",
	"cloudapis.googleapis.com",
	"cloudresourcemanager.googleapis.com",
	"dns.googleapis.com",
	"iam.googleapis.com",
	"iamcredentials.googleapis.com",
	"servicemanagement.googleapis.com",
	"serviceusage.googleapis.com",
	"storage-api.googleapis.com",
	"storage-component.googleapis.com",
}

// gcloud so we do not mispell "gcloud"
var gcloud string = "gcloud"

// enableServices interates through the serviceNames slice
// and uses gcloud to enable each of the services in the project
// specified.
func (d *deployer) enableServices() error {

	for _, service := range serviceNames {
		args := []string{
			"services",
			"enable",
			"--project", d.GCPProjectId,
			"--quiet",
			service,
		}

		if err := process.ExecJUnit(gcloud, args, os.Environ()); err != nil {
			return err
		}
	}

	return nil
}

// role is a slice of GCP roles that are required to
// create the service account.
var roles = []string{
	"compute.admin",
	"dns.admin",
	"iam.securityAdmin",
	"iam.serviceAccountAdmin",
	"iam.serviceAccountUser",
	"iam.serviceAccountKeyAdmin",
	"storage.admin",
	"servicemanagement.quotaViewer",
}

// createServiceAccount creates a service account,
// updates the permissions on said service account.
// And then creates a service account token that is used
// that is used by the openshift deployer.
// This is used in order to not create an OpenShift cluster
// instances with users permissions.
func (d *deployer) createServiceAccount() error {

	gcpDir := d.getServiceAccountTokenDir()
	if err := os.Mkdir(gcpDir, 0755); err != nil {
		return err
	}

	sa := d.getServiceAccountName()
	args := []string{
		"iam",
		"service-accounts",
		"create",
		"--project", d.GCPProjectId,
		sa,
	}

	if err := process.ExecJUnit(gcloud, args, os.Environ()); err != nil {
		return err
	}

	fullId := d.getServiceAccountEmail()

	for _, role := range roles {
		args := []string{
			"projects",
			"add-iam-policy-binding",
			d.GCPProjectId,
			fmt.Sprintf("--member=serviceAccount:%s", fullId),
			fmt.Sprintf("--role=roles/%s", role),
			"--quiet",
		}

		if err := process.ExecJUnit(gcloud, args, os.Environ()); err != nil {
			return err
		}
	}

	args = []string{
		"iam",
		"service-accounts",
		"keys",
		"create",
		d.getServiceAccountToken(),
		"--iam-account", fullId,
		"--project", d.GCPProjectId,
	}

	return process.ExecJUnit(gcloud, args, os.Environ())
}

// getServiceAccountTokenDir is a helper for getting a directory
// name since the directory is used by Up() and Down()
func (d *deployer) getServiceAccountTokenDir() string {
	dir := fmt.Sprintf(".gcp-%s", d.ClusterName)
	return filepath.Join(home, dir)
}

// getServiceAccountToken is a helper for getting the SA token
// name since the token is used by Up() and Down()
func (d *deployer) getServiceAccountToken() string {
	if d.ServiceAccountToken != "" {
		return d.ServiceAccountToken
	}
	return filepath.Join(d.getServiceAccountTokenDir(), "osServiceAccount.json")
}

// getBaseDir is a helper for getting the Base Directory
// name since the directory is used by Up() and Down()
func (d *deployer) getBaseDir() string {
	dir := fmt.Sprintf("openshift-%s", d.ClusterName)
	return filepath.Join(home, dir)
}

// getServiceAccountName is a helper for getting the SA
// name since the name is used by Up() and Down()
func (d *deployer) getServiceAccountName() string {
	return fmt.Sprintf("openshift-sa-%s", d.ClusterName)
}

// getServiceAccountName is a helper for getting the SA full name including the email
// since the name is used by Up() and Down()
func (d *deployer) getServiceAccountEmail() string {
	return fmt.Sprintf("%s@%s.iam.gserviceaccount.com", d.getServiceAccountName(), d.GCPProjectId)
}

// IsUp is required by the kubetest2 interface, but the
// openshift-install binary does this as part of the
// cluster creation.
func (d *deployer) IsUp() (up bool, err error) {
	return true, nil
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

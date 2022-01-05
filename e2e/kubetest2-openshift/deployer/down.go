/*
Copyright 2022 The Cockroach Authors

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

	"k8s.io/klog/v2"
	"sigs.k8s.io/kubetest2/pkg/process"
)

// Down destroys a openshift cluster, deletes the
// IAM user that was used to create the cluster,
// unless DoNotDeleteSA and also deletes the tokens and folders.
func (d *deployer) Down() error {

	if d.ClusterName == "" {
		return errors.New("flag cluster-name is not set")
	}

	// If we have a service account token defined make sure it exists
	// otherwise check that the service account token that was
	// created during the Up() func exists.
	if d.ServiceAccountToken != "" && !fileExists(d.ServiceAccountToken) {
		return errors.New("the sa-token flag is set incorrectly and does not point to a file")
	} else if !fileExists(d.getServiceAccountToken()) {
		return fmt.Errorf("unable to find the service account tocken: %s", d.getServiceAccountToken())
	}

	// check that the openshift-intall files exist
	baseDir := d.getBaseDir()
	meta := filepath.Join(baseDir, "metadata.json")
	if !fileExists(meta) {
		return fmt.Errorf(
			"unable to access metadata.json file: %s, do not have the directory for openshift-install to run",
			meta,
		)
	}

	args := []string{
		"destroy", "cluster",
		"--dir", baseDir,
		// TODO make this configurable
		//		"--log-level=debug",
	}

	os.Setenv("GOOGLE_CREDENTIALS", d.getServiceAccountToken())
	// we want to see the output so use process.ExecJUnit
	if err := process.ExecJUnit("openshift-install", args, os.Environ()); err != nil {
		klog.Fatalf("unable to destroy openshift cluster: %v", err)
		os.Unsetenv("GOOGLE_CREDENTIALS")
		return err
	}
	os.Unsetenv("GOOGLE_CREDENTIALS")

	if err := os.RemoveAll(baseDir); err != nil {
		klog.Fatalf("unable to delete the base directory: %v", err)
		return err
	}

	// If we have a sa token we are not going to remove the iam user and the token
	if d.DoNotDeleteSA {
		return nil
	}

	args = []string{
		"iam",
		"service-accounts",
		"delete",
		d.getServiceAccountEmail(),
		"--project", d.GCPProjectId,
		"--quiet",
	}

	if err := process.ExecJUnit(gcloud, args, os.Environ()); err != nil {
		klog.Fatalf("unable to remove service account: %v", err)
		return err
	}

	if err := os.RemoveAll(d.getServiceAccountTokenDir()); err != nil {
		klog.Fatalf("unable to delete the service account directory: %v", err)
		return err
	}

	return nil
}

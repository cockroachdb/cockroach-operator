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

package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	// certsDir is where the certs are store on disk in the cockroachdb-client-secure pod
	certsDir = "%2Fcockroach%2Fcockroach-certs%2F"

	// pURL is the connection string for use inside the cockroachdb-client-secure pod
	pURL = strings.Join(
		[]string{
			"postgresql://root@cockroachdb-public:26257",
			fmt.Sprintf("?sslcert=%sclient.root.crt", certsDir),
			fmt.Sprintf("&sslkey=%sclient.root.key", certsDir),
			"&sslmode=verify-full",
			fmt.Sprintf("&sslrootcert=%sca.crt", certsDir),
		},
		"",
	)
)

// Step defines an action to be taken during the release process
type Step interface {
	Apply(ExecFn) error
}

// StepFn is a function that implements Step.
type StepFn func(ExecFn) error

// Apply applies the function.
func (fn StepFn) Apply(ex ExecFn) error {
	return fn(ex)
}

// ExecFn describes a function that executes shell commands.
type ExecFn func(cmd string, args, env []string) error

// StartCluster starts a k3d cluster named `k3d-<name>` using the specified image
func StartCluster(name string, version string) Step {
	return StepFn(func(fn ExecFn) error {
		fmt.Println("Creating k3d cluster...")
		return fn(
			"k3d",
			[]string{"cluster", "create", name, "--image", fmt.Sprintf("rancher/k3s:v%s-k3s1", version)},
			os.Environ(),
		)
	})
}

// StopCluster stops the cluster named `k3d-<name>`.
func StopCluster(name string) Step {
	fmt.Println("Deleting k3d cluster...")
	return StepFn(func(fn ExecFn) error {
		return fn(
			"k3d",
			[]string{"cluster", "delete", name},
			os.Environ(),
		)
	})
}

// ApplyManifest performs a kubectl apply -f for the specified file.
func ApplyManifest(file string) Step {
	return StepFn(func(fn ExecFn) error {
		fmt.Println("Applying " + file)
		return fn(
			"kubectl",
			[]string{"apply", "-f", file},
			os.Environ(),
		)
	})
}

// WaitForDeploymentAvailable waits until the specified deployment is available. It will timeout after 2m.
func WaitForDeploymentAvailable(name, namespace string) Step {
	return StepFn(func(fn ExecFn) error {
		fmt.Println("Waiting for deployment to be available")
		return fn(
			"kubectl",
			[]string{"wait", "--for", "condition=Available", "deploy/" + name, "-n", namespace, "--timeout", "2m"},
			os.Environ(),
		)
	})
}

// WaitForSecret waits for a Kuebernetes secret to be available
func WaitForSecret(name, namespace string) Step {
	return StepFn(func(fn ExecFn) error {
		fmt.Println("Waiting for secret to be created")
		return retry(5, 10*time.Second, func() error {
			return fn(
				"kubectl",
				[]string{"get", "secret", name, "-n", namespace},
				os.Environ(),
			)
		})
	})
}

// WaitForStatefulSetRollout waits until the specified deployment is available. It will timeout after 2m.
func WaitForStatefulSetRollout(name string) Step {
	return StepFn(func(fn ExecFn) error {
		fmt.Println("Waiting for statefulset to be ready")
		return retry(20, 10*time.Second, func() error {
			return fn(
				"kubectl",
				[]string{"rollout", "status", "-w", "sts/" + name},
				os.Environ(),
			)
		})
	})
}

// WaitForPodReady waits until the named pod is in a ready state.
func WaitForPodReady(name string) Step {
	return StepFn(func(fn ExecFn) error {
		fmt.Println("Waiting for pod to be ready")
		return fn(
			"kubectl",
			[]string{"wait", "--for", "condition=Ready", "pod/" + name, "--timeout", "2m"},
			os.Environ(),
		)
	})
}

// InitBankWorkload initializes the bank workload
func InitBankWorkload() Step {
	return StepFn(func(fn ExecFn) error {
		return runWithClient(fn, "workload", "init", "bank")
	})
}

// RunBankWorkload runs the bank workload for the specified duration
func RunBankWorkload(d time.Duration) Step {
	return StepFn(func(fn ExecFn) error {
		return runWithClient(fn, "workload", "run", "bank", "--duration", d.String())
	})
}

func runWithClient(fn ExecFn, cmd ...string) error {
	args := append([]string{
		"exec",
		"-it",
		"cockroachdb-client-secure",
		"--",
		"cockroach",
	}, cmd...)

	args = append(args, pURL)
	return fn("kubectl", args, os.Environ())
}

func retry(maxAttempts int, durBetweenTries time.Duration, fn func() error) error {
	attempts := 0

	for {
		err := fn()
		if err == nil {
			return nil
		}

		attempts++
		if attempts == maxAttempts {
			return err
		}

		time.Sleep(durBetweenTries)
	}
}

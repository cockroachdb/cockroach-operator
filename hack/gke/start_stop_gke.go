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

package main

import (
	"flag"
	"os"

	"github.com/cockroachdb/cockroach-operator/pkg/testutil/exec"
)

// This binary is used to start and stop gke on the command line
// There is a bazezl target that allows you to do this
// See the Makefile for example

func main() {

	runFiles := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	if runFiles == "" {
		panic("unable to read BUILD_WORKSPACE_DIRECTORY environment variable")
	}

	action := flag.String("action", "", "start,stop or kubeconfig")
	name := flag.String("cluster-name", "", "name of the cluster")
	zone := flag.String("gcp-zone", "", "gcp zone")
	project := flag.String("gcp-project", "", "gcp project")
	flag.Parse()

	if *name == "" {
		panic("set cluster-name arg")
	}
	if *zone == "" {
		panic("set gcp-zone arg")
	}
	if *project == "" {
		panic("set gcp-project arg")
	}

	p := "hack/bin" + ":" + os.Getenv("PATH")
	os.Setenv("PATH", p)

	switch {
	case *action == "start":
		err := exec.StartGKEKubeTest2(*name, *zone, *project)
		if err != nil {
			panic(err)
		}
	case *action == "stop":
		err := exec.StopGKEKubeTest2(*name, *zone, *project)
		if err != nil {
			panic(err)
		}
	case *action == "kubeconfig":
		err := exec.GetGKEKubeConfig(*name, *zone, *project)
		if err != nil {
			panic(err)
		}
	default:
		panic("wrong action parameter")
	}

}

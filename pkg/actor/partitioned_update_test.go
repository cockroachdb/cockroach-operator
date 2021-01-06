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

package actor

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestDeployedInCluster(t *testing.T) {
	isInK8s := inK8s("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if isInK8s {
		t.Logf("%v", isInK8s)
		t.Fatal("we should not be running inside of k8s")
	}

	file, err := ioutil.TempFile("/tmp", "foo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	inK8s := inK8s(file.Name())

	if !inK8s {
		t.Fatal("we should find the file")
	}
}

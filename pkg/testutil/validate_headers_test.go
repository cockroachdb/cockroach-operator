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

package testutil

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const GOOD_MAKEFILE = `# Copyright 2021 The Cockroach Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

target:
	foo
`
const GOOD_GOFILE = `// +build
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

package testutil
`
const BAD_GOFILE = `/*
Copyright 2021 The Wrong Authors

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

package testutil
`

func writeFile(data string, dir string, fileName string) error {
	content := []byte(data)
	tmpfn := filepath.Join(dir, fileName)
	return ioutil.WriteFile(tmpfn, content, 0666)
}

func TestValidate(t *testing.T) {

	dir, err := ioutil.TempDir("", "validate-headers-test")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	var files = map[string]string{
		"bad_go_file.go":  BAD_GOFILE,
		"good_go_file.go": GOOD_GOFILE,
		"Makefile":        GOOD_MAKEFILE,
	}

	for fileName, data := range files {
		err := writeFile(data, dir, fileName)
		if err != nil {
			t.Fatal(err)
		}
	}

	v := NewValidateHeaders(nil, dir, "testdata", "")

	non, err := v.Validate()
	if err != nil {
		t.Fatal("error running Validate", err)
	}

	for _, filename := range *non {
		t.Logf("nonvalid file: %s", filename)
	}

	if non == nil {
		t.Fatal("test did not find any bad files")
	}

	n := *non

	if len(n) != 1 {
		t.Fatal("test found too may incorrect files")
	}

	base := path.Base(n[0])

	if base != "bad_go_file.go" {
		t.Fatal("bad_go_file.go was not detected as having incorrect header")
	}

}

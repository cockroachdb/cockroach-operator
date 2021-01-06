#!/bin/bash

# Copyright 2021 The Cockroach Authors
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

set -o errexit
set -o nounset
set -o pipefail

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then # Running inside bazel
  echo "Updating generated clients..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:update-codegen
  )
  exit 0
fi

module_name="github.com/cockroachdb/cockroach-operator"

# Generate deepcopy functions for all internal and external APIs
deepcopy_inputs=(
  api/v1alpha1
)

go_sdk=$PWD/external/go_sdk
go=$PWD/$1
clientgen=$PWD/$2
deepcopygen=$PWD/$3
informergen=$PWD/$4
listergen=$PWD/$5
defaultergen=$PWD/$6
conversiongen=$PWD/$7

shift 7

fake_gopath=""
fake_repopath=""
ensure-in-gopath() {
  export GOROOT=$go_sdk

  fake_gopath=$(mktemp -d -t codegen.gopath.XXXX)

  fake_repopath=$fake_gopath/src/github.com/cockroachdb/cockroach-operator
  mkdir -p "$fake_repopath"
  cp -R "$BUILD_WORKSPACE_DIRECTORY/." "$fake_repopath"

  export GOPATH=$fake_gopath
  cd "$fake_repopath"
  echo "Created fake GOPATH to run code generators in"
}

cleanup_gopath() {
  export GO111MODULE=off
  "$go" clean --modcache
  rm -rf "$fake_gopath" || true
}
trap cleanup_gopath EXIT

# clean will delete files matching name in path.
#
# When inside bazel test the files are read-only.
# Any attempts to write a file that already exists will fail.
# So resolve by deleting the files before generating them.
clean() {
  path=$1
  name=$2
  if [[ ! -d "$path" ]]; then
    return 0
  fi
  find "$path" -name "$name" -delete
}

mkcp() {
  src="$1"
  dst="$2"
  mkdir -p "$(dirname "$dst")"
  cp "$src" "$dst"
}
# Export mkcp for use in sub-shells
export -f mkcp

copyfiles() {
  # Don't copy data if the workspace directory is already within the GOPATH
  if [ "${BUILD_WORKSPACE_DIRECTORY:0:${#GOPATH}}" = "$GOPATH" ]; then
    return 0
  fi

  path=$1
  name=$2
  if [[ ! -d "$path" ]]; then
    return 0
  fi
  (
    cd "$GOPATH/src/$module_name/$path"

    find "." -name "$name" -exec bash -c "mkcp {} \"$BUILD_WORKSPACE_DIRECTORY/$path/{}\"" \;
  )
}

gen-deepcopy() {
  clean apis 'zz_generated.deepcopy.go'
  echo "Generating deepcopy methods..." >&2
  prefixed_inputs=( "${deepcopy_inputs[@]/#/$module_name/}" )
  joined=$( IFS=$','; echo "${prefixed_inputs[*]}" )
  "$deepcopygen" \
    --go-header-file hack/boilerplate/boilerplate.go.txt \
    --input-dirs "$joined" \
    --output-file-base zz_generated.deepcopy \
    --bounding-dirs "${module_name}"
  for dir in "${deepcopy_inputs[@]}"; do
    copyfiles "$dir" "zz_generated.deepcopy.go"
  done
}

runfiles="$(pwd)"
export GO111MODULE=off
ensure-in-gopath
old=${GOCACHE:-}
export GOCACHE=$(mktemp -d -t codegen.gocache.XXXX)
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org
export GOSUMDB=sum.golang.org
"$go_sdk/bin/go" mod vendor
export GO111MODULE=off
export GOCACHE=$old

gen-deepcopy

## Call update-bazel
export GO111MODULE=on
cd "$runfiles"
"$@"

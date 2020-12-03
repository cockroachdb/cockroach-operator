#!/usr/bin/env bash

# Copyright 2020 The Cockroach Authors
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
  echo "Generating  CSV..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:generate-csv
  )
  exit 0
fi

# This script should be run via `bazel run //hack:gen-csv`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}

echo "+++ Running operator-sdk"

# BUNDLE_METADATA_OPTS="$2"
VERSION="$4"
echo $VERSION
IMG="$5"
echo $IMG
PKG_MAN_OPTS="$6"
echo "bla: $PKG_MAN_OPTS"

operator-sdk generate kustomize manifests -q
cd manifests && kustomize edit set image cockroachdb/cockroach-operator=${IMG} && cd ..
kustomize build config/manifests | operator-sdk generate packagemanifests -q --version ${VERSION} ${PKG_MAN_OPTS}






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
    bazel run //hack:update-csv
  )
  exit 0
fi

opsdk=$(realpath "$1")
kstomize="$(realpath "$2")"
export PATH=$(dirname "$opsdk"):$PATH

# This script should be run via `bazel run //hack:gen-csv`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}

echo "+++ Running gen csv"

VERSION="$4"
echo $VERSION
IMG="$5"
echo "img=$IMG"
BUNDLE_METADATA_OPTS="$6"
echo "bundle opts:$BUNDLE_METADATA_OPTS"
"$opsdk" generate kustomize manifests -q 
cd manifests && kustomize edit set image cockroachdb/cockroach-operator=${IMG} && cd ..
"$kstomize" build config/manifests | "$opsdk" generate bundle -q --overwrite --version ${VERSION} ${BUNDLE_METADATA_OPTS}
"$opsdk" bundle validate ./bundle

FILE_NAMES=(bundle/manifests/cockroach-operator-sa_v1_serviceaccount.yaml \
bundle/tests/scorecard/config.yaml \
bundle/manifests/cockroach-operator.clusterserviceversion.yaml \
bundle/manifests/crdb.cockroachlabs.com_crdbclusters.yaml \
bundle/metadata/annotations.yaml \
config/manifests/bases/cockroach-operator.clusterserviceversion.yaml \
)
for YAML in "${FILE_NAMES[@]}"
do
   :
   cat "${REPO_ROOT}/hack/boilerplate/boilerplate.yaml.txt" "${REPO_ROOT}/${YAML}" > "${REPO_ROOT}/${YAML}.mod"
   mv "${REPO_ROOT}/${YAML}.mod" "${REPO_ROOT}/${YAML}"
done 

DOCKER_FILE_NAME="bundle.Dockerfile"

cat "${REPO_ROOT}/hack/boilerplate/boilerplate.Dockerfile.txt" "${REPO_ROOT}/${DOCKER_FILE_NAME}" > "${REPO_ROOT}/${DOCKER_FILE_NAME}.mod"
mv "${REPO_ROOT}/${DOCKER_FILE_NAME}.mod" "${REPO_ROOT}/${DOCKER_FILE_NAME}"
 





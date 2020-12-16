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
opm="$(realpath "$3")"
export PATH=$(dirname "$opsdk"):$PATH

# This script should be run via `bazel run //hack:gen-csv`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
echo "+++ Running gen csv"

RH_BUNDLE_VERSION="$4"
[[ -z "$RH_BUNDLE_VERSION" ]] && { echo "Error: RH_BUNDLE_VERSION not set"; exit 1; }
echo "RH_BUNDLE_VERSION=$RH_BUNDLE_VERSION"
RH_COCKROACH_OP_IMG="$5"
echo "RH_COCKROACH_OP_IMG=$RH_COCKROACH_OP_IMG"
[[ -z "$RH_COCKROACH_OP_IMG" ]] && { echo "Error: RH_COCKROACH_OP_IMG not set"; exit 1; }
RH_BUNDLE_METADATA_OPTS="$6"
echo "RH_BUNDLE_METADATA_OPTS=$RH_BUNDLE_METADATA_OPTS"
[[ -z "$RH_BUNDLE_METADATA_OPTS" ]] && { echo "Error: RH_BUNDLE_METADATA_OPTS not set"; exit 1; }
RH_COCKROACH_DATABASE_IMAGE="$8"
echo "RH_COCKROACH_DATABASE_IMAGE=$RH_COCKROACH_DATABASE_IMAGE"
[[ -z "$RH_COCKROACH_DATABASE_IMAGE" ]] && { echo "Error: RH_COCKROACH_DATABASE_IMAGE not set"; exit 1; }
"$opsdk" generate kustomize manifests -q 
"$kstomize" build config/manifests | "$opsdk" generate bundle -q --overwrite --version ${RH_BUNDLE_VERSION} ${RH_BUNDLE_METADATA_OPTS}
"$opsdk" bundle validate ./bundle
cat bundle/manifests/cockroach-operator.clusterserviceversion.yaml | sed -e "s+RH_COCKROACH_OP_IMAGE_PLACEHOLDER+${RH_COCKROACH_OP_IMG}+g" -e "s+RH_COCKROACH_DB_IMAGE_PLACEHOLDER+${RH_COCKROACH_DATABASE_IMAGE}+g" -e "s+CREATED_AT_PLACEHOLDER+"$(date +"%FT%H:%M:%SZ")"+g"> bundle/manifests/cockroach-operator.clusterserviceversion.yaml 
cd ${REPO_ROOT}
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
 





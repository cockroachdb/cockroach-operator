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
    bazel run //hack:update-pkg-manifest
  )
  exit 0
fi

opsdk=$(realpath "$1")
kstomize="$(realpath "$2")"
opm=$(realpath "$3")
export PATH=$(dirname "$opsdk"):$PATH
# This script should be run via `bazel run //hack:update-pkg-manifest
# It will be used on Openshift certification image bundle releases.
# It created a bundle in the package manifests format. 
# We are keeping the format that was initially used.
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
echo "+++ Running update package manifest for certification"
RH_BUNDLE_VERSION="$4"
[[ -z "$RH_BUNDLE_VERSION" ]] && { echo "Error: RH_BUNDLE_VERSION not set"; exit 1; }
echo "RH_BUNDLE_VERSION=$RH_BUNDLE_VERSION"
RH_COCKROACH_OP_IMG="$5"
echo "RH_COCKROACH_OP_IMG=$RH_COCKROACH_OP_IMG"
[[ -z "$RH_COCKROACH_OP_IMG" ]] && { echo "Error: RH_COCKROACH_OP_IMG not set"; exit 1; }
RH_PKG_MAN_OPTS="$6"
echo "RH_PKG_MAN_OPTS=$RH_PKG_MAN_OPTS"
[[ -z "$RH_PKG_MAN_OPTS" ]] && { echo "Error: RH_PKG_MAN_OPTS not set"; exit 1; }
RH_COCKROACH_DATABASE_IMAGE="$7"
echo "RH_COCKROACH_DATABASE_IMAGE=$RH_COCKROACH_DATABASE_IMAGE"
[[ -z "$RH_COCKROACH_DATABASE_IMAGE" ]] && { echo "Error: RH_COCKROACH_DATABASE_IMAGE not set"; exit 1; }
DEPLOY_PATH="deploy/certified-metadata-bundle/cockroach-operator"
DEPLOY_CERTIFICATION_PATH="deploy/certified-metadata-bundle"
if [ -d "${DEPLOY_PATH}/${RH_BUNDLE_VERSION}" ] 
then
    echo "Folder ${DEPLOY_PATH}/${RH_BUNDLE_VERSION} already exists. Please increase the version or remove the folder manually." 
    exit 1
fi
rm -rf ${DEPLOY_PATH}/${RH_BUNDLE_VERSION}
"$opsdk" generate kustomize manifests -q --verbose
"$kstomize" build config/manifests | "$opsdk" generate packagemanifests -q --version ${RH_BUNDLE_VERSION} ${RH_PKG_MAN_OPTS} --output-dir ${DEPLOY_PATH} --input-dir ${DEPLOY_PATH} --verbose
rm -rf ${DEPLOY_PATH}/${RH_BUNDLE_VERSION}/cockroach-operator-sa_v1_serviceaccount.yaml
cat ${DEPLOY_PATH}/${RH_BUNDLE_VERSION}/cockroach-operator.clusterserviceversion.yaml | sed -e "s+RH_COCKROACH_OP_IMAGE_PLACEHOLDER+${RH_COCKROACH_OP_IMG}+g" -e "s+RH_COCKROACH_DB_IMAGE_PLACEHOLDER+${RH_COCKROACH_DATABASE_IMAGE}+g" -e "s+CREATED_AT_PLACEHOLDER+"$(date +"%FT%H:%M:%SZ")"+g"> ${DEPLOY_PATH}/${RH_BUNDLE_VERSION}/cockroach-operator.v${RH_BUNDLE_VERSION}.clusterserviceversion.yaml
rm -rf ${DEPLOY_PATH}/${RH_BUNDLE_VERSION}/cockroach-operator.clusterserviceversion.yaml
# This is needed after csv generation
cd ${REPO_ROOT}
FILE_NAMES=(
  config/manifests/bases/cockroach-operator.clusterserviceversion.yaml \
)
for YAML in "${FILE_NAMES[@]}"
do
   :
   cat "${REPO_ROOT}/hack/boilerplate/boilerplate.yaml.txt" "${REPO_ROOT}/${YAML}" > "${REPO_ROOT}/${YAML}.mod"
   mv "${REPO_ROOT}/${YAML}.mod" "${REPO_ROOT}/${YAML}"
done 




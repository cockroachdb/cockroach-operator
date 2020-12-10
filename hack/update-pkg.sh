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
    bazel run //hack:update-pkg
  )
  exit 0
fi

opsdk=$(realpath "$1")
kstomize="$(realpath "$2")"
opm=$(realpath "$3")
export PATH=$(dirname "$opsdk"):$PATH
# This script should be run via `bazel run //hack:update-pkg
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
echo "+++ Running update package manifest "
VERSION="$4"
echo "VERSION:$VERSION"
IMG="$5"
echo $IMG
PKG_MAN_OPTS="$6"
echo "PKG_MAN_OPTS: $PKG_MAN_OPTS"
DEPLOY_PATH="deploy/certified-metadata-bundle/cockroach-operator"
DEPLOY_CERTIFICATION_PATH="deploy/certified-metadata-bundle"
if [ -d "${DEPLOY_PATH}/${VERSION}" ] 
then
    echo "Folder ${DEPLOY_PATH}/${VERSION} already exists. Please increase the version or remove the folder manually." 
    exit 1
fi
rm -rf ${DEPLOY_PATH}/${VERSION}
"$opsdk" generate kustomize manifests -q --verbose
cd manifests && "$kstomize" edit set image cockroachdb/cockroach-operator=${IMG} && cd ..
"$kstomize" build config/manifests | "$opsdk" generate packagemanifests -q --version ${VERSION} ${PKG_MAN_OPTS} --output-dir ${DEPLOY_PATH} --input-dir ${DEPLOY_PATH} --verbose
mv ${DEPLOY_PATH}/${VERSION}/cockroach-operator.clusterserviceversion.yaml ${DEPLOY_PATH}/${VERSION}/cockroach-operator.v${VERSION}.clusterserviceversion.yaml
rm -rf ${DEPLOY_PATH}/${VERSION}/cockroach-operator-sa_v1_serviceaccount.yaml





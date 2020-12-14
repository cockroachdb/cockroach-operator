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
  echo "Running opm to cretae the bundle.." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:opm-build-bundle
  )
  exit 0
fi
opm=$(realpath "$1")
export PATH=$(dirname "$opm"):$PATH
# This script should be run via `bazel //hack:opm-build-bundle`
# TO DO: this will be merged in one script that will release the OpenShift bundle 
# This script validates the bundle in the OpenShift format
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
echo "+++ Running opm to create bundle"
VERSION="$2"
echo "VERSION:$VERSION"
IMG="$3"
echo $IMG
PKG_MAN_OPTS="$4"
echo "PKG_MAN_OPTS: $PKG_MAN_OPTS"
DEPLOY_PATH="deploy/certified-metadata-bundle/cockroach-operator"
DEPLOY_CERTIFICATION_PATH="deploy/certified-metadata-bundle"
cd ${DEPLOY_PATH} &&  "$opm" alpha bundle generate -d ./${VERSION}/ -u ./${VERSION}/ -c beta,stable -e stable
cp ../annotations.yaml ./${VERSION}/metadata
sed "s/VERSION/${VERSION}/g" ../bundle.Dockerfile > ./bundle-${VERSION}.Dockerfile
cp ./bundle-${VERSION}.Dockerfile ./bundle.Dockerfile
# Move to latest folder for release -> I need a fixed folder name for the docker image that runs from bazel
rm -rf ./latest/*/*.yaml
rm -rf ./latest/*.yaml
cp -R ./${VERSION}/manifests/*.yaml ./latest/manifests
cp -R ./${VERSION}/manifests/*.yaml ./latest
cp -R ./${VERSION}/metadata/*.yaml ./latest/metadata








 





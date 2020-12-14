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
  echo "OPM INDEX build..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:opm-build-index
  )
  exit 0
fi
opm=$(realpath "$1")
export PATH=$(dirname "$opm"):$PATH
# This script should be run via `bazel run //opm-build-index`
# This script is DEV only and will be added in a DEV target
# TO DO: fix image bundle to simulate or use scratch
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
OLM_REPO=$2
OLM_BUNDLE_REPO=$3
TAG=$4
VERSION=$5
echo "Running with args OLM_REPO=$2 OLM_BUNDLE_REPO=$3 TAG=$4 VERSION=$5"
[[ -z "$OLM_REPO" ]] && { echo "Error: OLM_REPO not set"; exit 1; }
[[ -z "$RH_BUNDLE_REGISTRY" ]] && { echo "Error: RH_BUNDLE_REGISTRY not set"; exit 1; }
VERSIONS_LIST="$OLM_REPO:$TAG"
echo "Using tag ${OLM_BUNDLE_REPO}:${TAG}"
echo "Building index with $VERSIONS_LIST"
"$opm" index add -u docker --generate --bundles "$VERSIONS_LIST" --tag "${OLM_BUNDLE_REPO}:${TAG}"
if [ $? -ne 0 ]; then
    echo "fail to build opm"
    exit 1
fi
#build and push to the quay.io registry the image taged with the index
RH_BUNDLE_REGISTRY=${RH_BUNDLE_REGISTRY} \
RH_BUNDLE_IMAGE_REPOSITORY=${OLM_BUNDLE_REPO} \
RH_BUNDLE_VERSION=${RH_BUNDLE_VERSION} \
RH_DEPLOY_PATH=${RH_DEPLOY_PATH} \
RH_BUNDLE_IMAGE_TAG=${TAG} \
bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
  //:push_operator_bundle_image 
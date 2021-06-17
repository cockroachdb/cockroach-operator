#!/usr/bin/env bash

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
  echo "OPM INDEX build..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:push-openshift-images
  )
  exit 0
fi
# TODO rename script
# This script should be run via `bazel run //hack:push-openshift-images
# This script is DEV only and will be added in a DEV target

echo $1
opm=$(realpath "$1")
export PATH=$(dirname "$opm"):$PATH

REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
[[ -z "$APP_VERSION" ]] && { echo "Error: APP_VERSION not set"; exit 1; }
[[ -z "$DOCKER_REGISTRY" ]] && { echo "Error: DOCKER_REGISTRY not set"; exit 1; }

echo "Building and pushing openshift bundle and index"
echo "Running with args APP_VERSION=$APP_VERSION DOCKER_REGISTRY=$DOCKER_REGISTRY"

cd deploy/certified-metadata-bundle/cockroach-operator/
docker build -t  ${DOCKER_REGISTRY}/cockroachdb-operator-bundle:${APP_VERSION} -f  bundle.Dockerfile .
docker push ${DOCKER_REGISTRY}/cockroachdb-operator-bundle:${APP_VERSION}

cd -
echo "Running opm index with args APP_VERSION=$APP_VERSION DOCKER_REGISTRY=$DOCKER_REGISTRY"
"$opm" index add --bundles  ${DOCKER_REGISTRY}/cockroachdb-operator-bundle:${APP_VERSION} \
  --tag  ${DOCKER_REGISTRY}/cockroachdb-operator-index:${APP_VERSION} -c docker

docker push ${DOCKER_REGISTRY}/cockroachdb-operator-index:${APP_VERSION}

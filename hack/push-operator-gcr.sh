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

# "---------------------------------------------------------"
# "-                                                       -"
# "-  Push a new operator container                        -"
# "-                                                       -"
# "---------------------------------------------------------"

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")
# shellcheck disable=SC1090
source "${ROOT}/functions.sh"

# TODO figure out version and make sure it is set
VERSION=$(git rev-parse --short HEAD)

# TODO(chrislovecnm): need to test this in teamcity
if [[ -n "${GOOGLE_APPLICATION_CREDENTIALS:-}" ]]; then
  echo "Detected GOOGLE_APPLICATION_CREDENTIALS, activating..." >&2
  gcloud auth activate-service-account --key-file="${GOOGLE_APPLICATION_CREDENTIALS}"
fi

# enable googleapi for container registry on the project
enable-service containerregistry.googleapis.com

# TODO configure this so that we can override the default project
PROJECT=$(gcloud config get-value project)

if [ -z "${PROJECT}" ]; then
    echo "gcloud cli must be configured with a default project." 1>&2
    exit 1;
fi

# TODO(chrislovecnm): Update this once we have the right location.
GCR_REGISTRY="us.gcr.io/${PROJECT}"
IMAGE="${GCR_REGISTRY}/cockroach-operator:${VERSION}"

echo "building Dockerfile.ubi image"
docker build --no-cache --pull -t "${IMAGE}" -f "${ROOT}/../Dockerfile.ubi" .

gcloud auth configure-docker
# TODO not certain why we are not using the gcloud auth command, but are using the docker login
# echo "${GOOGLE_CREDENTIALS}" | docker login -u _json_key --password-stdin "https://${GCR_URL}"

docker push "${IMAGE}"

echo "pushed new image to container registry"
echo "${IMAGE}"

#!/bin/bash

# Copyright 2023 The Cockroach Authors
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

# The only argument this script should ever be called with is '--verify-only'
set -euo pipefail

cat <<EOF
NUMBER_COMMITS_ON_BRANCH $(git rev-list $(git rev-parse --abbrev-ref HEAD) | wc -l)

RH_BUNDLE_REGISTRY ${RH_BUNDLE_REGISTRY:-}
RH_BUNDLE_IMAGE_REPOSITORY ${RH_BUNDLE_IMAGE_REPOSITORY:-}
RH_BUNDLE_VERSION ${RH_BUNDLE_VERSION:-}

STABLE_CLUSTER ${K8S_CLUSTER:-}
STABLE_DOCKER_REGISTRY ${DOCKER_REGISTRY:-}
STABLE_DOCKER_TAG ${APP_VERSION:-}
STABLE_IMAGE_REGISTRY ${DEV_REGISTRY:-}
STABLE_IMAGE_REPOSITORY ${DOCKER_IMAGE_REPOSITORY:-}
EOF

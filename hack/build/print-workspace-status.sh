#!/bin/bash

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

# The only argument this script should ever be called with is '--verify-only'

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null && pwd )"
REPO_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../../" > /dev/null && pwd )"

source "${SCRIPT_ROOT}/version.sh"
kube::version::get_version_vars

APP_GIT_COMMIT=${APP_GIT_COMMIT:-$(git rev-parse HEAD)}
GIT_STATE=""
if [ ! -z "$(git status --porcelain)" ]; then
    GIT_STATE="dirty"
fi

cat <<EOF
STABLE_BUILD_GIT_COMMIT ${KUBE_GIT_COMMIT-}
STABLE_BUILD_SCM_STATUS ${KUBE_GIT_TREE_STATE-}
STABLE_BUILD_SCM_REVISION ${KUBE_GIT_VERSION-}
STABLE_BUILD_MAJOR_VERSION ${KUBE_GIT_MAJOR-}
STABLE_BUILD_MINOR_VERSION ${KUBE_GIT_MINOR-}
STABLE_DOCKER_TAG ${APP_VERSION:-${KUBE_GIT_VERSION/+/-}}
STABLE_DOCKER_REGISTRY ${DOCKER_REGISTRY:-us.gcr.io/chris-love-operator-playground}
STABLE_DOCKER_PUSH_REGISTRY ${DOCKER_PUSH_REGISTRY:-${DOCKER_REGISTRY:-us.gcr.io/chris-love-operator-playground}}
STABLE_IMAGE_REPOSITORY ${DOCKER_IMAGE_REPOSITORY:-cockroach-operator}

RH_BUNDLE_REGISTRY ${RH_BUNDLE_REGISTRY:-""}
RH_BUNDLE_IMAGE_REPOSITORY ${RH_BUNDLE_IMAGE_REPOSITORY:-cockroach-operator-bundle}

RH_DEPLOY_PATH ${RH_DEPLOY_PATH:-deploy/certified-metadata-bundle/cockroach-operator}
RH_BUNDLE_VERSION ${RH_BUNDLE_VERSION:-""}
RH_BUNDLE_IMAGE_TAG ${RH_BUNDLE_IMAGE_TAG:-${RH_BUNDLE_VERSION:-""}}
  
IMAGE_REGISTRY ${DEV_REGISTRY:-us.gcr.io/chris-love-operator-playground}

CLUSTER ${K8S_CLUSTER:-gke_chris-love-operator-playground_us-central1-a_test}
NUMBER_COMMITS_ON_BRANCH $(git rev-list $(git rev-parse --abbrev-ref HEAD) | wc -l)

gitCommit ${KUBE_GIT_COMMIT-}
gitTreeState ${KUBE_GIT_TREE_STATE-}
gitVersion ${KUBE_GIT_VERSION-}
gitMajor ${KUBE_GIT_MAJOR-}
gitMinor ${KUBE_GIT_MINOR-}
buildDate $(date \
  ${SOURCE_DATE_EPOCH:+"--date=@${SOURCE_DATE_EPOCH}"} \
 -u +'%Y-%m-%dT%H:%M:%SZ')
EOF

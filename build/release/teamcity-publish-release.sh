#!/usr/bin/env bash
# Copyright 2024 The Cockroach Authors
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

set -euxo pipefail

source "$(dirname "${0}")/teamcity-support.sh"

REGISTRY="docker.io"
REPO="cockroachdb/cockroach-operator"
if ! [[ -z "${DRY_RUN}" ]] ; then REPO="${REPO}-misc"; fi

OPERATOR_IMG="${REGISTRY}/${REPO}:${TAG}"

main() {
  docker_login "${REGISTRY}" "${OPERATOR_DOCKER_ID}" "${OPERATOR_DOCKER_ACCESS_TOKEN}"

  validate_image
  publish_to_registry
}

validate_image() {
  tc_start_block "Ensure image should be pushed"
  
  if docker_image_exists "${OPERATOR_IMG}"; then
    echo "Docker image ${OPERATOR_IMG} already exists!"

    if [[ -z "${FORCE}" ]] ; then
      echo "Use FORCE=1 to force push the docker image."
      echo "Alternatively you can delete the tag in Docker Hub."
      exit 1
    fi
    echo "Forcing docker push..."
  fi

  tc_end_block "Ensure image should be pushed"
}

publish_to_registry() {
  tc_start_block "Make and push docker image"

  make \
    DOCKER_REGISTRY="${REGISTRY}" \
    DOCKER_IMAGE_REPOSITORY="${REPO}" \
    release/image

  tc_end_block "Make and push docker image"
}

main "$@"

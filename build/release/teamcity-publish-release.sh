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

set -euxo pipefail

source "$(dirname "${0}")/teamcity-support.sh"

tc_start_block "Variable Setup"
VERSION="v"$(cat version.txt)
# Matching the version name regex from within the cockroach code except
# for the `metadata` part at the end because Docker tags don't support
# `+` in the tag name.
# https://github.com/cockroachdb/cockroach/blob/4c6864b44b9044874488cfedee3a31e6b23a6790/pkg/util/version/version.go#L75
image_tag="$(echo "${VERSION}" | grep -E -o '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[-.0-9A-Za-z]+)?$')"
#                                         ^major           ^minor           ^patch         ^preRelease

if [[ -z "$image_tag" ]] ; then
  echo "Invalid VERSION \"${VERSION}\". Must be of the format \"vMAJOR.MINOR.PATCH(-PRERELEASE)?\"."
  exit 1
fi

docker_registry="docker.io"
operator_image_repository="cockroachdb/cockroach-operator"

if ! [[ -z "${DRY_RUN}" ]] ; then
  operator_image_repository="cockroachdb/cockroach-operator-misc"
fi

tc_end_block "Variable Setup"

tc_start_block "Make and push docker images"
configure_docker_creds
docker_login "$docker_registry" "$OPERATOR_DOCKER_ID" "$OPERATOR_DOCKER_ACCESS_TOKEN"

if docker_image_exists "$docker_registry/$operator_image_repository:$image_tag"; then
  echo "Docker image $docker_registry/$operator_image_repository:$image_tag already exists"
  if [[ -z "${FORCE}" ]] ; then
    echo "Use FORCE=1 to force push the docker image."
    echo "Alternatively you can delete the tag in Docker Hub."
    exit 1
  fi
  echo "Forcing docker push..."
fi

make \
  DOCKER_REGISTRY="$docker_registry" \
  DOCKER_IMAGE_REPOSITORY="$operator_image_repository" \
  release/image
tc_end_block "Make and push docker images"

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
build_name="$(echo "${VERSION}" | grep -E -o '^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[-.0-9A-Za-z]+)?$')"
#                                         ^major           ^minor           ^patch         ^preRelease

if [[ -z "$build_name" ]] ; then
    echo "Invalid VERSION \"${VERSION}\". Must be of the format \"vMAJOR.MINOR.PATCH(-PRERELEASE)?\"."
    exit 1
fi

rhel_registry="scan.connect.redhat.com"
src_docker_registry="docker.io/cockroachdb"
operator_rhel_docker_image_repository="ospid-cf721588-ad8a-4618-938c-5191c5e10ae4"
bundle_rhel_docker_image_repository="ospid-857fe786-3eb7-4508-aafd-cc74c1b1dc24"

if [[ -z "${DRY_RUN}" ]] ; then
  src_docker_image_repository="cockroach-operator"
else
  src_docker_image_repository="cockroach-operator-misc"
  build_name="${build_name}-dryrun"
fi
tc_end_block "Variable Setup"


tc_start_block "Make and push docker images"
configure_docker_creds
docker_login "$rhel_registry" "$OPERATOR_REDHAT_REGISTRY_USER" "$OPERATOR_REDHAT_REGISTRY_KEY"

# sanity check
if docker_image_exists "$rhel_registry/$operator_rhel_docker_image_repository:$build_name"; then
  echo "Docker image $rhel_registry/$operator_rhel_docker_image_repository:$build_name already exists"
  if [[ -z "${FORCE}" ]] ; then
    echo "Use FORCE=1 to force push the docker image."
    echo "Alternatively you can delete the tag in RedHat Connect and rerun the script."
    exit 1
  fi
  echo "Forcing docker push..."
fi

docker pull "$src_docker_registry/$src_docker_image_repository:$build_name"
docker tag "$src_docker_registry/$src_docker_image_repository:$build_name" "$rhel_registry/$operator_rhel_docker_image_repository:$build_name"
docker push "$rhel_registry/$operator_rhel_docker_image_repository:$build_name"

docker_login "$rhel_registry" "$OPERATOR_BUNDLE_REDHAT_REGISTRY_USER" "$OPERATOR_BUNDLE_REDHAT_REGISTRY_KEY"

make \
  RH_BUNDLE_IMAGE_REPOSITORY=$bundle_rhel_docker_image_repository \
  RH_BUNDLE_IMAGE_TAG=$build_name \
  RH_BUNDLE_REGISTRY=$rhel_registry \
  RH_BUNDLE_VERSION=$build_name \
  release/bundle-image

tc_end_block "Make and push docker images"

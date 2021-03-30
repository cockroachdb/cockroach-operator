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
# TODO: the following check can be removed after
# https://github.com/cockroachdb/cockroach-operator/pull/378 is merged
# Using the version.txt file is the preferred method, because it the change has
# to be reviewed. Settin an arbitrary version may break the OpenShift metadata
# build process.
if [ -e version.txt ]; then
  echo "Using version from version.txt"
  VERSION="v"$(cat version.txt)
else
  echo "Using version set by TeamCity"
  VERSION=$NAME
fi
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

docker_registry="docker.io/cockroachdb"

if [[ -z "${DRY_RUN}" ]] ; then
  docker_image_repository="cockroachdb-operator"
else
  docker_image_repository="cockroachdb-operator-misc"
  if [[ -z "$(echo ${build_name} | grep -E -o '^v[0-9]+\.[0-9]+\.[0-9]+$')" ]] ; then
    # Using `.` to match how we usually format the pre-release portion of the
    # version string using '.' separators.
    # ex: v20.2.0-rc.2.dryrun
    build_name="${build_name}.dryrun"
  else
    # Using `-` to put dryrun in the pre-release portion of the version string.
    # ex: v20.2.0-dryrun
    build_name="${build_name}-dryrun"
  fi
fi

tc_end_block "Variable Setup"

tc_start_block "Make and push docker images"
configure_docker_creds
docker_login

if docker_image_exists "$docker_registry/$docker_image_repository:$build_name"; then
  echo "Docker image $docker_registry/$docker_image_repository:$build_name already exists"
  if [[ -z "${FORCE}" ]] ; then
    echo "Use FORCE=1 to force push the docker image."
    echo "Alternatively you can delete the tag in Docker Hub."
    exit 1
  fi
  echo "Forcing docker push..."
fi

make \
  DOCKER_REGISTRY="$docker_registry" \
  DOCKER_IMAGE_REPOSITORY="$docker_image_repository" \
  release/image
tc_end_block "Make and push docker images"

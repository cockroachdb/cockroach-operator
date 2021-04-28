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

rhel_registry="scan.connect.redhat.com"
rh_bundle_image_repository="ospid-857fe786-3eb7-4508-aafd-cc74c1b1dc24/cockroachdb-operator-bundle"
image="$rhel_registry/$rh_bundle_image_repository:$image_tag"

if ! [[ -z "${DRY_RUN}" ]] ; then
  image="${image}-dryrun"
fi
tc_end_block "Variable Setup"


tc_start_block "Make and push docker images"
configure_docker_creds
docker_login "$rhel_registry" "$OPERATOR_BUNDLE_REDHAT_REGISTRY_USER" "$OPERATOR_BUNDLE_REDHAT_REGISTRY_KEY"

# TODO(rail): switch to bazel generated images when it supports "FROM: scratch"
cd deploy/certified-metadata-bundle/cockroach-operator
docker build -t $image -f bundle.Dockerfile .
docker push $image

tc_end_block "Make and push docker images"

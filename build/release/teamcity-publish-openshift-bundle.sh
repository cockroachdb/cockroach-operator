#!/usr/bin/env bash
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

set -euxo pipefail

source "$(dirname "${0}")/teamcity-support.sh"

# Default values are defined for the certified bundle.
RH_PROJECT="5f5a433f9d6546ed7aa8634d"
RH_REGISTRY="scan.connect.redhat.com"
RH_REPO="ospid-857fe786-3eb7-4508-aafd-cc74c1b1dc24/cockroachdb-operator-bundle"
BUNDLE_DIR="bundle/cockroachdb-certified"

# If this is the marketplace bundle, update accordingly.
if ! [[ -z "${MARKETPLACE}" ]]; then
  RH_PROJECT="61765afbdd607bfc82e643b8"
  RH_REPO="ospid-61765afbdd607bfc82e643b8/cockroachdb-operator-bundle-marketplace"
  BUNDLE_DIR="bundle/cockroachdb-certified-rhmp"
fi

# If it's a dry run, add -dryrun to the image
if ! [[ -z "${DRY_RUN}" ]]; then RH_REPO="${RH_REPO}-dryrun"; fi

IMAGE="${RH_REGISTRY}/${RH_REPO}:${TAG}"

main() {
  docker_login "${RH_REGISTRY}" "${OPERATOR_REDHAT_REGISTRY_USER}" "${OPERATOR_REDHAT_REGISTRY_KEY}"

  generate_bundle
  publish_bundle_image
  run_preflight
}

generate_bundle() {
  # create the certified and marketplace bundles
  tc_start_block "Generate bundle"
  make release/generate-bundle
  tc_end_block "Generate bundle"
}

publish_bundle_image() {
  tc_start_block "Make and push bundle image"

  pushd "${BUNDLE_DIR}"
  docker build -t "${IMAGE}" .
  docker push "${IMAGE}"
  popd

  tc_end_block "Make and push bundle image"
}

run_preflight() {
  bazel build //hack/bin:preflight
  PFLT_PYXIS_API_TOKEN="${REDHAT_API_TOKEN}" bazel-bin/hack/bin/preflight \
    check operator "${IMAGE}" --docker-config ~/.docker/config.json
}

main "$@"

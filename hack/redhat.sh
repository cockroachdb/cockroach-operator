#!/usr/bin/env bash

# Copyright 2026 The Cockroach Authors
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
set -euo pipefail

# include bazel binaries in the path
PATH="bazel-bin/hack/bin:${PATH}"

# Global preflight settings
export PFLT_DOCKERCONFIG="${HOME}/.docker/config.json"
export PFLT_LOGLEVEL=debug

OPERATOR="cockroachdb/cockroach-operator"
REGISTRY="gcr.io/${GCP_PROJECT}"
VERSION="$(cat version.txt)"
IMAGE="${OPERATOR}:v${VERSION}"
BUNDLE_IMAGE="${OPERATOR}-bundle:v${VERSION}"
BUNDLE_INDEX="${OPERATOR}-index:v${VERSION}"
RHMP_BUNDLE_IMAGE="${OPERATOR}-bundle-rhmp:v${VERSION}"
RHMP_BUNDLE_INDEX="${OPERATOR}-index-rhmp:v${VERSION}"

main() {
  # Switch to the build directory. The bundle directory is not part of source
  # controller and therefore isn't a bazel target. This means when we run this
  # script, there's no way to reference the Dockerfile created by the call to
  # make release/generate-bundle (prerequisite of make test/preflight). By cd'ing
  # into the build directory, we'll have access to _all_ the files.
  if [[ -n "${BUILD_WORKSPACE_DIRECTORY}" ]]; then
    cd "${BUILD_WORKSPACE_DIRECTORY}"
  fi

  case "${1:-}" in
    operator)
      publish_operator_image
      preflight_operator;;
    bundle)
			publish_bundle_image "${REGISTRY}/${BUNDLE_IMAGE}" "bundle/cockroachdb-certified"
			publish_bundle_index "${REGISTRY}/${BUNDLE_IMAGE}" "${REGISTRY}/${BUNDLE_INDEX}"
			preflight_bundle "${REGISTRY}/${BUNDLE_IMAGE}" "${REGISTRY}/${BUNDLE_INDEX}"
			ensure_success;;
    marketplace)
			publish_bundle_image "${REGISTRY}/${RHMP_BUNDLE_IMAGE}" "bundle/cockroachdb-certified-rhmp"
			publish_bundle_index "${REGISTRY}/${RHMP_BUNDLE_IMAGE}" "${REGISTRY}/${RHMP_BUNDLE_INDEX}"
			preflight_bundle "${REGISTRY}/${RHMP_BUNDLE_IMAGE}" "${REGISTRY}/${RHMP_BUNDLE_INDEX}"
			ensure_success;;
    *)
      echo "ERROR: Unknown command: ${1}" 1>&2
      echo "Usage bazel run //hack:redhat-preflight -- <operator|bundle|marketplace>." 1>&2
      exit 1;;
  esac
}

publish_operator_image() {
  echo "Publishing operator image to local repo..."
  APP_VERSION="v${VERSION}" \
  DOCKER_REGISTRY="${REGISTRY}" \
  DOCKER_IMAGE_REPOSITORY="${IMAGE%:*}" \
  bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //:push_operator_image
}

preflight_operator() {
  echo "Running preflight checks on operator image..."
  preflight check container "${REGISTRY}/${IMAGE}" \
  --docker-config "${HOME}/.docker/config.json"
}

publish_bundle_image() {
	local img="${1}"
	local dir="${2}"

  echo "Publishing ${img}..."
  pushd "${dir}"
  docker build -t "${img}" .
  docker push "${img}"
  popd
}

publish_bundle_index() {
	local bundle_img="${1}"
	local index_img="${2}"

  echo "Publishing ${index_img}..."
  opm index add \
  --container-tool docker \
  --bundles "${bundle_img}" \
  --tag "${index_img}"
}

preflight_bundle() {
	local bundle_img="${1}"
	local index_img="${2}"

  echo "Running preflight checks on bundle image..."
	echo "IMAGE: ${bundle_img}"

	PFLT_INDEXIMAGE="${index_img}" preflight check operator "${bundle_img}"
}

ensure_success() {
	if [[ $(cat artifacts/results.json | jq -r .passed) == 'false' ]]; then
		# error already displayed
		exit 1
	fi
}

main "$@"

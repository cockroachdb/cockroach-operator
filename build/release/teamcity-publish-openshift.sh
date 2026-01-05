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

set -euxo pipefail

source "$(dirname "${0}")/teamcity-support.sh"

RH_PROJECT_ID="5e6027425c5456060d5f6084"
RH_REGISTRY="quay.io"
RH_OPERATOR_IMG="${RH_REGISTRY}/redhat-isv-containers/${RH_PROJECT_ID}:${TAG}"

OPERATOR_IMG="docker.io/cockroachdb/cockroach-operator:${TAG}"
if ! [[ -z "${DRY_RUN}" ]] ; then
  OPERATOR_IMG="docker.io/cockroachdb/cockroach-operator-misc:${TAG}-dryrun"
fi

main() {
  docker_login "${RH_REGISTRY}" "${OPERATOR_REDHAT_REGISTRY_USER}" "${OPERATOR_REDHAT_REGISTRY_KEY}"

  publish_to_redhat
  run_preflight
}

publish_to_redhat() {
  tc_start_block "Tag and release docker image"
  docker pull "${OPERATOR_IMG}"
  docker tag "${OPERATOR_IMG}" "${RH_OPERATOR_IMG}"
  docker push "${RH_OPERATOR_IMG}"
  tc_end_block "Tag and release docker image"
}

run_preflight() {
  bazel build //hack/bin:preflight
  PFLT_PYXIS_API_TOKEN="${REDHAT_API_TOKEN}" bazel-bin/hack/bin/preflight \
    check container "${RH_OPERATOR_IMG}" \
    --certification-project-id="${RH_PROJECT_ID}" \
    --docker-config=/home/agent/.docker/config.json \
    --submit
}

main "$@"

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
set -euo pipefail

PATH="bazel-bin/hack/bin:${PATH}"

APP_VERSION="${APP_VERSION:-v$(cat version.txt)}"
CLUSTER_NAME="dev"
NODE_IMAGE="rancher/k3s:v1.23.3-k3s1"
REGISTRY_NAME="registry.localhost"
REGISTRY_PORT=5000
DEV_REGISTRY="${REGISTRY_NAME}:${REGISTRY_PORT}"

main() {
  case "${1:-}" in
    up)
    create_cluster
    install_operator
    wait_for_ready;;
    down)
      k3d cluster delete "${CLUSTER_NAME}";;
    *) echo "Unknown command. Usage $0 <up|down>" 1>&2; exit 1;;
  esac
}

create_cluster() {
  k3d cluster create "${CLUSTER_NAME}" \
    --image "${NODE_IMAGE}" \
    --registry-create "${REGISTRY_NAME}:${REGISTRY_PORT}"
}

install_operator() {
  # Can't seem to figure out how to leverage the stamp variables here. So for
  # now I've added a defined make variable which can be used for substitution
  # in //config/default/BUILD.bazel.
  # ${REGISTRY_NAME} must be mapped to localhost or 127.0.0.1 to push the image to the k3d registry.
  bazel run //hack/crdbversions:crdbversions -- -operator-image ${DEV_REGISTRY}/cockroach-operator -operator-version ${APP_VERSION} -crdb-versions $(PWD)/crdb-versions.yaml -repo-root $(PWD)
  K8S_CLUSTER="k3d-${CLUSTER_NAME}" \
    DEV_REGISTRY="${DEV_REGISTRY}" \
    bazel run \
    --stamp \
    --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
    --define APP_VERSION=${APP_VERSION} \
    //config/default:install.apply
}

wait_for_ready() {
  echo "Waiting for deployment to be available..."
  kubectl wait \
    --for=condition=Available \
    --timeout=2m \
    -n cockroach-operator-system \
    deploy/cockroach-operator-manager
}

main "$@"

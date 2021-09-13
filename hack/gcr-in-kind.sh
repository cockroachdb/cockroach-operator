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
set -euo pipefail

# ensure docker login stores results in our config
export DOCKER_CONFIG="$(mktemp -d)"

KIND_BIN="bazel-bin/hack/bin/kind"
KIND_CLUSTER="test"

remove_config() {
  rm -rf "${DOCKER_CONFIG}"
}

# ensure the file is deleted when we're done
trap remove_config EXIT

# copies your access token to the docker config file
generate_config() {
  echo '{"auths":{"gcr.io":{}}}' > "${DOCKER_CONFIG}/config.json"

  gcloud auth print-access-token | docker login \
    -u oauth2accesstoken \
    --password-stdin \
    https://gcr.io \
    2>/dev/null >/dev/null
}

# copies the docker config to all nodes and restarts kubelet
update_nodes() {
  for node in $(${KIND_BIN} get nodes --name ${KIND_CLUSTER}); do
    docker cp "${DOCKER_CONFIG}/config.json" "${node}:/var/lib/kubelet/config.json"
    docker exec "${node}" systemctl restart kubelet.service
  done
}

main() {
  generate_config
  update_nodes
}

main "$@"

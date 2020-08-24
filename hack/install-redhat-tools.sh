#!/usr/bin/env bash

# Copyright 2020 The Cockroach Authors
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

set -o errexit
set -o nounset
set -o pipefail

#
# Install script that installs the different tooling for generating the bundle for RedHat Marketplace
#
RELEASE_VERSION="v0.18.0"
OPM_VERSION="v1.12.5"
COURIER_VERSION="v2.1.7"
BIN_DIR="${GOPATH}/bin"

# TODO move this into bazel

echo "Installing operator-sdk"
cd /tmp
curl -LO https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
chmod +x operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu 
cp operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu ${BIN_DIR}/operator-sdk 
rm operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu

echo "Building operator-courier-sdk"
git clone https://github.com/operator-framework/operator-courier
cd operator-courier
git fetch && git fetch --tags
git checkout ${COURIER_VERSION}
docker build --tag operator-courier/operator-courier:${COURIER_VERSION} ./
cd ..
rm -rf operator-courier

echo "Installing opm"
curl -LO https://github.com/operator-framework/operator-registry/releases/download/${OPM_VERSION}/linux-amd64-opm
chmod +x linux-amd64-opm 
cp linux-amd64-opm ${BIN_DIR}/opm 
rm linux-amd64-opm
curl -LO https://github.com/mikefarah/yq/releases/download/3.3.2/yq_linux_amd64
chmod +x yq_linux_amd64
mv yq_linux_amd64 ${BIN_DIR}/opm

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
shopt -s globstar

################################################################################
# This script generates an OPM bundle for distribution to OpenShift.
#
# It generates an overlay for creating the bundle image consisting of the
# manifests and crds for the current version.
################################################################################

if [[ -z "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
  echo 'Must be run via "make release/opm-build-bundle"' >&2
  exit 1
fi

export PATH="$(pwd)/hack/bin:${PATH}"

main() {
  local version="${1}"
  local image="${2}"
  local package_opts="${3}"
  local cert_path="deploy/certified-metadata-bundle"
  local deploy_path="${cert_path}/cockroach-operator"

  echo "+++ Running opm to create bundle"
  echo "VERSION=${version}"

	mkdir -p "${deploy_path}"
  cd "${BUILD_WORKSPACE_DIRECTORY}/${deploy_path}"
  generate_bundle "${version}"
  copy_files "${version}"
}

generate_bundle() {
  opm alpha bundle generate -d "./${1}/" -u "./${1}/" -c beta,stable -e stable
  sed "s/VERSION/${1}/g" ../bundle.Dockerfile > ./bundle-${1}.Dockerfile
}

copy_files() {
  cp ./bundle-${1}.Dockerfile ./bundle.Dockerfile
	rm latest/**/*.yaml
  cp -R ./${1}/manifests/*.yaml ./latest/manifests
  cp -R ./${1}/manifests/*.yaml ./latest
  cp -R ./${1}/metadata/*.yaml ./latest/metadata
}

main "$@"

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

set -o nounset
set -o errexit
set -o pipefail

# enable **/*.yaml
shopt -s globstar

if [[ -z "${TEST_WORKSPACE:-}" ]]; then
  echo 'Must be run with bazel via "bazel test //hack:verify-crds"' >&2
  exit 1
fi

FILES=(apis cmd config external go.mod go.sum hack pkg)

echo "Verifying generated CRD manifests are up-to-date..." >&2

create_working_dir() {
  local dir="${TEST_TMPDIR}/files"
  mkdir -p "${dir}"

  # copy necessary files
  cp -RL ${FILES[@]} "${dir}"

  # copy config to config_
  pushd "${dir}" >/dev/null
  cp -RHL config config_
}

# TODO: Ideally we'd back able to share this with the makefile
generate_crds() {
  HOME="${TEST_TMPDIR}/home" "${1}" crd:trivialVersions=true \
    rbac:roleName=role \
    webhook \
    paths=./... \
    output:crd:artifacts:config=config/crd/bases

  hack/boilerplaterize hack/boilerplate/boilerplate.yaml.txt config/**/*.yaml
}

run_diff() {
  # Avoid diff -N so we handle empty files correctly
  echo "diffing"
  diff=$(diff -upr "config_" "config" 2>/dev/null || true)

  if [[ -n "${diff}" ]]; then
    echo "${diff}" >&2
    echo >&2
    echo "generated CRDs are out of date. Please run 'make dev/update-crds" >&2
    exit 1
  fi

  echo "SUCCESS: generated CRDs up-to-date"
}

main() {
  create_working_dir
  generate_crds "${2}"
  run_diff
}

main "$@"

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
shopt -s extglob

if [[ -z "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
  echo 'Must be run via "make release/update-pkg-bundle"' >&2
  exit 1
fi

export PATH="$(pwd)/hack/bin:${PATH}"

main() {
  local rh_bundle_version="${1}"
  local rh_operator_image="${2}"
  local rh_package_options="${3}"
  local rh_crdb_image="${4}"
  local cert_path="deploy/certified-metadata-bundle"
  local deploy_path="${cert_path}/cockroach-operator"

  echo "+++ Running update package manifest for certification"
  echo "RH_BUNDLE_VERSION=${rh_bundle_version}"
  echo "RH_OPERATOR_IMAGE=${rh_operator_image}"
  echo "RH_PACKAGE_OPTS=${rh_package_options}"
  echo "RH_CRDB_IMAGE=${rh_crdb_image}"
  echo "CERTIFICATION_PATH=${cert_path}"
  echo "DEPLOY_PATH=${deploy_path}"

  cd "${BUILD_WORKSPACE_DIRECTORY}"
  ensure_unique_deployment "${deploy_path}/${rh_bundle_version}"
  generate_package_bundle "${rh_bundle_version}" "${rh_package_options}" "${deploy_path}"
  generate_csv "${deploy_path}/${rh_bundle_version}/manifests" "${rh_operator_image}"
  combine_files "${deploy_path}/${rh_bundle_version}" "${rh_bundle_version}" "${deploy_path}"
}

ensure_unique_deployment() {
  if [ -d "${1}" ]; then
    echo "Folder ${1} already exists. Please increase the version or remove the folder manually." >&2
    exit 1
  fi
}

generate_package_bundle() {
  # Generate CSV in config/manifests and add boilerplate back (lost when regenerating)
  operator-sdk generate kustomize manifests -q --apis-dir apis
  hack/boilerplaterize hack/boilerplate/boilerplate.yaml.txt config/manifests/**/*.yaml

  # Generate the new package bundle
  kustomize build config/manifests | operator-sdk generate bundle -q \
    --version "${1}" \
    ${2} \
    --output-dir "${3}/${1}"

  # For some reason, the package name specified in the annotations is "cockroachdb-certified". We need to manually
  # replace that in here to maintain compatibility with previous versions.
  sed -i '' "s/package.v1=cockroach-operator/package.v1=cockroachdb-certified/g" bundle.Dockerfile
  sed -i '' "s/package.v1: cockroach-operator/package.v1: cockroachdb-certified/g" ${3}/${1}/metadata/annotations.yaml

  # There's no way to specify where bundle.Dockerfile ends up, so we do some post-processing on it here.
  sed "s#${3}/##g" bundle.Dockerfile > ${3}/bundle.Dockerfile
  rm bundle.Dockerfile
}

generate_csv() {
  # replace RH_COCKROACH_OP_IMAGE_PLACEHOLDER with the proper image and CREATED_AT_PLACEHOLDER with the current time
  cat ${1}/cockroach-operator.clusterserviceversion.yaml | sed \
    "s+RH_COCKROACH_OP_IMAGE_PLACEHOLDER+${2}+g; s+CREATED_AT_PLACEHOLDER+"$(date +"%FT%H:%M:%SZ")"+g" > ${1}/csv.yaml

  # for each RH_COCKROACH_DB_IMAGE_PLACEHOLDER_* set to the corresponding connect image
  local version env img
  for v in $(faq -r '.CrdbVersions' crdb-versions.yaml | cut -d ' ' -f2); do
    version=${v//./_}
    env="RH_COCKROACH_DB_IMAGE_PLACEHOLDER_${version}"
    img="registry.connect.redhat.com/cockroachdb/cockroach:${v}"
    sed -i '' -e "s+${env}+${img}+g" "${1}/csv.yaml"
  done
}

combine_files() {
  pushd "${1}/manifests" >/dev/null

  local csv="cockroach-operator.v${2}.clusterserviceversion.yaml"

  # sticks all the necessary cluster permissions into the operator's CSV yaml
  faq -f yaml -o yaml \
    --slurp '.[0].spec.install.spec.clusterPermissions+= [{serviceAccountName: .[3].metadata.name, rules: .[1].rules }] | .[0]' \
    csv.yaml cockroach-*.yaml webhook-service*.yaml \
    > "${csv}"

  # remove the "replaces" attribute. This is a new bundle
  sed '/replaces: .*/d' "${csv}" > tmp-csv && mv tmp-csv "${csv}"

  # delete all except the new csv and crd files
  rm -v !("${csv}"|"crdb.cockroachlabs.com_crdbclusters.yaml")

  popd >/dev/null

  # update the latest directory with all the new stuff
  rm -r ${3}/latest/*
  cp -R ${1}/* ${3}/latest
}

main "$@"

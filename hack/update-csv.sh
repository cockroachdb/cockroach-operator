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

set -o errexit
set -o nounset
set -o pipefail

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then # Running inside bazel
  echo "Generating  CSV..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:update-csv
  )
  exit 0
fi

opsdk=$(realpath "$1")
kstomize="$(realpath "$2")"
opm="$(realpath "$3")"
faq="$(realpath "$4")"
export PATH=$(dirname "$opsdk"):$PATH

# This script should be run via `bazel run //hack:gen-csv`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}
echo "+++ Running gen csv for development mode"
[[ -z "$5" ]] && { echo "Error: RH_BUNDLE_VERSION not set"; exit 1; }
RH_BUNDLE_VERSION="$5"
echo "RH_BUNDLE_VERSION=$RH_BUNDLE_VERSION"
[[ -z "$6" ]] && { echo "Error: RH_COCKROACH_OP_IMG not set"; exit 1; }
RH_COCKROACH_OP_IMG="$6"
echo "RH_COCKROACH_OP_IMG=$RH_COCKROACH_OP_IMG"
[[ -z "$7" ]] && { echo "Error: RH_BUNDLE_METADATA_OPTS not set"; exit 1; }
RH_BUNDLE_METADATA_OPTS="$7"
echo "RH_BUNDLE_METADATA_OPTS=$RH_BUNDLE_METADATA_OPTS"
[[ -z "$9" ]] && { echo "Error: RH_COCKROACH_DATABASE_IMAGE not set"; exit 1; }
RH_COCKROACH_DATABASE_IMAGE="$9"
echo "RH_COCKROACH_DATABASE_IMAGE=$RH_COCKROACH_DATABASE_IMAGE"
"$opsdk" generate kustomize manifests -q 
"$kstomize" build config/manifests | "$opsdk" generate bundle -q --overwrite --version ${RH_BUNDLE_VERSION} ${RH_BUNDLE_METADATA_OPTS}
"$opsdk" bundle validate ./bundle
cat bundle/manifests/cockroach-operator.clusterserviceversion.yaml | sed -e "s+RH_COCKROACH_OP_IMAGE_PLACEHOLDER+${RH_COCKROACH_OP_IMG}+g" -e "s+RH_COCKROACH_DB_IMAGE_PLACEHOLDER+${RH_COCKROACH_DATABASE_IMAGE}+g" -e "s+CREATED_AT_PLACEHOLDER+"$(date +"%FT%H:%M:%SZ")"+g"> bundle/manifests/cockroach-operator.clusterserviceversion.yaml 
cd  bundle/manifests && "$faq" -f yaml -o yaml --slurp '.[0].spec.install.spec.clusterPermissions+= [{serviceAccountName: .[2].metadata.name, rules: .[1].rules }] | .[0]' cockroach-operator.clusterserviceversion.yaml cockroach-database-role_rbac.authorization.k8s.io_v1_clusterrole.yaml cockroach-database-sa_v1_serviceaccount.yaml > csv.yaml
mv csv.yaml cockroach-operator.clusterserviceversion.yaml 
shopt -s extglob
rm -v !("cockroach-operator.clusterserviceversion.yaml"|"crdb.cockroachlabs.com_crdbclusters.yaml") 
shopt -u extglob
cd ${REPO_ROOT}
FILE_NAMES=(bundle/tests/scorecard/config.yaml \
bundle/manifests/cockroach-operator.clusterserviceversion.yaml \
bundle/manifests/crdb.cockroachlabs.com_crdbclusters.yaml \
bundle/metadata/annotations.yaml \
config/manifests/bases/cockroach-operator.clusterserviceversion.yaml \
)
for YAML in "${FILE_NAMES[@]}"
do
   :
   cat "${REPO_ROOT}/hack/boilerplate/boilerplate.yaml.txt" "${REPO_ROOT}/${YAML}" > "${REPO_ROOT}/${YAML}.mod"
   mv "${REPO_ROOT}/${YAML}.mod" "${REPO_ROOT}/${YAML}"
done 

DOCKER_FILE_NAME="bundle.Dockerfile"

cat "${REPO_ROOT}/hack/boilerplate/boilerplate.Dockerfile.txt" "${REPO_ROOT}/${DOCKER_FILE_NAME}" > "${REPO_ROOT}/${DOCKER_FILE_NAME}.mod"
mv "${REPO_ROOT}/${DOCKER_FILE_NAME}.mod" "${REPO_ROOT}/${DOCKER_FILE_NAME}"
 





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

if [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then # Running inside bazel
  echo "Generating  CSV..." >&2
elif ! command -v bazel &>/dev/null; then
  echo "Install bazel at https://bazel.build" >&2
  exit 1
else
  (
    set -o xtrace
    bazel run //hack:update-pkg
  )
  exit 0
fi

# This script should be run via `bazel run //hack:gen-csv`
REPO_ROOT=${BUILD_WORKSPACE_DIRECTORY}
cd "${REPO_ROOT}"
echo ${REPO_ROOT}

echo "+++ Running operator-sdk"
VERSION="$4"
echo "VERSION:$VERSION"
IMG="$5"
echo $IMG
PKG_MAN_OPTS="$6"
echo "PKG_MAN_OPTS: $PKG_MAN_OPTS"
DEPLOY_PATH="deploy/certified-metadata-bundle/cockroach-operator/"
DEPLOY_CERTIFICATION_PATH="deploy/certified-metadata-bundle"

if [ -d "${DEPLOY_PATH}/${VERSION}" ] 
then
    echo "Folder ${DEPLOY_PATH}/${VERSION} already exists. Please increase the version or remove the folder manually." 
fi

operator-sdk generate kustomize manifests -q
cd manifests && kustomize edit set image cockroachdb/cockroach-operator=${IMG} && cd ..
kustomize build config/manifests | operator-sdk generate packagemanifests -q --version ${VERSION} ${PKG_MAN_OPTS} --output-dir ${DEPLOY_PATH} --input-dir ${DEPLOY_PATH}



# Keep the original format so we can rollback anytime
mv ${DEPLOY_PATH}/${VERSION}/cockroach-operator.clusterserviceversion.yaml ${DEPLOY_PATH}/${VERSION}/cockroach-operator.v${VERSION}.clusterserviceversion.yaml
[ ! -d ${DEPLOY_PATH}/${VERSION}/manifests ] && mkdir ${DEPLOY_PATH}/${VERSION}/manifests
cp ${DEPLOY_PATH}/${VERSION}/*.* ${DEPLOY_PATH}/${VERSION}/manifests
[ ! -d ${DEPLOY_PATH}/${VERSION}/metadata ] && mkdir  ${DEPLOY_PATH}/${VERSION}/metadata
cp ${DEPLOY_CERTIFICATION_PATH}/annotations.yaml ${DEPLOY_PATH}/${VERSION}/metadata
sed "s/VERSION/${VERSION}/g" ${DEPLOY_CERTIFICATION_PATH}/bundle.Dockerfile > ${DEPLOY_PATH}/bundle.v${VERSION}.Dockerfile
cp ${DEPLOY_PATH}/bundle.v${VERSION}.Dockerfile ${DEPLOY_PATH}/bundle.Dockerfile

cd  ${DEPLOY_PATH}  && ./opm alpha bundle generate -d ./${VERSION}/ -u ./${VERSION}/ -p  -c beta,stable -e stable



# # Add licence to the generated files... for certification maybe we will need to remove 
# FILE_NAMES=(${DEPLOY_PATH}/${VERSION}/manifests/cockroach-operator-sa_v1_serviceaccount.yaml \
# ${DEPLOY_PATH}/${VERSION}/manifests/cockroach-operator.v${VERSION}.clusterserviceversion.yaml \
# ${DEPLOY_PATH}/${VERSION}/manifests/crdb.cockroachlabs.com_crdbclusters.yaml \
# config/manifests/bases/cockroach-operator.clusterserviceversion.yaml \
# )
# for YAML in "${FILE_NAMES[@]}"
# do
#    :
#    cat "${REPO_ROOT}/hack/boilerplate/boilerplate.yaml.txt" "${REPO_ROOT}/${YAML}" > "${REPO_ROOT}/${YAML}.mod"
#    mv "${REPO_ROOT}/${YAML}.mod" "${REPO_ROOT}/${YAML}"
# done 


# Move to latest folder for release -> I need a fixed folder name for the docker image that runs from bazel
rm -rf ${DEPLOY_PATH}/latest/*/*.yaml
cp -R ${DEPLOY_PATH}/${VERSION}/* ${DEPLOY_PATH}/latest








 





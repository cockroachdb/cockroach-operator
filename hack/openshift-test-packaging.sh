#!/usr/bin/env bash

# Copyright 2023 The Cockroach Authors
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


# "---------------------------------------------------------"
# "-                                                       -"
# "-  This script reads the current release version and    -"
# "-  increments that version.  Then a make target is run  -"
# "-  that builds and pushes the the operator image, and   -"
# "-  openshift build and index files. The same make       -"
# "-  target then runs an e2e test to ensure that the      -"
# "-  operator is installed on openshift and that we can   -"
# "-  can use example.yaml to create a cluster.            -"
# "-  This script takes two arguments. Example:            -"
# "-  -r us.gcr.io/chris-love-operator-playground          -"
# "-  -k clove                                             -"
# "-  It assumes that the openshift cluser kubeconfig      -"
# "-  lives in the standard directory.                     -"
# "-                                                       -"
# "-  If this script fails run cleanup-packaging.sh        -"
# "---------------------------------------------------------"

#
# TODO we might want to improve this buy sending in the kubeconfig
# file and not just the name of the server, but the Makefile
# I think uses that path as well.
#

set -o errexit
set -o nounset
set -o pipefail

command -v perl >/dev/null 2>&1 || { \
 echo >&2 "I require perl but it's not installed.  Aborting."; exit 1; }

# return a number that increments the current version
# in the version.txt file.
increment_version() {
  if [ ! -f "version.txt" ]; then
    echo "$KUBECONFIG does not exist."
    usage
  fi
  local v=$(cat version.txt)
  local rgx='^((?:[0-9]+\.)*)([0-9]+)($)'
  val=`echo -e "$v" | perl -pe 's/^.*'$rgx'.*$/$2/'`
  echo "$v" | perl -pe s/$rgx.*$'/${1}'`printf %0${#val}s $(($val+1))`/
}

usage() { echo "Usage: $0 [-o <openshift cluster name> -r <registry name>]" 1>&2; exit 1; }
# parse flags using getopts
while getopts ":o:r:" opt; do
  case ${opt} in
    o)
      OPENSHIFT=$OPTARG
      ;;
    r)
      DOCKER_REGISTRY=$OPTARG
      ;;
    \?)
      echo "Invalid flag on command line: $OPTARG" 1>&2
      ;;
    *)
      usage
      ;;
  esac
done
shift $((OPTIND -1))

# If user did not pass in -r flag then fail
if [ -z "${DOCKER_REGISTRY}" ]; then
    echo "-r flage not set"
    usage
fi

# If user did not pass in -o flag then fail
if [ -z "${OPENSHIFT}" ]; then
    echo "-o flage not set"
    usage
fi

VERSION=$(increment_version)
echo $VERSION
APP_VERSION=v$VERSION
RH_OPERATOR_IMAGE=${DOCKER_REGISTRY}/cockroachdb-operator:${APP_VERSION}
KUBECONFIG=${HOME}/openshift-${OPENSHIFT}/auth/kubeconfig

if [ ! -f "$KUBECONFIG" ]; then
    echo "The kubeconfig file does not exist: '$KUBECONFIG'"
    usage
fi

echo "Running make target with:"
echo "VERSION: $VERSION"
echo "RH_BUNDLE_VERSION=$VERSION"
echo "RH_OPERATOR_IMAGE=$RH_OPERATOR_IMAGE"
echo "DOCKER_REGISTRY=$DOCKER_REGISTRY"
echo "CLUSTER_NAME=$OPENSHIFT"
echo "KUBECONFIG=$KUBECONFIG"
APP_VERSION=$APP_VERSION \
  VERSION=$VERSION \
  RH_BUNDLE_VERSION=$VERSION \
  RH_OPERATOR_IMAGE=$RH_OPERATOR_IMAGE \
  DOCKER_REGISTRY=$DOCKER_REGISTRY \
  CLUSTER_NAME=$OPENSHIFT \
  KUBECONFIG=$KUBECONFIG \
  make -j 1 test/e2e/testrunner-openshift-packaging

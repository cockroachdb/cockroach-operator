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

set -x
set -o errexit
set -o nounset
set -o pipefail

#
# Creates a service account named open-shift-sa-$NAME and then adds the role owner
# to that services account. This script sets the current gcloud project to the
# project provided and does not reset the gcloud project value.
#

command -v gcloud >/dev/null 2>&1 || { \
 echo >&2 "I require gcloud but it's not installed.  Aborting."; exit 1; }

PROJ=
NAME=

usage() { echo "Usage: $0 [-p <gcp project> -n <name>]" 1>&2; exit 1; }

while getopts ":p:n:" opt; do
  case ${opt} in
    p)
      PROJ=$OPTARG
      ;;
    n)
      NAME=$OPTARG
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

if [ -z "${PROJ}" ]; then
    echo "Project not in cli args"
    usage
fi
if [ -z "${NAME}" ]; then
    echo "name not in cli args"
    usage
fi

SA=open-shift-sa-$NAME
gcloud config set project $PROJ

FULLID="${SA}@${PROJ}.iam.gserviceaccount.com"
gcloud iam service-accounts delete $FULLID

rm -rf $HOME/.gcp

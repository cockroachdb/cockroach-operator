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

#
# Creates a service account named open-shift-sa-$NAME and then adds the role owner
# to that services account. This script sets the current gcloud project to the
# project provided and does not reset the gcloud project value.
#

command -v gcloud >/dev/null 2>&1 || { \
 echo >&2 "I require gcloud but it's not installed.  Aborting."; exit 1; }
command -v openshift-install >/dev/null 2>&1 || { \
 echo >&2 "I require gcloud but it's not installed.  Aborting."; exit 1; }

PROJ=
PULL_SECRET=
DNS_ZONE=
NAME=
REGION=

usage() { echo "Usage: $0 [-p <gcp project> -s <pull secret file> -z <dns zone> -n <name> -r <gcp region>]" 1>&2; exit 1; }

while getopts ":p:s:z:n:r:" opt; do
  case ${opt} in
    r)
      REGION=$OPTARG
      ;;
    p)
      PROJ=$OPTARG
      ;;
    s)
      PULL_SECRET=$OPTARG
      ;;
    z)
      DNS_ZONE=$OPTARG
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

if [ -z "${REGION}" ]; then
    echo "Region not in cli args"
    usage
fi
if [ -z "${PROJ}" ]; then
    echo "Project not in cli args"
    usage
fi
if [ -z "${PULL_SECRET}" ]; then
    echo "pull secret file not in cli args"
    usage
fi
if [ -z "${DNS_ZONE}" ]; then
    echo "dns zone not in cli args"
    usage
fi
if [ -z "${NAME}" ]; then
    echo "name not in cli args"
    usage
fi

mkdir $HOME/.gcp-$NAME $HOME/openshift-$NAME

SA=open-shift-sa-$NAME
gcloud config set project $PROJ
gcloud services enable compute.googleapis.com
gcloud services enable cloudapis.googleapis.com
gcloud services enable cloudresourcemanager.googleapis.com
gcloud services enable dns.googleapis.com
gcloud services enable iam.googleapis.com
gcloud services enable iamcredentials.googleapis.com
gcloud services enable servicemanagement.googleapis.com
gcloud services enable serviceusage.googleapis.com
gcloud services enable storage-api.googleapis.com
gcloud services enable storage-component.googleapis.com

gcloud iam service-accounts create $SA
FULLID="${SA}@${PROJ}.iam.gserviceaccount.com"

# Compute Admin
# DNS Administrator
# Security Admin
# Service Account Admin
# Service Account User
# Storage Admin
declare -a ROLES=(
"compute.admin"
"dns.admin"
"iam.securityAdmin"
"iam.serviceAccountAdmin"
"iam.serviceAccountUser"
"iam.serviceAccountKeyAdmin"
"storage.admin"
"servicemanagement.quotaViewer"
)

for R in "${ROLES[@]}"
do
   gcloud projects add-iam-policy-binding $PROJ --member=serviceAccount:$FULLID --role=roles/${R}
done

# TODO make account token name unique
# TODO set a variable for install-config
gcloud iam service-accounts keys create ~/.gcp-$NAME/osServiceAccount.json --iam-account $FULLID
PULL_SECRET=$(<${PULL_SECRET})

cat <<EOT >> $HOME/openshift-$NAME/install-config.yaml
apiVersion: v1
baseDomain: ${DNS_ZONE}
metadata:
  name: ${NAME}-openshift
platform:
  gcp:
    projectID: ${PROJ}
    region: ${REGION}
pullSecret: '${PULL_SECRET}'
EOT

GOOGLE_CREDENTIALS=$HOME/.gcp-$NAME/osServiceAccount.json \
 openshift-install create cluster --dir $HOME/openshift-$NAME --log-level=debug 

cat << EOF
========

OpenShift should be configured properly.

To configure oc export or set your KUBECONFIG variable.  

For example:

$ export KUBECONFIG=$HOME/openshift-$NAME/auth/kubeconfig 

To test that oc is connecting properly run

$ oc whoami

=======
EOF

#!/usr/bin/env bash

# Copyright 2024 The Cockroach Authors
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
# "-  Common commands for some scripts                     -"
# "-                                                       -"
# "---------------------------------------------------------"

# gcloud and kubectl are required for this POC
command -v gcloud >/dev/null 2>&1 || { \
 echo >&2 "I require gcloud but it's not installed.  Aborting."; exit 1; }
command -v kubectl >/dev/null 2>&1 || { \
 echo >&2 "I require kubectl but it's not installed.  Aborting."; exit 1; }
command -v kustomize >/dev/null 2>&1 || { \
 echo >&2 "I require kustomize but it's not installed.  Aborting."; exit 1; }

usage() { echo "Usage: $0 [-c <cluster name> -r <REGION> -z <ZONE>]" 1>&2; exit 1; }
# parse -c flag for the CLUSTER_NAME using getopts
while getopts ":c:i:z:r:" opt; do
  case ${opt} in
    c)
      CLUSTER_NAME=$OPTARG
      ;;
    i)
      IMAGE_NAME=$OPTARG
      ;;
    z)
      ZONE=$OPTARG
      ;;
    r)
      REGION=$OPTARG
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

export IMAGE_NAME

# If user did not pass in -c flag then fail
if [ -z "${CLUSTER_NAME}" ]; then
    usage
fi

# If user did not pass in -z flag then get the default zone and use it or die
if [ -z "${ZONE}" ]; then
  # Get the default zone and use it or die
  ZONE=$(gcloud config get-value compute/zone)
  if [ -z "${ZONE}" ]; then
    echo "gcloud cli must be configured with a default zone." 1>&2
    echo "run 'gcloud config set compute/zone ZONE'." 1>&2
    echo "replace 'ZONE' with the zone name like us-west1-a." 1>&2
    echo "Or provide 'ZONE' as input to script as follows:" 1>&2
    usage
    exit 1;
  fi
fi

#If user did not pass in -r flag then get the default region and use it or die
if [ -z "${REGION}" ]; then
  REGION=$(gcloud config get-value compute/region)
  if [ -z "${REGION}" ]; then
    echo "gcloud cli must be configured with a default region." 1>&2
    echo "run 'gcloud config set compute/region REGION'." 1>&2
    echo "replace 'REGION' with the region name like us-west1." 1>&2
    echo "Or provide 'REGION' as input to script as follows:" 1>&2
    usage
    exit 1;
  fi
fi

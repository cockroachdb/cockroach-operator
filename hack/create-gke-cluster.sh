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
# "-  Create starts a GKE Cluster
# "-                                                       -"
# "---------------------------------------------------------"

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")
CLUSTER_NAME=""
# TODO maybe set default zone?
ZONE=""

REGION=""

# shellcheck disable=SC1090
source "$ROOT"/common.sh
# shellcheck disable=SC1090
source "$ROOT"/functions.sh

# enable googleapi for container registry on the project
enable-service compute.googleapis.com
enable-service container.googleapis.com

GKE_VERSION=$(gcloud container get-server-config \
  --format="value(validMasterVersions[0])" --region=$REGION)

# Get a comma separated list of zones from the default region
ZONESINREGION=""
for FILTEREDZONE in $(gcloud compute zones list --filter="region:$REGION" --format="value(name)" --limit 3)
do
  # Get a least 3 zones to run 3 nodes in
  ZONESINREGION+="$FILTEREDZONE,"
done
#Remove the last comma from the starting
ZONESINREGION=${ZONESINREGION%?}

# Create a GKE cluster
echo "Creating cluster"
gcloud container clusters create "$CLUSTER_NAME" \
  --zone "$ZONE" \
  --node-locations "$ZONESINREGION" \
  --cluster-version "$GKE_VERSION" \
  --machine-type "n1-standard-4" \
  --num-nodes=1 \
  --enable-network-policy \
  --enable-ip-alias

gcloud container clusters get-credentials "$CLUSTER_NAME" --zone "$ZONE"

echo "Cluster created"
echo "kubectl context set to cluster: $CLUSTER_NAME"

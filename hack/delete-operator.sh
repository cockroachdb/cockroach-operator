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

# "---------------------------------------------------------"
# "-                                                       -"
# "-  delete the operator 				   -"
# "-                                                       -"
# "---------------------------------------------------------"

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")
CLUSTER_NAME=""
# TODO maybe set default zone?
ZONE=""

# shellcheck disable=SC1090
source "$ROOT"/common.sh

gcloud container clusters get-credentials "$CLUSTER_NAME" --zone "$ZONE"

kubectl delete -f "${ROOT}/../config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml"
kubectl delete -f "${ROOT}/../manifests/operator.yaml"

# TODO create function to test if deployment still exists

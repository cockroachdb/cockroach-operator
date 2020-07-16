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

# "---------------------------------------------------------"
# "-                                                       -"
# "-  update yaml crd headers                              -"
# "-                                                       -"
# "---------------------------------------------------------"

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")
FILE_NAMES=(config/rbac/role.yaml config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml)

for YAML in "${FILE_NAMES[@]}"
do
   :
   cat "$ROOT/boilerplate/boilerplate.yaml.txt" "$ROOT/../$YAML" > "$ROOT/../$YAML.mod"
   mv "$ROOT/../$YAML.mod" "$ROOT/../$YAML"
done

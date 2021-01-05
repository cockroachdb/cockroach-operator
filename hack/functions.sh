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
# "-  common functions                                     -"
# "-                                                       -"
# "---------------------------------------------------------"

# check to see if a gcp service is enabled, and if it is not
# enable the service
function enable-service() {
  SERVICE=$1
  if [[ $(gcloud services list --format="value(NAME)" \
                                --filter="NAME:$SERVICE" 2>&1) != \
                                "$SERVICE" ]]; then
    echo "Enabling $SERVICE"
    gcloud services enable "$SERVICE"
  else
    echo "$SERVICE is already enabled"
  fi
}


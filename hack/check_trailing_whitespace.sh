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

# This code is based on existing work
# https://github.com/GoogleCloudPlatform/gke-terraform-generator/tree/master/test

check_trailing_whitespace

# This function makes sure that there is no trailing whitespace
# in any files in the project.
# There are some exclusions
function check_trailing_whitespace() {
  echo "Checking for trailing whitespace"
  grep -rn '[[:blank:]]$' --exclude-dir=".terraform" --exclude-dir=".docker-build" --exclude-dir="bazel-*" --exclude="*.png" --exclude-dir=".git" --exclude="*.pyc" .
  rc=$?
  if [ $rc = 0 ]; then
    exit 1
  fi
}

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

# Check that the .packages file contains all packages

ROOT=$(dirname "${BASH_SOURCE[0]}")
packages_file="${ROOT}/.packages"
if ! diff -u "${packages_file}" <(go list github.com/cockroachdb/cockroach-operator/...); then
	{
		echo
		echo "FAIL: ./hack/verify-packages.sh failed as the ./hack/.packages file is not in up to date."
		echo
		echo "FAIL: please execute the following command:  'go list github.com/cockroachdb/cockroach-operator/... > hack/.packages'"
		echo
	} >&2
	false
fi


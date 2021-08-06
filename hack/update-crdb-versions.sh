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

set -o nounset
set -o pipefail


# TODO(rail): we may need to add pagination handling in case we pass 500 versions
# Use anonymous API to get the list of published images from the RedHat Catalog.
URL="https://catalog.redhat.com/api/containers/v1/repositories/registry/registry.connect.redhat.com/repository/cockroachdb/cockroach/images?exclude=data.repositories.comparison.advisory_rpm_mapping,data.brew,data.cpe_ids,data.top_layer_id&page_size=500&page=0"


cat > crdb-versions.yaml << EOF
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
#
# Supported CockroachDB versions.
#
# This file contains a list of CockroachDB versions that are supported by the
# operator. hack/crdbversions/main.go uses this list to generate various
# manifests.
# Please update this file when CockroachDB releases new versions.

CrdbVersions:
EOF

# Skip unsupported versions and the latest tag
for version in $(curl $URL | jq -r '.data[] .repositories[] .tags[] .name' | grep -v ^v19 | grep -v latest | grep -v ubi$ | sort --version-sort); do
    echo "  - $version" >> crdb-versions.yaml
done

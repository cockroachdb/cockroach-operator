#!/bin/bash

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

bazel test //api/... //pkg/... \
	--test_env=TEST_ASSET_ETCD=$(bazel info execution_root)/external/etcd_bin/file/etcd \
	--test_env=TEST_ASSET_KUBECTL=$(bazel info execution_root)/external/kubectl_bin/file/kubectl \
	--test_env=TEST_ASSET_KUBE_APISERVER=$(bazel info execution_root)/external/kube_apiserver_bin/file/kube-apiserver \
	--test_env=KUBEBUILDER_ATTACH_CONTROL_PLANE_OUTPUT=true

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

# 
# This project requires the use of bazel.
# Install instuctions https://docs.bazel.build/versions/master/install.html
#

# values used in workspace-status.sh

DOCKER_REGISTRY?=us.gcr.io/chris-love-operator-playground
DOCKER_IMAGE_REPOSITORY?=cockroach-operator
APP_VERSION?=v1.0.0-alpha.2

# 
# Testing targets
# 
.PHONY: test/all
test/all:
	bazel test //api/... //pkg/...

.PHONY: test/api
test/api:
	bazel test //api/...

.PHONY: test/pkg
test/pkg:
	bazel test //pkg/...

# This runs the all of the verify scripts and
# takes a bit of time.
.PHONY: test/verify
test/verify:
	bazel test //hack/...

# This target uses kind to start a k8s cluster  and runs the e2e tests
# against that cluster.
.PHONY: test/e2e-short
test/e2e: 
	bazel test //e2e/... --test_arg=--test.short

# This target uses kind to start a k8s cluster  and runs the e2e tests
# against that cluster.
.PHONY: test/e2e
test/e2e: 
	bazel test //e2e/...

# 
# Different dev targets
#
.PHONY: dev/build
dev/build:
	bazel build //...

.PHONY: dev/fmt
dev/fmt:
	bazel run //hack:update-gofmt

.PHONY: dev/generate
dev/generate:
	bazel run //hack:update-codegen //hack:update-crds

# TODO move to bazel / go
.PHONY: dev/goimports
dev/goimports:
	@echo "Running goimports"
	@python3 hack/update_goimports.py

#
# Targets that allow to install the operator on an existing cluster
#
.PHONY: k8s/apply
k8s/apply:
	DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
	DOCKER_IMAGE_REPOSITORY=$(DOCKER_IMAGE_REPOSITORY) \
	APP_VERSION=$(APP_VERSION) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//manifests:install_operator.apply

.PHONY: k8s/delete
k8s/delete:
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//manifests:install_operator.delete
	
#
# Release targets
#
.PHONY: release/image
release/image:
	DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
	DOCKER_IMAGE_REPOSITORY=$(DOCKER_IMAGE_REPOSITORY) \
	APP_VERSION=$(APP_VERSION) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//:push_operator_image 

#
# Dev target that updates bazel files and dependecies
#
.PHONY: dev/syncdeps
dev/syncdeps:
	bazel run //hack:update-deps \
	bazel run //hack:update-bazel \
	bazel run //:gazelle -- update-repos -from_file=go.mod


		

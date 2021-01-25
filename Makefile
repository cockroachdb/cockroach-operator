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
# This project requires the use of bazel.
# Install instuctions https://docs.bazel.build/versions/master/install.html
#

# values used in workspace-status.sh

DOCKER_REGISTRY?=cockroachdb
DOCKER_IMAGE_REPOSITORY?=cockroachdb-operator
# Default bundle image tag
APP_VERSION?=v1.6.12-rc.2

# 
# Testing targets
# 
.PHONY: test/all
test/all:
	bazel test //api/... //pkg/... --test_arg=--test.v

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
test/e2e-short: 
	bazel test //e2e/... --test_arg=--test.short

# This target uses kind to start a k8s cluster  and runs the e2e tests
# against that cluster.
.PHONY: test/e2e
test/e2e: 
	bazel test --stamp //e2e/...

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

#RED HAT IMAGE BUNDLE
RH_BUNDLE_REGISTRY?=registry.connect.redhat.com/cockroachdb
RH_BUNDLE_IMAGE_REPOSITORY?=cockroachdb-operator-bundle
RH_BUNDLE_VERSION?=1.1.26
RH_DEPLOY_PATH="deploy/certified-metadata-bundle"
RH_DEPLOY_FULL_PATH="$(RH_DEPLOY_PATH)/cockroach-operator/"
RH_COCKROACH_DATABASE_IMAGE=registry.connect.redhat.com/cockroachdb/cockroach:v20.2.3
RH_OPERATOR_IMAGE?=registry.connect.redhat.com/cockroachdb/cockroachdb-operator:v1.6.12-rc.1

# Generate package manifests.
# Options for "packagemanifests".
CHANNEL?=beta
FROM_BUNDLE_VERSION?=1.0.1
IS_CHANNEL_DEFAULT?=0

ifneq ($(origin FROM_BUNDLE_VERSION), undefined)
PKG_FROM_VERSION := --from-version=$(FROM_BUNDLE_VERSION)
endif
ifneq ($(origin CHANNEL), undefined)
PKG_CHANNELS := --channel=$(CHANNEL)
endif
ifeq ($(IS_CHANNEL_DEFAULT), 1)
PKG_IS_DEFAULT_CHANNEL := --default-channel
endif
PKG_MAN_OPTS ?= "$(PKG_FROM_VERSION) $(PKG_CHANNELS) $(PKG_IS_DEFAULT_CHANNEL)"

# Build the packagemanifests
.PHONY: release/update-pkg-manifest
release/update-pkg-manifest:dev/generate
	bazel run  //hack:update-pkg-manifest  -- $(RH_BUNDLE_VERSION) $(RH_OPERATOR_IMAGE) $(PKG_MAN_OPTS) $(RH_COCKROACH_DATABASE_IMAGE)


#  Build the packagemanifests
.PHONY: release/opm-build-bundle
release/opm-build-bundle:
	bazel run  //hack:opm-build-bundle  -- $(RH_BUNDLE_VERSION) $(RH_OPERATOR_IMAGE) $(PKG_MAN_OPTS)

#
# Release bundle image
#
.PHONY: release/bundle-image
release/bundle-image:
	RH_BUNDLE_REGISTRY=$(RH_BUNDLE_REGISTRY) \
	RH_BUNDLE_IMAGE_REPOSITORY=$(RH_BUNDLE_IMAGE_REPOSITORY) \
	RH_BUNDLE_VERSION=$(RH_BUNDLE_VERSION) \
	RH_DEPLOY_PATH=$(RH_DEPLOY_FULL_PATH) \
	RH_BUNDLE_IMAGE_TAG=$(RH_BUNDLE_VERSION) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//:push_operator_bundle_image 


OLM_REPO ?= 
OLM_BUNDLE_REPO ?= cockroachdb-operator-index
OLM_PACKAGE_NAME ?= cockroachdb-certified
TAG ?= $(RH_BUNDLE_VERSION)
#
# dev opm index build for quay repo
#
.PHONY: dev/opm-build-index
dev/opm-build-index:
	RH_BUNDLE_REGISTRY=$(RH_BUNDLE_REGISTRY) \
	RH_BUNDLE_IMAGE_REPOSITORY=$(RH_BUNDLE_IMAGE_REPOSITORY) \
	RH_BUNDLE_VERSION=$(RH_BUNDLE_VERSION) \
	RH_DEPLOY_PATH=$(RH_DEPLOY_FULL_PATH) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//hack:opm-build-index $(OLM_REPO) $(OLM_BUNDLE_REPO) $(TAG) $(RH_BUNDLE_VERSION) $(RH_COCKROACH_DATABASE_IMAGE)

CHANNELS?=beta,stable
DEFAULT_CHANNEL?=stable
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Build the bundle image.
.PHONY: gen-csv
gen-csv: dev/generate
	bazel run  //hack:update-csv  -- $(RH_BUNDLE_VERSION) $(RH_OPERATOR_IMAGE) $(BUNDLE_METADATA_OPTS) $(RH_COCKROACH_DATABASE_IMAGE)
		


		

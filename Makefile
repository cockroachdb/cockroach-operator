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
VERSION ?= 0.0.10
# Default bundle image tag
APP_VERSION?=v1.0.0-alpha.3
DEFAULT_CHANNEL=alpha
CHANNELS=alpha

# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

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

kustomize:
ifeq (, $(shell which kustomize))
	@{ \
	set -e ;\
	KUSTOMIZE_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$KUSTOMIZE_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 ;\
	rm -rf $$KUSTOMIZE_GEN_TMP_DIR ;\
	}
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif
# Current Operator version

BUNDLE_IMG ?= cockroach-operator-bundle:$(VERSION)
# Default bundle image tag
# BUNDLE_IMG ?= cockroach-operator:$(VERSION)
# IMG="us.gcr.io/chris-love-operator-playground/cockroach-operator:v1.0.0-alpha.1"
IMG="quay.io/alinalion/cockroach-operator:v1.0.0-alpha.3"
# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: dev/generate
	operator-sdk generate kustomize manifests -q
	cd manifests && $(KUSTOMIZE) edit set image cockroach-operator=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t quay.io/alinalion/$(BUNDLE_IMG) .
	docker push quay.io/alinalion/$(BUNDLE_IMG)

# Generate package manifests.
# Options for "packagemanifests".
ifneq ($(origin FROM_VERSION), undefined)
PKG_FROM_VERSION := --from-version=$(FROM_VERSION)
endif
ifneq ($(origin CHANNEL), undefined)
PKG_CHANNELS := --channel=$(CHANNEL)
endif
ifeq ($(IS_CHANNEL_DEFAULT), 1)
PKG_IS_DEFAULT_CHANNEL := --default-channel
endif
PKG_MAN_OPTS ?= $(FROM_VERSION) $(PKG_CHANNELS) $(PKG_IS_DEFAULT_CHANNEL)

# Build packagemanifests.
.PHONY: packagemanifests
packagemanifests: dev/generate
	operator-sdk generate kustomize manifests -q
	cd manifests && $(KUSTOMIZE) edit set image cockroach-operator=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate packagemanifests -q --version $(VERSION) $(PKG_MAN_OPTS)
# Build the bundle image.
.PHONY: gen-csv
gen-csv:
	bazel run  //hack:update-csv  -- $(VERSION) $(IMG) $(BUNDLE_METADATA_OPTS)


		

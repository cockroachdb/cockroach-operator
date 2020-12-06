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

DOCKER_REGISTRY?=quay.io/alinalion
DOCKER_IMAGE_REPOSITORY?=cockroach-operator
VERSION ?= 0.0.27
# Default bundle image tag
APP_VERSION?=v1.0.5-alpha.3

IMG=$(DOCKER_REGISTRY)/$(DOCKER_IMAGE_REPOSITORY):$(APP_VERSION)
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
# Default bundle image tag
BUNDLE_IMG ?= cockroach-operator:$(VERSION)

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t quay.io/alinalion/$(BUNDLE_IMG) .
	docker push quay.io/alinalion/$(BUNDLE_IMG)

# Generate package manifests.
# Options for "packagemanifests".
CHANNEL?=beta
FROM_VERSION?=0.0.26
IS_CHANNEL_DEFAULT?=0

ifneq ($(origin FROM_VERSION), undefined)
PKG_FROM_VERSION := --from-version=$(FROM_VERSION)
endif
ifneq ($(origin CHANNEL), undefined)
PKG_CHANNELS := --channel=$(CHANNEL)
endif
ifeq ($(IS_CHANNEL_DEFAULT), 1)
PKG_IS_DEFAULT_CHANNEL := --default-channel
endif
PKG_MAN_OPTS ?= "$(PKG_FROM_VERSION) $(PKG_CHANNELS) $(PKG_IS_DEFAULT_CHANNEL)"


# Build the packagemanifests
.PHONY: update-pkg
update-pkg:
	bazel run  //hack:update-pkg  -- $(VERSION) $(IMG) $(PKG_MAN_OPTS)

# Build the bundle image.
.PHONY: gen-csv
gen-csv: dev/generate
	bazel run  //hack:update-csv  -- $(VERSION) $(IMG) $(BUNDLE_METADATA_OPTS)


##@ OPM

OLM_REPO ?= quay.io/alinalion/cockroach-operator-manifest
OLM_BUNDLE_REPO ?= quay.io/alinalion/cockroach-operator-bundle
OLM_PACKAGE_NAME ?= cockroach-operator
TAG ?= $(VERSION)

# opm-bundle-all: # used to bundle all the versions available
# 	./scripts/opm_bundle_all.sh $(OLM_REPO) $(OLM_PACKAGE_NAME) $(VERSION)
# opm-bundle-last-beta: ## Bundle latest for beta
# 	# $(operator-sdk) bundle create -g --directory "./deploy/olm-catalog/redhat-marketplace-operator/manifests" -c stable,beta --default-channel stable --package $(OLM_PACKAGE_NAME)
# 	$(docker) build -f custom-bundle.Dockerfile -t "$(OLM_REPO):$(TAG)" --build-arg channels=beta .
# 	$(docker) tag "$(OLM_REPO):$(TAG)" "$(OLM_REPO):$(VERSION)"
# 	$(docker) push "$(OLM_REPO):$(TAG)"
# 	$(docker) push "$(OLM_REPO):$(VERSION)"

# opm-bundle-last-stable: ## Bundle latest for stable
# 	$(operator-sdk) bundle create -g --directory "./deploy/olm-catalog/redhat-marketplace-operator/manifests" -c stable,beta --default-channel stable --package $(OLM_PACKAGE_NAME)
# 	$(docker) build -f custom-bundle.Dockerfile -t "$(OLM_REPO):$(TAG)" --build-arg channels=stable,beta .
# 	$(docker) tag "$(OLM_REPO):$(TAG)" "$(OLM_REPO):$(VERSION)"
# 	$(docker) push "$(OLM_REPO):$(TAG)"
# 	$(docker) push "$(OLM_REPO):$(VERSION)"

# opm-index-base: ## Create an index base
# 	git fetch --tags
# 	./scripts/opm_build_index.sh $(OLM_REPO) $(OLM_BUNDLE_REPO) $(TAG) $(VERSION)

# install-test-registry: ## Install the test registry
# 	kubectl apply -f ./deploy/olm-catalog/test-registry.yaml

		


		

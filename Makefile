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

# Make will use bash instead of sh
SHELL := /usr/bin/env bash

REGISTRY_PREFIX ?= us.gcr.io
GENERATOR_IMG ?= cockroach-operator/code-generator
TEST_RUNNER_IMG ?= cockroach-operator/test-runner
UBI_IMG ?= cockroach-operator/cockroach-operator-ubi
VERSION ?= latest
DATE_STAMP=$(shell date "+%Y%m%d-%H%M%S")
TEST_ARGS ?=

LOCAL_GOPATH := $(shell go env GOPATH)

TOOLS_WRAPPER := docker run --rm -v $(CURDIR):/ws -v $(LOCAL_GOPATH)/pkg:/go/pkg -v $(CURDIR)/.docker-build:/root/.cache/go-build

CONTROLLER_GEN = $(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(GENERATOR_IMG) controller-gen

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Run tests
native-test:
	go test -v ./api/... ./pkg/... -coverprofile cover.out

native-test-short:
	go test -v -short ./api/... ./pkg/... -coverprofile cover.out

test: generate manifests
	$(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) make native-test

test-short: generate manifests
	$(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) make native-test-short

e2e-test:
	$(TOOLS_WRAPPER) -v ${HOME}/.kube:/root/.kube -v ${HOME}/.config/gcloud:/root/.config/gcloud -e USE_EXISTING_CLUSTER=true $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) go test -v $(TEST_ARGS) ./e2e/... 2>&1 | tee  e2e-test-output.$(DATE_STAMP).log


e2e-test-gke:
	hack/create-gke-cluster.sh -c test
	$(TOOLS_WRAPPER) -v ${HOME}/.kube:/root/.kube -v ${HOME}/.config/gcloud:/root/.config/gcloud -e USE_EXISTING_CLUSTER=true $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) go test -v ./e2e/... > e2e-test-output.$(DATE_STAMP).log
	hack/delete-gke-cluster.sh -c test

run:
	go run ./cmd/cockroach-operator/main.go

fmt:
	go fmt ./...

vet:
	go vet ./...

mod/tidy:
	go mod tidy

# Generate manifests e.g. CRD, RBAC etc.
manifests:
	@$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	@hack/update-crd-manifest-headers.sh

# Generate code
generate:
	@$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate/boilerplate.go.txt paths=./api/...

docker/build/code-gen:
	@echo "===========> Building $(GENERATOR_IMG) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(GENERATOR_IMG):$(VERSION) -f Dockerfile.code-gen .

docker/build/test-runner:
	@echo "===========> Building $(TEST_RUNNER_IMG) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG):$(VERSION) -f Dockerfile.test-runner .

docker/build/operator-ubi:
	@echo "===========> Building $(UBI_IMG) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(UBI_IMG):$(VERSION) -f Dockerfile.ubi .

# Linting
# Removing a couple of items check_python check_docker check_base_files check_terraform
# This target does not run on teamcity yet because of python library issues.
lint: check_shell check_golang check_headers check_trailing_whitespace check_headers

# The .PHONY directive tells make that this isn't a real target and so
# the presence of a file named 'check_shell' won't cause this target to stop
# working
.PHONY: check_shell
check_shell:
	@source hack/make.sh && check_shell

.PHONY: check_python
check_python:
	@source hack/make.sh && check_python

.PHONY: check_golang
check_golang:
	@source hack/make.sh && golang

.PHONY: check_terraform
check_terraform:
	@source hack/make.sh && check_terraform

.PHONY: check_docker
check_docker:
	@source hack/make.sh && docker

.PHONY: check_base_files
check_base_files:
	@source hack/make.sh && basefiles

.PHONY: check_shebangs
check_shebangs:
	@source hack/make.sh && check_bash

.PHONY: check_trailing_whitespace
check_trailing_whitespace:
	@source hack/make.sh && check_trailing_whitespace

.PHONY: check_headers
check_headers:
	@echo "Checking file headers"
	@python3 hack/verify_boilerplate.py

.PHONY: goimports
goimports:
	@echo "Running goimports"
	@python3 hack/update_goimports.py

.PHONY: bazel/build
bazel/build:
	@bazel build //...

bazel/gazelle-mod:
	@bazel run //:gazelle -- update-repos -from_file=go.mod

bazel/gazelle-update:
	@bazel run //:gazelle -- update

bazel/test:
	@bazel test //api/... //pkg/...


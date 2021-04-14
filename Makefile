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
DOCKER_REGISTRY?=us.gcr.io/chris-love-operator-playground
DOCKER_IMAGE_REPOSITORY?=cockroachdb-operator
VERSION?=1.9.62
APP_VERSION?=v$(VERSION)
GCP_PROJECT?=chris-love-operator-playground
GCP_ZONE?=us-central1-a
CLUSTER_NAME?=bazel-test
DEV_REGISTRY?=gcr.io/$(GCP_PROJECT)
COCKROACH_DATABASE_VERSION=v20.2.5

#
# Unit Testing Targets
#
.PHONY: test/all
test/all:
	bazel test //apis/... //pkg/... //hack/... --test_arg=--test.v

.PHONY: test/apis
test/apis:
	bazel test //apis/...

.PHONY: test/pkg
test/pkg:
	bazel test //pkg/...

# This runs the all of the verify scripts and
# takes a bit of time.
.PHONY: test/verify
test/verify:
	bazel test //hack/...

# Run only e2e stort tests
# We can use this to only run one specific test
.PHONY: test/e2e-short
test/e2e-short:
	bazel test //e2e/... --test_arg=--test.short

#
# End to end testing targets
#
# kind: use make test/e2e/kind
# gke: use make test/e2e/gke
#
# kubetest2 binaries from the kubernetes testing team is used
# by the e2e tests.  We maintain the binaries and the binaries are
# downloaded from google storage by bazel.  See hack/bin/deps.bzl
# Once the repo releases binaries we should vendor the tag or
# download the built binaries.

# This target is used by kubetest2-tester-exec when running a kind test
# This target exportis the kubeconfig from kind and then runs
# k8s:k8s -type kind which checks to see if kind is up and running.
# Then bazel e2e testing is run.
.PHONY: test/e2e/testrunner-kind
test/e2e/testrunner-kind:
	bazel-bin/hack/bin/kind export kubeconfig --name $(CLUSTER_NAME)
	bazel run //hack/k8s:k8s -- -type kind
	bazel test --stamp //e2e/...

# Use this target to run e2e tests using a kind k8s cluster.
# This target uses kind to start a k8s cluster  and runs the e2e tests
# against that cluster.
# This is the main entrypoint for running the e2e tests on kind.
# This target runs kubetest2 kind that starts a kind cluster
# Then kubetest2 tester exec is run which runs the make target
# test/e2e/testrunner-kind.
# After the tests run the cluster is deleted.
# If you need a unique cluster name override CLUSTER_NAME.
.PHONY: test/e2e/kind
test/e2e/kind:
	bazel build //hack/bin/...
	PATH=${PATH}:bazel-bin/hack/bin kubetest2 kind --cluster-name=$(CLUSTER_NAME) \
		--up --down -v 10 --test=exec -- make test/e2e/testrunner-kind

# This target is used by kubetest2-tester-exec when running a gke test
# k8s:k8s -type gke which checks to see if gke is up and running.
# Then bazel e2e testing is run.
# This target also installs the operator in the default namespace
# you may need to overrirde the DOCKER_IMAGE_REPOSITORY to match
# the GKEs project repo.
.PHONY: test/e2e/testrunner-gke
test/e2e/testrunner-gke:
	bazel run //hack/k8s:k8s -- -type gke
	K8S_CLUSTER=gke_$(GCP_PROJECT)_$(GCP_ZONE)_$(CLUSTER_NAME) \
	DEV_REGISTRY=$(DEV_REGISTRY) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//manifests:install_operator.apply
	bazel test --stamp --test_arg=--pvc=true //e2e/...

# Use this target to run e2e tests with a gke cluster.
# This target uses kind to start a gke k8s cluster  and runs the e2e tests
# against that cluster.
# This is the main entrypoint for running the e2e tests on gke kind.
# This target runs kubetest2 gke that starts a gke cluster
# Then kubetest2 tester exec is run which runs the make target
# test/e2e/testrunner-gke.
# After the tests run the cluster is deleted.
# If you need a unique cluster name override CLUSTER_NAME.
# You will probably want to override GCP_ZONE and GCP_PROJECT as well.
# The gcloud binary is used to start the cluster and is not installed by bazel.
# You also need gcp permission to start a cluster and upload containers to the
# projects registry.
.PHONY: test/e2e/gke
test/e2e/gke:
	bazel build //hack/bin/...
	PATH=${PATH}:bazel-bin/hack/bin bazel-bin/hack/bin/kubetest2 gke --cluster-name=$(CLUSTER_NAME) \
		--zone=$(GCP_ZONE) --project=$(GCP_PROJECT) \
		--version latest --up --down -v 10 --ignore-gcp-ssh-key \
		--test=exec -- make test/e2e/testrunner-gke

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

#
# Targets that allow to install the operator on an existing cluster
#
.PHONY: k8s/apply
k8s/apply:
	K8S_CLUSTER=gke_$(GCP_PROJECT)_$(GCP_ZONE)_$(CLUSTER_NAME) \
	DEV_REGISTRY=$(DEV_REGISTRY) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//manifests:install_operator.apply

.PHONY: k8s/delete
k8s/delete:
	K8S_CLUSTER=gke_$(GCP_PROJECT)_$(GCP_ZONE)_$(CLUSTER_NAME) \
	DEV_REGISTRY=$(DEV_REGISTRY) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//manifests:install_operator.delete

#
# Dev target that updates bazel files and dependecies
#
.PHONY: dev/syncdeps
dev/syncdeps:
	bazel run //hack:update-deps \
	bazel run //hack:update-bazel \
	bazel run //:gazelle -- update-repos -from_file=go.mod

#
# Release targets
#

.PHONY: release/versionbump
release/versionbump:
	$(MAKE) CHANNEL=beta IS_DEFAULT_CHANNEL=0 release/update-pkg-manifest && \
	sed -i -e 's,\(image: cockroachdb/cockroach-operator:\).*,\1$(APP_VERSION),' manifests/operator.yaml && \
	git add . && \
	git commit -m "Bump version to $(VERSION)"


.PHONY: release/image
release/image:
	DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
	DOCKER_IMAGE_REPOSITORY=$(DOCKER_IMAGE_REPOSITORY) \
	APP_VERSION=$(APP_VERSION) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//:push_operator_image

#
# RedHat OpenShift targets
#

#RED HAT IMAGE BUNDLE
RH_BUNDLE_REGISTRY?=registry.connect.redhat.com/cockroachdb
RH_BUNDLE_IMAGE_REPOSITORY?=cockroachdb-operator-bundle
RH_BUNDLE_VERSION?=$(VERSION)
RH_DEPLOY_PATH="deploy/certified-metadata-bundle"
RH_DEPLOY_FULL_PATH="$(RH_DEPLOY_PATH)/cockroach-operator/"
RH_COCKROACH_DATABASE_IMAGE=registry.connect.redhat.com/cockroachdb/cockroach:$(COCKROACH_DATABASE_VERSION)
RH_OPERATOR_IMAGE?=registry.connect.redhat.com/cockroachdb/cockroachdb-operator:$(APP_VERSION)

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

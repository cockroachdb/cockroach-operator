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

# load env vars from .envrc if it exists
ifneq ("$(wildcard .envrc)","")
include .envrc
endif

SHELL:=/usr/bin/env bash -O globstar

# values used in workspace-status.sh
DOCKER_REGISTRY?=cockroachdb
DOCKER_IMAGE_REPOSITORY?=cockroachdb-operator
VERSION?=$(shell cat version.txt)
APP_VERSION?=v$(VERSION)
GCP_PROJECT?=chris-love-operator-playground
GCP_ZONE?=us-central1-a
CLUSTER_NAME?=bazel-test
DEV_REGISTRY?=gcr.io/$(GCP_PROJECT)
COCKROACH_DATABASE_VERSION=v20.2.5

# used for running e2e tests with OpenShift
PULL_SECRET?=
GCP_REGION?=
BASE_DOMAIN?=
EXTRA_KUBETEST2_PARAMS?=

#
# Unit Testing Targets
#
.PHONY: test/all
test/all: | test/apis test/pkg test/verify

.PHONY: test/apis
test/apis:
	bazel test //apis/... --test_arg=-test.v

.PHONY: test/pkg
test/pkg:
	bazel test //pkg/... --test_arg=-test.v

# This runs the all of the verify scripts and
# takes a bit of time.
.PHONY: test/verify
test/verify:
	bazel test //hack/... --test_arg=-test.v

.PHONY: test/lint
test/lint:
	bazel run //hack:verify-gofmt

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
# An example of calling this is using make test/e2e/testrunner-kind-upgrades
test/e2e/testrunner-kind-%: PACKAGE=$*
test/e2e/testrunner-kind-%:
	bazel-bin/hack/bin/kind export kubeconfig --name $(CLUSTER_NAME)
	bazel run //hack/k8s:k8s -- -type kind
	bazel test --stamp //e2e/$(PACKAGE)/... --test_arg=-test.v --test_arg=-test.parallel=8 --test_arg=parallel=true

# Use this target to run e2e tests using a kind k8s cluster.
# This target uses kind to start a k8s cluster  and runs the e2e tests
# against that cluster.
# This is the main entrypoint for running the e2e tests on kind.
# This target runs kubetest2 kind that starts a kind cluster
# Then kubetest2 tester exec is run which runs the make target
# test/e2e/testrunner-kind.
# After the tests run the cluster is deleted.
# If you need a unique cluster name override CLUSTER_NAME.
test/e2e/kind-%: PACKAGE=$*
test/e2e/kind-%:
	bazel build //hack/bin/...
	PATH=${PATH}:bazel-bin/hack/bin kubetest2 kind --cluster-name=$(CLUSTER_NAME) \
		--up --down -v 10 --test=exec -- make test/e2e/testrunner-kind-$(PACKAGE)
	
# This target is used by kubetest2-eks to run e2e tests.
.PHONY: test/e2e/testrunner-eks
test/e2e/testrunner-eks:
	KUBECONFIG=$(TMPDIR)/$(CLUSTER_NAME)-eks.kubeconfig.yaml bazel-bin/hack/bin/kubectl create -f hack/eks-storageclass.yaml
	bazel test --stamp //e2e/upgrades/...  --action_env=KUBECONFIG=$(TMPDIR)/$(CLUSTER_NAME)-eks.kubeconfig.yaml
	bazel test --stamp //e2e/create/...  --action_env=KUBECONFIG=$(TMPDIR)/$(CLUSTER_NAME)-eks.kubeconfig.yaml
	bazel test --stamp //e2e/decomission/...  --action_env=KUBECONFIG=$(TMPDIR)/$(CLUSTER_NAME)-eks.kubeconfig.yaml

# Use this target to run e2e tests with a eks cluster.
# This target uses kind to start a eks k8s cluster  and runs the e2e tests
# against that cluster.
.PHONY: test/e2e/eks
test/e2e/eks:
	bazel build //hack/bin/... //e2e/kubetest2-eks/...
	PATH=${PATH}:bazel-bin/hack/bin:bazel-bin/e2e/kubetest2-eks/kubetest2-eks_/ \
	bazel-bin/hack/bin/kubetest2 eks --cluster-name=$(CLUSTER_NAME)  --up --down -v 10 \
		--test=exec -- make test/e2e/testrunner-eks

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
	# TODO this is not working because we create the cluster role binding now
	# for openshift.  We need to move this to a different target
	#bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
	#	//manifests:install_operator.apply
	bazel test --stamp //e2e/upgrades/...
	bazel test --stamp //e2e/create/...
	bazel test --stamp --test_arg=--pvc=true //e2e/pvcresize/...
	bazel test --stamp //e2e/decomission/...

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

.PHONY: test/e2e/testrunner-openshift
test/e2e/testrunner-openshift:
	bazel test --stamp //e2e/upgrades/...  --action_env=KUBECONFIG=$(HOME)/openshift-$(CLUSTER_NAME)/auth/kubeconfig
	bazel test --stamp //e2e/create/...  --action_env=KUBECONFIG=$(HOME)/openshift-$(CLUSTER_NAME)/auth/kubeconfig
	bazel test --stamp //e2e/decomission/...  --action_env=KUBECONFIG=$(HOME)/openshift-$(CLUSTER_NAME)/auth/kubeconfig

# Use this target to run e2e tests with a openshift cluster.
# This target uses kind to start a openshift cluster and runs the e2e tests
# against that cluster. A full TLD is required to creat an openshift clutser.
# This target runs kubetest2 openshift that starts a openshift cluster
# Then kubetest2 tester exec is run which runs the make target
# test/e2e/testrunner-openshift.  After the tests run the cluster is deleted.
# See the instructions in the kubetes2-openshift on running the
# provider.
.PHONY: test/e2e/openshift
test/e2e/openshift:
	bazel build //hack/bin/... //e2e/kubetest2-openshift/...
	PATH=${PATH}:bazel-bin/hack/bin:bazel-bin/e2e/kubetest2-openshift/kubetest2-openshift_/ \
	     bazel-bin/hack/bin/kubetest2 openshift --cluster-name=$(CLUSTER_NAME) \
	     --gcp-project-id=$(GCP_PROJECT) \
	     --gcp-region=$(GCP_REGION) \
	     --base-domain=$(BASE_DOMAIN) \
	     --pull-secret-file=$(PULL_SECRET) \
	     $(EXTRA_KUBETEST2_PARAMS) \
	     --up --down --test=exec -- make test/e2e/testrunner-openshift

# This testrunner launchs the openshift packaging e2e test
# and requires an existing openshift cluster and the kubeconfig
# located in the usual place.
.PHONY: test/e2e/testrunner-openshift-packaging
test/e2e/testrunner-openshift-packaging: test/openshift-package
	bazel build //hack/bin:oc
	bazel test --stamp //e2e/openshift/... --cache_test_results=no \
		--action_env=KUBECONFIG=$(HOME)/openshift-$(CLUSTER_NAME)/auth/kubeconfig \
		--action_env=APP_VERSION=$(APP_VERSION) \
		--action_env=DOCKER_REGISTRY=$(DOCKER_REGISTRY)

#
# Different dev targets
#
.PHONY: dev/build
dev/build: dev/syncdeps
	bazel build //...

.PHONY: dev/fmt
dev/fmt:
	@echo +++ Running gofmt
	@bazel run //hack/bin:gofmt -- -s -w $(shell pwd)

.PHONY: dev/generate
dev/generate: | dev/update-codegen dev/update-crds

.PHONY: dev/update-codegen
dev/update-codegen:
	@bazel run //hack:update-codegen

# TODO: Be sure to update hack/verify-crds.sh if/when this changes
.PHONY: dev/update-crds
dev/update-crds:
	@bazel run //hack/bin:controller-gen \
		crd:trivialVersions=true \
		rbac:roleName=cockroach-operator-role \
		webhook \
		paths=./... \
		output:crd:artifacts:config=config/crd/bases
	@hack/boilerplaterize hack/boilerplate/boilerplate.yaml.txt config/**/*.yaml

.PHONY: dev/syncbazel
dev/syncbazel:
	@bazel run //:gazelle -- fix -external=external -go_naming_convention go_default_library
	@bazel run //hack/bin:kazel -- --cfg-path hack/build/.kazelcfg.json

.PHONY: dev/syncdeps
dev/syncdeps:
	@bazel run //:gazelle -- update-repos \
		-from_file=go.mod \
		-to_macro=hack/build/repos.bzl%_go_dependencies \
		-build_file_generation=on \
		-build_file_proto_mode=disable \
		-prune
	@make dev/syncbazel

.PHONY: dev/up
dev/up:
	@hack/dev.sh up

.PHONY: dev/down
dev/down:
	@hack/dev.sh down
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
# Release targets
#

# This target reads the current version from version.txt, increments the patch
# part of the version, saves the result in the same file, and calls make with
# the next release-specific target in a separate shell in order to reread the
# new version.
.PHONY: release/versionbump
release/versionbump:
	bazel run //hack/versionbump:versionbump -- patch $(VERSION) > $(PWD)/version.txt
	$(MAKE) release/gen-files

# Generate various config files, which usually contain the current operator
# version, latest CRDB version, a list of supported CRDB versions, etc.
.PHONY: release/gen-templates
release/gen-templates:
	bazel run //hack/crdbversions:crdbversions -- -operator-version $(APP_VERSION) -crdb-versions $(PWD)/crdb-versions.yaml -repo-root $(PWD)

# Generate various manifest files for OpenShift. We run this target after the
# operator version is changed. The results are committed to Git.
.PHONY: release/gen-files
release/gen-files: release/gen-templates
	$(MAKE) release/update-pkg-manifest && \
	$(MAKE) release/opm-build-bundle && \
	git add . && \
	git commit -m "Bump version to $(VERSION)"


.PHONY: release/image
release/image:
	# TODO this bazel clean is here because we need to pull the latest image from redhat registry every time
	# but this removes all caching and makes compile time for developers LONG.
	bazel clean --expunge
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
	RH_BUNDLE_IMAGE_TAG=$(APP_VERSION) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//:push_operator_bundle_image

# This target:
# 1. Updates CRD and CSV files
# 2. Pushes the operator image to a registry
# 3. Builds the OpenShift bundle
# 4. Pushes the OpenShift bundle to a registry
# 5. Run opm to create and push the OpenShift index container to a registry
# 6. Removes the newly created OpenShift files so that it can run again.
#
# The following env variables are used for the above process.
#
# APP_VERSION
# VERSION
# RH_BUNDLE_VERSION
# RH_OPERATOR_IMAGE
# DOCKER_REGISTRY
#
# See hack/openshift-test-packaging.sh for more information on running this target.
.PHONY: test/openshift-package
test/openshift-package: release/update-pkg-manifest release/image release/opm-build-bundle test/push-openshift-images
	VERSION=$(VERSION) \
	hack/cleanup-packaging.sh

# This target pushes the OpenShift bundle, then uses opm to push the index bundle.
.PHONY: test/push-openshift-images
test/push-openshift-images:
	APP_VERSION=$(APP_VERSION) \
	DOCKER_REGISTRY=$(DOCKER_REGISTRY) \
	bazel run --stamp --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 \
		//hack:push-openshift-images

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

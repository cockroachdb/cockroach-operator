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
APP_VERSION?=v1.0.0-rc.0

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

# This target uses kind to start a cluster and runs the e2e tests
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

VERSION ?= 1.0.0
VERSION_CSV_FILE := ./deploy/olm-catalog/cockroach-operator/$(VERSION)/cockroach-operator.v$(VERSION).clusterserviceversion.yaml
CSV_CHANNEL ?= beta # change to stable for release
CSV_DEFAULT_CHANNEL ?= "false" # change to true for release
CHANNELS ?= beta
INTERNAL_CRDS='["crdbclusters.crdb.cockroachlabs.com"]'
CREATED_TIME ?= $(shell date +"%FT%H:%M:%SZ")
MANIFEST_CSV_FILE := ./deploy/olm-catalog/cockroach-operator/manifests/cockroach-operator.clusterserviceversion.yaml
OPERATOR_IMAGE ?= registry.connect.redhat.com/cockroachdb/cockroachdb-operator:v1.0.0-rc.2

#VERSION ?= $(shell go run scripts/version/main.go)
# FROM_VERSION ?= $(shell go run scripts/version/main.go last)

#
# Generate CSV 
# TODO move this to bazel
#
generate-csv: dev/generate
	cp config/crd/bases/crdb.cockroachlabs.com_crdbclusters.yaml deploy/crds/crdb.cockroachlabs.com_crdbclusters.yaml
	cp config/rbac/role.yaml deploy/role.yaml
	operator-sdk generate csv \
		--csv-version=$(VERSION) \
		--csv-channel=$(CSV_CHANNEL) \
		--default-channel=$(CSV_DEFAULT_CHANNEL) \
		--operator-name=cockroach-operator \
		--make-manifests=false \
		--update-crds \
		--apis-dir=api
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.containerImage' $(OPERATOR_IMAGE)
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.createdAt' $(CREATED_TIME)
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.capabilities' "Full Lifecycle"
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.categories' "Database"
	yq w -i $(VERSION_CSV_FILE) --tag '!!str' 'metadata.annotations.certified' true
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.description' "CockroachDB Operator"
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.repository' "https://github.com/cockroachdb/cockroach-operator"
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.support' "Cockroach Labs"
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations."operators.operatorframework.io/internal-objects"' $(INTERNAL_CRDS)
	yq w -i $(VERSION_CSV_FILE) 'metadata.annotations.alm-examples' '[{"apiVersion": "crdb.cockroachlabs.com/v1alpha1", "kind": "CrdbCluster", "metadata": {"name": "crdb-tls-enabled"}, "spec": {"dataStore": {"emptyDir": {}}, "tlsEnabled": true, "nodes": 3}}]'
	yq w -i $(VERSION_CSV_FILE) 'spec.maintainers[0].email' "support@cockroachlabs.com"
	yq w -i $(VERSION_CSV_FILE) 'spec.maintainers[0].name' "Cockroach Labs Support"
	yq d -i $(VERSION_CSV_FILE) 'spec.install.spec.deployments[*].spec.template.spec.containers[*].env(name==WATCH_NAMESPACE).valueFrom'
	yq w -i $(VERSION_CSV_FILE) 'spec.install.spec.deployments[*].spec.template.spec.containers[*].env(name==WATCH_NAMESPACE).value' ''
	hack/bundle-csv.sh $(VERSION) $(OPERATOR_IMAGE)

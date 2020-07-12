REGISTRY_PREFIX ?= us.gcr.io
GENERATOR_IMG ?= cockroach-operator/code-generator
TEST_RUNNER_IMG ?= cockroach-operator/test-runner
UBI_IMG ?= cockroach-operator/cockroach-operator-ubi
VERSION ?= latest

LOCAL_GOPATH := $(shell go env GOPATH)

TOOLS_WRAPPER := docker run --rm -v $(CURDIR):/ws -v $(LOCAL_GOPATH)/pkg:/go/pkg -v $(CURDIR)/.docker-build:/root/.cache/go-build

CONTROLLER_GEN = $(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(GENERATOR_IMG) controller-gen

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Run tests
native-test: fmt vet
	go test ./api/... ./pkg/... -coverprofile cover.out

native-test-short: fmt vet
	go test -short ./api/... ./pkg/... -coverprofile cover.out

test: generate manifests
	$(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) make native-test

test-short: generate manifests
	$(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) make native-test-short

e2e-test:
	$(TOOLS_WRAPPER) -v ${HOME}/.kube:/root/.kube -v ${HOME}/.config/gcloud:/root/.config/gcloud -e USE_EXISTING_CLUSTER=true $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) go test -v ./e2e/...

run: fmt vet
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
lint: check_shell check_python check_golang check_terraform check_docker \
	check_base_files check_headers check_trailing_whitespace

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

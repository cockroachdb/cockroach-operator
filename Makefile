REGISTRY_PREFIX ?= us.gcr.io
GENERATOR_IMG ?= cockroach-operator/code-generator
TEST_RUNNER_IMG ?= cockroach-operator/test-runner
UBI_IMG ?= cockroach-operator/cockroach-operator-ubi
VERSION ?= latest
DATE_STAMP=$(shell date "+%Y%m%d-%H%M%S")

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
	#$(TOOLS_WRAPPER) -v ${HOME}/.kube:/root/.kube -v ${HOME}/.config/gcloud:/root/.config/gcloud -e USE_EXISTING_CLUSTER=true $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) go test -v  -run TestCreatesSecureClusterWithGeneratedCert ./e2e/... 
	$(TOOLS_WRAPPER) -v ${HOME}/.kube:/root/.kube -v ${HOME}/.config/gcloud:/root/.config/gcloud -e USE_EXISTING_CLUSTER=true $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG) go test -v  ./e2e/... 2>&1 | tee  e2e-test-output.$(DATE_STAMP).log

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

# Generate code
generate:
	@$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

docker/build/code-gen:
	@echo "===========> Building $(GENERATOR_IMG) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(GENERATOR_IMG):$(VERSION) -f Dockerfile.code-gen .

docker/build/test-runner:
	@echo "===========> Building $(TEST_RUNNER_IMG) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(TEST_RUNNER_IMG):$(VERSION) -f Dockerfile.test-runner .

docker/build/operator-ubi:
	@echo "===========> Building $(UBI_IMG) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(UBI_IMG):$(VERSION) -f Dockerfile.ubi .

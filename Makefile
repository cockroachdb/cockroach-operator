REGISTRY_PREFIX ?= us.gcr.io
GENERATOR_IMAGE_NAME ?= crdb-opercode-generator
VERSION ?= latest

TOOLS_WRAPPER = docker run --rm -v $(CURDIR):/go/src/github.com/cockroachlabs/crdb-operator
CONTROLLER_GEN = $(TOOLS_WRAPPER) $(REGISTRY_PREFIX)/$(GENERATOR_IMAGE_NAME) controller-gen

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Run tests
test: generate fmt vet manifests
	go test ./... -coverprofile cover.out

run: fmt vet
	go run ./main.go

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
	@echo "===========> Building $(IMAGE_NAME) docker image"
	docker build --pull -t $(REGISTRY_PREFIX)/$(GENERATOR_IMAGE_NAME):$(VERSION) -f Docker.code-gen .


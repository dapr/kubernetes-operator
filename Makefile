
PROJECT_NAME ?= kubernetes-operator
PROJECT_VERSION ?= 0.0.8

CONTAINER_REGISTRY ?= docker.io
CONTAINER_REGISTRY_ORG ?= dapr
CONTAINER_IMAGE_VERSION ?= $(PROJECT_VERSION)
CONTAINER_IMAGE ?= $(CONTAINER_REGISTRY)/$(CONTAINER_REGISTRY_ORG)/$(PROJECT_NAME):$(CONTAINER_IMAGE_VERSION)


BUNDLE_NAME ?= dapr-kubernetes-operator
BUNDLE_VERSION ?= $(PROJECT_VERSION)
BUNDLE_CONTAINER_IMAGE ?= $(CONTAINER_REGISTRY)/$(CONTAINER_REGISTRY_ORG)/$(PROJECT_NAME)-bundle:$(BUNDLE_VERSION)

CATALOG_VERSION ?= latest
CATALOG_CONTAINER_IMAGE ?= $(CONTAINER_REGISTRY)/$(CONTAINER_REGISTRY_ORG)/$(PROJECT_NAME)-catalog:$(CATALOG_VERSION)

LINT_GOGC ?= 10
LINT_TIMEOUT ?= 10m

MKFILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
PROJECT_PATH := $(patsubst %/,%,$(dir $(MKFILE_PATH)))
LOCALBIN := $(PROJECT_PATH)/bin

HELM_CHART_REPO ?= https://dapr.github.io/helm-charts
HELM_CHART ?= dapr
HELM_CHART_VERSION ?= 1.13.3
HELM_CHART_URL ?= https://raw.githubusercontent.com/dapr/helm-charts/master/dapr-$(HELM_CHART_VERSION).tgz

OPENSHIFT_VERSIONS ?= v4.12

## Tool Versions
CODEGEN_VERSION ?= v0.30.0
KUSTOMIZE_VERSION ?= v5.4.2
CONTROLLER_TOOLS_VERSION ?= v0.15.0
KIND_VERSION ?= v0.22.0
LINTER_VERSION ?= v1.59.0
OPERATOR_SDK_VERSION ?= v1.34.2
OPM_VERSION ?= v1.43.1
GOVULNCHECK_VERSION ?= latest
KO_VERSION ?= latest

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize
LINTER ?= $(LOCALBIN)/golangci-lint
GOIMPORT ?= $(LOCALBIN)/goimports
YQ ?= $(LOCALBIN)/yq
KIND ?= $(LOCALBIN)/kind
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
OPM ?= $(LOCALBIN)/opm
GOVULNCHECK ?= $(LOCALBIN)/govulncheck
KO ?= $(LOCALBIN)/ko
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen


# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build


ifndef ignore-not-found
  ignore-not-found = false
endif

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-\/]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Dapr

.PHONY: update/dapr
update/dapr: ## Update the helm chart.
	$(PROJECT_PATH)/hack/scripts/update_helm_chart.sh $(PROJECT_PATH) $(HELM_CHART_URL)

##@ Development

.PHONY: manifests
manifests: codegen-tools-install ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(PROJECT_PATH)/hack/scripts/gen_crd.sh $(PROJECT_PATH)

.PHONY: generate
generate: codegen-tools-install ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(PROJECT_PATH)/hack/scripts/gen_res.sh $(PROJECT_PATH)
	$(PROJECT_PATH)/hack/scripts/gen_client.sh $(PROJECT_PATH)

.PHONY: fmt
fmt: goimport ## Run go fmt, gomiport against code.
	$(GOIMPORT) -l -w .
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet ## Run tests.
	DAPR_HELM_CHART_VERSION="$(HELM_CHART_VERSION)" go test -ldflags="$(GOLDFLAGS)" -v ./pkg/... ./internal/...

.PHONY: test/e2e/operator
test/e2e/operator: manifests generate fmt vet ## Run e2e operator tests.
	DAPR_HELM_CHART_VERSION="$(HELM_CHART_VERSION)" go test -ldflags="$(GOLDFLAGS)" -p 1 -v ./test/e2e/operator/...

.PHONY: test/e2e/olm
test/e2e/olm: ## Run e2e catalog tests.
	DAPR_HELM_CHART_VERSION="$(HELM_CHART_VERSION)" go test -ldflags="$(GOLDFLAGS)" -p 1 -v ./test/e2e/olm/...

.PHONY: test/e2e/app
test/e2e/app: ko ## Deploy test app.
	KO_DOCKER_REPO=kind.local $(LOCALBIN)/ko build -B ./test/e2e/support/dapr-test-app

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -ldflags="$(GOLDFLAGS)" -o bin/dapr-control-plane cmd/main.go

.PHONY: run
run: ## Run a controller from your host.
	go run -ldflags="$(GOLDFLAGS)" cmd/main.go run \
		--leader-election=false \
		--zap-devel \
		--health-probe-bind-address ":0" \
		--metrics-bind-address ":0"

.PHONY: run/local
run/local: install ## Install and Run a controller from your host.
	go run -ldflags="$(GOLDFLAGS)" cmd/main.go run \
		--leader-election=false \
		--zap-devel \
		--health-probe-bind-address ":0" \
		--metrics-bind-address ":0"


.PHONY: deps
deps:  ## Tidy up deps.
	go mod tidy


.PHONY: check
check: check/lint  check/vuln

.PHONY: check/lint
check/lint: golangci-lint
	@echo "run golangci-lint"
	@$(LINTER) run \
		--config .golangci.yml \
		--out-format tab \
		--exclude-dirs etc \
		--timeout $(LINT_TIMEOUT) \
		--verbose

.PHONY: check/lint/fix
check/lint/fix: golangci-lint
	@$(LINTER) run \
		--config .golangci.yml \
		--out-format tab \
		--exclude-dirs etc \
		--timeout $(LINT_TIMEOUT) \
		--fix

.PHONY: check/vuln
check/vuln: govulncheck
	@echo "run govulncheck"
	@$(GOVULNCHECK) ./...

.PHONY: docker/build
docker/build: test ## Build docker image with the manager.
	$(CONTAINER_TOOL) build -t $(CONTAINER_IMAGE) .

.PHONY: docker/push
docker/push: ## Push docker image with the manager.
	$(CONTAINER_TOOL) push $(CONTAINER_IMAGE)

.PHONY: docker/push/kind
docker/push/kind: docker/build ## Load docker image in kind.
	kind load docker-image $(CONTAINER_IMAGE)

##@ Deployment

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) apply -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(CONTAINER_IMAGE)
	$(KUSTOMIZE) build config/deploy/standalone | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/deploy/standalone | $(KUBECTL) delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy/kind
deploy/kind: manifests kustomize kind ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(CONTAINER_IMAGE)
	kind load docker-image $(CONTAINER_IMAGE)
	$(KUSTOMIZE) build config/deploy/standalone | kubectl apply -f -

.PHONY: deploy/e2e/controller
deploy/e2e/controller: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(CONTAINER_IMAGE)
	$(KUSTOMIZE) build config/deploy/e2e | kubectl apply -f -
	
	kubectl wait \
		--namespace=dapr-system \
		--for=condition=ready \
		pod \
		--selector=control-plane=dapr-control-plane \
		--timeout=90s

.PHONY: deploy/e2e/ingress
deploy/e2e/ingress: 
	$(PROJECT_PATH)/hack/scripts/deploy_ingress.sh

.PHONY: deploy/e2e/olm
deploy/e2e/olm: 
	$(PROJECT_PATH)/hack/scripts/deploy_olm.sh

##@ Bundles

.PHONY: bundle/info
bundle/info: ## Dump bundle info.
	@echo $(CONTAINER_IMAGE)
	@echo $(BUNDLE_CONTAINER_IMAGE)

.PHONY: bundle/generate_
bundle/generate: generate manifests kustomize operator-sdk yq ## Generate bundle.
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(CONTAINER_IMAGE)
	$(PROJECT_PATH)/hack/scripts/gen_bundle.sh \
		$(PROJECT_PATH) \
		$(BUNDLE_NAME) \
		$(BUNDLE_VERSION) \
		$(OPENSHIFT_VERSIONS)

.PHONY: bundle/build
bundle/build: ## Build bundle image.
	$(CONTAINER_TOOL) build \
		-t $(BUNDLE_CONTAINER_IMAGE) \
		-f $(PROJECT_PATH)/bundle/bundle.Dockerfile \
		$(PROJECT_PATH)/bundle

.PHONY: bundle/push
bundle/push: ## Push bundle image.
	$(CONTAINER_TOOL) push $(BUNDLE_CONTAINER_IMAGE)

.PHONY: catalog/build
catalog/build: opm ## Build catalog image.
	$(OPM) index add \
		--container-tool $(CONTAINER_TOOL) \
		--mode semver \
		--tag $(CATALOG_CONTAINER_IMAGE) \
		--bundles $(BUNDLE_CONTAINER_IMAGE)

.PHONY: catalog/push
catalog/push: ## Push catalog image.
	$(CONTAINER_TOOL) push $(CATALOG_CONTAINER_IMAGE)

.PHONY: olm/install
olm/install: operator-sdk ## Install olm.
	cd bin && $(OPERATOR_SDK) olm install

##@ Build Dependencies

## Location to install dependencies to
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || \
	GOBIN=$(LOCALBIN) GO111MODULE=on go install sigs.k8s.io/kustomize/kustomize/v5@$(KUSTOMIZE_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)


.PHONY: golangci-lint
golangci-lint: $(LINTER)
$(LINTER): $(LOCALBIN)
	@test -s $(LOCALBIN)/golangci-lint || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(LINTER_VERSION)

.PHONY: goimport
goimport: $(GOIMPORT)
$(GOIMPORT): $(LOCALBIN)
	@test -s $(LOCALBIN)/goimport || \
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@latest

.PHONY: yq
yq: $(YQ)
$(YQ): $(LOCALBIN)
	@test -s $(LOCALBIN)/yq || \
	GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@latest

.PHONY: kind
kind: $(KIND)
$(KIND): $(LOCALBIN)
	@test -s $(LOCALBIN)/kind || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/kind@$(KIND_VERSION)

.PHONY: ko
ko: $(KO)
$(KO): $(LOCALBIN)
	@test -s $(LOCALBIN)/ko || \
	GOBIN=$(LOCALBIN) go install github.com/google/ko@$(KO_VERSION)

.PHONY: govulncheck
govulncheck: $(GOVULNCHECK)
$(GOVULNCHECK): $(LOCALBIN)
	@test -s $(GOVULNCHECK) || \
	GOBIN=$(LOCALBIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)

.PHONY: codegen-tools-install
codegen-tools-install: $(LOCALBIN)
	@echo "Installing code gen tools"
	$(PROJECT_PATH)/hack/scripts/install_gen_tools.sh $(PROJECT_PATH) $(CODEGEN_VERSION) $(CONTROLLER_TOOLS_VERSION)

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)
$(OPERATOR_SDK): $(LOCALBIN)
	@echo "Installing operator-sdk"
	$(PROJECT_PATH)/hack/scripts/install_operator_sdk.sh $(PROJECT_PATH) $(OPERATOR_SDK_VERSION)

.PHONY: opm
opm: $(OPM)
$(OPM): $(LOCALBIN)
	@echo "Installing opm"
	$(PROJECT_PATH)/hack/scripts/install_opm.sh $(PROJECT_PATH) $(OPM_VERSION)




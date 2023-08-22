#!/bin/sh

if [ $# -ne 3 ]; then
    echo "project root, codegen version and is expected"
fi

PROJECT_ROOT="$1"
CODEGEN_VERSION="$2"
CONTROLLER_TOOLS_VERSION="$3"

if [ ! -f "${PROJECT_ROOT}/bin/applyconfiguration-gen" ]; then
  GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/applyconfiguration-gen@"${CODEGEN_VERSION}"
fi

if [ ! -f "${PROJECT_ROOT}/bin/client-gen" ]; then
  GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/client-gen@"${CODEGEN_VERSION}"
fi

if [ ! -f "${PROJECT_ROOT}/bin/lister-gen" ]; then
  GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/lister-gen@"${CODEGEN_VERSION}"
fi

if [ ! -f "${PROJECT_ROOT}/bin/informer-gen" ]; then
  GOBIN="${PROJECT_ROOT}/bin" go install k8s.io/code-generator/cmd/informer-gen@"${CODEGEN_VERSION}"
fi

if [ ! -f "${PROJECT_ROOT}/bin/controller-gen" ]; then
  GOBIN="${PROJECT_ROOT}/bin" go install sigs.k8s.io/controller-tools/cmd/controller-gen@"${CONTROLLER_TOOLS_VERSION}"
fi
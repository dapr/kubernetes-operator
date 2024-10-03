#!/usr/bin/env bash

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"
TMP_DIR=$( mktemp -d -t dapr-client-gen-XXXXXXXX )

mkdir -p "${TMP_DIR}/client"
mkdir -p "${PROJECT_ROOT}/pkg/client"

echo "tmp dir: $TMP_DIR"


echo "Generating openapi schema"
go run k8s.io/kube-openapi/cmd/openapi-gen \
  --output-file zz_generated.openapi.go \
  --output-dir "pkg/generated/openapi" \
  --output-pkg "github.com/dapr/kubernetes-operator/pkg/generated/openapi" \
  github.com/dapr/kubernetes-operator/api/operator/v1alpha1 \
  k8s.io/apimachinery/pkg/apis/meta/v1 \
  k8s.io/apimachinery/pkg/runtime \
  k8s.io/apimachinery/pkg/version

echo "Generate ApplyConfiguration"
go run k8s.io/code-generator/cmd/applyconfiguration-gen \
  --openapi-schema <(go run ${PROJECT_ROOT}/cmd/main.go modelschema) \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${TMP_DIR}/client/applyconfiguration" \
  --output-pkg=github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration \
  github.com/dapr/kubernetes-operator/api/operator/v1alpha1

echo "Generate client"
go run k8s.io/code-generator/cmd/client-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${TMP_DIR}/client/clientset" \
  --input-base=github.com/dapr/kubernetes-operator/api \
  --input=operator/v1alpha1 \
  --fake-clientset=false \
  --clientset-name "versioned"  \
  --apply-configuration-package=github.com/dapr/kubernetes-operator/pkg/client/applyconfiguration \
  --output-pkg=github.com/dapr/kubernetes-operator/pkg/client/clientset

echo "Generate lister"
go run k8s.io/code-generator/cmd/lister-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${TMP_DIR}/client/listers" \
  --output-pkg=github.com/dapr/kubernetes-operator/pkg/client/listers \
  github.com/dapr/kubernetes-operator/api/operator/v1alpha1

echo "Generate informer"
go run k8s.io/code-generator/cmd/informer-gen \
  --go-header-file="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  --output-dir="${TMP_DIR}/client/informers" \
  --versioned-clientset-package=github.com/dapr/kubernetes-operator/pkg/client/clientset/versioned \
  --listers-package=github.com/dapr/kubernetes-operator/pkg/client/listers \
  --output-pkg=github.com/dapr/kubernetes-operator/pkg/client/informers \
  github.com/dapr/kubernetes-operator/api/operator/v1alpha1

# This should not be needed but for some reasons, the applyconfiguration-gen tool
# sets a wrong APIVersion for the Dapr type (operator/v1alpha1 instead of the one with
# the domain operator.dapr.io/v1alpha1).
#
# See: https://github.com/kubernetes/code-generator/issues/150
sed -i \
  's/WithAPIVersion(\"operator\/v1alpha1\")/WithAPIVersion(\"operator.dapr.io\/v1alpha1\")/g' \
  "${TMP_DIR}"/client/applyconfiguration/operator/v1alpha1/daprcontrolplane.go
sed -i \
  's/WithAPIVersion(\"operator\/v1alpha1\")/WithAPIVersion(\"operator.dapr.io\/v1alpha1\")/g' \
  "${TMP_DIR}"/client/applyconfiguration/operator/v1alpha1/daprinstance.go
sed -i \
  's/WithAPIVersion(\"operator\/v1alpha1\")/WithAPIVersion(\"operator.dapr.io\/v1alpha1\")/g' \
  "${TMP_DIR}"/client/applyconfiguration/operator/v1alpha1/daprcruisecontrol.go

cp -r \
  "${TMP_DIR}"/client/* \
  "${PROJECT_ROOT}"/pkg/client


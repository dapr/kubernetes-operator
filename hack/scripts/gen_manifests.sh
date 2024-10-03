#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"

go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  rbac:roleName=dapr-control-plane-role \
  crd \
  paths="{./api/operator/...,./internal/...}" \
  output:crd:artifacts:config="${PROJECT_ROOT}/config/crd/bases"

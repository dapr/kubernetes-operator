#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"

go run sigs.k8s.io/controller-tools/cmd/controller-gen \
  object:headerFile="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  paths="{./api/operator/...,./internal/...}"
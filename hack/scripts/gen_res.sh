#!/bin/sh

if [ $# -ne 1 ]; then
    echo "project root is expected"
fi

PROJECT_ROOT="$1"

"${PROJECT_ROOT}"/bin/controller-gen \
  object:headerFile="${PROJECT_ROOT}/hack/boilerplate.go.txt" \
  paths="./..."
#!/bin/sh

if [ $# -ne 2 ]; then
    echo "project root, opm version are expected"
fi

PROJECT_ROOT="$1"
OPM_VERSION="$2"

if [ ! -f "${PROJECT_ROOT}/bin/opm" ]; then
  OS=$(go env GOOS)
  ARCH=$(go env GOARCH)

  curl -sSLo \
    "${PROJECT_ROOT}/bin/opm" \
    "https://github.com/operator-framework/operator-registry/releases/download/${OPM_VERSION}/${OS}-${ARCH}-opm"

  chmod +x "${PROJECT_ROOT}/bin/opm"
fi

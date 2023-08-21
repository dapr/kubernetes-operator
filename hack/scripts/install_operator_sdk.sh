#!/bin/sh

if [ $# -ne 2 ]; then
    echo "project root, sdk version are expected"
fi

PROJECT_ROOT="$1"
OPERATOR_SDK_VERSION="$2"

if [ ! -f "${PROJECT_ROOT}/bin/operator-sdk" ]; then
  OS=$(go env GOOS)
  ARCH=$(go env GOARCH)

  curl -sSLo \
    "${PROJECT_ROOT}/bin/operator-sdk" \
    "https://github.com/operator-framework/operator-sdk/releases/download/${OPERATOR_SDK_VERSION}/operator-sdk_${OS}_${ARCH}"

  chmod +x "${PROJECT_ROOT}/bin/operator-sdk"
fi

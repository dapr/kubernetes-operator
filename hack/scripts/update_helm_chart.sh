#!/bin/sh

if [ $# -ne 2 ]; then
    echo "project root, helm chart url are required"
fi

PROJECT_ROOT="$1"
HELM_CHART_URL="$2"

rm -rf "${PROJECT_ROOT}/helm-charts/dapr"
mkdir -p "${PROJECT_ROOT}/helm-charts/dapr"

curl --location --silent "${HELM_CHART_URL}" \
      | tar xzf - \
          --directory "${PROJECT_ROOT}/helm-charts/dapr" \
          --strip-components=1


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

rm -rf "${PROJECT_ROOT}/config/crd/dapr"
cp -r "${PROJECT_ROOT}/helm-charts/dapr/crds"  "${PROJECT_ROOT}/config/crd/dapr"

cd "${PROJECT_ROOT}/config/crd/dapr" || exit

touch "kustomization.yaml"

for f in "${PROJECT_ROOT}"/helm-charts/dapr/crds/*.yaml; do
  kustomize edit add resource "$(basename ${f})"
done

# remove CRDs from the helm chart so they won't get installed
rm -rf "${PROJECT_ROOT}/helm-charts/dapr/crds"
#!/bin/sh

if [ $# -ne 3 ]; then
    echo "project root, bundle name, bundle version are expected"
fi

PROJECT_ROOT="$1"
BUNDLE_NAME="$2"
BUNDLE_VERSION="$3"

rm -rf "${PROJECT_ROOT}/bundle/dapr-operator-ng"

mkdir -p "${PROJECT_ROOT}/bundle"
cd "${PROJECT_ROOT}/bundle" || exit

echo "Project root   : ${PROJECT_ROOT}"
echo "Bundle Name    : ${BUNDLE_NAME}"
echo "Bundle Version : ${BUNDLE_VERSION}"

echo "Generate bundle"

${PROJECT_ROOT}/bin/kustomize build "${PROJECT_ROOT}/config/manifests" | ${PROJECT_ROOT}/bin/operator-sdk generate bundle \
  --use-image-digests \
  --overwrite \
  --package "${BUNDLE_NAME}" \
  --version "${BUNDLE_VERSION}" \
  --channels "alpha" \
  --default-channel "alpha" \
  --output-dir "${BUNDLE_NAME}"

echo "Patch bundle metadata"

${PROJECT_ROOT}/bin/yq -i \
  '.metadata.annotations.containerImage = .spec.install.spec.deployments[0].spec.template.spec.containers[0].image' \
   "${PROJECT_ROOT}/bundle/dapr-operator-ng/manifests/${BUNDLE_NAME}.clusterserviceversion.yaml"

echo "Validate bundle"

${PROJECT_ROOT}/bin/operator-sdk bundle validate "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}"
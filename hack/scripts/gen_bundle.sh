#!/bin/sh

if [ $# -ne 3 ]; then
    echo "project root, bundle name, bundle version, openshift versions are expected"
fi

export PROJECT_ROOT="$1"
export BUNDLE_NAME="$2"
export BUNDLE_VERSION="$3"
export OPENSHIFT_VERSIONS="$4"

rm -rf "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}"

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
  "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}/manifests/${BUNDLE_NAME}.clusterserviceversion.yaml"

${PROJECT_ROOT}/bin/yq -i \
  '.annotations."com.redhat.openshift.versions" = env(OPENSHIFT_VERSIONS)' \
  "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}/metadata/annotations.yaml"

echo "Validate bundle"

${PROJECT_ROOT}/bin/operator-sdk bundle validate "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}"

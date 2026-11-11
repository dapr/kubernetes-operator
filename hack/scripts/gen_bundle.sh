#!/bin/sh

if [ $# -ne 7 ]; then
    echo "project root, bundle name, bundle version, openshift versions, channels, default channel, replaces version are expected"
    exit 1
fi

export PROJECT_ROOT="$1"
export BUNDLE_NAME="$2"
export BUNDLE_VERSION="$3"
export OPENSHIFT_VERSIONS="$4"
export BUNDLE_CHANNELS="$5"
export BUNDLE_DEFAULT_CHANNEL="$6"
export BUNDLE_REPLACES_VERSION="$7"

rm -rf "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}"

mkdir -p "${PROJECT_ROOT}/bundle"
cd "${PROJECT_ROOT}/bundle" || exit

echo "Project root   : ${PROJECT_ROOT}"
echo "Bundle Name    : ${BUNDLE_NAME}"
echo "Bundle Version : ${BUNDLE_VERSION}"
echo "Channels       : ${BUNDLE_CHANNELS} (default: ${BUNDLE_DEFAULT_CHANNEL})"
echo "Replaces       : ${BUNDLE_REPLACES_VERSION:-none}"

echo "Generate bundle"

${KUSTOMIZE} build "${PROJECT_ROOT}/config/manifests" | ${PROJECT_ROOT}/bin/operator-sdk generate bundle \
  --use-image-digests \
  --overwrite \
  --package "${BUNDLE_NAME}" \
  --version "${BUNDLE_VERSION}" \
  --channels "${BUNDLE_CHANNELS}" \
  --default-channel "${BUNDLE_DEFAULT_CHANNEL}" \
  --output-dir "${BUNDLE_NAME}"

echo "Patch bundle metadata"

${YQ} -i \
  '.metadata.annotations.containerImage = .spec.install.spec.deployments[0].spec.template.spec.containers[0].image' \
  "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}/manifests/${BUNDLE_NAME}.clusterserviceversion.yaml"

${YQ} -i \
  '.annotations."com.redhat.openshift.versions" = env(OPENSHIFT_VERSIONS)' \
  "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}/metadata/annotations.yaml"

if [ -n "${BUNDLE_REPLACES_VERSION}" ]; then
  ${YQ} -i \
    '.spec.replaces = strenv(BUNDLE_NAME) + ".v" + strenv(BUNDLE_REPLACES_VERSION)' \
    "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}/manifests/${BUNDLE_NAME}.clusterserviceversion.yaml"
fi

echo "Validate bundle"

${PROJECT_ROOT}/bin/operator-sdk bundle validate "${PROJECT_ROOT}/bundle/${BUNDLE_NAME}"

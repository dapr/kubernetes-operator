#!/bin/sh +x

operator_dir="$1"
operator_bundle="$2"
operator_version="$3"
catalog_image="$4"

echo $operator_dir
echo $operator_bundle
echo $operator_version
echo $catalog_image

tmp_dir=$(mktemp -d -t ci-XXXXXXXXXX)
mkdir -p ${tmp_dir}/dapr

${operator_dir}/bin/opm generate dockerfile ${tmp_dir}/dapr

${operator_dir}/bin/opm init dapr-helm-operator \
    --default-channel=preview \
    --icon=${operator_dir}/hack/operator-icon.svg \
    --output yaml \
    > ${tmp_dir}/dapr/operator.yaml

${operator_dir}/bin/opm render ${operator_bundle} \
    --output=yaml \
    > ${tmp_dir}/dapr/operator.yaml

cat << EOF >> ${tmp_dir}/dapr/operator.yaml
---
schema: olm.channel
package: dapr-help-operator
name: preview
entries:
  - name: dapr-helm-operator.${operator_version}
EOF

#opm validate ${tmp_dir}

#docker build -f ${tmp_dir}/dapr.Dockerfile -t ${catalog_image} ${tmp_dir}
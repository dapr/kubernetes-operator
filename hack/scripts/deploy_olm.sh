#!/bin/sh

set -e

make olm/install

kubectl wait \
  --namespace=olm \
  --for=condition=ready \
  pod \
  --selector=app=olm-operator \
  --timeout=90s

kubectl wait \
  --namespace=olm \
  --for=condition=ready \
  pod \
  --selector=app=catalog-operator \
  --timeout=90s
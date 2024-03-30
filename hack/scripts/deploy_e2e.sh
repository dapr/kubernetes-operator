#!/bin/sh

set -e

make deploy/e2e

kubectl wait \
  --namespace=dapr-system \
  --for=condition=ready \
  pod \
  --selector=control-plane=dapr-control-plane \
  --timeout=90s
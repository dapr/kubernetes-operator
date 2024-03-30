#!/bin/sh

set -e

kubectl apply --server-side -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml

# it may take a while to have apply the
# resource, hence the kubectl wait may
# fail
sleep 5

kubectl wait \
  --namespace=ingress-nginx \
  --for=condition=ready \
  pod \
  --selector=app.kubernetes.io/component=controller \
  --timeout=90s
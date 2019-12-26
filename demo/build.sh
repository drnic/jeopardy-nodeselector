#!/bin/bash

# This script should not be required to be re-run unless demo/certificate.yaml
# is changed. It generates a certificate pair for an HTTPS server that might
# be run either on:
# * 127.0.0.1/localhost, or
# * as service jeopary-nodeselector in default namespace.
# This should allow people to quickly experiment with the webhook
# before ultimately deciding where it should be installed and with what
# certificates. It also allows for local integration testing.
#
# certificates.yaml uses cert-manager to generate the secret
# For installation see https://cert-manager.io/docs/installation/kubernetes/

set -eu

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

# from certificate.yaml
secret=jeopary-nodeselector-demo-certs

(
  echo "--> Clean up from previous build.sh"
  set -x; set +e
  kubectl delete -f certificate.yaml
  kubectl delete secrets $secret
  rm -f *.{crt,key}*
)

(
  echo "--> Create self-signed certificate with cert-manager"
  set -x
  kubectl apply -f certificate.yaml
  sleep 3
  kubectl get secret $secret -n default
)

(
  echo "--> Get certificate components from secret"
  set -x
  kubectl get secrets $secret -ojsonpath='{.data.ca\.crt}' | base64 --decode > \
    ca.crt
  kubectl get secrets $secret -ojsonpath='{.data.ca\.crt}' > \
    ca.crt.base64
  kubectl get secrets $secret -ojsonpath='{.data.tls\.crt}' | base64 --decode > \
    tls.crt
  kubectl get secrets $secret -ojsonpath='{.data.tls\.key}' | base64 --decode > \
    tls.key
)

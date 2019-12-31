#!/bin/bash

# This script should not be required to be re-run unless demo/certificate.yaml
# is changed. It generates a certificate pair for an HTTPS server that might
# be run either on:
# * 127.0.0.1/localhost, or
# * as service jeopardy-nodeselector in default namespace.
# This should allow people to quickly experiment with the webhook
# before ultimately deciding where it should be installed and with what
# certificates. It also allows for local integration testing.
#
# certificates.yaml uses cert-manager to generate the secret
# For installation see https://cert-manager.io/docs/installation/kubernetes/

set -eu

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

# from src/certificate.yaml
secret=jeopardy-nodeselector-demo-build

(
  echo "--> Clean up from previous build.sh"
  set -x; set +e
  kubectl delete -f src/certificate.yaml
  kubectl delete secrets $secret
  rm -f *.{crt,key}*
  rm -f *.yaml
  rm -f deploy/*.yaml
)

(
  echo "--> Create self-signed certificate with cert-manager"
  set -x
  kubectl apply -f src/certificate.yaml
  sleep 3
  kubectl get secret $secret -n default
)

(
  echo "--> Get certificate components from secret"
  set -x
  kubectl get secrets $secret -ojsonpath='{.data.ca\.crt}' | base64 --decode > \
    ca.crt
  kubectl get secrets $secret -ojsonpath='{.data.tls\.crt}' | base64 --decode > \
    tls.crt
  kubectl get secrets $secret -ojsonpath='{.data.tls\.key}' | base64 --decode > \
    tls.key
)

(
  echo "--> Create deployment files"
  echo "+ deploy/deployment.yaml"
  cp src/deployment.yaml deploy/

  echo "+ deploy/serviceaccount.yaml"
  cp src/serviceaccount.yaml deploy/

  echo "+ deploy/cert-secret.yaml"
  cat > deploy/cert-secret.yaml <<YAML
apiVersion: v1
type: kubernetes.io/tls
data:
  ca.crt:  $(base64 < ca.crt)
  tls.crt: $(base64 < tls.crt)
  tls.key: $(base64 < tls.key)
kind: Secret
metadata:
  name: jeopardy-nodeselector-demo-certs
  namespace: default
YAML

  echo "+ deploy/webhook-config.yaml"
  cat > deploy/webhook-config.yaml <<YAML
apiVersion: v1
kind: Namespace
metadata:
  # Create a namespace that we'll match on
  name: multiarch-test
  labels:
    multiarch: "true"
---
apiVersion: admissionregistration.k8s.io/v1beta1
kind: MutatingWebhookConfiguration
metadata:
  name: jeopardy-nodeselector
webhooks:
  - name: jeopardy-nodeselector.starkandwayne.com
    sideEffects: None
    # "Equivalent" provides insurance against API version upgrades/changes - e.g.
    # extensions/v1beta1 Ingress -> networking.k8s.io/v1beta1 Ingress
    # matchPolicy: Equivalent
    rules:
      - apiGroups:
          - "*"
        apiVersions:
          - "*"
        operations:
          - "CREATE"
          - "UPDATE"
        resources:
          - "pods"
          - "deployments"
    namespaceSelector:
      matchExpressions:
        # Any Namespace with a label matching the below will have its
        # annotations validated by this admission controller
        - key: "multiarch"
          operator: In
          values: ["true"]
    failurePolicy: Fail
    clientConfig:
      service:
        # This is the hostname our certificate needs in its Subject Alternative
        # Name array - name.namespace.svc
        # If the certificate does NOT have this name, TLS validation will fail.
        name: jeopardy-nodeselector
        namespace: default
        path: "/jeopardy-nodeselector/multiarch"
      caBundle: "$(base64 < ca.crt)"
YAML
)

(
  echo "--> Create all-in-one demo.yaml deployment file"
  rm -f demo.yaml
  for f in deploy/*.yaml
  do
    [[ -e "$f" ]] || break  # handle the case of no *.yaml files
    echo "---" >> demo.yaml
    cat $f >> demo.yaml
  done
)
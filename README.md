# Jeopardy Node Selector

```script
"I'll take 'Random Docker images off the Internet' for $200"
"binami/nginx"
"What is an image that only runs on amd64?"
"Correct!"

"I'll take 'Random Docker images off the Internet' for $1000"
"nginx"
"What is an image that only runs on multiple architectures?"
"Correct!"
```

Look up each OCI in pods/deployments/statefulsets/jobs and ensure the pods are restricted to nodes that support the image platform architecture.

Active in any namespace with label `jeopardy-nodeselector: enabled`.

## Quick Demo

To deploy the webhook server into the `default` namespace, including a self-signed certificate, and webhook registration:

```plain
kubectl apply -f demo/demo.yaml
```

This creates a namespace `multiarch-test`, and the demo webhook will only activate on resources created or updated within this namespace (technically, within any namespace labelled `jeopardy-nodeselector: enabled`).

The `demo/test-resources/` folder contains various pods, deployments, etc that deploy into the `multiarch-test` namespace:

```plain
$ kubectl apply -f demo/test-resources/
daemonset.apps/ds-whats-my-arch created
deployment.apps/multiarch-nginx created
deployment.apps/singlearch-nginx created
job.batch/job-whats-my-arch created
pod/multiarch-nginx created
pod/singlearch-nginx created
```

Check that it's `nodeSelector` has been assigned automatically to each resulting pod:

```plain
$ kubectl describe pod -n multiarch-test
...
Node-Selectors:  kubernetes.io/arch=amd64
Events:
  Type    Reason     Age        From                 Message
  ----    ------     ----       ----                 -------
  Normal  Scheduled  <unknown>  default-scheduler       Successfully assigned multiarch-test/nginx-amd64-684b5dd9bd-cv6qb to my-amd64-node
  Normal  Pulling    5s         kubelet, my-amd64-node  Pulling image "bitnami/nginx"
```

In addition to the pod having its `spec.nodeSelector` mutated, the original resource (deployment, job) should have also been mutated. This is not visible via `kubectl describe <resource>`. Instead we need to trawl through YAML or JSON:

```plain
$ kubectl get job -n multiarch-test -oyaml
...
spec:
  template:
    spec:
      nodeSelector:
        kubernetes.io/arch: arm
...
```

### Complex demo - Ghost/MariaDB

You can now deploy complex sets of things without worrying whether you need to determine nodeSelectors. For example, the Ghost helm chart uses images that only run on `amd64`. But you don't need to know this anymore:

```plain
helm install ghost stable/ghost \
    -n multiarch-test \
    --set "ghostHost=ghost.multiarch-test.svc.cluster.local"
kubectl get pods -n multiarch-test -l release=ghost -owide
NAME                     READY   STATUS    NODE
ghost-mariadb-0          1/1     Running   lattepanda
ghost-57f665d946-bh56s   1/1     Running   lattepanda
```

Both the `mariadb` and `ghost` pods are assigned to the `amd64` lattepanda node.

### Demo Cleanup

To clean up the webhook service, configuration, and the `multiarch-test` namespace:

```plain
kubectl delete -f demo/demo.yaml
```

## Helm Chart Installation

### Requirements

The Helm chart currently requires [Cert Manager](https://cert-manager.io) to create a self-signed certificate pair for the server, using a known self-signed CA (which is hard-coded into the webhook configuration as `caBundle`).

If you install the Helm chart and Cert Manager is not yet installed you'll get an error like:

```plain
Error: UPGRADE FAILED: unable to recognize "": no matches for kind "Issuer" in version "cert-manager.io/v1alpha2"
```

### Steps

```plain
kubectl create ns jeopardy-nodeselector
helm upgrade --install jeopardy-nodeselector . \
    -n jeopardy-nodeselector \
    --set "certificate.static=true"
```

### Uninstall

To deregister the webhook configuration, delete the running server, the `multiarch-test` namespace, and other configuration:

```plain
helm delete jeopardy-nodeselector -n jeopardy-nodeselector
kubectl delete ns jeopardy-nodeselector
```

## Local demo

In one terminal, run the webhook server with some pre-created self-signed certificates:

```plain
go run cmd/main.go
```

In another terminal, interact with the webhook server (as if you are a Kubernetes API Server requesting permission to mutate a resource):

```plain
$ curl https://localhost:8443 --cacert demo/ca.crt
Jeopardy Node Selector to ensure OCI runs on suppported nodes
Available routes:
/
/healthz
/jeopardy-nodeselector
/jeopardy-nodeselector/multiarch
```

## Build

To build the OCI (docker image) for multiple architectures:

```plain
make build push manifest
```

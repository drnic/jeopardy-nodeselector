# Jeopardy Node Selector

**STATUS:** Still in initial development. The demo below does install; though currently it hard-codes all `nodeSelector` to `kubernetes.io/arch=amd64`. Will finish it soon.

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

## Quick Demo

To deploy the webhook server into the `default` namespace, including a self-signed certificate, and webhook registration:

```plain
kubectl apply -f demo/demo.yaml
```

To test the webhook with an "amd64-only" image into the newly created `multiarch-test` namespace:

```plain
kubectl run --image bitnami/nginx nginx-amd64 -n multiarch-test
```

Check that it's `nodeSelector` has been assigned automatically:

```plain
$ kubectl describe pod -n multiarch-test
...
Node-Selectors:  kubernetes.io/arch=amd64
...
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
docker buildx build --progress=plain \
  --platform linux/amd64,linux/arm/v7,linux/arm64 \
  --push \
  --tag docker.io/drnic/jeopardy-nodeselector \
  .
```

For just a single architecture:

```plain
docker build -t drnic/jeopardy-nodeselector .
docker push drnic/jeopardy-nodeselector
```
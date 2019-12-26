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

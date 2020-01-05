* now active on namespaces labelled jeopardy-nodeselector=enabled
* fail fast if cert-manager.io not installed in cluster
    ```
    Error: UPGRADE FAILED: unable to recognize "": no matches for kind "Issuer" in version "cert-manager.io/v1alpha2"
    ```
    The webhook configuration requires `metadata.annotations.cert-manager\.io/inject-ca-from-secret` to setup CA
* README suggests installing Helm release into `jeopardy-nodeselector` namespace, so as to ensure its never labelled
  to self-mutate future releases

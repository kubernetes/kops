# Walkthrough of functionality


Quick start:

```bash
# Generate certs, kubeconfig, yaml files
# Check out the script to better understand how all the pieces fit together!
./scripts/create-kubeconfig.sh

# Start Server (using generated server certs)
go run ./cmd/discovery-server --tls-cert walkthrough/server.crt --tls-key walkthrough/server.key &

# Verify server is running and serving the DiscoveryEndpoint resource
kubectl --kubeconfig=walkthrough/universe1/client1.kubeconfig api-resources

# List DiscoveryEndpoints
kubectl --kubeconfig=walkthrough/universe1/client1.kubeconfig get discoveryendpoints --all-namespaces

# Register (Apply)
# The `metadata.name` MUST match the Common Name (CN) of your client certificate (e.g., `client1`), or the server will reject it with 403 Forbidden.
kubectl --kubeconfig=walkthrough/universe1/client1.kubeconfig apply -f walkthrough/universe1/client1-discoveryendpoint.yaml  --server-side=true --validate=false

# List DiscoveryEndpoints
kubectl --kubeconfig=walkthrough/universe1/client1.kubeconfig get discoveryendpoints --all-namespaces
```


## Using curl

The kubernetes API is a well-structured REST API, so we don't have to use kubectl.

If you want to test the API with curl, you must include the **Universe ID** in the URL path.

**Export the Universe ID:**
```bash
export UNIVERSE_ID=$(openssl x509 -in walkthrough/universe1/ca.crt -noout -fingerprint -sha256 | sed 's/SHA256 Fingerprint=//' | sed 's/://g' | tr '[:upper:]' '[:lower:]')
echo "UNIVERSE_ID is ${UNIVERSE_ID}"
```

```bash
curl --cert walkthrough/universe1/client1-bundle.crt --key walkthrough/universe1/client1.key --cacert walkthrough/server.crt \
  "https://localhost:8443/${UNIVERSE_ID}/apis/discovery.kops.k8s.io/v1alpha1/namespaces/default/discoveryendpoints"
```

## OIDC discovery

The goal here is to allow anonymous access to OIDC endpoints, so let's verify that:

**Getting the OpenID Configuration**
```bash
curl --cacert walkthrough/server.crt  "https://localhost:8443/${UNIVERSE_ID}/.well-known/openid-configuration"
```

**Getting the JWKS keys**
```bash
curl --cacert walkthrough/server.crt  "https://localhost:8443/${UNIVERSE_ID}/openid/v1/jwks"
```

Note that we do not need a client certificate to get this data.  Data from the DiscoveryEndpoints is published publicly.
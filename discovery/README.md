# Discovery Service

A public discovery service using mTLS for authentication and "Universe" isolation, emulating a Kubernetes API.

## Concept

- **Universe**: Defined by the SHA256 Fingerprint of a Custom CA Certificate.
- **Client**: Identified by a Client Certificate signed by that Custom CA.
- **DiscoveryEndpoint**: The resource type representing a registered client.
- **Isolation**: Clients can only see `DiscoveryEndpoint` objects signed by the same Custom CA.

## Usage

### Run Server

```bash
go run ./cmd/discovery-server --tls-cert server.crt --tls-key server.key --listen :8443
```

(You can generate a self-signed server certificate for testing, see the [walkthrough](docs/walkthrough.md) ).

### Client Requirement

Clients must authenticate using mTLS.
**Important**: The client MUST provide the full certificate chain, including the Root CA, because the server does not have pre-configured trust stores for these custom universes.
The server identifies the Universe from the SHA256 hash of the Root CA certificate found in the TLS chain.

### Quick start

See `docs/walkthrough.md` for detailed instructions.


## OIDC Discovery

The discovery server also serves OIDC discovery information publicly, allowing external systems (like AWS IAM) to discover the cluster's identity provider configuration.

- `GET /<universe-id>/.well-known/openid-configuration`: Returns the OIDC discovery document.
- `GET /<universe-id>/openid/v1/jwks`: Returns the JWKS.

This information is populated by clients uploading `DiscoveryEndpoint` resources containing the `oidc` spec.

## Building and Running
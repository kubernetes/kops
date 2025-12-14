# Discovery Service Project

## Overview
This project implements a public discovery service designed for decentralized, secure peer discovery. The core innovation is the use of **Custom Certificate Authorities (CAs)** to define isolated "Universes". Clients register and discover peers within their own Universe, identified and secured purely by mTLS.

The service emulates a Kubernetes API, allowing interaction via `kubectl`, including support for **Server-Side Apply**.

## Key Concepts

### 1. The "Universe"
- A **Universe** is an isolated scope for peer discovery.
- It is cryptographically defined by the **SHA256 hash of the Root CA's Public Key**.
- Any client possessing a valid certificate signed by a specific CA belongs to that CA's Universe.
- Different CAs = Different Universes. There is no crossover.

### 2. Authentication & Authorization
- **Mechanism**: Mutual TLS (mTLS).
- **Client Identity**: Derived from the **Common Name (CN)** of the leaf certificate.
- **Universe Context**: Derived from the **Root CA** presented in the TLS handshake.
- **Requirement**: Clients **MUST** present the full certificate chain (Leaf + Root CA) during the handshake. The server does not maintain a pre-configured trust store for these custom CAs; it uses the presented chain to determine the scope.

### 3. API Resources
- **DiscoveryEndpoint** (`discovery.kops.k8s.io/v1alpha1`): Represents a peer in the discovery network. Can optionally hold OIDC configuration (Issuer URL, JWKS).
- **Validation**: A client with CN `client1` can only Create/Update a `DiscoveryEndpoint` named `client1`.
- **Apply Support**: The server supports `PATCH` requests to facilitate `kubectl apply --server-side`.

### 4. OIDC Discovery
The server acts as an OIDC Discovery Provider for the Universe.
- **Public Endpoints**:
  - `/.well-known/openid-configuration`: Returns the OIDC discovery document.
  - `/openid/v1/jwks`: Returns the JSON Web Key Set (JWKS).
- **Data Source**: These endpoints serve data uploaded by clients via the `DiscoveryEndpoint` resource.

## Architecture

### Project Structure
- `cmd/discovery-server/`: Main entry point. Wires up the HTTP server with TLS configuration.
- `pkg/discovery/`:
  - `auth.go`: logic for inspecting TLS `PeerCertificates` to extract the Universe ID (CA hash) and Client ID.
  - `store.go`: In-memory thread-safe storage (`MemoryStore`) mapping Universe IDs to lists of `DiscoveryEndpoint` objects.
  - `server.go`: HTTP handlers implementing the K8s API emulation for `/apis/discovery.kops.k8s.io/v1alpha1`.
  - `k8s_types.go`: Definitions of `DiscoveryEndpoint`, `DiscoveryEndpointList`, `TypeMeta`, `ObjectMeta` etc.

### Data Model
- **DiscoveryEndpoint**: The core resource. Contains `Spec.Addresses` and metadata.
- **Universe**: Contains a map of `DiscoveryEndpoint` objects (keyed by name).
- **Unified Types**: The API type `DiscoveryEndpoint` is used directly for in-memory storage, ensuring zero conversion overhead.

## Security Model
- **Trust Delegation**: The server delegates trust to the CA. If you hold the CA key, you control the Universe.
- **Isolation**: The server ensures that a client presenting a cert chain for `CA_A` cannot read or write data to the Universe defined by `CA_B`.
- **Ephemeral**: The current implementation uses in-memory storage. Data is lost on restart.

## Building and Running

### Build
```bash
go build ./cmd/discovery-server
```

### Run

See docs/walkthrough.md for instructions on testing functionality.

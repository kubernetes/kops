#!/bin/bash

# Copyright 2025 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

CERTS_DIR="walkthrough"
SERVER_KEY="${CERTS_DIR}/server.key"
SERVER_CRT="${CERTS_DIR}/server.crt"
UNIVERSE_CA_KEY="${CERTS_DIR}/universe1/ca.key"
UNIVERSE_CA_CRT="${CERTS_DIR}/universe1/ca.crt"
CLIENT_KEY="${CERTS_DIR}/universe1/client1.key"
CLIENT_CRT="${CERTS_DIR}/universe1/client1.crt"
CLIENT_BUNDLE="${CERTS_DIR}/universe1/client1-bundle.crt"
CLIENT_KUBECONFIG="${CERTS_DIR}/universe1/client1.kubeconfig"

mkdir -p "$CERTS_DIR/universe1"

# 1. Generate Server Certificate with SANs (if missing)
if [ ! -f "$SERVER_CRT" ] || [ ! -f "$SERVER_KEY" ]; then
    echo "Generating Server Certificate with SANs..."

    # Create OpenSSL config for SANs
    cat > "${CERTS_DIR}/server.conf" <<EOF
[req]
distinguished_name = req_distinguished_name
x509_extensions = v3_req
prompt = no
[req_distinguished_name]
CN = localhost
[v3_req]
keyUsage = keyEncipherment, dataEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
IP.1 = 127.0.0.1
EOF

    openssl req -new -newkey rsa:2048 -days 365 -nodes -x509 \
        -keyout "$SERVER_KEY" \
        -out "$SERVER_CRT" \
        -config "${CERTS_DIR}/server.conf" \
        -extensions v3_req

    rm "${CERTS_DIR}/server.conf"
    echo "Created $SERVER_CRT and $SERVER_KEY"
fi

# 2. Generate Universe CA (if missing)
if [ ! -f "$UNIVERSE_CA_CRT" ]; then
    echo "Generating Universe CA..."
    openssl req -new -x509 -days 365 -nodes \
        -out "$UNIVERSE_CA_CRT" \
        -keyout "$UNIVERSE_CA_KEY" \
        -subj "/CN=MyPrivateUniverse"
    echo "Created $UNIVERSE_CA_CRT"
fi

# 3. Generate Client Certificate (if missing)
if [ ! -f "$CLIENT_BUNDLE" ]; then
    echo "Generating Client Certificate..."

    openssl req -new -nodes \
        -out "${CERTS_DIR}/client1.csr" \
        -keyout "$CLIENT_KEY" \
        -subj "/CN=client1"

    openssl x509 -req -in "${CERTS_DIR}/client1.csr" \
        -CA "$UNIVERSE_CA_CRT" \
        -CAkey "$UNIVERSE_CA_KEY" \
        -CAcreateserial \
        -out "$CLIENT_CRT" \
        -days 365

    # Bundle Leaf + CA
    cat "$CLIENT_CRT" "$UNIVERSE_CA_CRT" > "$CLIENT_BUNDLE"
    rm "${CERTS_DIR}/client1.csr"
    echo "Created $CLIENT_BUNDLE"
fi

# Resolve absolute paths for kubeconfig
# (Function to handle macOS/Linux differences for readlink)
get_abs_path() {
    echo "$(cd "$(dirname "$1")"; pwd)/$(basename "$1")"
}

ABS_SERVER_CRT=$(get_abs_path "$SERVER_CRT")
ABS_CLIENT_BUNDLE=$(get_abs_path "$CLIENT_BUNDLE")
ABS_CLIENT_KEY=$(get_abs_path "$CLIENT_KEY")

# 4. Generate Kubeconfig
UNIVERSE_ID=$(openssl x509 -in "$UNIVERSE_CA_CRT" -noout -fingerprint -sha256 | sed 's/SHA256 Fingerprint=//' | sed 's/://g' | tr '[:upper:]' '[:lower:]')
echo "Universe ID: ${UNIVERSE_ID}"

cat > "$CLIENT_KUBECONFIG" <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority: $ABS_SERVER_CRT
    server: https://localhost:8443/${UNIVERSE_ID}
  name: discovery-universe
contexts:
- context:
    cluster: discovery-universe
    user: client
  name: default
current-context: default
users:
- name: client
  user:
    client-certificate: $ABS_CLIENT_BUNDLE
    client-key: $ABS_CLIENT_KEY
EOF

# 5. Generate DiscoveryEndpoint resource
cat <<EOF > "${CERTS_DIR}/universe1/client1-discoveryendpoint.yaml"
apiVersion: discovery.kops.k8s.io/v1alpha1
kind: DiscoveryEndpoint
metadata:
  name: client1
spec:
  addresses:
  - 10.0.0.1
  oidc:
    issuerURL: https://issuer.example.com
    keys:
    - kty: RSA
      kid: example-key-id
      use: sig
      n: example-modulus
      e: AQAB
EOF

# 6. Output Instructions
echo "Generated $OUTPUT"
echo ""
echo "To use:"
echo "1. Run server: ./discovery-server --tls-cert $SERVER_CRT --tls-key $SERVER_KEY"
echo "2. Run kubectl: kubectl --kubeconfig=$OUTPUT get discoveryendpoints"

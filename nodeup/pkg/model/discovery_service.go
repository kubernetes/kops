/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package model

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// DiscoveryService registers with the discovery service.
type DiscoveryService struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &DiscoveryService{}

// Build is responsible for configuring the discovery service registration tasks.
func (b *DiscoveryService) Build(c *fi.NodeupModelBuilderContext) error {
	ctx := c.Context()

	if !b.IsMaster {
		return nil
	}
	discoveryServiceOptions := b.DiscoveryServiceOptions()
	if discoveryServiceOptions == nil {
		return nil
	}

	nodeName, err := b.NodeName()
	if err != nil {
		return err
	}

	// TODO: Currently we enforce that the node name must match the certificate CN, but ... it probably doesn't matter.
	// certificateName := nodeName + "." + b.ClusterName()
	certificateName := nodeName

	namespace := strings.ReplaceAll(b.ClusterName(), ".", "-")
	id := types.NamespacedName{
		Namespace: namespace,
		Name:      nodeName,
	}

	issueCert := &nodetasks.IssueCert{
		Name:      "discovery-service-client",
		Signer:    fi.DiscoveryCAID,
		KeypairID: b.NodeupConfig.KeypairIDs[fi.DiscoveryCAID],
		Type:      "client",
		Subject: nodetasks.PKIXName{
			CommonName: certificateName,
		},
		AlternateNames: []string{certificateName},
	}
	c.AddTask(issueCert)

	certResource, keyResource, caResource := issueCert.GetResources()

	jwks, err := findJWKSForServiceAccount(ctx, b.NodeupConfig.KeypairIDs, b.KeyStore)
	if err != nil {
		return err
	}
	registerTask := &nodetasks.DiscoveryServiceRegisterTask{
		Name:              "register",
		DiscoveryService:  discoveryServiceOptions.URL,
		ClientCert:        certResource,
		ClientKey:         keyResource,
		ClientCA:          caResource,
		RegisterName:      id.Name,
		RegisterNamespace: id.Namespace,
		JWKS:              jwks,
	}
	c.AddTask(registerTask)

	return nil
}

func findJWKSForServiceAccount(ctx context.Context, keypairIDs map[string]string, keystore fi.KeystoreReader) ([]nodetasks.JSONWebKey, error) {
	var jwks []nodetasks.JSONWebKey

	name := "service-account"
	keypairID := keypairIDs[name]
	if keypairID == "" {
		// kOps bug where KeypairID was not populated for the node role.
		return nil, fmt.Errorf("no keypair ID for %q", name)
	}

	keyset, err := keystore.FindKeyset(ctx, name)
	if err != nil {
		return nil, err
	}
	if keyset == nil {
		return nil, fmt.Errorf("keyset %q not found", name)
	}

	for _, item := range keyset.Items {
		if item.DistrustTimestamp != nil {
			continue
		}
		if item.Certificate == nil || item.Certificate.Subject.CommonName != "service-account" {
			continue
		}

		publicKey := item.Certificate.PublicKey

		jwk := nodetasks.JSONWebKey{}

		{
			jwk.KeyID = item.Id
			// publicKeyDERBytes, err := x509.MarshalPKIXPublicKey(publicKey)
			// if err != nil {
			// 	return nil, fmt.Errorf("failed to serialize public key to DER format: %v", err)
			// }

			// hasher := crypto.SHA256.New()
			// hasher.Write(publicKeyDERBytes)
			// publicKeyDERHash := hasher.Sum(nil)

			// jwk.KeyID = base64.RawURLEncoding.EncodeToString(publicKeyDERHash)
		}

		switch publicKey := publicKey.(type) {
		case *rsa.PublicKey:
			jwk.Algorithm = "RS256"
			jwk.Use = "sig"
			jwk.N = base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
			jwk.E = base64.RawURLEncoding.EncodeToString(uint64ToBytes(uint64(publicKey.E)))
			jwk.KeyType = "RSA"

		default:
			return nil, fmt.Errorf("unsupported public key type for service-account: %T", publicKey)
		}

		jwks = append(jwks, jwk)
	}
	sort.Slice(jwks, func(i, j int) bool {
		return jwks[i].KeyID < jwks[j].KeyID
	})

	return jwks, nil
}

func uint64ToBytes(n uint64) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint64(data, n)
	return bytes.TrimLeft(data, "\x00")
}

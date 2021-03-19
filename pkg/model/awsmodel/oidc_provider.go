/*
Copyright 2019 The Kubernetes Authors.

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

package awsmodel

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"

	"gopkg.in/square/go-jose.v2"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// OIDCProviderBuilder configures IAM OIDC Provider
type OIDCProviderBuilder struct {
	*model.KopsModelContext

	KeyStore  fi.CAStore
	Lifecycle *fi.Lifecycle
}

type oidcDiscovery struct {
	Issuer                string   `json:"issuer"`
	JWKSURI               string   `json:"jwks_uri"`
	AuthorizationEndpoint string   `json:"authorization_endpoint"`
	ResponseTypes         []string `json:"response_types_supported"`
	SubjectTypes          []string `json:"subject_types_supported"`
	SigningAlgs           []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported       []string `json:"claims_supported"`
}

var _ fi.ModelBuilder = &OIDCProviderBuilder{}

const (
	defaultAudience = "amazonaws.com"
)

func (b *OIDCProviderBuilder) Build(c *fi.ModelBuilderContext) error {

	if !featureflag.PublicJWKS.Enabled() {
		return nil
	}

	serviceAccountIssuer, err := iam.ServiceAccountIssuer(&b.Cluster.Spec)
	if err != nil {
		return err
	}

	signingKeyTaskObject, found := c.Tasks["Keypair/master"]
	if !found {
		return fmt.Errorf("keypair/master task not found")
	}

	fingerprints := getFingerprints()

	thumbprints := []*string{}

	for _, fingerprint := range fingerprints {
		thumbprints = append(thumbprints, fi.String(fingerprint))
	}

	skTask := signingKeyTaskObject.(*fitasks.Keypair)

	keys := &OIDCKeys{
		SigningKey: skTask,
	}

	discovery, err := buildDiscoveryJSON(serviceAccountIssuer)
	if err != nil {
		return err
	}
	keysFile := &fitasks.ManagedFile{
		Contents:  keys,
		Lifecycle: b.Lifecycle,
		Location:  fi.String("oidc/keys.json"),
		Name:      fi.String("keys.json"),
		Base:      fi.String(b.Cluster.Spec.PublicDataStore),
		Public:    fi.Bool(true),
	}
	c.AddTask(keysFile)

	discoveryFile := &fitasks.ManagedFile{
		Contents:  fi.NewBytesResource(discovery),
		Lifecycle: b.Lifecycle,
		Location:  fi.String("oidc/.well-known/openid-configuration"),
		Name:      fi.String("discovery.json"),
		Base:      fi.String(b.Cluster.Spec.PublicDataStore),
		Public:    fi.Bool(true),
	}
	c.AddTask(discoveryFile)

	c.AddTask(&awstasks.IAMOIDCProvider{
		Name:        fi.String(b.ClusterName()),
		Lifecycle:   b.Lifecycle,
		URL:         fi.String(serviceAccountIssuer),
		ClientIDs:   []*string{fi.String(defaultAudience)},
		Tags:        b.CloudTags(b.ClusterName(), false),
		Thumbprints: thumbprints,
	})

	return nil
}

func buildDiscoveryJSON(issuerURL string) ([]byte, error) {
	d := oidcDiscovery{
		Issuer:                fmt.Sprintf("%v/", issuerURL),
		JWKSURI:               fmt.Sprintf("%v/keys.json", issuerURL),
		AuthorizationEndpoint: "urn:kubernetes:programmatic_authorization",
		ResponseTypes:         []string{"id_token"},
		SubjectTypes:          []string{"public"},
		SigningAlgs:           []string{"RS256"},
		ClaimsSupported:       []string{"sub", "iss"},
	}
	return json.MarshalIndent(d, "", "")
}

type KeyResponse struct {
	Keys []jose.JSONWebKey `json:"keys"`
}

type OIDCKeys struct {
	SigningKey *fitasks.Keypair
}

// GetDependencies adds CA to the list of dependencies
func (o *OIDCKeys) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return []fi.Task{
		o.SigningKey,
	}
}
func (o *OIDCKeys) Open() (io.Reader, error) {

	certBytes, err := fi.ResourceAsBytes(o.SigningKey.Certificate())
	if err != nil {
		return nil, fmt.Errorf("failed to get cert: %w", err)
	}
	block, _ := pem.Decode(certBytes)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse cert: %w", err)
	}

	publicKey := cert.PublicKey

	publicKeyDERBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize public key to DER format: %v", err)
	}

	hasher := crypto.SHA256.New()
	hasher.Write(publicKeyDERBytes)
	publicKeyDERHash := hasher.Sum(nil)

	keyID := base64.RawURLEncoding.EncodeToString(publicKeyDERHash)

	keys := []jose.JSONWebKey{
		{
			Key:       publicKey,
			KeyID:     keyID,
			Algorithm: string(jose.RS256),
			Use:       "sig",
		},
	}

	keyResponse := KeyResponse{Keys: keys}
	jsonBytes, err := json.MarshalIndent(keyResponse, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json: %w", err)
	}

	return bytes.NewReader(jsonBytes), nil
}

func getFingerprints() []string {

	//These strings are the sha1 of the two possible S3 root CAs.
	return []string{
		"9e99a48a9960b14926bb7f3b02e22da2b0ab7280",
		"a9d53002e97e00e043244f3d170d6f4c414104fd",
	}

}

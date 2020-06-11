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
	"encoding/json"
	"fmt"

	"k8s.io/kops/pkg/model"
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
	provider := b.Cluster.Spec.ServiceOIDCProvider
	if provider == nil {
		return nil
	}

	issuerURL := provider.IssuerURL

	saSigner := &fitasks.Keypair{
		Name:      fi.String("service-oidc-ca"),
		Lifecycle: b.Lifecycle,
		Subject:   "cn=service-oidc-ca",
		Type:      "ca",
	}
	c.AddTask(saSigner)

	discovery, err := buildDicoveryJSON(issuerURL)
	if err != nil {
		return err
	}
	discoveryFile := &fitasks.ManagedFile{
		Contents:  fi.WrapResource(fi.NewBytesResource(discovery)),
		Lifecycle: b.Lifecycle,
		Location:  fi.String("discovery.json"),
		Name:      fi.String("discovery.json"),
	}
	c.AddTask(discoveryFile)

	// TODO create keys.json from https://github.com/aws/amazon-eks-pod-identity-webhook/blob/master/hack/self-hosted/main.go
	keysContents := ""
	keysFile := &fitasks.ManagedFile{
		Contents:  fi.WrapResource(fi.NewStringResource(keysContents)),
		Lifecycle: b.Lifecycle,
		Location:  fi.String("keys.json"),
		Name:      fi.String("keys.json"),
	}
	c.AddTask(keysFile)

	thumbprints := make([]*string, len(provider.IssuerCAThumbprints))
	for i, t := range provider.IssuerCAThumbprints {
		thumbprints[i] = &t
	}
	oidcProvider := &awstasks.IAMOIDCProvider{
		Name:        fi.String(b.ClusterName()),
		Lifecycle:   b.Lifecycle,
		URL:         fi.String(issuerURL),
		ClientIDs:   []*string{fi.String(defaultAudience)},
		Thumbprints: thumbprints,
	}
	c.AddTask(oidcProvider)

	return nil
}

func buildDicoveryJSON(issuerURL string) ([]byte, error) {
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

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

var _ fi.ModelBuilder = &OIDCProviderBuilder{}

const (
	stsAudience   = "sts.amazonaws.com"
	discoveryJSON = `
{
		"issuer": "%v/",
		"jwks_uri": "%v/keys.json",
		"authorization_endpoint": "urn:kubernetes:programmatic_authorization",
		"response_types_supported": [
						"id_token"
		],
		"subject_types_supported": [
						"public"
		],
		"id_token_signing_alg_values_supported": [
						"RS256"
		],
		"claims_supported": [
						"sub",
						"iss"
		]
}`
)

func (b *OIDCProviderBuilder) Build(c *fi.ModelBuilderContext) error {
	provider := b.Cluster.Spec.ServiceOIDCProvider
	if provider == nil {
		return nil
	}

	issuerURL := provider.IssuerURL

	saSigner := &fitasks.Keypair{
		Name:      fi.String("service-account-signer"),
		Lifecycle: b.Lifecycle,
		Subject:   "cn=service-account-signer",
		Type:      "ca",
	}
	c.AddTask(saSigner)

	discoveryContents := fmt.Sprintf(discoveryJSON, issuerURL, issuerURL)
	discoveryFile := &fitasks.ManagedFile{
		Contents:  fi.WrapResource(fi.NewStringResource(discoveryContents)),
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
		ClientIDs:   []*string{fi.String(stsAudience)},
		Thumbprints: thumbprints,
	}
	c.AddTask(oidcProvider)

	return nil
}

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

var _ fi.ModelBuilder = &OIDCProviderBuilder{}

const (
	defaultAudience = "amazonaws.com"
)

func (b *OIDCProviderBuilder) Build(c *fi.ModelBuilderContext) error {
	var thumbprints []fi.Resource
	var issuerURL string

	if featureflag.PublicJWKS.Enabled() {
		serviceAccountIssuer, err := iam.ServiceAccountIssuer(b.ClusterName(), &b.Cluster.Spec)
		if err != nil {
			return err
		}
		issuerURL = serviceAccountIssuer

		caTaskObject, found := c.Tasks["Keypair/ca"]
		if !found {
			return fmt.Errorf("keypair/ca task not found")
		}

		caTask := caTaskObject.(*fitasks.Keypair)
		fingerprint := caTask.CertificateSHA1Fingerprint()

		thumbprints = []fi.Resource{fingerprint}
	}

	if issuerURL == "" {
		return nil
	}

	c.AddTask(&awstasks.IAMOIDCProvider{
		Name:        fi.String(b.ClusterName()),
		Lifecycle:   b.Lifecycle,
		URL:         fi.String(issuerURL),
		ClientIDs:   []*string{fi.String(defaultAudience)},
		Thumbprints: thumbprints,
	})

	return nil
}

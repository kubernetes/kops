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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

// OIDCProviderBuilder configures IAM OIDC Provider
type OIDCProviderBuilder struct {
	*AWSModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &OIDCProviderBuilder{}

const (
	defaultAudience = "amazonaws.com"
)

func (b *OIDCProviderBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if b.Cluster.Spec.ServiceAccountIssuerDiscovery == nil ||
		!b.Cluster.Spec.ServiceAccountIssuerDiscovery.EnableAWSOIDCProvider {
		return nil
	}

	fingerprints := getFingerprints()

	audiences := []string{defaultAudience}
	if b.Cluster.Spec.ServiceAccountIssuerDiscovery.AdditionalAudiences != nil {
		audiences = append(audiences, b.Cluster.Spec.ServiceAccountIssuerDiscovery.AdditionalAudiences...)
	}

	c.AddTask(&awstasks.IAMOIDCProvider{
		Name:        fi.PtrTo(b.ClusterName()),
		Lifecycle:   b.Lifecycle,
		URL:         b.Cluster.Spec.KubeAPIServer.ServiceAccountIssuer,
		ClientIDs:   audiences,
		Tags:        b.CloudTags(b.ClusterName(), false),
		Thumbprints: fingerprints,
	})

	return nil
}

func getFingerprints() []string {
	// These strings are the sha1 of the two possible S3 root CAs.
	return []string{
		"9e99a48a9960b14926bb7f3b02e22da2b0ab7280",
		"a9d53002e97e00e043244f3d170d6f4c414104fd",
	}
}

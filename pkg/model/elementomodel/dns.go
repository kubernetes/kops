/*
Copyright 2026 The Kubernetes Authors.

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

package elementomodel

import "k8s.io/kops/upup/pkg/fi"

// DNSModelBuilder is the provider-native integration point for Elemento-managed
// DNS records that must exist before nodeup starts.
type DNSModelBuilder struct {
	*ElementoModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.Cluster.PublishesDNSRecords() {
		return nil
	}

	// TODO: Create Elemento DNS tasks here for the records needed during early bootstrap.
	// At minimum this should cover:
	// - api.<cluster> when the cluster uses DNS instead of an API load balancer hostname
	// - api.internal.<cluster> so kubeconfig, service-account issuer discovery, and
	//   internal control-plane traffic can resolve before in-cluster components reconcile
	// - kops-controller.internal.<cluster> so worker nodeup can reach the config server
	//
	// TODO: Replace these comments with calls to the Elemento DNS API once the SDK
	// surface for zone/record management is available.
	_ = c
	return nil
}

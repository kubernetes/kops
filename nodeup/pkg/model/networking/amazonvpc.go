/*
Copyright 2020 The Kubernetes Authors.

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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"

	"k8s.io/kops/upup/pkg/fi"
)

// AmazonVPCBuilder writes AmazonVPC's assets
type AmazonVPCBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &AmazonVPCBuilder{}

// Build is responsible for configuring the network cni
func (b *AmazonVPCBuilder) Build(c *fi.ModelBuilderContext) error {

	if b.Cluster.Spec.Networking.AmazonVPC != nil {
		b.AddCNIBinAssets(c, []string{"loopback"})
	}
	return nil
}

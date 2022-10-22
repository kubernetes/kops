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

package model

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

type WarmPoolBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &WarmPoolBuilder{}

func (b *WarmPoolBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	// Check if the cloud provider is AWS
	if b.CloudProvider != kops.CloudProviderAWS {
		return nil
	}

	// Pre-pull container images during pre-initialization
	if b.NodeupConfig != nil && b.ConfigurationMode == "Warming" {
		for _, image := range b.NodeupConfig.WarmPoolImages {
			c.AddTask(&nodetasks.PullImageTask{
				Name:    image,
				Runtime: b.Cluster.Spec.ContainerRuntime,
			})
		}
	}

	return nil
}

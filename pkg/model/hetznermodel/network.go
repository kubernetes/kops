/*
Copyright 2022 The Kubernetes Authors.

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

package hetznermodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetznertasks"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*HetznerModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	network := &hetznertasks.Network{
		Name:      fi.String(b.ClusterName()),
		Lifecycle: b.Lifecycle,
	}

	if b.Cluster.Spec.NetworkID == "" {
		network.IPRange = b.Cluster.Spec.NetworkCIDR
		network.Region = b.Region
		network.Subnets = []string{
			b.Cluster.Spec.NetworkCIDR,
		}
		network.Labels = map[string]string{
			hetzner.TagKubernetesClusterName: b.ClusterName(),
		}
	} else {
		network.ID = fi.String(b.Cluster.Spec.NetworkID)
	}

	c.AddTask(network)

	return nil
}

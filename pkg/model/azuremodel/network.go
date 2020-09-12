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

package azuremodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// NetworkModelBuilder configures a Virtual Network and subnets.
type NetworkModelBuilder struct {
	*AzureModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

// Build builds tasks for creating a virtual network and subnets.
func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	networkTask := &azuretasks.VirtualNetwork{
		Name:          fi.String(b.NameForVirtualNetwork()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		CIDR:          fi.String(b.Cluster.Spec.NetworkCIDR),
		Tags:          map[string]*string{},
	}
	c.AddTask(networkTask)

	for _, subnetSpec := range b.Cluster.Spec.Subnets {
		subnetTask := &azuretasks.Subnet{
			Name:           fi.String(subnetSpec.Name),
			Lifecycle:      b.Lifecycle,
			ResourceGroup:  b.LinkToResourceGroup(),
			VirtualNetwork: b.LinkToVirtualNetwork(),
			CIDR:           fi.String(subnetSpec.CIDR),
		}
		c.AddTask(subnetTask)
	}

	rtTask := &azuretasks.RouteTable{
		Name:          fi.String(b.NameForRouteTable()),
		Lifecycle:     b.Lifecycle,
		ResourceGroup: b.LinkToResourceGroup(),
		Tags:          map[string]*string{},
	}
	c.AddTask(rtTask)

	return nil
}

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

package alimodel

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/alitasks"
)

// NetworkModelBuilder configures VPC network objects
type NetworkModelBuilder struct {
	*ALIModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	sharedVPC := b.Cluster.SharedVPC()
	vpcName := b.ClusterName()
	tags := b.CloudTags(vpcName, sharedVPC)

	// VPC that holds everything for the cluster
	{
		vpcTags := tags
		if sharedVPC {
			// We don't tag a shared VPC
			vpcTags = nil
		}

		vpc := &alitasks.VPC{
			Name:      s(vpcName),
			Lifecycle: b.Lifecycle,
			Shared:    fi.Bool(sharedVPC),
			Tags:      vpcTags,
		}

		if b.Cluster.Spec.NetworkID != "" {
			vpc.ID = s(b.Cluster.Spec.NetworkID)
		}

		if b.Cluster.Spec.NetworkCIDR != "" {
			vpc.CIDR = s(b.Cluster.Spec.NetworkCIDR)
		}
		c.AddTask(vpc)
	}

	natGateway := &alitasks.NatGateway{
		Name:      s(b.GetNameForNatGateway()),
		Lifecycle: b.Lifecycle,
		VPC:       b.LinkToVPC(),
	}
	c.AddTask(natGateway)

	eip := &alitasks.EIP{
		Name:       s(b.GetNameForEIP()),
		Lifecycle:  b.Lifecycle,
		NatGateway: b.LinkToNatGateway(),
		Available:  fi.Bool(false),
	}
	c.AddTask(eip)

	for i := range b.Cluster.Spec.Subnets {
		subnetSpec := &b.Cluster.Spec.Subnets[i]

		vswitch := &alitasks.VSwitch{
			Name:      s(b.GetNameForVSwitch(subnetSpec.Name)),
			Lifecycle: b.Lifecycle,
			VPC:       b.LinkToVPC(),
			ZoneId:    s(subnetSpec.Zone),
			CidrBlock: s(subnetSpec.CIDR),
			Shared:    fi.Bool(false),
		}

		if subnetSpec.ProviderID != "" {
			vswitch.VSwitchId = s(subnetSpec.ProviderID)
			vswitch.Shared = fi.Bool(true)
		}

		c.AddTask(vswitch)

		if subnetSpec.Type == kops.SubnetTypePrivate {
			vswitchSNAT := &alitasks.VSwitchSNAT{
				Name:       s(b.GetNameForVSwitchSNAT(subnetSpec.Name)),
				Lifecycle:  b.Lifecycle,
				NatGateway: b.LinkToNatGateway(),
				VSwitch:    b.LinkToVSwitch(subnetSpec.Name),
				EIP:        b.LinkToEIP(),
			}

			if subnetSpec.ProviderID != "" {
				vswitchSNAT.Shared = fi.Bool(true)
			}

			c.AddTask(vswitchSNAT)
		}
	}

	return nil
}

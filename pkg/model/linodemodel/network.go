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

package linodemodel

import (
	"fmt"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

// NetworkModelBuilder configures the Linode (Akamai) VPC for the cluster.
type NetworkModelBuilder struct {
	*LinodeModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if len(b.Cluster.Spec.Networking.Subnets) == 0 {
		return fmt.Errorf("linode VPC requires at least one subnet")
	}

	seenSubnetNames := map[string]string{}
	region := ""
	for i, subnet := range b.Cluster.Spec.Networking.Subnets {
		subnetName := subnet.Name
		if subnetName == "" {
			subnetName = fmt.Sprintf("subnet %d", i)
		}

		if subnet.Name == "" {
			return fmt.Errorf("linode subnet %q requires a name", subnetName)
		}
		if subnet.Region == "" {
			return fmt.Errorf("linode subnet %q requires a region", subnetName)
		}
		if subnet.CIDR == "" {
			return fmt.Errorf("linode subnet %q requires a CIDR", subnetName)
		}

		normalizedSubnetName := linode.NormalizeLinodeLabel(b.ClusterName() + "-" + subnet.Name)
		if previousSubnetName, found := seenSubnetNames[normalizedSubnetName]; found {
			return fmt.Errorf("linode subnets %q and %q normalize to the same label %q", previousSubnetName, subnetName, normalizedSubnetName)
		}
		seenSubnetNames[normalizedSubnetName] = subnetName

		if region == "" {
			region = subnet.Region
			continue
		}
		if subnet.Region != region {
			return fmt.Errorf("linode subnets must all use the same region; found %q and %q", region, subnet.Region)
		}
	}

	name := linode.NormalizeLinodeLabel(b.ClusterName())
	description := fmt.Sprintf("kOps VPC for %s", b.ClusterName())
	vpcTask := &linodetasks.VPC{
		Name:        new(name),
		Lifecycle:   b.Lifecycle,
		Description: new(description),
		Region:      new(region),
	}
	c.AddTask(vpcTask)

	for _, subnet := range b.Cluster.Spec.Networking.Subnets {
		subnetName := linode.NormalizeLinodeLabel(b.ClusterName() + "-" + subnet.Name)
		subnetTask := &linodetasks.Subnet{
			Name:      new(subnetName),
			Lifecycle: b.Lifecycle,
			IPv4:      new(subnet.CIDR),
			VPC:       vpcTask,
		}
		c.AddTask(subnetTask)
	}

	return nil
}

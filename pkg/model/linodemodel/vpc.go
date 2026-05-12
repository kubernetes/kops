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

// VPCModelBuilder configures the Linode (Akamai) VPC for the cluster.
type VPCModelBuilder struct {
	*LinodeModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &VPCModelBuilder{}

func (b *VPCModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	region := ""
	for _, subnet := range b.Cluster.Spec.Networking.Subnets {
		if subnet.Region != "" {
			region = subnet.Region
			break
		}
	}
	if region == "" {
		return fmt.Errorf("linode VPC requires at least one subnet with a region")
	}

	name := linode.NormalizeLinodeVPCLabel(b.ClusterName())
	description := fmt.Sprintf("kOps VPC for %s", b.ClusterName())
	c.AddTask(&linodetasks.VPC{
		Name:        fi.PtrTo(name),
		Lifecycle:   b.Lifecycle,
		Description: fi.PtrTo(description),
		Region:      fi.PtrTo(region),
	})

	return nil
}

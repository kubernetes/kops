/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
)

const (
	DefaultVolumeSizeNode    = 128
	DefaultVolumeSizeMaster  = 64
	DefaultVolumeSizeBastion = 32
	DefaultVolumeType        = "gp2"
	DefaultVolumeIops        = 100
)

// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*AWSModelContext

	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		name := b.AutoscalingGroupName(ig)

		// LaunchConfiguration
		var launchConfiguration *awstasks.LaunchConfiguration
		{
			volumeSize := fi.Int32Value(ig.Spec.RootVolumeSize)
			if volumeSize == 0 {
				switch ig.Spec.Role {
				case kops.InstanceGroupRoleMaster:
					volumeSize = DefaultVolumeSizeMaster
				case kops.InstanceGroupRoleNode:
					volumeSize = DefaultVolumeSizeNode
				case kops.InstanceGroupRoleBastion:
					volumeSize = DefaultVolumeSizeBastion
				default:
					return fmt.Errorf("this case should not get hit, kops.Role not found %s", ig.Spec.Role)
				}
			}
			volumeType := fi.StringValue(ig.Spec.RootVolumeType)
			volumeIops := fi.Int32Value(ig.Spec.RootVolumeIops)

			switch volumeType {
			case "io1":
				if volumeIops == 0 {
					volumeIops = DefaultVolumeIops
				}
			default:
				volumeType = DefaultVolumeType
			}

			t := &awstasks.LaunchConfiguration{
				Name:      s(name),
				Lifecycle: b.Lifecycle,

				SecurityGroups: []*awstasks.SecurityGroup{
					b.LinkToSecurityGroup(ig.Spec.Role),
				},
				IAMInstanceProfile: b.LinkToIAMInstanceProfile(ig),
				ImageID:            s(ig.Spec.Image),
				InstanceType:       s(ig.Spec.MachineType),

				RootVolumeSize:         i64(int64(volumeSize)),
				RootVolumeType:         s(volumeType),
				RootVolumeOptimization: ig.Spec.RootVolumeOptimization,
			}

			if volumeType == "io1" {
				t.RootVolumeIops = i64(int64(volumeIops))
			}

			if ig.Spec.Tenancy != "" {
				t.Tenancy = s(ig.Spec.Tenancy)
			}

			for _, id := range ig.Spec.AdditionalSecurityGroups {
				sgTask := &awstasks.SecurityGroup{
					Name:   fi.String(id),
					ID:     fi.String(id),
					Shared: fi.Bool(true),
				}
				if err := c.EnsureTask(sgTask); err != nil {
					return err
				}
				t.SecurityGroups = append(t.SecurityGroups, sgTask)
			}

			var err error

			if t.SSHKey, err = b.LinkToSSHKey(); err != nil {
				return err
			}

			if t.UserData, err = b.BootstrapScript.ResourceNodeUp(ig, b.Cluster.Spec.EgressProxy); err != nil {
				return err
			}

			if fi.StringValue(ig.Spec.MaxPrice) != "" {
				spotPrice := fi.StringValue(ig.Spec.MaxPrice)
				t.SpotPrice = spotPrice
			}

			{
				// TODO: Wrapper / helper class to analyze clusters
				subnetMap := make(map[string]*kops.ClusterSubnetSpec)
				for i := range b.Cluster.Spec.Subnets {
					subnet := &b.Cluster.Spec.Subnets[i]
					subnetMap[subnet.Name] = subnet
				}

				var subnetType kops.SubnetType
				for _, subnetName := range ig.Spec.Subnets {
					subnet := subnetMap[subnetName]
					if subnet == nil {
						return fmt.Errorf("InstanceGroup %q uses subnet %q that does not exist", ig.ObjectMeta.Name, subnetName)
					}
					if subnetType != "" && subnetType != subnet.Type {
						return fmt.Errorf("InstanceGroup %q cannot be in subnets of different Type", ig.ObjectMeta.Name)
					}
					subnetType = subnet.Type
				}

				associatePublicIP := true
				switch subnetType {
				case kops.SubnetTypePublic, kops.SubnetTypeUtility:
					associatePublicIP = true
					if ig.Spec.AssociatePublicIP != nil {
						associatePublicIP = *ig.Spec.AssociatePublicIP
					}

				case kops.SubnetTypePrivate:
					associatePublicIP = false
					if ig.Spec.AssociatePublicIP != nil {
						// This isn't meaningful - private subnets can't have public ip
						//associatePublicIP = *ig.Spec.AssociatePublicIP
						if *ig.Spec.AssociatePublicIP {
							glog.Warningf("Ignoring AssociatePublicIP=true for private InstanceGroup %q", ig.ObjectMeta.Name)
						}
					}

				default:
					return fmt.Errorf("unknown subnet type %q", subnetType)
				}
				t.AssociatePublicIP = &associatePublicIP
			}
			c.AddTask(t)

			launchConfiguration = t
		}

		// AutoscalingGroup
		{
			t := &awstasks.AutoscalingGroup{
				Name:      s(name),
				Lifecycle: b.Lifecycle,

				LaunchConfiguration: launchConfiguration,
			}

			minSize := int32(1)
			maxSize := int32(1)
			if ig.Spec.MinSize != nil {
				minSize = fi.Int32Value(ig.Spec.MinSize)
			} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
				minSize = 2
			}
			if ig.Spec.MaxSize != nil {
				maxSize = *ig.Spec.MaxSize
			} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
				maxSize = 2
			}

			t.MinSize = i64(int64(minSize))
			t.MaxSize = i64(int64(maxSize))

			subnets, err := b.GatherSubnets(ig)
			if err != nil {
				return err
			}
			if len(subnets) == 0 {
				return fmt.Errorf("could not determine any subnets for InstanceGroup %q; subnets was %s", ig.ObjectMeta.Name, ig.Spec.Subnets)
			}
			for _, subnet := range subnets {
				t.Subnets = append(t.Subnets, b.LinkToSubnet(subnet))
			}

			tags, err := b.CloudTagsForInstanceGroup(ig)
			if err != nil {
				return fmt.Errorf("error building cloud tags: %v", err)
			}
			t.Tags = tags

			c.AddTask(t)
		}
	}

	return nil
}

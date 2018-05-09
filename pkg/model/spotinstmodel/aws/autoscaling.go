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

package aws

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	awstasks "k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks/aws"
)

const (
	DefaultVolumeSizeNode    = 128
	DefaultVolumeSizeMaster  = 64
	DefaultVolumeSizeBastion = 32
	DefaultVolumeType        = "gp2"
)

// AutoscalingGroupModelBuilder configures Group objects
type AutoscalingGroupModelBuilder struct {
	*ModelContext
	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		group := &awstasks.AutoscalingGroup{
			Lifecycle:            b.Lifecycle,
			Name:                 fi.String(b.AutoscalingGroupName(ig)),
			Product:              b.Cluster.Spec.CloudConfig.SpotinstProduct,
			Orientation:          b.Cluster.Spec.CloudConfig.SpotinstOrientation,
			ImageID:              fi.String(ig.Spec.Image),
			OnDemandInstanceType: fi.String(strings.Split(ig.Spec.MachineType, ",")[0]),
			SpotInstanceTypes:    strings.Split(ig.Spec.MachineType, ","),
			IAMInstanceProfile:   b.LinkToIAMInstanceProfile(ig),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToSecurityGroup(ig.Spec.Role),
			},
		}

		// Root volume.
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
			if volumeType == "" {
				volumeType = DefaultVolumeType
			}

			group.RootVolumeSize = fi.Int64(int64(volumeSize))
			group.RootVolumeType = fi.String(volumeType)
			group.RootVolumeOptimization = ig.Spec.RootVolumeOptimization
		}

		// Tenancy.
		{
			if ig.Spec.Tenancy != "" {
				group.Tenancy = fi.String(ig.Spec.Tenancy)
			}
		}

		// Risk.
		{
			var risk float64
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				risk = 0
			case kops.InstanceGroupRoleNode:
				risk = 100
			case kops.InstanceGroupRoleBastion:
				risk = 0
			default:
				return fmt.Errorf("this case should not get hit, kops.Role not found %s", ig.Spec.Role)
			}
			group.Risk = &risk
		}

		// Security groups.
		{
			for _, id := range ig.Spec.AdditionalSecurityGroups {
				sgTask := &awstasks.SecurityGroup{
					Name:   fi.String(id),
					ID:     fi.String(id),
					Shared: fi.Bool(true),
				}
				if err := c.EnsureTask(sgTask); err != nil {
					return err
				}
				group.SecurityGroups = append(group.SecurityGroups, sgTask)
			}
		}

		// SSH Key.
		{
			sshKey, err := b.LinkToSSHKey()
			if err != nil {
				return err
			}
			group.SSHKey = sshKey
		}

		// User data.
		{
			userData, err := b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
			if err != nil {
				return err
			}
			group.UserData = userData
		}

		// Public IP.
		{
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
					if *ig.Spec.AssociatePublicIP {
						glog.Warningf("Ignoring AssociatePublicIP=true for private InstanceGroup %q", ig.ObjectMeta.Name)
					}
				}
			default:
				return fmt.Errorf("unknown subnet type %q", subnetType)
			}
			group.AssociatePublicIP = &associatePublicIP
		}

		// Subnets.
		{
			subnets, err := b.GatherSubnets(ig)
			if err != nil {
				return err
			}
			if len(subnets) == 0 {
				return fmt.Errorf("could not determine any subnets for InstanceGroup %q; subnets was %s", ig.ObjectMeta.Name, ig.Spec.Subnets)
			}
			for _, subnet := range subnets {
				group.Subnets = append(group.Subnets, b.LinkToSubnet(subnet))
			}
		}

		// Capacity.
		{
			minSize := int32(1)
			if ig.Spec.MinSize != nil {
				minSize = fi.Int32Value(ig.Spec.MinSize)
			} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
				minSize = 2
			}

			maxSize := int32(1)
			if ig.Spec.MaxSize != nil {
				maxSize = *ig.Spec.MaxSize
			} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
				maxSize = 10
			}

			group.MinSize = fi.Int64(int64(minSize))
			group.MaxSize = fi.Int64(int64(maxSize))
		}

		// Tags.
		{
			tags, err := b.CloudTagsForInstanceGroup(ig)
			if err != nil {
				return fmt.Errorf("error building cloud tags: %v", err)
			}
			tags[awsup.TagClusterName] = b.ClusterName()
			tags["Name"] = b.AutoscalingGroupName(ig)
			group.Tags = tags
		}

		// Integration.
		{
			if ig.Spec.Role != kops.InstanceGroupRoleBastion {
				group.IntegrationClusterIdentifier = fi.String(b.ClusterName())

				if ig.Spec.Role == kops.InstanceGroupRoleNode {
					group.IntegrationAutoScaleEnabled = fi.Bool(true)
					group.IntegrationAutoScaleNodeLabels = ig.Spec.NodeLabels
				}
			}
		}

		c.AddTask(group)
	}

	return nil
}

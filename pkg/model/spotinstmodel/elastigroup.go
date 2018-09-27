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

package spotinstmodel

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/spotinsttasks"
)

// ElastigroupModelBuilder configures Elastigroup objects
type ElastigroupModelBuilder struct {
	*SpotinstModelContext

	BootstrapScript   *model.BootstrapScript
	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &ElastigroupModelBuilder{}

func (b *ElastigroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		glog.V(2).Infof("Building instance group: %s", b.AutoscalingGroupName(ig))

		group := &spotinsttasks.Elastigroup{
			Lifecycle:            b.Lifecycle,
			Name:                 fi.String(b.AutoscalingGroupName(ig)),
			ImageID:              fi.String(ig.Spec.Image),
			Monitoring:           fi.Bool(false),
			OnDemandInstanceType: fi.String(strings.Split(ig.Spec.MachineType, ",")[0]),
			SpotInstanceTypes:    strings.Split(ig.Spec.MachineType, ","),
			SecurityGroups: []*awstasks.SecurityGroup{
				b.LinkToSecurityGroup(ig.Spec.Role),
			},
		}

		// Cloud config.
		{
			if cfg := b.Cluster.Spec.CloudConfig; cfg != nil {
				// Product.
				if cfg.SpotinstProduct != nil {
					group.Product = cfg.SpotinstProduct
				}

				// Orientation.
				if cfg.SpotinstOrientation != nil {
					group.Orientation = cfg.SpotinstOrientation
				}
			}
		}

		// Instance profile.
		{
			iprof, err := b.LinkToIAMInstanceProfile(ig)
			if err != nil {
				return err
			}
			group.IAMInstanceProfile = iprof
		}

		// Root volume.
		{
			volumeSize := fi.Int32Value(ig.Spec.RootVolumeSize)
			if volumeSize == 0 {
				var err error
				volumeSize, err = defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
				if err != nil {
					return err
				}
			}

			volumeType := fi.StringValue(ig.Spec.RootVolumeType)
			if volumeType == "" {
				volumeType = awsmodel.DefaultVolumeType
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
				return fmt.Errorf("spotinst: kops.Role not found %s", ig.Spec.Role)
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

		// Load balancer.
		{
			var lb *awstasks.LoadBalancer
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				if b.UseLoadBalancerForAPI() {
					lb = b.LinkToELB("api")
				}
			case kops.InstanceGroupRoleBastion:
				lb = b.LinkToELB(model.BastionELBSecurityGroupPrefix)
			}
			if lb != nil {
				group.LoadBalancer = lb
			}
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
					return fmt.Errorf("spotinst: InstanceGroup %q uses subnet %q that does not exist", ig.ObjectMeta.Name, subnetName)
				}
				if subnetType != "" && subnetType != subnet.Type {
					return fmt.Errorf("spotinst: InstanceGroup %q cannot be in subnets of different Type", ig.ObjectMeta.Name)
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
				return fmt.Errorf("spotinst: unknown subnet type %q", subnetType)
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
				return fmt.Errorf("spotinst: could not determine any subnets for InstanceGroup %q; subnets was %s", ig.ObjectMeta.Name, ig.Spec.Subnets)
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
				return fmt.Errorf("spotinst: error building cloud tags: %v", err)
			}
			tags[awsup.TagClusterName] = b.ClusterName()
			tags["Name"] = b.AutoscalingGroupName(ig)
			group.Tags = tags
		}

		// Integration.
		{
			if ig.Spec.Role != kops.InstanceGroupRoleBastion {
				group.IntegrationClusterIdentifier = fi.String(b.ClusterName())

				var disableAutoScaler bool
				if cfg := b.Cluster.Spec.CloudConfig; cfg != nil && cfg.SpotinstDisableAutoScaler != nil {
					disableAutoScaler = fi.BoolValue(cfg.SpotinstDisableAutoScaler)
				}

				if ig.Spec.Role == kops.InstanceGroupRoleNode && !disableAutoScaler {
					group.IntegrationAutoScaleEnabled = fi.Bool(true)

					if len(ig.Spec.NodeLabels) > 0 {
						nodeLabels := make(map[string]string)
						for k, v := range ig.Spec.NodeLabels {
							if !strings.HasPrefix(k, kops.NodeLabelInstanceGroup) {
								nodeLabels[k] = v
							}
						}
						if len(nodeLabels) > 0 {
							group.IntegrationAutoScaleNodeLabels = nodeLabels
						}
					}
				}
			}
		}

		c.AddTask(group)
	}

	return nil
}

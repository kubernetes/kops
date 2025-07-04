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

package gcemodel

import (
	"fmt"
	"slices"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/model/iam"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

const (
	DefaultVolumeType = "pd-standard"
)

// TODO: rework these parts to be more GCE native. ie: Managed Instance Groups > ASGs
// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*GCEModelContext

	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &AutoscalingGroupModelBuilder{}

// Build the GCE instance template object for an InstanceGroup
// We are then able to extract out the fields when running with the clusterapi.
func (b *AutoscalingGroupModelBuilder) buildInstanceTemplate(c *fi.CloudupModelBuilderContext, ig *kops.InstanceGroup, subnet *kops.ClusterSubnetSpec) (*gcetasks.InstanceTemplate, error) {
	// Indented to keep diff manageable
	// TODO: Remove spurious indent
	{
		var err error
		name := b.SafeObjectName(ig.ObjectMeta.Name)

		startupScript, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return nil, err
		}

		{
			var volumeSize int32
			var volumeType string
			if ig.Spec.RootVolume != nil {
				volumeSize = fi.ValueOf(ig.Spec.RootVolume.Size)
				volumeType = fi.ValueOf(ig.Spec.RootVolume.Type)
			}
			if volumeSize == 0 {
				volumeSize, err = defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
				if err != nil {
					return nil, err
				}
			}
			if volumeType == "" {
				volumeType = DefaultVolumeType
			}

			namePrefix := gce.LimitedLengthName(name, gcetasks.InstanceTemplateNamePrefixMaxLength)
			network, err := b.LinkToNetwork()
			if err != nil {
				return nil, err
			}
			t := &gcetasks.InstanceTemplate{
				Name:           s(name),
				NamePrefix:     s(namePrefix),
				Lifecycle:      b.Lifecycle,
				Network:        network,
				MachineType:    s(ig.Spec.MachineType),
				BootDiskType:   s(volumeType),
				BootDiskSizeGB: i64(int64(volumeSize)),
				BootDiskImage:  s(ig.Spec.Image),

				Preemptible:          fi.PtrTo(fi.ValueOf(ig.Spec.GCPProvisioningModel) == "SPOT"),
				GCPProvisioningModel: ig.Spec.GCPProvisioningModel,

				HasExternalIP: fi.PtrTo(subnet.Type == kops.SubnetTypePublic || subnet.Type == kops.SubnetTypeUtility || ig.IsBastion()),

				Scopes: []string{
					"compute-rw",
					"monitoring",
					"logging-write",
					"cloud-platform",
				},
				Metadata: map[string]fi.Resource{
					gcemetadata.MetadataKeyClusterName:           fi.NewStringResource(b.ClusterName()),
					nodeidentitygce.MetadataKeyInstanceGroupName: fi.NewStringResource(ig.Name),
				},
			}

			if startupScript != nil {
				if !fi.ValueOf(b.Cluster.Spec.CloudProvider.GCE.UseStartupScript) {
					// Use "user-data" instead of "startup-script", for compatibility with cloud-init
					t.Metadata["user-data"] = startupScript
				} else {
					t.Metadata["startup-script"] = startupScript
				}
			}

			if ig.Spec.Role == kops.InstanceGroupRoleNode {
				autoscalerEnvVars := "os_distribution=ubuntu;arch=amd64;os=linux"
				if strings.HasPrefix(ig.Spec.Image, "cos-cloud/") {
					autoscalerEnvVars = "os_distribution=cos;arch=amd64;os=linux"
				}

				if len(ig.Spec.NodeLabels) > 0 {
					var nodeLabels string
					sortedLabelKeys := make([]string, len(ig.Spec.NodeLabels))
					i := 0
					for k := range ig.Spec.NodeLabels {
						sortedLabelKeys[i] = k
						i++
					}
					slices.SortStableFunc(sortedLabelKeys, func(a, b string) int {
						return strings.Compare(a, b)
					})
					for _, k := range sortedLabelKeys {
						nodeLabels += k + "=" + ig.Spec.NodeLabels[k] + ","
					}
					nodeLabels, _ = strings.CutSuffix(nodeLabels, ",")

					autoscalerEnvVars += ";node_labels=" + nodeLabels
				}

				if len(ig.Spec.Taints) > 0 {
					autoscalerEnvVars += ";node_taints=" + strings.Join(ig.Spec.Taints, ",")
				}

				t.Metadata["kube-env"] = fi.NewStringResource("AUTOSCALER_ENV_VARS: " + autoscalerEnvVars)
			}

			stackType := "IPV4_ONLY"
			if b.IsIPv6Only() {
				// The subnets are dual-mode; IPV6_ONLY is not yet supported.
				// This means that VMs will get an IPv4 and a /96 IPv6.
				// However, pods will still be IPv6 only.
				stackType = "IPV4_IPV6"

				// // Ipv6AccessType must be set when enabling IPv6.
				// // EXTERNAL is currently the only supported value
				// ipv6AccessType := "EXTERNAL"
				// t.Ipv6AccessType = &ipv6AccessType
			}
			t.StackType = &stackType

			nodeRole, err := iam.BuildNodeRoleSubject(ig.Spec.Role, false)
			if err != nil {
				return nil, err
			}

			storagePaths, err := iam.WriteableVFSPaths(b.Cluster, nodeRole)
			if err != nil {
				return nil, err
			}
			if len(storagePaths) == 0 {
				t.Scopes = append(t.Scopes, "storage-ro")
			} else {
				klog.Warningf("enabling storage-rw for etcd backups")
				t.Scopes = append(t.Scopes, "storage-rw")
			}

			if len(b.SSHPublicKeys) > 0 {
				var gFmtKeys []string
				for _, key := range b.SSHPublicKeys {
					gFmtKeys = append(gFmtKeys, fmt.Sprintf("%s: %s", fi.SecretNameSSHPrimary, key))
				}

				t.Metadata["ssh-keys"] = fi.NewStringResource(strings.Join(gFmtKeys, "\n"))
			}

			switch ig.Spec.Role {
			case kops.InstanceGroupRoleControlPlane:
				// Grant DNS permissions
				// TODO: migrate to IAM permissions instead of oldschool scopes?
				t.Scopes = append(t.Scopes, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleControlPlane))
				t.Tags = append(t.Tags, b.GCETagForRole("master"))

			case kops.InstanceGroupRoleNode:
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleNode))

			case kops.InstanceGroupRoleBastion:
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleBastion))
			}
			clusterLabel := gce.LabelForCluster(b.ClusterName())
			roleLabel := gce.GceLabelNameRolePrefix + ig.Spec.Role.ToLowerString()
			t.Labels = map[string]string{
				clusterLabel.Key:              clusterLabel.Value,
				roleLabel:                     ig.Spec.Role.ToLowerString(),
				gce.GceLabelNameInstanceGroup: ig.ObjectMeta.Name,
			}
			if ig.Spec.Role == kops.InstanceGroupRoleControlPlane {
				t.Labels[gce.GceLabelNameRolePrefix+"master"] = "master"
			}

			if gce.UsesIPAliases(b.Cluster) {
				t.CanIPForward = fi.PtrTo(false)

				nodeCIDRMaskSize := int32(24)
				if b.Cluster.Spec.KubeControllerManager.NodeCIDRMaskSize != nil {
					nodeCIDRMaskSize = *b.Cluster.Spec.KubeControllerManager.NodeCIDRMaskSize
				}
				t.AliasIPRanges = map[string]string{
					b.NameForIPAliasRange("pods"): fmt.Sprintf("/%d", nodeCIDRMaskSize),
				}
			} else {
				t.CanIPForward = fi.PtrTo(true)
			}
			t.Subnet = b.LinkToSubnet(subnet)

			t.ServiceAccounts = append(t.ServiceAccounts, b.LinkToServiceAccount(ig))

			//labels, err := b.CloudTagsForInstanceGroup(ig)
			//if err != nil {
			//	return fmt.Errorf("error building cloud tags: %v", err)
			//}
			//t.Labels = labels

			t.GuestAccelerators = []gcetasks.AcceleratorConfig{}
			for _, accelerator := range ig.Spec.GuestAccelerators {
				t.GuestAccelerators = append(t.GuestAccelerators, gcetasks.AcceleratorConfig{
					AcceleratorCount: accelerator.AcceleratorCount,
					AcceleratorType:  accelerator.AcceleratorType,
				})
			}

			return t, nil
		}
	}
}

func (b *AutoscalingGroupModelBuilder) splitToZones(ig *kops.InstanceGroup) (map[string]int, error) {
	// Indented to keep diff manageable
	// TODO: Remove spurious indent
	{
		// AutoscalingGroup
		zones, err := b.FindZonesForInstanceGroup(ig)
		if err != nil {
			return nil, err
		}

		// TODO: Duplicated from aws - move to defaults?
		minSize := 1
		if ig.Spec.MinSize != nil {
			minSize = int(fi.ValueOf(ig.Spec.MinSize))
		} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
			minSize = 2
		}

		// We have to assign instances to the various zones
		// TODO: Switch to regional managed instance group
		// But we can't yet use RegionInstanceGroups:
		// 1) no support in terraform
		// 2) we can't steer to specific zones AFAICT, only to all zones in the region

		targetSizes := make([]int, len(zones))
		totalSize := 0
		for i := range zones {
			targetSizes[i] = minSize / len(zones)
			totalSize += targetSizes[i]
		}
		i := 0
		for {
			if totalSize >= minSize {
				break
			}
			targetSizes[i]++
			totalSize++

			i++
			if i > len(targetSizes) {
				i = 0
			}
		}

		instanceCountByZone := make(map[string]int)
		for i, zone := range zones {
			instanceCountByZone[zone] = targetSizes[i]
		}
		return instanceCountByZone, nil
	}
}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		subnets, err := b.GatherSubnets(ig)
		if err != nil {
			return err
		}

		// On GCE, instance groups cannot have multiple subnets.
		// Because subnets are regional on GCE, this should not be limiting.
		// (IGs can in theory support multiple zones, but in practice we don't recommend this)
		if len(subnets) != 1 {
			return fmt.Errorf("instanceGroup %q has multiple subnets", ig.Name)
		}
		subnet := subnets[0]

		instanceTemplate, err := b.buildInstanceTemplate(c, ig, subnet)
		if err != nil {
			return err
		}
		c.AddTask(instanceTemplate)

		instanceCountByZone, err := b.splitToZones(ig)
		if err != nil {
			return err
		}

		for zone, targetSize := range instanceCountByZone {
			name := gce.NameForInstanceGroupManager(b.Cluster.ObjectMeta.Name, ig.ObjectMeta.Name, zone)
			updatePolicy := &gcetasks.UpdatePolicy{
				MaxSurgeFixed:       1,
				MaxUnavailableFixed: 1,
				MinimalAction:       "REPLACE",
				Type:                "OPPORTUNISTIC",
			}

			t := &gcetasks.InstanceGroupManager{
				Name:                        s(name),
				Lifecycle:                   b.Lifecycle,
				Zone:                        s(zone),
				TargetSize:                  fi.PtrTo(int64(targetSize)),
				UpdatePolicy:                updatePolicy,
				BaseInstanceName:            s(ig.ObjectMeta.Name),
				InstanceTemplate:            instanceTemplate,
				ListManagedInstancesResults: "PAGINATED",
			}

			// Attach masters to load balancer if we're using one
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleControlPlane:
				if b.UseLoadBalancerForAPI() {
					lbSpec := b.Cluster.Spec.API.LoadBalancer
					if lbSpec != nil {
						switch lbSpec.Type {
						case kops.LoadBalancerTypePublic:
							t.TargetPools = append(t.TargetPools, b.LinkToTargetPool("api"))
						case kops.LoadBalancerTypeInternal:
							klog.Warningf("Not hooking the instance group manager up to anything.")
						}
					}
				}
			}

			c.AddTask(t)
		}
	}

	return nil
}

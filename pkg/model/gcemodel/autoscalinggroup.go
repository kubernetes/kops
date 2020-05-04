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
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/defaults"
	"k8s.io/kops/pkg/model/iam"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

const (
	DefaultVolumeType = "pd-standard"
)

// TODO: rework these parts to be more GCE native. ie: Managed Instance Groups > ASGs
// AutoscalingGroupModelBuilder configures AutoscalingGroup objects
type AutoscalingGroupModelBuilder struct {
	*GCEModelContext

	BootstrapScript *model.BootstrapScript
	Lifecycle       *fi.Lifecycle
}

var _ fi.ModelBuilder = &AutoscalingGroupModelBuilder{}

func (b *AutoscalingGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		name := b.SafeObjectName(ig.ObjectMeta.Name)

		startupScript, err := b.BootstrapScript.ResourceNodeUp(ig, b.Cluster)
		if err != nil {
			return err
		}

		// InstanceTemplate
		var instanceTemplate *gcetasks.InstanceTemplate
		{
			volumeSize := fi.Int32Value(ig.Spec.RootVolumeSize)
			if volumeSize == 0 {
				volumeSize, err = defaults.DefaultInstanceGroupVolumeSize(ig.Spec.Role)
				if err != nil {
					return err
				}
			}
			volumeType := fi.StringValue(ig.Spec.RootVolumeType)
			if volumeType == "" {
				volumeType = DefaultVolumeType
			}

			namePrefix := gce.LimitedLengthName(name, gcetasks.InstanceTemplateNamePrefixMaxLength)
			t := &gcetasks.InstanceTemplate{
				Name:           s(name),
				NamePrefix:     s(namePrefix),
				Lifecycle:      b.Lifecycle,
				Network:        b.LinkToNetwork(),
				MachineType:    s(ig.Spec.MachineType),
				BootDiskType:   s(volumeType),
				BootDiskSizeGB: i64(int64(volumeSize)),
				BootDiskImage:  s(ig.Spec.Image),

				// TODO: Support preemptible nodes?
				Preemptible: fi.Bool(false),

				Scopes: []string{
					"compute-rw",
					"monitoring",
					"logging-write",
				},
				Metadata: map[string]*fi.ResourceHolder{
					"startup-script": startupScript,
					//"config": resources/config.yaml $nodeset.Name
					"cluster-name": fi.WrapResource(fi.NewStringResource(b.ClusterName())),
					nodeidentitygce.MetadataKeyInstanceGroupName: fi.WrapResource(fi.NewStringResource(ig.Name)),
				},
			}

			storagePaths, err := iam.WriteableVFSPaths(b.Cluster, ig.Spec.Role)
			if err != nil {
				return err
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

				t.Metadata["ssh-keys"] = fi.WrapResource(fi.NewStringResource(strings.Join(gFmtKeys, "\n")))
			}

			switch ig.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				// Grant DNS permissions
				// TODO: migrate to IAM permissions instead of oldschool scopes?
				t.Scopes = append(t.Scopes, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleMaster))

			case kops.InstanceGroupRoleNode:
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleNode))
			}

			if gce.UsesIPAliases(b.Cluster) {
				t.CanIPForward = fi.Bool(false)

				t.AliasIPRanges = map[string]string{
					b.NameForIPAliasRange("pods"): "/24",
				}
				t.Subnet = b.LinkToIPAliasSubnet()
			} else {
				t.CanIPForward = fi.Bool(true)
			}

			if b.Cluster.Spec.CloudConfig.GCEServiceAccount != "" {
				klog.Infof("VMs using Service Account: %v", b.Cluster.Spec.CloudConfig.GCEServiceAccount)
				// b.Cluster.Spec.GCEServiceAccount = c.GCEServiceAccount
			} else {
				klog.Warning("VMs will be configured to use the GCE default compute Service Account! This is an anti-pattern")
				klog.Warning("Use a pre-created Service Account with the flag: --gce-service-account=account@projectname.iam.gserviceaccount.com")
				b.Cluster.Spec.CloudConfig.GCEServiceAccount = "default"
			}

			klog.Infof("gsa: %v", b.Cluster.Spec.CloudConfig.GCEServiceAccount)
			t.ServiceAccounts = []string{b.Cluster.Spec.CloudConfig.GCEServiceAccount}
			//labels, err := b.CloudTagsForInstanceGroup(ig)
			//if err != nil {
			//	return fmt.Errorf("error building cloud tags: %v", err)
			//}
			//t.Labels = labels

			c.AddTask(t)

			instanceTemplate = t
		}

		// AutoscalingGroup
		zones, err := b.FindZonesForInstanceGroup(ig)
		if err != nil {
			return err
		}

		// TODO: Duplicated from aws - move to defaults?
		minSize := 1
		if ig.Spec.MinSize != nil {
			minSize = int(fi.Int32Value(ig.Spec.MinSize))
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

		for i, targetSize := range targetSizes {
			zone := zones[i]

			name := gce.NameForInstanceGroupManager(b.Cluster, ig, zone)

			t := &gcetasks.InstanceGroupManager{
				Name:             s(name),
				Lifecycle:        b.Lifecycle,
				Zone:             s(zone),
				TargetSize:       fi.Int64(int64(targetSize)),
				BaseInstanceName: s(ig.ObjectMeta.Name),
				InstanceTemplate: instanceTemplate,
			}

			// Attach masters to load balancer if we're using one
			switch ig.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				if b.UseLoadBalancerForAPI() {
					t.TargetPools = append(t.TargetPools, b.LinkToTargetPool("api"))
				}
			}

			c.AddTask(t)
		}

		//{{ if HasTag "_master_lb" }}
		//# Attach ASG to ELB
		//loadBalancerAttachment/masters.{{ $m.Name }}.{{ SafeClusterName }}:
		//loadBalancer: loadBalancer/api.{{ ClusterName }}
		//autoscalingGroup: autoscalingGroup/{{ $m.Name }}.{{ ClusterName }}
		//{{ end }}

	}

	return nil
}

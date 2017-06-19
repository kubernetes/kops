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

package gcemodel

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

const (
	DefaultVolumeSize = 100
	DefaultVolumeType = "pd-standard"
)

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

		startupScript, err := b.BootstrapScript.ResourceNodeUp(ig, b.Cluster.Spec.EgressProxy)
		if err != nil {
			return err
		}

		// InstanceTemplate
		var instanceTemplate *gcetasks.InstanceTemplate
		{
			volumeSize := fi.Int32Value(ig.Spec.RootVolumeSize)
			if volumeSize == 0 {
				volumeSize = DefaultVolumeSize
			}
			volumeType := fi.StringValue(ig.Spec.RootVolumeType)
			if volumeType == "" {
				volumeType = DefaultVolumeType
			}

			t := &gcetasks.InstanceTemplate{
				Name:           s(name),
				Lifecycle:      b.Lifecycle,
				Network:        b.LinkToNetwork(),
				MachineType:    s(ig.Spec.MachineType),
				BootDiskType:   s(volumeType),
				BootDiskSizeGB: i64(int64(volumeSize)),
				BootDiskImage:  s(ig.Spec.Image),

				CanIPForward: fi.Bool(true),

				// TODO: Support preemptible nodes?
				Preemptible: fi.Bool(false),

				Scopes: []string{
					"compute-rw",
					"monitoring",
					"logging-write",
					"storage-ro",
				},

				Metadata: map[string]*fi.ResourceHolder{
					"startup-script": startupScript,
					//"config": resources/config.yaml $nodeset.Name
					"cluster-name": fi.WrapResource(fi.NewStringResource(b.ClusterName())),
				},
			}

			switch ig.Spec.Role {
			case kops.InstanceGroupRoleMaster:
				// Grant DNS permissions
				t.Scopes = append(t.Scopes, "https://www.googleapis.com/auth/ndev.clouddns.readwrite")
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleMaster))

			case kops.InstanceGroupRoleNode:
				t.Tags = append(t.Tags, b.GCETagForRole(kops.InstanceGroupRoleNode))
			}

			//labels, err := b.CloudTagsForInstanceGroup(ig)
			//if err != nil {
			//	return fmt.Errorf("error building cloud tags: %v", err)
			//}
			//t.Labels = labels

			c.AddTask(t)

			instanceTemplate = t
		}

		// AutoscalingGroup
		zones := sets.NewString()
		for _, subnetName := range ig.Spec.Subnets {
			subnet := b.FindSubnet(subnetName)
			if subnet == nil {
				return fmt.Errorf("subnet %q not found", subnetName)
			}
			if subnet.Zone == "" {
				return fmt.Errorf("subnet %q has not Zone", subnetName)
			}
			zones.Insert(subnet.Zone)
		}

		zoneList := zones.List()
		targetSizes := make([]int, len(zoneList), len(zoneList))
		totalSize := 0

		// TODO: Duplicated from aws - move to defaults?
		minSize := 1
		if ig.Spec.MinSize != nil {
			minSize = int(fi.Int32Value(ig.Spec.MinSize))
		} else if ig.Spec.Role == kops.InstanceGroupRoleNode {
			minSize = 2
		}

		for i := range zoneList {
			targetSizes[i] = minSize / zones.Len()
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
			zone := zoneList[i]

			// TODO: Switch to regional managed instance group

			name := b.SafeObjectName(zone + "." + ig.ObjectMeta.Name)

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

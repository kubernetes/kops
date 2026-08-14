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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

// InstanceModelBuilder configures the Akamai (Linode) instances (aka Linodes) for the cluster.
type InstanceModelBuilder struct {
	*LinodeModelContext
	Lifecycle              fi.Lifecycle
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
}

var _ fi.CloudupModelBuilder = &InstanceModelBuilder{}

func (b *InstanceModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	for _, ig := range b.InstanceGroups {
		var sshKeyTasks []*linodetasks.SSHKey

		subnets, err := b.GatherSubnets(ig)
		if err != nil {
			return err
		}
		if len(subnets) != 1 {
			return fmt.Errorf("expected exactly one subnet for InstanceGroup %q; subnets was %s", ig.Name, ig.Spec.Subnets)
		}
		subnetSpec := subnets[0]
		subnetTaskName := linode.NormalizeLinodeLabel(b.ClusterName() + "-" + subnetSpec.Name)
		subnetTask, err := findSubnetTask(c, subnetTaskName, ig)
		if err != nil {
			return err
		}

		for _, task := range c.Tasks {
			if sshKey, ok := task.(*linodetasks.SSHKey); ok {
				sshKeyTasks = append(sshKeyTasks, sshKey)
			}
		}

		userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return err
		}

		tagsMap, err := b.CloudTagsForInstanceGroup(ig)
		if err != nil {
			return err
		}
		tags := make([]string, 0, len(tagsMap))
		for k, v := range tagsMap {
			tags = append(tags, fmt.Sprintf("%s:%s", k, v))
		}

		instanceGroup := linodetasks.Instance{
			Name:                   new(ig.Name),
			Lifecycle:              b.Lifecycle,
			Region:                 subnetSpec.Region,
			Type:                   ig.Spec.MachineType,
			Subnet:                 subnetTask,
			RequirePublicInterface: requirePublicInterface(subnetSpec, ig),
			AuthorizedKeys:         sshKeyTasks,
			Count:                  int(fi.ValueOf(ig.Spec.MinSize)),
			Image:                  ig.Spec.Image,
			UserData:               userData,
			Tags:                   tags,
		}

		c.AddTask(&instanceGroup)
	}

	return nil
}

// requirePublicInterface checks whether the instance group requires a public interface based on the subnet type and instance group settings.
func requirePublicInterface(subnet *kops.ClusterSubnetSpec, ig *kops.InstanceGroup) *bool {
	requirePublic := false

	switch subnet.Type {
	case kops.SubnetTypePublic, kops.SubnetTypeUtility:
		requirePublic = true
		if ig.Spec.AssociatePublicIP != nil {
			requirePublic = fi.ValueOf(ig.Spec.AssociatePublicIP)
		}
	case kops.SubnetTypePrivate, kops.SubnetTypeDualStack:
		requirePublic = false
	}

	return &requirePublic
}

// findInstanceTask searches for the instance task corresponding to the given instance group name in the provided context.
func findSubnetTask(c *fi.CloudupModelBuilderContext, subnetTaskName string, ig *kops.InstanceGroup) (*linodetasks.Subnet, error) {
	for _, task := range c.Tasks {
		subnet, ok := task.(*linodetasks.Subnet)
		if !ok {
			continue
		}
		if fi.ValueOf(subnet.Name) == subnetTaskName {
			return subnet, nil
		}
	}

	return nil, fmt.Errorf("subnet task %q not found for InstanceGroup %q", subnetTaskName, ig.Name)
}

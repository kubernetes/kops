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
	"slices"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
	"k8s.io/kops/upup/pkg/fi/cloudup/linodetasks"
)

// InstanceModelBuilder configures Linode (Akamai) instances for each instance group.
type InstanceModelBuilder struct {
	*LinodeModelContext

	BootstrapScriptBuilder *model.BootstrapScriptBuilder
	Lifecycle              fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &InstanceModelBuilder{}

func (b *InstanceModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if b.BootstrapScriptBuilder == nil {
		return fmt.Errorf("bootstrap script builder is required")
	}

	if len(b.SSHPublicKeys) == 0 {
		return fmt.Errorf("SSH public key is required for Linode (Akamai) instances")
	}
	sshPublicKeyResource := fi.Resource(fi.NewBytesResource(b.SSHPublicKeys[0]))

	for _, ig := range b.InstanceGroups {
		subnets, err := b.GatherSubnets(ig)
		if err != nil {
			return fmt.Errorf("error building Linode (Akamai) instance task for %q: %w", ig.Name, err)
		}
		if len(subnets) == 0 || subnets[0].Region == "" {
			return fmt.Errorf("error building Linode (Akamai) instance task for %q: subnet region is required", ig.Name)
		}

		userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return fmt.Errorf("error building bootstrap script for %q: %w", ig.Name, err)
		}

		name := b.AutoscalingGroupName(ig)

		cloudTags, err := b.CloudTagsForInstanceGroup(ig)
		if err != nil {
			return fmt.Errorf("error building cloud tags for %q: %w", ig.Name, err)
		}

		tagKeys := make([]string, 0, len(cloudTags))
		for k := range cloudTags {
			tagKeys = append(tagKeys, k)
		}
		slices.Sort(tagKeys)

		tags := make([]string, 0, len(cloudTags)+3)
		tags = append(tags, linode.BuildLinodeTag(kops.LabelClusterName, b.ClusterName()))
		tags = append(tags, linode.BuildLinodeTag(linode.TagKubernetesInstanceGroup, ig.Name))
		tags = append(tags, linode.BuildLinodeTag(linode.TagKubernetesInstanceRole, string(ig.Spec.Role)))
		for _, k := range tagKeys {
			tags = append(tags, linode.BuildLinodeTag(k, cloudTags[k]))
		}

		t := &linodetasks.Instance{
			Name:          fi.PtrTo(name),
			Lifecycle:     b.Lifecycle,
			Region:        fi.PtrTo(subnets[0].Region),
			Type:          fi.PtrTo(ig.Spec.MachineType),
			Image:         fi.PtrTo(ig.Spec.Image),
			Count:         int(fi.ValueOf(ig.Spec.MinSize)),
			Tags:          tags,
			AuthorizedKey: &sshPublicKeyResource,
			UserData:      &userData,
		}
		c.AddTask(t)
	}

	return nil
}

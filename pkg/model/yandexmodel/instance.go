/*
Copyright 2022 The Kubernetes Authors.

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

package yandexmodel

import (
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandextasks"
)

// InstanceModelBuilder configures network objects
type InstanceModelBuilder struct {
	*YandexModelContext
	Lifecycle              fi.Lifecycle
	BootstrapScriptBuilder *model.BootstrapScriptBuilder
}

var _ fi.ModelBuilder = &InstanceModelBuilder{}

func (b *InstanceModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// TODO(YuraBeznos): serviceAccountId already defined in cloudup/yandex/cloud.go but not sure how to get it from there
	// TODO(YuraBeznos): find a better way to get serviceAccountId
	serviceAccountId := b.Cluster.Spec.CloudConfig.GCEServiceAccount
	for _, ig := range b.InstanceGroups {
		// TODO(YuraBeznos): dirty hack with only one subnet, must be rewritten
		subnets := b.Cluster.Spec.Subnets
		var subnet *yandextasks.Subnet
		for _, s := range subnets {
			if ig.Spec.Subnets[0] == s.Name {
				subnet = b.LinkToSubnet(&s)
				break
			}
		}

		userData, err := b.BootstrapScriptBuilder.ResourceNodeUp(c, ig)
		if err != nil {
			return err
		}
		instance := &yandextasks.Instance{
			FolderId:         b.Cluster.Spec.Project,
			SSHPublicKeys:    b.SSHPublicKeys,
			ServiceAccountId: serviceAccountId,
			Name:             fi.String(ig.Name),
			Lifecycle:        b.Lifecycle,
			ZoneId:           ig.Spec.Zones[0], // TODO(YuraBeznos): should be a better way to define a zone
			UserData:         userData,
			PlatformId:       "standard-v1",
			Subnet:           subnet,
			Description:      fi.String(b.ClusterName() + " " + ig.Name),
			Labels: map[string]string{
				yandex.TagKubernetesClusterName:   b.ClusterName(),
				yandex.TagKubernetesInstanceGroup: ig.Name,
			},
		}

		c.AddTask(instance)
	}

	return nil
}

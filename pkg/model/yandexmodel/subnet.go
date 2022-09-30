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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandextasks"
)

// SubnetModelBuilder configures network objects
type SubnetModelBuilder struct {
	*YandexModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &SubnetModelBuilder{}

func (b *SubnetModelBuilder) Build(c *fi.ModelBuilderContext) error {
	for i := range b.Cluster.Spec.Subnets {
		subnet := &b.Cluster.Spec.Subnets[i]

		network, err := b.LinkToNetwork()
		if err != nil {
			return nil
		}

		t := &yandextasks.Subnet{
			Name:        b.LinkToSubnet(subnet).Name,
			Lifecycle:   b.Lifecycle,
			Network:     network, //networkId
			FolderId:    b.Cluster.Spec.Project,
			Description: fi.String(b.ClusterName()), //b.Description, TODO(YuraBeznos): make Description configurable
			ZoneId:      subnet.Zone,
			//subnet.Labels = map[string]string{
			//	yandex.TagKubernetesClusterName: b.ClusterName(),
		}
		if subnet.CIDR != "" {
			t.V4CidrBlocks = []string{subnet.CIDR}
		}
		c.AddTask(t)
	}
	return nil
}

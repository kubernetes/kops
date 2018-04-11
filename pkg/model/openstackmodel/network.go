/*
Copyright 2018 The Kubernetes Authors.

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

package openstackmodel

import (
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*OpenstackModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	clusterName := b.ClusterName()

	{
		t := &openstacktasks.Network{
			Name:      s(clusterName),
			ID:        s(b.Cluster.Spec.NetworkID),
			Lifecycle: b.Lifecycle,
		}

		c.AddTask(t)
	}

	{
		t := &openstacktasks.Router{
			Name:      s(strings.Replace(clusterName, ".", "-", -1)),
			Lifecycle: b.Lifecycle,
		}

		c.AddTask(t)
	}

	for _, sp := range b.Cluster.Spec.Subnets {
		t := &openstacktasks.Subnet{
			Name:      s(sp.Name),
			Network:   b.LinkToNetwork(),
			CIDR:      s(sp.CIDR),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	return nil
}

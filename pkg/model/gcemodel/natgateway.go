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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// NATGatewayModeBuilder adds model objects to support Cloud NAT Gateway
//
// Cloud NAT provides internet access for instances without external IPs.

type NatGatewayModelBuilder struct {
	*GCEModelContext
	Lifecycle         *fi.Lifecycle
	SecurityLifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &NatGatewayModelBuilder{}

func (b *NatGatewayModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// TODO: get region from cluster info
	region := "us-east1"

	// Do not configure Cloud NAT for public clusters
	if b.Cluster.Spec.Topology.Masters != "private" {
		return nil
	}

	{
		t := &gcetasks.Router{
			Name:      s(b.SafeObjectName(region + "-" + b.ClusterName())),
			Lifecycle: b.Lifecycle,
			Network:   s(b.LinkToNetwork().URL(b.Cluster.Spec.Project)),
			Region:    s(region),
		}
		c.AddTask(t)
	}

	return nil
}

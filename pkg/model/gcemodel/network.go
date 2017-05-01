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

// NetworkModelBuilder configures network objects
type NetworkModelBuilder struct {
	*GCEModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &NetworkModelBuilder{}

func (b *NetworkModelBuilder) Build(c *fi.ModelBuilderContext) error {
	network := &gcetasks.Network{
		Name:      s("default"),
		Lifecycle: b.Lifecycle,
		CIDR:      s(b.Cluster.Spec.NetworkCIDR), // Default: s("10.240.0.0/16"),
	}
	c.AddTask(network)

	return nil
}

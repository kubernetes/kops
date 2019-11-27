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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

type GCEModelContext struct {
	*model.KopsModelContext
}

// LinkToNetwork returns the GCE Network object the cluster is located in
func (c *GCEModelContext) LinkToNetwork() *gcetasks.Network {
	return &gcetasks.Network{Name: s(c.NameForNetwork())}
}

// NameForNetwork returns the name for the GCE Network the cluster is located in
func (c *GCEModelContext) NameForNetwork() string {
	networkName := c.Cluster.Spec.NetworkID
	if networkName == "" {
		networkName = "default"
	}
	return networkName
}

// NameForIPAliasRange returns the name for the secondary IP range attached to a subnet
func (c *GCEModelContext) NameForIPAliasRange(key string) string {
	// We include the cluster name so we could share a subnet...
	// but there's a 5 IP alias range limit per subnet anwyay, so
	// this is rather pointless and in practice we just use a
	// separate subnet per cluster
	return c.SafeObjectName(key)
}

// NameForIPAliasSubnet returns the name for the GCE subnet used for ip aliases
func (c *GCEModelContext) NameForIPAliasSubnet() string {
	return c.SafeObjectName("default")
}

// LinkToIPAliasSubnet returns the GCE subnet object used for ip aliases
func (c *GCEModelContext) LinkToIPAliasSubnet() *gcetasks.Subnet {
	return &gcetasks.Subnet{Name: s(c.NameForIPAliasSubnet())}
}

// SafeObjectName returns the object name and cluster name escaped for GCE
func (c *GCEModelContext) SafeObjectName(name string) string {
	return gce.SafeObjectName(name, c.Cluster.ObjectMeta.Name)
}

func (c *GCEModelContext) GCETagForRole(role kops.InstanceGroupRole) string {
	return components.GCETagForRole(c.Cluster.ObjectMeta.Name, role)
}

func (c *GCEModelContext) LinkToTargetPool(id string) *gcetasks.TargetPool {
	return &gcetasks.TargetPool{Name: s(c.NameForTargetPool(id))}
}

func (c *GCEModelContext) NameForTargetPool(id string) string {
	return c.SafeObjectName(id)
}

func (c *GCEModelContext) NameForForwardingRule(id string) string {
	return c.SafeObjectName(id)
}

func (c *GCEModelContext) NameForIPAddress(id string) string {
	return c.SafeObjectName(id)
}

func (c *GCEModelContext) NameForFirewallRule(id string) string {
	return c.SafeObjectName(id)
}

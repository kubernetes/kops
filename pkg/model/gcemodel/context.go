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
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

type GCEModelContext struct {
	ProjectID string

	*model.KopsModelContext
}

// LinkToNetwork returns the GCE Network object the cluster is located in
func (c *GCEModelContext) LinkToNetwork() (*gcetasks.Network, error) {
	if c.Cluster.Spec.Networking.NetworkID == "" {
		return &gcetasks.Network{Name: s(c.SafeTruncatedClusterName())}, nil
	}
	name, project, err := gce.ParseNameAndProjectFromNetworkID(c.Cluster.Spec.Networking.NetworkID)
	if err != nil {
		return nil, err
	}

	network := &gcetasks.Network{Name: s(name)}
	if project != "" {
		network.Project = &project
	}
	return network, nil
}

// NameForIPAliasRange returns the name for the secondary IP range attached to a subnet
func (c *GCEModelContext) NameForIPAliasRange(key string) string {
	// We include the cluster name so we could share a subnet...
	// but there's a 5 IP alias range limit per subnet anwyay, so
	// this is rather pointless and in practice we just use a
	// separate subnet per cluster
	return c.SafeSuffixedObjectName(key)
}

// LinkToSubnet returns a link to the GCE subnet object
func (c *GCEModelContext) LinkToSubnet(subnet *kops.ClusterSubnetSpec) *gcetasks.Subnet {
	name := subnet.ID
	if name == "" {
		name = gce.ClusterSuffixedName(subnet.Name, c.Cluster.ObjectMeta.Name, 63)
	}

	return &gcetasks.Subnet{Name: s(name)}
}

// SafeObjectName returns the object name and cluster name escaped for GCE
func (c *GCEModelContext) SafeObjectName(name string) string {
	return gce.SafeObjectName(name, c.Cluster.ObjectMeta.Name)
}

// SafeSuffixedObjectName returns the object name and cluster name escaped for GCE, limited to 63 chars
func (c *GCEModelContext) SafeSuffixedObjectName(name string) string {
	return gce.ClusterSuffixedName(name, c.Cluster.ObjectMeta.Name, 63)
}

// SafeClusterName returns the cluster name escaped for use as a GCE resource name
func (c *GCEModelContext) SafeClusterName() string {
	return gce.SafeClusterName(c.Cluster.ObjectMeta.Name)
}

// SafeTruncatedClusterName returns the cluster name escaped and truncated for use as a GCE resource name
func (c *GCEModelContext) SafeTruncatedClusterName() string {
	return gce.SafeTruncatedClusterName(c.Cluster.ObjectMeta.Name, 63)
}

// GCETagForRole returns the (network) tag for GCE instances in the given instance group role.
func (c *GCEModelContext) GCETagForRole(role kops.InstanceGroupRole) string {
	return gce.TagForRole(c.Cluster.ObjectMeta.Name, role)
}

func (c *GCEModelContext) LinkToTargetPool(id string) *gcetasks.TargetPool {
	return &gcetasks.TargetPool{Name: s(c.NameForTargetPool(id))}
}

func (c *GCEModelContext) NameForTargetPool(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForHealthCheck(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForBackendService(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForForwardingRule(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForIPAddress(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForPoolHealthcheck(id string) string {
	return c.SafeObjectName(id)
}

func (c *GCEModelContext) NameForHealthcheck(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForRouter(id string) string {
	return c.SafeSuffixedObjectName(id)
}

func (c *GCEModelContext) NameForFirewallRule(id string) string {
	return gce.ClusterSuffixedName(id, c.Cluster.ObjectMeta.Name, 63)
}

func (c *GCEModelContext) NetworkingIsIPAlias() bool {
	return c.Cluster.Spec.Networking.GCP != nil
}

func (c *GCEModelContext) NetworkingIsGCERoutes() bool {
	return c.Cluster.Spec.Networking.Kubenet != nil
}

// LinkToServiceAccount returns a link to the GCE ServiceAccount object for VMs in the given role
func (c *GCEModelContext) LinkToServiceAccount(ig *kops.InstanceGroup) *gcetasks.ServiceAccount {
	if c.Cluster.Spec.CloudProvider.GCE.ServiceAccount != "" {
		// This is a legacy setting because the nodes & control-plane run under the same serviceaccount
		klog.Warningf("using legacy spec.cloudProvider.gce.serviceAccount=%q setting", c.Cluster.Spec.CloudProvider.GCE.ServiceAccount)
		return &gcetasks.ServiceAccount{
			Name:   s("shared"),
			Email:  &c.Cluster.Spec.CloudProvider.GCE.ServiceAccount,
			Shared: fi.PtrTo(true),
		}
	}

	role := ig.Spec.Role

	name := ""
	switch role {
	case kops.InstanceGroupRoleAPIServer, kops.InstanceGroupRoleControlPlane:
		name = gce.ControlPlane

	case kops.InstanceGroupRoleBastion:
		name = gce.Bastion

	case kops.InstanceGroupRoleNode:
		name = gce.Node

	default:
		klog.Fatalf("unknown role %q", role)
	}

	accountID := gce.ServiceAccountName(name, c.ClusterName())
	projectID := c.ProjectID

	email := accountID + "@" + projectID + ".iam.gserviceaccount.com"

	return &gcetasks.ServiceAccount{Name: s(name), Email: s(email)}
}

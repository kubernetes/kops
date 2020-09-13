/*
Copyright 2020 The Kubernetes Authors.

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

package testutils

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// BuildMinimalCluster a generic minimal cluster
func BuildMinimalCluster(clusterName string) *kops.Cluster {
	c := &kops.Cluster{}
	c.ObjectMeta.Name = clusterName
	c.Spec.KubernetesVersion = "1.14.6"
	c.Spec.Subnets = []kops.ClusterSubnetSpec{
		{Name: "subnet-us-mock-1a", Zone: "us-mock-1a", CIDR: "172.20.1.0/24", Type: kops.SubnetTypePrivate},
	}

	c.Spec.MasterPublicName = fmt.Sprintf("api.%v", clusterName)
	c.Spec.MasterInternalName = fmt.Sprintf("internal.api.%v", clusterName)
	c.Spec.KubernetesAPIAccess = []string{"0.0.0.0/0"}
	c.Spec.SSHAccess = []string{"0.0.0.0/0"}

	// Default to public topology
	c.Spec.Topology = &kops.TopologySpec{
		Masters: kops.TopologyPublic,
		Nodes:   kops.TopologyPublic,
		DNS: &kops.DNSSpec{
			Type: kops.DNSTypePublic,
		},
	}

	c.Spec.IAM = &kops.IAMSpec{}

	c.Spec.Networking = &kops.NetworkingSpec{}

	c.Spec.NetworkCIDR = "172.20.0.0/16"
	c.Spec.Subnets = []kops.ClusterSubnetSpec{
		{Name: "subnet-us-mock-1a", Zone: "us-mock-1a", CIDR: "172.20.1.0/24"},
		{Name: "subnet-us-mock-1b", Zone: "us-mock-1b", CIDR: "172.20.2.0/24"},
		{Name: "subnet-us-mock-1c", Zone: "us-mock-1c", CIDR: "172.20.3.0/24"},
	}

	c.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Spec.CloudProvider = "aws"

	c.Spec.ConfigBase = "memfs://unittest-bucket/" + clusterName

	c.Spec.DNSZone = "test.com"

	c.Spec.SSHKeyName = fi.String("test")

	addEtcdClusters(c)

	return c
}

func addEtcdClusters(c *kops.Cluster) {
	subnetNames := sets.NewString()
	for _, z := range c.Spec.Subnets {
		subnetNames.Insert(z.Name)
	}
	etcdZones := subnetNames.List()

	for _, etcdCluster := range []string{"main", "events"} {
		etcd := kops.EtcdClusterSpec{}
		etcd.Name = etcdCluster
		for _, zone := range etcdZones {
			m := kops.EtcdMemberSpec{}
			m.Name = zone
			m.InstanceGroup = fi.String("master-" + zone)
			etcd.Members = append(etcd.Members, m)
		}
		c.Spec.EtcdClusters = append(c.Spec.EtcdClusters, etcd)
	}
}

func BuildMinimalNodeInstanceGroup(name string, subnets ...string) kops.InstanceGroup {
	g := kops.InstanceGroup{}
	g.ObjectMeta.Name = name
	g.Spec.Role = kops.InstanceGroupRoleNode
	g.Spec.Subnets = subnets

	return g
}

func BuildMinimalBastionInstanceGroup(name string, subnets ...string) kops.InstanceGroup {
	g := kops.InstanceGroup{}
	g.ObjectMeta.Name = name
	g.Spec.Role = kops.InstanceGroupRoleNode
	g.Spec.Subnets = subnets

	return g
}

func BuildMinimalMasterInstanceGroup(subnet string) kops.InstanceGroup {
	g := kops.InstanceGroup{}
	g.ObjectMeta.Name = "master-" + subnet
	g.Spec.Role = kops.InstanceGroupRoleMaster
	g.Spec.Subnets = []string{subnet}

	return g
}

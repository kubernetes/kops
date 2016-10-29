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

package cloudup

import (
	"fmt"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/pkg/util/sets"
	"strings"
	"testing"
)

func buildMinimalCluster() *api.Cluster {
	c := &api.Cluster{}
	c.Name = "testcluster.test.com"
	c.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-mock-1a", CIDR: "172.20.1.0/24"},
		{Name: "us-mock-1b", CIDR: "172.20.2.0/24"},
		{Name: "us-mock-1c", CIDR: "172.20.3.0/24"},
	}
	// Default to public topology
	c.Spec.Topology = &api.TopologySpec{
		Masters: api.TopologyPublic,
		Nodes: api.TopologyPublic,
	}
	c.Spec.NetworkCIDR = "172.20.0.0/16"
	c.Spec.NonMasqueradeCIDR = "100.64.0.0/10"
	c.Spec.CloudProvider = "aws"

	c.Spec.ConfigBase = "s3://unittest-bucket/"

	// Required to stop a call to cloud provider
	// TODO: Mock cloudprovider
	c.Spec.DNSZone = "test.com"

	return c
}

func addEtcdClusters(c *api.Cluster) {
	zones := sets.NewString()
	for _, z := range c.Spec.Zones {
		zones.Insert(z.Name)
	}
	etcdZones := zones.List()

	for _, etcdCluster := range EtcdClusters {
		etcd := &api.EtcdClusterSpec{}
		etcd.Name = etcdCluster
		for _, zone := range etcdZones {
			m := &api.EtcdMemberSpec{}
			m.Name = zone
			m.Zone = fi.String(zone)
			etcd.Members = append(etcd.Members, m)
		}
		c.Spec.EtcdClusters = append(c.Spec.EtcdClusters, etcd)
	}
}

func TestPopulateCluster_Default_NoError(t *testing.T) {
	c := buildMinimalCluster()

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	addEtcdClusters(c)

	_, err = PopulateClusterSpec(c)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}
}

func TestPopulateCluster_Docker_Spec(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.Docker = &api.DockerConfig{
		MTU:              fi.Int(5678),
		InsecureRegistry: fi.String("myregistry.com:1234"),
	}

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	addEtcdClusters(c)

	full, err := PopulateClusterSpec(c)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}

	if fi.IntValue(full.Spec.Docker.MTU) != 5678 {
		t.Fatalf("Unexpected Docker MTU: %v", full.Spec.Docker.MTU)
	}

	if fi.StringValue(full.Spec.Docker.InsecureRegistry) != "myregistry.com:1234" {
		t.Fatalf("Unexpected Docker InsecureRegistry: %v", full.Spec.Docker.InsecureRegistry)
	}
}

func build(c *api.Cluster) (*api.Cluster, error) {
	err := c.PerformAssignments()
	if err != nil {
		return nil, fmt.Errorf("error from PerformAssignments: %v", err)
	}

	addEtcdClusters(c)
	full, err := PopulateClusterSpec(c)
	if err != nil {
		return nil, fmt.Errorf("Unexpected error from PopulateCluster: %v", err)
	}
	return full, nil
}

func TestPopulateCluster_Kubenet(t *testing.T) {
	c := buildMinimalCluster()

	full, err := build(c)
	if err != nil {
		t.Fatalf("error during build: %v", err)
	}

	if full.Spec.Kubelet.NetworkPluginName != "kubenet" {
		t.Fatalf("Unexpected NetworkPluginName: %v", full.Spec.Kubelet.NetworkPluginName)
	}

	if fi.BoolValue(full.Spec.Kubelet.ReconcileCIDR) != true {
		t.Fatalf("Unexpected ReconcileCIDR: %v", full.Spec.Kubelet.ReconcileCIDR)
	}

	if fi.BoolValue(full.Spec.KubeControllerManager.ConfigureCloudRoutes) != true {
		t.Fatalf("Unexpected ConfigureCloudRoutes: %v", full.Spec.KubeControllerManager.ConfigureCloudRoutes)
	}
}

func TestPopulateCluster_CNI(t *testing.T) {
	c := buildMinimalCluster()

	c.Spec.Kubelet = &api.KubeletConfigSpec{
		ConfigureCBR0:     fi.Bool(false),
		NetworkPluginName: "cni",
		NonMasqueradeCIDR: c.Spec.NonMasqueradeCIDR,
		CloudProvider:     c.Spec.CloudProvider,
	}

	full, err := build(c)
	if err != nil {
		t.Fatalf("error during build: %v", err)
	}

	if full.Spec.Kubelet.NetworkPluginName != "cni" {
		t.Fatalf("Unexpected NetworkPluginName: %v", full.Spec.Kubelet.NetworkPluginName)
	}

	if fi.BoolValue(full.Spec.Kubelet.ReconcileCIDR) != true {
		t.Fatalf("Unexpected ReconcileCIDR: %v", full.Spec.Kubelet.ReconcileCIDR)
	}

	if fi.BoolValue(full.Spec.Kubelet.ConfigureCBR0) != false {
		t.Fatalf("Unexpected ConfigureCBR0: %v", full.Spec.Kubelet.ConfigureCBR0)
	}

	if fi.BoolValue(full.Spec.KubeControllerManager.ConfigureCloudRoutes) != true {
		t.Fatalf("Unexpected ConfigureCloudRoutes: %v", full.Spec.KubeControllerManager.ConfigureCloudRoutes)
	}
}

func TestPopulateCluster_Custom_CIDR(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.NetworkCIDR = "172.20.2.0/24"
	c.Spec.Zones = []*api.ClusterZoneSpec{
		{Name: "us-mock-1a", CIDR: "172.20.2.0/27"},
		{Name: "us-mock-1b", CIDR: "172.20.2.32/27"},
		{Name: "us-mock-1c", CIDR: "172.20.2.64/27"},
	}

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	addEtcdClusters(c)

	full, err := PopulateClusterSpec(c)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}
	if full.Spec.NetworkCIDR != "172.20.2.0/24" {
		t.Fatalf("Unexpected NetworkCIDR: %v", full.Spec.NetworkCIDR)
	}
}

func TestPopulateCluster_IsolateMasters(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.IsolateMasters = fi.Bool(true)

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	addEtcdClusters(c)

	full, err := PopulateClusterSpec(c)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}
	if fi.BoolValue(full.Spec.MasterKubelet.EnableDebuggingHandlers) != false {
		t.Fatalf("Unexpected EnableDebuggingHandlers: %v", fi.BoolValue(full.Spec.MasterKubelet.EnableDebuggingHandlers))
	}
	if fi.BoolValue(full.Spec.MasterKubelet.ReconcileCIDR) != false {
		t.Fatalf("Unexpected ReconcileCIDR: %v", fi.BoolValue(full.Spec.MasterKubelet.ReconcileCIDR))
	}
}

func TestPopulateCluster_IsolateMastersFalse(t *testing.T) {
	c := buildMinimalCluster()
	// default: c.Spec.IsolateMasters = fi.Bool(false)

	err := c.PerformAssignments()
	if err != nil {
		t.Fatalf("error from PerformAssignments: %v", err)
	}

	addEtcdClusters(c)

	full, err := PopulateClusterSpec(c)
	if err != nil {
		t.Fatalf("Unexpected error from PopulateCluster: %v", err)
	}
	if fi.BoolValue(full.Spec.MasterKubelet.EnableDebuggingHandlers) != true {
		t.Fatalf("Unexpected EnableDebuggingHandlers: %v", fi.BoolValue(full.Spec.MasterKubelet.EnableDebuggingHandlers))
	}
	if fi.BoolValue(full.Spec.MasterKubelet.ReconcileCIDR) != true {
		t.Fatalf("Unexpected ReconcileCIDR: %v", fi.BoolValue(full.Spec.MasterKubelet.ReconcileCIDR))
	}
}

func TestPopulateCluster_Name_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Name = ""

	expectErrorFromPopulateCluster(t, c, "Name")
}

func TestPopulateCluster_Zone_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.Zones = nil

	expectErrorFromPopulateCluster(t, c, "Zone")
}

func TestPopulateCluster_NetworkCIDR_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.NetworkCIDR = ""

	expectErrorFromPopulateCluster(t, c, "NetworkCIDR")
}

func TestPopulateCluster_NonMasqueradeCIDR_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.NonMasqueradeCIDR = ""

	expectErrorFromPopulateCluster(t, c, "NonMasqueradeCIDR")
}

func TestPopulateCluster_CloudProvider_Required(t *testing.T) {
	c := buildMinimalCluster()
	c.Spec.CloudProvider = ""

	expectErrorFromPopulateCluster(t, c, "CloudProvider")
}

func expectErrorFromPopulateCluster(t *testing.T, c *api.Cluster, message string) {
	_, err := PopulateClusterSpec(c)
	if err == nil {
		t.Fatalf("Expected error from PopulateCluster")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

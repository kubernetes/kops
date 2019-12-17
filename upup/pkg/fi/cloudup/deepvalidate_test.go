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

package cloudup

import (
	"fmt"
	"strings"
	"testing"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/upup/pkg/fi"
)

func TestDeepValidate_OK(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1a"))
	err := validation.DeepValidate(c, groups, true)
	if err != nil {
		t.Fatalf("Expected no error from DeepValidate, got %v", err)
	}
}

func TestDeepValidate_NoNodeZones(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1a"))
	expectErrorFromDeepValidate(t, c, groups, "must configure at least one Node InstanceGroup")
}

func TestDeepValidate_NoMasterZones(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1a"))
	expectErrorFromDeepValidate(t, c, groups, "must configure at least one Master InstanceGroup")
}

func TestDeepValidate_BadZone(t *testing.T) {
	t.Skipf("Zone validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Subnets = []api.ClusterSubnetSpec{
		{Name: "subnet-badzone", Zone: "us-mock-1z", CIDR: "172.20.1.0/24"},
	}
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1z"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1z"))
	expectErrorFromDeepValidate(t, c, groups, "Zone is not a recognized AZ")
}

func TestDeepValidate_MixedRegion(t *testing.T) {
	t.Skipf("Region validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Subnets = []api.ClusterSubnetSpec{
		{Name: "mock1a", Zone: "us-mock-1a", CIDR: "172.20.1.0/24"},
		{Name: "west1b", Zone: "us-west-1b", CIDR: "172.20.2.0/24"},
	}
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1a", "subnet-us-west-1b"))

	expectErrorFromDeepValidate(t, c, groups, "Clusters cannot span multiple regions")
}

func TestDeepValidate_RegionAsZone(t *testing.T) {
	t.Skipf("Region validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Subnets = []api.ClusterSubnetSpec{
		{Name: "mock1", Zone: "us-mock-1", CIDR: "172.20.1.0/24"},
	}
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1"))

	expectErrorFromDeepValidate(t, c, groups, "Region is not a recognized EC2 region: \"us-east-\" (check you have specified valid zones?)")
}

func TestDeepValidate_NotIncludedZone(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1d"))

	expectErrorFromDeepValidate(t, c, groups, "not configured as a Subnet in the cluster")
}

func TestDeepValidate_DuplicateZones(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.Subnets = []api.ClusterSubnetSpec{
		{Name: "dup1", Zone: "us-mock-1a", CIDR: "172.20.1.0/24"},
		{Name: "dup1", Zone: "us-mock-1a", CIDR: "172.20.2.0/24"},
	}
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("dup1"))
	groups = append(groups, buildMinimalNodeInstanceGroup("dup1"))
	expectErrorFromDeepValidate(t, c, groups, "subnets with duplicate name \"dup1\" found")
}

func TestDeepValidate_ExtraMasterZone(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.Subnets = []api.ClusterSubnetSpec{
		{Name: "mock1a", Zone: "us-mock-1a", CIDR: "172.20.1.0/24"},
		{Name: "mock1b", Zone: "us-mock-1b", CIDR: "172.20.2.0/24"},
	}
	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1a", "subnet-us-mock-1b", "subnet-us-mock-1c"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1a", "subnet-us-mock-1b"))

	expectErrorFromDeepValidate(t, c, groups, "is not configured as a Subnet in the cluster")
}

func TestDeepValidate_EvenEtcdClusterSize(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.EtcdClusters = []*api.EtcdClusterSpec{
		{
			Name: "main",
			Members: []*api.EtcdMemberSpec{
				{Name: "us-mock-1a", InstanceGroup: fi.String("us-mock-1a")},
				{Name: "us-mock-1b", InstanceGroup: fi.String("us-mock-1b")},
			},
		},
	}

	var groups []*api.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-mock-1a", "subnet-us-mock-1b", "subnet-us-mock-1c", "subnet-us-mock-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-mock-1a"))

	expectErrorFromDeepValidate(t, c, groups, "Should be an odd number of master-zones for quorum. Use --zones and --master-zones to declare node zones and master zones separately")
}

func expectErrorFromDeepValidate(t *testing.T, c *api.Cluster, groups []*api.InstanceGroup, message string) {
	err := validation.DeepValidate(c, groups, true)
	if err == nil {
		t.Fatalf("Expected error %q from DeepValidate (strict=true), not no error raised", message)
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

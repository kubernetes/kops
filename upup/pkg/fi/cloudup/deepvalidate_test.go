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

	kopsapi "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

func TestDeepValidate_OK(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	for _, subnet := range c.Spec.Networking.Subnets {
		groups = append(groups, buildMinimalMasterInstanceGroup(subnet.Name))
		groups = append(groups, buildMinimalNodeInstanceGroup(subnet.Name))
	}
	err := validation.DeepValidate(c, groups, true, vfs.Context, nil)
	if err != nil {
		t.Fatalf("Expected no error from DeepValidate, got %v", err)
	}
}

func TestDeepValidate_NoNodeZones(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	err := validation.DeepValidate(c, groups, true, vfs.Context, nil)
	if err != nil {
		t.Fatalf("Expected no error from DeepValidate, got %v", err)
	}
}

func TestDeepValidate_NoMasterZones(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))
	expectErrorFromDeepValidate(t, c, groups, "must configure at least one ControlPlane InstanceGroup")
}

func TestDeepValidate_BadZone(t *testing.T) {
	t.Skipf("Zone validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "subnet-badzone", Zone: "us-test-1z", CIDR: "172.20.1.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1z"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1z"))
	expectErrorFromDeepValidate(t, c, groups, "Zone is not a recognized AZ")
}

func TestDeepValidate_MixedRegion(t *testing.T) {
	t.Skipf("Region validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "test1a", Zone: "us-test-1a", CIDR: "172.20.1.0/24"},
		{Name: "west1b", Zone: "us-west-1b", CIDR: "172.20.2.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a", "subnet-us-west-1b"))

	expectErrorFromDeepValidate(t, c, groups, "Clusters cannot span multiple regions")
}

func TestDeepValidate_RegionAsZone(t *testing.T) {
	t.Skipf("Region validation not checked by DeepValidate")
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "test1", Zone: "us-test-1", CIDR: "172.20.1.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1"))

	expectErrorFromDeepValidate(t, c, groups, "Region is not a recognized EC2 region: \"us-east-\" (check you have specified valid zones?)")
}

func TestDeepValidate_NotIncludedZone(t *testing.T) {
	c := buildDefaultCluster(t)
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1d"))

	expectErrorFromDeepValidate(t, c, groups, "spec.networking.subnets[0]: Not found: \"subnet-us-test-1d\"")
}

func TestDeepValidate_DuplicateZones(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "dup1", Zone: "us-test-1a", CIDR: "172.20.1.0/24"},
		{Name: "dup1", Zone: "us-test-1a", CIDR: "172.20.2.0/24"},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("dup1"))
	groups = append(groups, buildMinimalNodeInstanceGroup("dup1"))
	expectErrorFromDeepValidate(t, c, groups, "spec.networking.subnets[1].name: Duplicate value: \"dup1\"")
}

func TestDeepValidate_ExtraMasterZone(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.Networking.Subnets = []kopsapi.ClusterSubnetSpec{
		{Name: "test1a", Zone: "us-test-1a", CIDR: "172.20.1.0/24", Type: kopsapi.SubnetTypePublic},
		{Name: "test1b", Zone: "us-test-1b", CIDR: "172.20.2.0/24", Type: kopsapi.SubnetTypePublic},
	}
	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1b"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1c"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a", "subnet-us-test-1b"))

	expectErrorFromDeepValidate(t, c, groups, "spec.networking.subnets[0]: Not found: \"subnet-us-test-1a\"")
}

func TestDeepValidate_EvenEtcdClusterSize(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.EtcdClusters = []kopsapi.EtcdClusterSpec{
		{
			Name: "main",
			Members: []kopsapi.EtcdMemberSpec{
				{Name: "us-test-1a", InstanceGroup: fi.PtrTo("us-test-1a")},
				{Name: "us-test-1b", InstanceGroup: fi.PtrTo("us-test-1b")},
			},
		},
	}

	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1b"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1c"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))

	expectErrorFromDeepValidate(t, c, groups, "Should be an odd number of control-plane-zones for quorum. Use --zones and --control-plane-zones to declare node zones and control-plane zones separately")
}

func TestDeepValidate_MissingEtcdMember(t *testing.T) {
	c := buildDefaultCluster(t)
	c.Spec.EtcdClusters = []kopsapi.EtcdClusterSpec{
		{
			Name: "main",
			Members: []kopsapi.EtcdMemberSpec{
				{Name: "us-test-1a", InstanceGroup: fi.PtrTo("us-test-1a")},
				{Name: "us-test-1b", InstanceGroup: fi.PtrTo("us-test-1b")},
				{Name: "us-test-1c", InstanceGroup: fi.PtrTo("us-test-1c")},
			},
		},
	}

	var groups []*kopsapi.InstanceGroup
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1a"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1b"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1c"))
	groups = append(groups, buildMinimalMasterInstanceGroup("subnet-us-test-1d"))
	groups = append(groups, buildMinimalNodeInstanceGroup("subnet-us-test-1a"))

	expectErrorFromDeepValidate(t, c, groups, "spec.metadata.name: Forbidden: InstanceGroup \"master-subnet-us-test-1a\" with role ControlPlane must have a member in etcd cluster \"main\"")
}

func expectErrorFromDeepValidate(t *testing.T, c *kopsapi.Cluster, groups []*kopsapi.InstanceGroup, message string) {
	err := validation.DeepValidate(c, groups, true, vfs.Context, nil)
	if err == nil {
		t.Fatalf("Expected error %q from DeepValidate (strict=true), not no error raised", message)
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

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

package main

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/amazon-ec2-instance-selector/v2/pkg/cli"
	"k8s.io/kops/pkg/apis/kops"
)

func TestGetFilters(t *testing.T) {
	commandline := cli.New("test", "test", "test", "test", nil)
	commandline.Flags = map[string]interface{}{
		flexible:  true,
		allowList: regexp.MustCompile(".*"),
		denyList:  regexp.MustCompile("t3.nano"),
	}

	filters := getFilters(&commandline, "us-east-2")
	if !*filters.Flexible {
		t.Fatalf("Flexible should be true")
	}
	if filters.AllowList == nil || filters.DenyList == nil {
		t.Fatalf("allowList and denyList should not be nil")
	}
}

func TestGetInstanceSelectorOpts(t *testing.T) {
	count := int32(1)
	outputStr := "json"
	commandline := cli.New("test", "test", "test", "test", nil)
	commandline.Flags = map[string]interface{}{
		instanceGroupName:  "test",
		instanceGroupCount: count,
		nodeCountMax:       count,
		nodeCountMin:       count,
		nodeVolumeSize:     &count,
		nodeSecurityGroups: []string{"sec1"},
		output:             outputStr,
		dryRun:             true,
		clusterAutoscaler:  true,
	}

	instanceSelectorOpts := getInstanceSelectorOpts(&commandline)
	if instanceSelectorOpts.NodeCountMax != count || instanceSelectorOpts.NodeCountMin != count ||
		*instanceSelectorOpts.NodeVolumeSize != count || len(instanceSelectorOpts.NodeSecurityGroups) != int(count) ||
		instanceSelectorOpts.InstanceGroupCount != int(count) {
		t.Fatalf("node count max/min, volume size, and count of secrurity groups should return %d", count)
	}
	if instanceSelectorOpts.Output != outputStr {
		t.Fatalf("Output should be %s but got %s", outputStr, instanceSelectorOpts.Output)
	}
	if !instanceSelectorOpts.DryRun || !instanceSelectorOpts.ClusterAutoscaler {
		t.Fatalf("dryRun and clusterAutoscaler should be true, got false instead")
	}
	if instanceSelectorOpts.InstanceGroupName != "test" {
		t.Fatalf("instance group name should have been test")
	}
}

func TestGetClusterZones(t *testing.T) {
	subnets := []kops.ClusterSubnetSpec{
		{
			Name: "subnet-1234",
			Zone: "us-east-2a",
		},
		{
			Name: "subnet-987",
			Zone: "us-east-2b",
		},
	}
	zones, err := getClusterZones(subnets)
	if err != nil {
		t.Fatalf("an error occurred getting cluster zones")
	}
	if len(zones) != 2 {
		t.Fatalf("should have received two zones from the cluster subnets but only received %d", len(zones))
	}

	// Cross-Region Failure Case
	subnets[0].Zone = "us-west-2a"
	_, err = getClusterZones(subnets)
	if err == nil {
		t.Fatalf("should receive an error when passing in subnets that span multiple regions")
	}

}

func TestCreateInstanceGroup(t *testing.T) {
	zones := []string{"us-east-2a", "us-east-2b", "us-east-2c"}
	actualIG := createInstanceGroup("testGroup", "clusterTest", zones)
	if actualIG.Spec.Role != kops.InstanceGroupRoleNode {
		t.Fatalf("instance group should have the \"%s\" role but got %s", kops.InstanceGroupRoleNode, actualIG.Spec.Role)
	}
	if !reflect.DeepEqual(actualIG.Spec.Subnets, zones) {
		t.Fatalf("instance group should have all the zones passed in but got %s", actualIG.Spec.Subnets)
	}
}

func TestDecorateWithInstanceGroupSpecs(t *testing.T) {
	count := int32(1)
	instanceGroupOpts := InstanceSelectorOptions{
		NodeCountMax:       count,
		NodeCountMin:       count,
		NodeVolumeSize:     &count,
		NodeSecurityGroups: []string{"sec-1", "sec-2"},
	}
	actualIG := decorateWithInstanceGroupSpecs(&kops.InstanceGroup{}, instanceGroupOpts)
	if *actualIG.Spec.MaxSize != instanceGroupOpts.NodeCountMax {
		t.Fatalf("expected instance group MaxSize to be %d but got %d", instanceGroupOpts.NodeCountMax, actualIG.Spec.MaxSize)
	}
	if *actualIG.Spec.MinSize != instanceGroupOpts.NodeCountMin {
		t.Fatalf("expected instance group MinSize to be %d but got %d", instanceGroupOpts.NodeCountMin, actualIG.Spec.MinSize)
	}
	if *actualIG.Spec.RootVolumeSize != *instanceGroupOpts.NodeVolumeSize {
		t.Fatalf("expected instance group RootVolumeSize to be %d but got %d", instanceGroupOpts.NodeVolumeSize, actualIG.Spec.RootVolumeSize)
	}
	if !reflect.DeepEqual(actualIG.Spec.AdditionalSecurityGroups, instanceGroupOpts.NodeSecurityGroups) {
		t.Fatalf("expected instance group MaxSize to be %d but got %d", instanceGroupOpts.NodeCountMax, actualIG.Spec.MaxSize)
	}
}

func TestDecorateWithMixedInstancesPolicy(t *testing.T) {
	selectedInstanceTypes := []string{"m3.medium", "m4.medium", "m5.medium"}
	usageClasses := []string{"spot", "on-demand"}
	for _, usageClass := range usageClasses {
		actualIG, err := decorateWithMixedInstancesPolicy(&kops.InstanceGroup{}, usageClass, selectedInstanceTypes)
		if err != nil {
			t.Fatalf("decorateWithMixedInstancesPolicy returned an error: %v", err)
		}
		if actualIG.Spec.MixedInstancesPolicy == nil {
			t.Fatal("MixedInstancesPolicy should not be nil")
		}
		if !reflect.DeepEqual(actualIG.Spec.MixedInstancesPolicy.Instances, selectedInstanceTypes) {
			t.Fatalf("Instances in MixedInstancePolicy should match selectedInstanceTypes: actual: %v expected: %v", actualIG.Spec.MixedInstancesPolicy, selectedInstanceTypes)
		}
		if usageClass == "spot" && *actualIG.Spec.MixedInstancesPolicy.SpotAllocationStrategy != "capacity-optimized" {
			t.Fatal("Spot MixedInstancePolicy should use capacity-optimizmed allocation strategy")
		}
	}
}

func TestDecorateWithClusterAutoscalerLabels(t *testing.T) {
	initialIG := kops.InstanceGroup{}
	initialIG.ObjectMeta.Name = "testInstanceGroup"

	actualIG := decorateWithClusterAutoscalerLabels(&initialIG)
	if _, ok := actualIG.Spec.CloudLabels["k8s.io/cluster-autoscaler/enabled"]; !ok {
		t.Fatalf("cloudLabels for cluster autoscaler should have been added to the instance group spec")
	}
}

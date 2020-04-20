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
)

func buildMinimalNodeInstanceGroup(subnets ...string) *kopsapi.InstanceGroup {
	g := &kopsapi.InstanceGroup{}
	g.ObjectMeta.Name = "nodes"
	g.Spec.Role = kopsapi.InstanceGroupRoleNode
	g.Spec.Subnets = subnets

	return g
}

func buildMinimalMasterInstanceGroup(subnets ...string) *kopsapi.InstanceGroup {
	g := &kopsapi.InstanceGroup{}
	g.ObjectMeta.Name = "master"
	g.Spec.Role = kopsapi.InstanceGroupRoleMaster
	g.Spec.Subnets = subnets

	return g
}

func TestPopulateInstanceGroup_Name_Required(t *testing.T) {
	cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.ObjectMeta.Name = ""

	channel := &kopsapi.Channel{}

	expectErrorFromPopulateInstanceGroup(t, cluster, g, channel, "objectMeta.name")
}

func TestPopulateInstanceGroup_Role_Required(t *testing.T) {
	cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.Spec.Role = ""

	channel := &kopsapi.Channel{}

	expectErrorFromPopulateInstanceGroup(t, cluster, g, channel, "spec.role")
}

func expectErrorFromPopulateInstanceGroup(t *testing.T, cluster *kopsapi.Cluster, g *kopsapi.InstanceGroup, channel *kopsapi.Channel, message string) {
	_, err := PopulateInstanceGroupSpec(cluster, g, channel)
	if err == nil {
		t.Fatalf("Expected error from PopulateInstanceGroup")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

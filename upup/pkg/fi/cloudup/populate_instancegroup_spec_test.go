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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
)

func buildMinimalNodeInstanceGroup(subnets ...string) *kopsapi.InstanceGroup {
	g := &kopsapi.InstanceGroup{}
	g.ObjectMeta.Name = "nodes"
	g.Spec.Role = kopsapi.InstanceGroupRoleNode
	g.Spec.MinSize = fi.Int32(1)
	g.Spec.MaxSize = fi.Int32(1)
	g.Spec.Image = "my-image"
	g.Spec.Subnets = subnets

	return g
}

func buildMinimalMasterInstanceGroup(subnet string) *kopsapi.InstanceGroup {
	g := &kopsapi.InstanceGroup{}
	g.ObjectMeta.Name = "master-" + subnet
	g.Spec.Role = kopsapi.InstanceGroupRoleMaster
	g.Spec.MinSize = fi.Int32(1)
	g.Spec.MaxSize = fi.Int32(1)
	g.Spec.Image = "my-image"
	g.Spec.Subnets = []string{subnet}

	return g
}

func TestPopulateInstanceGroup_Name_Required(t *testing.T) {
	_, cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.ObjectMeta.Name = ""

	channel := &kopsapi.Channel{}

	expectErrorFromPopulateInstanceGroup(t, cluster, g, channel, "objectMeta.name")
}

func TestPopulateInstanceGroup_Role_Required(t *testing.T) {
	_, cluster := buildMinimalCluster()
	g := buildMinimalNodeInstanceGroup()
	g.Spec.Role = ""

	channel := &kopsapi.Channel{}

	expectErrorFromPopulateInstanceGroup(t, cluster, g, channel, "spec.role")
}

// TestPopulateInstanceGroup_AddTaintsCollision ensures we handle IGs with a user configured taint that kOps also adds by default
func TestPopulateInstanceGroup_AddTaintsCollision(t *testing.T) {
	_, cluster := buildMinimalCluster()
	input := buildMinimalNodeInstanceGroup()
	input.Spec.Taints = []string{"nvidia.com/gpu:NoSchedule"}
	input.Spec.MachineType = "g4dn.xlarge"
	cluster.Spec.Containerd.NvidiaGPU = &kopsapi.NvidiaGPUConfig{Enabled: fi.Bool(true)}

	channel := &kopsapi.Channel{}

	cloud, err := BuildCloud(cluster)
	if err != nil {
		t.Fatalf("error from BuildCloud: %v", err)
	}
	output, err := PopulateInstanceGroupSpec(cluster, input, cloud, channel)
	if err != nil {
		t.Fatalf("error from PopulateInstanceGroupSpec: %v", err)
	}
	if len(output.Spec.Kubelet.Taints) != 1 {
		t.Errorf("Expected only 1 taint, got %d", len(output.Spec.Taints))
	}
}

// TestPopulateInstanceGroup_AddTaintsCollision2 ensures we handle taints that are configured in multiple parts of the spec and multiple resources.
// This one also adds a second taint that we should see in the final result
func TestPopulateInstanceGroup_AddTaintsCollision3(t *testing.T) {
	taint := "e2etest:NoSchedule"
	taint2 := "e2etest:NoExecute"
	_, cluster := buildMinimalCluster()
	cluster.Spec.Kubelet = &kopsapi.KubeletConfigSpec{
		Taints: []string{taint, taint2},
	}
	input := buildMinimalNodeInstanceGroup()
	input.Spec.Taints = []string{taint}
	input.Spec.Kubelet = &kopsapi.KubeletConfigSpec{
		Taints: []string{taint},
	}

	channel := &kopsapi.Channel{}

	cloud, err := BuildCloud(cluster)
	if err != nil {
		t.Fatalf("error from BuildCloud: %v", err)
	}
	output, err := PopulateInstanceGroupSpec(cluster, input, cloud, channel)
	if err != nil {
		t.Fatalf("error from PopulateInstanceGroupSpec: %v", err)
	}
	if len(output.Spec.Kubelet.Taints) != 2 {
		t.Errorf("Expected only 2 taints, got %d", len(output.Spec.Kubelet.Taints))
	}
}

func TestPopulateInstanceGroup_AddTaints(t *testing.T) {
	_, cluster := buildMinimalCluster()
	input := buildMinimalNodeInstanceGroup()
	input.Spec.MachineType = "g4dn.xlarge"
	cluster.Spec.Containerd.NvidiaGPU = &kopsapi.NvidiaGPUConfig{Enabled: fi.Bool(true)}

	channel := &kopsapi.Channel{}

	cloud, err := BuildCloud(cluster)
	if err != nil {
		t.Fatalf("error from BuildCloud: %v", err)
	}
	output, err := PopulateInstanceGroupSpec(cluster, input, cloud, channel)
	if err != nil {
		t.Fatalf("error from PopulateInstanceGroupSpec: %v", err)
	}
	if len(output.Spec.Taints) != 1 {
		t.Errorf("Expected only 1 taint, got %d", len(output.Spec.Taints))
	}
}

func expectErrorFromPopulateInstanceGroup(t *testing.T, cluster *kopsapi.Cluster, g *kopsapi.InstanceGroup, channel *kopsapi.Channel, message string) {
	cloud, err := BuildCloud(cluster)
	if err != nil {
		t.Fatalf("error from BuildCloud: %v", err)
	}

	_, err = PopulateInstanceGroupSpec(cluster, g, cloud, channel)
	if err == nil {
		t.Fatalf("Expected error from PopulateInstanceGroup")
	}
	actualMessage := fmt.Sprintf("%v", err)
	if !strings.Contains(actualMessage, message) {
		t.Fatalf("Expected error %q, got %q", message, actualMessage)
	}
}

func TestMachineArchitecture(t *testing.T) {
	tests := []struct {
		machineType string
		arch        architectures.Architecture
		err         error
	}{
		{
			machineType: "t2.micro",
			arch:        architectures.ArchitectureAmd64,
			err:         nil,
		},
		{
			machineType: "t3.micro",
			arch:        architectures.ArchitectureAmd64,
			err:         nil,
		},
		{
			machineType: "a1.large",
			arch:        architectures.ArchitectureArm64,
			err:         nil,
		},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s-%s", test.machineType, test.arch), func(t *testing.T) {
			_, cluster := buildMinimalCluster()
			cloud, err := BuildCloud(cluster)
			if err != nil {
				t.Fatalf("error from BuildCloud: %v", err)
			}

			arch, err := MachineArchitecture(cloud, test.machineType)
			if err != test.err {
				t.Errorf("actual error %q differs from expected error %q", err, test.err)
				return
			}

			if arch != test.arch {
				t.Errorf("actual architecture %q differs from expected architecture %q", arch, test.arch)
				return
			}
		})
	}
}

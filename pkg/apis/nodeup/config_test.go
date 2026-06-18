/*
Copyright 2026 The Kubernetes Authors.

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

package nodeup

import (
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

// TestNewConfigDefaultMachineType verifies that NewConfig populates DefaultMachineType
// for exactly the networking modes whose kubelet maxPods calculation in nodeup falls
// back to it when the IMDS instance-type lookup fails: AmazonVPC and Cilium-ENI.
func TestNewConfigDefaultMachineType(t *testing.T) {
	grid := []struct {
		name       string
		networking kops.NetworkingSpec
		want       string
	}{
		{
			name:       "AmazonVPC",
			networking: kops.NetworkingSpec{AmazonVPC: &kops.AmazonVPCNetworkingSpec{}},
			want:       "m5.large",
		},
		{
			name:       "Cilium with ENI IPAM",
			networking: kops.NetworkingSpec{Cilium: &kops.CiliumNetworkingSpec{IPAM: kops.CiliumIpamEni}},
			want:       "m5.large",
		},
		{
			name:       "Cilium without ENI IPAM",
			networking: kops.NetworkingSpec{Cilium: &kops.CiliumNetworkingSpec{}},
			want:       "",
		},
		{
			name:       "Calico",
			networking: kops.NetworkingSpec{Calico: &kops.CalicoNetworkingSpec{}},
			want:       "",
		},
	}

	for _, tc := range grid {
		t.Run(tc.name, func(t *testing.T) {
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.32.0",
					KubeAPIServer:     &kops.KubeAPIServerConfig{},
					Networking:        tc.networking,
				},
			}
			cluster.Name = "test.example.com"

			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:        kops.InstanceGroupSubRoleNode.Role(),
					MachineType: "m5.large,m5.xlarge",
				},
			}

			config, _ := NewConfig(cluster, ig)

			if tc.want == "" {
				if config.DefaultMachineType != nil {
					t.Errorf("DefaultMachineType = %q, want nil", *config.DefaultMachineType)
				}
				return
			}
			if config.DefaultMachineType == nil {
				t.Fatalf("DefaultMachineType = nil, want %q", tc.want)
			}
			if *config.DefaultMachineType != tc.want {
				t.Errorf("DefaultMachineType = %q, want %q", *config.DefaultMachineType, tc.want)
			}
		})
	}
}

func TestNewConfigGVisorWorkerOnly(t *testing.T) {
	ptrToBool := func(v bool) *bool {
		return &v
	}

	newCluster := func(containerd *kops.ContainerdConfig) *kops.Cluster {
		cluster := &kops.Cluster{
			Spec: kops.ClusterSpec{
				Containerd:            containerd,
				KubernetesVersion:     "1.32.0",
				KubeAPIServer:         &kops.KubeAPIServerConfig{},
				KubeControllerManager: &kops.KubeControllerManagerConfig{},
				KubeScheduler:         &kops.KubeSchedulerConfig{},
				Networking:            kops.NetworkingSpec{Calico: &kops.CalicoNetworkingSpec{}},
			},
		}
		cluster.Name = "test.example.com"
		return cluster
	}

	for _, test := range []struct {
		name                string
		role                kops.InstanceGroupRole
		containerd          *kops.ContainerdConfig
		instanceGroupConfig bool
		wantGVisor          bool
	}{
		{
			name:       "cluster config ignored on worker",
			role:       kops.InstanceGroupSubRoleNode.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on control plane",
			role:       kops.InstanceGroupSubRoleControlPlane.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on apiserver",
			role:       kops.InstanceGroupSubRoleAPIServer.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on etcd",
			role:       kops.InstanceGroupSubRoleEtcd.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on scheduler",
			role:       kops.InstanceGroupSubRoleScheduler.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on cloud controller manager",
			role:       kops.InstanceGroupSubRoleCloudControllerManager.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on kube controller manager",
			role:       kops.InstanceGroupSubRoleKubeControllerManager.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:       "cluster config on bastion",
			role:       kops.InstanceGroupSubRoleBastion.Role(),
			containerd: &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
		},
		{
			name:                "instance group config on worker",
			role:                kops.InstanceGroupSubRoleNode.Role(),
			containerd:          &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
			instanceGroupConfig: true,
			wantGVisor:          true,
		},
		{
			name:                "instance group config on control plane",
			role:                kops.InstanceGroupSubRoleControlPlane.Role(),
			containerd:          &kops.ContainerdConfig{GVisor: &kops.GVisorConfig{Enabled: ptrToBool(true)}},
			instanceGroupConfig: true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			clusterContainerd := test.containerd
			igContainerd := (*kops.ContainerdConfig)(nil)
			if test.instanceGroupConfig {
				clusterContainerd = nil
				igContainerd = test.containerd
			}
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:       test.role,
					Containerd: igContainerd,
				},
			}

			config, _ := NewConfig(newCluster(clusterContainerd), ig)

			if got := config.GVisor != nil; got != test.wantGVisor {
				t.Errorf("GVisor config presence = %v, want %v", got, test.wantGVisor)
			}
			if config.ContainerdConfig == nil {
				t.Fatalf("ContainerdConfig = nil, want non-nil")
			}
			if got := config.ContainerdConfig.GVisor != nil; got != test.wantGVisor {
				t.Errorf("ContainerdConfig.GVisor presence = %v, want %v", got, test.wantGVisor)
			}
		})
	}
}

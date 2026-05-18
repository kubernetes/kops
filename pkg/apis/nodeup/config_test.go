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
					Role:        kops.InstanceGroupRoleNode,
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

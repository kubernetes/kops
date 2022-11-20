/*
Copyright 2021 The Kubernetes Authors.

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

package commands

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestUnsetClusterBadInput(t *testing.T) {
	fields := []string{
		"bad-unset-input",
	}

	err := UnsetClusterFields(fields, &kops.Cluster{})
	if err == nil {
		t.Errorf("expected a field parsing error, but received none")
	}
}

func TestUnsetClusterFields(t *testing.T) {
	grid := []struct {
		Fields []string
		Input  kops.Cluster
		Output kops.Cluster
	}{
		{
			Fields: []string{
				"spec.kubernetesVersion",
				"spec.kubelet.authorizationMode",
				"spec.kubelet.authenticationTokenWebhook",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.8.2",
					Kubelet: &kops.KubeletConfigSpec{
						AuthorizationMode:          "Webhook",
						AuthenticationTokenWebhook: fi.PtrTo(true),
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{},
				},
			},
		},
		{
			Fields: []string{
				"spec.api.dns",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						DNS: &kops.DNSAccessSpec{},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{},
			},
		},
		{
			Fields: []string{"spec.kubelet.authorizationMode"},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{
						AuthorizationMode: "Webhook",
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{},
				},
			},
		},
		{
			Fields: []string{"spec.kubelet.authenticationTokenWebhook"},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{
						AuthenticationTokenWebhook: fi.PtrTo(false),
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{},
				},
			},
		},
		{
			Fields: []string{"spec.docker.selinuxEnabled"},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Docker: &kops.DockerConfig{
						SelinuxEnabled: fi.PtrTo(true),
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Docker: &kops.DockerConfig{},
				},
			},
		},
		{
			Fields: []string{"spec.kubernetesVersion"},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.2.3",
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{},
			},
		},
		{
			Fields: []string{"spec.api.publicName"},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "api.example.com",
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{},
			},
		},
		{
			Fields: []string{"spec.kubeDNS.provider"},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeDNS: &kops.KubeDNSConfig{
						Provider: "CoreDNS",
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeDNS: &kops.KubeDNSConfig{},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.nodePortAccess",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					NodePortAccess: []string{"10.0.0.0/8", "192.168.0.0/16"},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].version",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Version: "v3.2.1"},
						{Name: "two", Version: "v3.2.1"},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one"},
						{Name: "two"},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].provider",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Provider: kops.EtcdProviderTypeManager},
						{Name: "two", Provider: kops.EtcdProviderTypeManager},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one"},
						{Name: "two"},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*]",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Image: "etcd-manager:v1.2.3"},
						{Name: "two", Image: "etcd-manager:v1.2.3"},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{},
						{},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.ipam",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							IPAM: "on",
						},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.enableHostReachableServices",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							EnableHostReachableServices: true,
						},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.enableNodePort",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							EnableNodePort: true,
						},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.masquerade",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							Masquerade: fi.PtrTo(false),
						},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.kubeProxy.enabled",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{
						Enabled: fi.PtrTo(true),
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.agentPrometheusPort",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							AgentPrometheusPort: 1234,
						},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
		},
	}

	for _, g := range grid {
		c := g.Input

		err := UnsetClusterFields(g.Fields, &c)
		if err != nil {
			t.Errorf("unexpected error from unsetClusterFields %v: %v", g.Fields, err)
			continue
		}

		if !reflect.DeepEqual(c, g.Output) {
			t.Errorf("unexpected output from unsetClusterFields %v.  expected=%v, actual=%v", g.Fields, g.Output, c)
			continue
		}

	}
}

func TestUnsetCiliumFields(t *testing.T) {
	grid := []struct {
		Fields []string
		Input  kops.Cluster
		Output kops.Cluster
	}{
		{
			Fields: []string{
				"cluster.spec.networking.cilium.ipam",
				"cluster.spec.networking.cilium.enableNodePort",
				"cluster.spec.networking.cilium.masquerade",
				"cluster.spec.kubeProxy.enabled",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{
						Enabled: fi.PtrTo(false),
					},
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							IPAM:           "eni",
							EnableNodePort: true,
							Masquerade:     fi.PtrTo(false),
						},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{},
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{},
					},
				},
			},
		},
	}

	for _, g := range grid {
		c := g.Input

		err := UnsetClusterFields(g.Fields, &c)
		if err != nil {
			t.Errorf("unexpected error from unsetClusterFields %v: %v", g.Fields, err)
			continue
		}

		if !reflect.DeepEqual(c, g.Output) {
			t.Errorf("unexpected output from unsetClusterFields %v.  expected=%v, actual=%v", g.Fields, g.Output, c)
			continue
		}

	}
}

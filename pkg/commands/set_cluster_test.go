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

package commands

import (
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestSetClusterFields(t *testing.T) {
	grid := []struct {
		Fields []string
		Input  kops.Cluster
		Output kops.Cluster
	}{
		{
			Fields: []string{
				"spec.kubernetesVersion=1.8.2",
				"spec.kubelet.authorizationMode=Webhook",
				"spec.kubelet.authenticationTokenWebhook=true",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "1.8.2",
					Kubelet: &kops.KubeletConfigSpec{
						AuthorizationMode:          "Webhook",
						AuthenticationTokenWebhook: fi.Bool(true),
					},
				},
			},
		},
		{
			Fields: []string{"spec.kubelet.authorizationMode=Webhook"},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{
						AuthorizationMode: "Webhook",
					},
				},
			},
		},
		{
			Fields: []string{"spec.kubelet.authenticationTokenWebhook=false"},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Kubelet: &kops.KubeletConfigSpec{
						AuthenticationTokenWebhook: fi.Bool(false),
					},
				},
			},
		},
		{
			Fields: []string{"spec.docker.selinuxEnabled=true"},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Docker: &kops.DockerConfig{
						SelinuxEnabled: fi.Bool(true),
					},
				},
			},
		},
		{
			Fields: []string{"spec.kubernetesVersion=v1.2.3"},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubernetesVersion: "v1.2.3",
				},
			},
		},
		{
			Fields: []string{"spec.masterPublicName=api.example.com"},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					MasterPublicName: "api.example.com",
				},
			},
		},
		{
			Fields: []string{"spec.kubeDNS.provider=CoreDNS"},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeDNS: &kops.KubeDNSConfig{
						Provider: "CoreDNS",
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.nodePortAccess=10.0.0.0/8,192.168.0.0/16",
			},
			Input: kops.Cluster{},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					NodePortAccess: []string{"10.0.0.0/8", "192.168.0.0/16"},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].enableEtcdTLS=true",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", EnableEtcdTLS: true},
						{Name: "two", EnableEtcdTLS: false},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", EnableEtcdTLS: true},
						{Name: "two", EnableEtcdTLS: true},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].enableTLSAuth=true",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", EnableTLSAuth: true},
						{Name: "two", EnableTLSAuth: false},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", EnableTLSAuth: true},
						{Name: "two", EnableTLSAuth: true},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].version=v3.2.1",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Version: "v2.0.0"},
						{Name: "two"},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Version: "v3.2.1"},
						{Name: "two", Version: "v3.2.1"},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].provider=Manager",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Provider: kops.EtcdProviderTypeLegacy},
						{Name: "two"},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Provider: kops.EtcdProviderTypeManager},
						{Name: "two", Provider: kops.EtcdProviderTypeManager},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.etcdClusters[*].image=etcd-manager:v1.2.3",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Image: "foo"},
						{Name: "two"},
					},
				},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{Name: "one", Image: "etcd-manager:v1.2.3"},
						{Name: "two", Image: "etcd-manager:v1.2.3"},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.ipam=on",
			},
			Input: kops.Cluster{},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							Ipam: "on",
						},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.enableNodePort=true",
			},
			Input: kops.Cluster{},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							EnableNodePort: true,
						},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.disableMasquerade=true",
			},
			Input: kops.Cluster{},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							DisableMasquerade: true,
						},
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.kubeProxy.enabled=true",
			},
			Input: kops.Cluster{},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{
						Enabled: fi.Bool(true),
					},
				},
			},
		},
		{
			Fields: []string{
				"cluster.spec.networking.cilium.agentPrometheusPort=1234",
			},
			Input: kops.Cluster{},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							AgentPrometheusPort: 1234,
						},
					},
				},
			},
		},
	}

	for _, g := range grid {
		var igs []*kops.InstanceGroup
		c := g.Input

		err := SetClusterFields(g.Fields, &c, igs)
		if err != nil {
			t.Errorf("unexpected error from setClusterFields %v: %v", g.Fields, err)
			continue
		}

		if !reflect.DeepEqual(c, g.Output) {
			t.Errorf("unexpected output from setClusterFields %v.  expected=%v, actual=%v", g.Fields, g.Output, c)
			continue
		}

	}
}

func TestSetCiliumFields(t *testing.T) {

	grid := []struct {
		Fields []string
		Input  kops.Cluster
		Output kops.Cluster
	}{
		{
			Fields: []string{
				"cluster.spec.networking.cilium.ipam=eni",
				"cluster.spec.networking.cilium.enableNodePort=true",
				"cluster.spec.networking.cilium.disableMasquerade=true",
				"cluster.spec.kubeProxy.enabled=false",
			},
			Input: kops.Cluster{
				Spec: kops.ClusterSpec{},
			},
			Output: kops.Cluster{
				Spec: kops.ClusterSpec{
					KubeProxy: &kops.KubeProxyConfig{
						Enabled: fi.Bool(false),
					},
					Networking: &kops.NetworkingSpec{
						Cilium: &kops.CiliumNetworkingSpec{
							Ipam:              "eni",
							EnableNodePort:    true,
							DisableMasquerade: true,
						},
					},
				},
			},
		},
	}

	for _, g := range grid {
		var igs []*kops.InstanceGroup
		c := g.Input

		err := SetClusterFields(g.Fields, &c, igs)
		if err != nil {
			t.Errorf("unexpected error from setClusterFields %v: %v", g.Fields, err)
			continue
		}

		if !reflect.DeepEqual(c, g.Output) {
			t.Errorf("unexpected output from setClusterFields %v.  expected=%v, actual=%v", g.Fields, g.Output, c)
			continue
		}

	}
}

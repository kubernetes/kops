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
	"reflect"
	"testing"

	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
)

func TestRemoveSharedPrefix(t *testing.T) {
	grid := []struct {
		Input  []string
		Output []string
	}{
		{
			Input:  []string{"a", "b", "c"},
			Output: []string{"a", "b", "c"},
		},
		{
			Input:  []string{"za", "zb", "zc"},
			Output: []string{"a", "b", "c"},
		},
		{
			Input:  []string{"zza", "zzb", "zzc"},
			Output: []string{"a", "b", "c"},
		},
		{
			Input:  []string{"zza", "zzb", ""},
			Output: []string{"zza", "zzb", ""},
		},
		{
			Input:  []string{"us-test-1a-1", "us-test-1a-2", "us-test-1a-3"},
			Output: []string{"etcd-1", "etcd-2", "etcd-3"},
		},
		{
			Input:  []string{"us-test-1a-1", "us-test-1b-1", "us-test-1a-2", "us-test-1b-2", "us-test-1a-3"},
			Output: []string{"a-1", "b-1", "a-2", "b-2", "a-3"},
		},
	}
	for _, g := range grid {
		actual := trimCommonPrefix(g.Input)
		if !reflect.DeepEqual(actual, g.Output) {
			t.Errorf("unexpected result from %q.  actual=%v, expected=%v", g.Input, actual, g.Output)
		}
	}
}

func TestCreateEtcdCluster(t *testing.T) {
	masters := []*api.InstanceGroup{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master1",
			},
		},
	}
	name := "foo"
	etcd := createEtcdCluster(name, masters, false, "")

	if name != etcd.Name {
		t.Errorf("Expected: %v, Got: %v", name, etcd.Name)
	}
}

func TestSetupNetworking(t *testing.T) {
	tests := []struct {
		options  NewClusterOptions
		actual   api.Cluster
		expected api.Cluster
	}{
		{
			options: NewClusterOptions{
				Networking: "kubenet",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Kubenet: &api.KubenetNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "external",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						External: &api.ExternalNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "cni",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						CNI: &api.CNINetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "kopeio",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Kopeio: &api.KopeioNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "kopeio-vxlan",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Kopeio: &api.KopeioNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "weave",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Weave: &api.WeaveNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "flannel",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Flannel: &api.FlannelNetworkingSpec{
							Backend: "vxlan",
						},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "flannel-vxlan",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Flannel: &api.FlannelNetworkingSpec{
							Backend: "vxlan",
						},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "flannel-udp",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Flannel: &api.FlannelNetworkingSpec{
							Backend: "udp",
						},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "calico",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Calico: &api.CalicoNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "canal",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						Canal: &api.CanalNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "kube-router",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					KubeProxy: &api.KubeProxyConfig{
						Enabled: fi.PtrTo(false),
					},
					Networking: &api.NetworkingSpec{
						Kuberouter: &api.KuberouterNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "amazonvpc",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						AmazonVPC: &api.AmazonVPCNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "amazon-vpc-routed-eni",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						AmazonVPC: &api.AmazonVPCNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "cilium",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					KubeProxy: &api.KubeProxyConfig{
						Enabled: fi.PtrTo(false),
					},
					Networking: &api.NetworkingSpec{
						Cilium: &api.CiliumNetworkingSpec{
							EnableNodePort: true,
						},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "cilium-etcd",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					KubeProxy: &api.KubeProxyConfig{
						Enabled: fi.PtrTo(false),
					},
					Networking: &api.NetworkingSpec{
						Cilium: &api.CiliumNetworkingSpec{
							EnableNodePort: true,
							EtcdManaged:    true,
						},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "gce",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: &api.NetworkingSpec{
						GCE: &api.GCENetworkingSpec{},
					},
				},
			},
		},
	}

	for _, test := range tests {
		actual := api.Cluster{}
		err := setupNetworking(&test.options, &actual)
		if err != nil {
			t.Errorf("error during network setup: %v", err)
		}
		expectedYaml, err := yaml.Marshal(test.expected)
		if err != nil {
			t.Errorf("error converting expected cluster spec to yaml: %v", err)
		}
		actualYaml, err := yaml.Marshal(actual)
		if err != nil {
			t.Errorf("error converting actual cluster spec to yaml: %v", err)
		}
		if string(expectedYaml) != string(actualYaml) {
			diffString := diff.FormatDiff(string(expectedYaml), string(actualYaml))
			t.Errorf("unexpected cluster networking setup:\n%s", diffString)
		}
	}
}

func TestDefaultImage(t *testing.T) {
	tests := []struct {
		cluster      *api.Cluster
		architecture architectures.Architecture
		expected     string
	}{
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						AWS: &api.AWSSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     "099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-20221018",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						AWS: &api.AWSSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureArm64,
			expected:     "099720109477/ubuntu/images/hvm-ssd/ubuntu-focal-20.04-arm64-server-20221018",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						Azure: &api.AzureSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     "Canonical:0001-com-ubuntu-server-focal:20_04-lts-gen2:20.04.202210180",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						GCE: &api.GCESpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     "ubuntu-os-cloud/ubuntu-2004-focal-v20221018",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						DO: &api.DOSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultDOImage,
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						Hetzner: &api.HetznerSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultHetznerImage,
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.25.0",
					CloudProvider: api.CloudProviderSpec{
						Scaleway: &api.ScalewaySpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultScalewayImage,
		},
	}

	channel, err := api.LoadChannel("file://tests/channels/channel.yaml")
	if err != nil {
		t.Fatalf("unable to load test channel: %v", err)
	}

	for _, test := range tests {
		actual, err := defaultImage(test.cluster, channel, test.architecture)
		if err != nil {
			t.Error(err)
			continue
		}
		if actual != test.expected {
			t.Errorf("unexpected default image for cluster %s: expected=%q, actual=%q", fi.DebugAsJsonString(test.cluster.Spec), test.expected, actual)
		}
	}
}

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
	"strings"
	"testing"

	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/util/pkg/vfs"
	"sigs.k8s.io/yaml"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
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

func TestSetupKarpenterNodes(t *testing.T) {
	grid := []struct {
		desc        string
		nodeCount   int32
		nodeSizes   []string
		static      bool
		machineType string
		mixedTypes  []string
	}{
		{
			desc: "dynamic",
		},
		{
			desc:      "static",
			nodeCount: 4,
			static:    true,
		},
		{
			desc:        "single node size",
			nodeSizes:   []string{"m6g.large"},
			machineType: "m6g.large",
		},
		{
			desc:        "multiple node sizes",
			nodeSizes:   []string{"m6g.large", "m6gd.large"},
			machineType: "m6g.large",
			mixedTypes:  []string{"m6g.large", "m6gd.large"},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			groups, err := setupKarpenterNodes(&NewClusterOptions{NodeCount: g.nodeCount, NodeSizes: g.nodeSizes})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(groups) != 1 {
				t.Fatalf("expected one InstanceGroup, got %d", len(groups))
			}

			ig := groups[0]
			if ig.Spec.Manager != api.InstanceManagerKarpenter {
				t.Errorf("expected Karpenter manager, got %q", ig.Spec.Manager)
			}
			if g.static {
				if fi.ValueOf(ig.Spec.MinSize) != g.nodeCount {
					t.Errorf("expected minSize %d, got %v", g.nodeCount, ig.Spec.MinSize)
				}
			} else if ig.Spec.MinSize != nil {
				t.Errorf("expected minSize to be omitted, got %v", ig.Spec.MinSize)
			}
			if ig.Spec.MachineType != g.machineType {
				t.Errorf("expected machineType %q, got %q", g.machineType, ig.Spec.MachineType)
			}
			if g.mixedTypes == nil {
				if ig.Spec.MixedInstancesPolicy != nil {
					t.Errorf("expected no MixedInstancesPolicy, got %v", ig.Spec.MixedInstancesPolicy)
				}
			} else if !reflect.DeepEqual(ig.Spec.MixedInstancesPolicy.Instances, g.mixedTypes) {
				t.Errorf("expected mixed instances %v, got %v", g.mixedTypes, ig.Spec.MixedInstancesPolicy.Instances)
			}
		})
	}
}

func TestSetupNetworking(t *testing.T) {
	tests := []struct {
		options  NewClusterOptions
		expected api.Cluster
	}{
		{
			options: NewClusterOptions{
				Networking: "kubenet",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
						Kopeio: &api.KopeioNetworkingSpec{},
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
						Calico: &api.CalicoNetworkingSpec{},
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
						Enabled: new(false),
					},
					Networking: api.NetworkingSpec{
						KubeRouter: &api.KuberouterNetworkingSpec{},
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
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
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
						Enabled: new(false),
					},
					Networking: api.NetworkingSpec{
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
						Enabled: new(false),
					},
					Networking: api.NetworkingSpec{
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
					Networking: api.NetworkingSpec{
						GCP: &api.GCPNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "gcp",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: api.NetworkingSpec{
						GCP: &api.GCPNetworkingSpec{},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Networking: "kindnet",
			},
			expected: api.Cluster{
				Spec: api.ClusterSpec{
					Networking: api.NetworkingSpec{
						Kindnet: &api.KindnetNetworkingSpec{},
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

func TestSetupTopology(t *testing.T) {
	tests := []struct {
		options  NewClusterOptions
		skeleton api.Cluster
		expected api.Cluster
	}{
		{
			options: NewClusterOptions{
				Topology: api.TopologyPrivate,
				Bastion:  true,
			},
			skeleton: api.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					Networking: api.NetworkingSpec{
						Topology: &api.TopologySpec{
							DNS: api.DNSTypeNone,
						},
					},
				},
			},
			expected: api.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					Networking: api.NetworkingSpec{
						Topology: &api.TopologySpec{
							DNS: api.DNSTypeNone,
						},
					},
				},
			},
		},
		{
			options: NewClusterOptions{
				Topology: api.TopologyPrivate,
				DNSType:  string(api.DNSTypePublic),
				Bastion:  true,
			},
			skeleton: api.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					Networking: api.NetworkingSpec{
						Topology: &api.TopologySpec{
							DNS: api.DNSTypePublic,
						},
					},
				},
			},
			expected: api.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					Networking: api.NetworkingSpec{
						Topology: &api.TopologySpec{
							DNS: api.DNSTypePublic,
							Bastion: &api.BastionSpec{
								PublicName: "bastion.test",
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		_, err := setupTopology(&test.options, &test.skeleton, nil)
		if err != nil {
			t.Errorf("error during topology setup: %v", err)
		}
		expectedYaml, err := yaml.Marshal(test.expected)
		if err != nil {
			t.Errorf("error converting expected cluster spec to yaml: %v", err)
		}
		actualYaml, err := yaml.Marshal(&test.skeleton)
		if err != nil {
			t.Errorf("error converting actual cluster spec to yaml: %v", err)
		}
		if string(expectedYaml) != string(actualYaml) {
			diffString := diff.FormatDiff(string(expectedYaml), string(actualYaml))
			t.Errorf("unexpected cluster topology setup:\n%s", diffString)
		}
	}
}

func TestNewClusterValidatesKubernetesFeatureGates(t *testing.T) {
	vfsContext := vfs.NewTestingVFSContext()
	basePath, err := vfsContext.BuildVfsPath("memfs://tests")
	if err != nil {
		t.Fatalf("error building test state store: %v", err)
	}
	clientset := vfsclientset.NewVFSClientset(vfsContext, basePath)

	tests := []struct {
		name    string
		gates   []string
		wantErr string
	}{
		{
			name:    "should reject empty feature gate if entry is blank",
			gates:   []string{""},
			wantErr: "must not be empty",
		},
		{
			name:    "should reject feature gate if entry is sign only plus",
			gates:   []string{"+"},
			wantErr: "must include a feature name",
		},
		{
			name:    "should reject feature gate if entry is sign only minus",
			gates:   []string{"-"},
			wantErr: "must include a feature name",
		},
		{
			name:    "should continue past feature gate validation if entry is valid",
			gates:   []string{"+ReadWriteOncePod"},
			wantErr: "must specify at least one zone",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opt := &NewClusterOptions{
				ClusterName:            "test.example.com",
				Channel:                "file://tests/channels/channel.yaml",
				KubernetesVersion:      "v1.32.0",
				KubernetesFeatureGates: test.gates,
			}

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("NewCluster panicked: %v", r)
				}
			}()

			_, err := NewCluster(opt, clientset)
			if err == nil {
				t.Fatalf("expected error containing %q", test.wantErr)
			}
			if !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("unexpected error %q, expected to contain %q", err, test.wantErr)
			}
		})
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
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						AWS: &api.AWSSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     "099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-resolute-26.04-amd64-server-20221018",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						AWS: &api.AWSSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureArm64,
			expected:     "099720109477/ubuntu/images/hvm-ssd-gp3/ubuntu-resolute-26.04-arm64-server-20221018",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						Azure: &api.AzureSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     "Canonical:ubuntu-26_04-lts:server:26.04.202210180",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						GCE: &api.GCESpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     "ubuntu-os-cloud/ubuntu-2604-resolute-amd64-v20221018",
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						DO: &api.DOSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultDOImageNoble,
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						Hetzner: &api.HetznerSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultHetznerImageNoble,
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						Scaleway: &api.ScalewaySpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultScalewayImageNoble,
		},
		{
			cluster: &api.Cluster{
				Spec: api.ClusterSpec{
					KubernetesVersion: "v1.32.0",
					CloudProvider: api.CloudProviderSpec{
						Linode: &api.LinodeSpec{},
					},
				},
			},
			architecture: architectures.ArchitectureAmd64,
			expected:     defaultLinodeImageNoble,
		},
	}

	channel, err := api.LoadChannel(vfs.NewTestingVFSContext(), "file://tests/channels/channel.yaml")
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

func TestSetupZonesLinodeSingleRegion(t *testing.T) {
	opt := &NewClusterOptions{Zones: []string{"us-east"}}
	cluster := &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider: api.CloudProviderSpec{Linode: &api.LinodeSpec{}},
		},
	}

	zones := sets.NewString("us-east")
	zoneToSubnetsMap, err := setupZones(opt, cluster, zones)
	if err != nil {
		t.Fatalf("setupZones returned error: %v", err)
	}

	subnets := zoneToSubnetsMap["us-east"]
	if got, want := len(subnets), 1; got != want {
		t.Fatalf("unexpected subnet count for region: got %d, want %d", got, want)
	}

	subnet := subnets[0]
	if got, want := subnet.Name, "us-east"; got != want {
		t.Fatalf("unexpected subnet name: got %q, want %q", got, want)
	}
	if got, want := subnet.Region, "us-east"; got != want {
		t.Fatalf("unexpected subnet region: got %q, want %q", got, want)
	}
	if got, want := subnet.Zone, "us-east"; got != want {
		t.Fatalf("unexpected subnet zone: got %q, want %q", got, want)
	}
}

func TestSetupZonesLinodeSingleRegionOnly(t *testing.T) {
	opt := &NewClusterOptions{Zones: []string{"us-east", "eu-west"}}
	cluster := &api.Cluster{
		Spec: api.ClusterSpec{
			CloudProvider: api.CloudProviderSpec{Linode: &api.LinodeSpec{}},
		},
	}

	_, err := setupZones(opt, cluster, sets.NewString("us-east", "eu-west"))
	if err == nil {
		t.Fatalf("expected error when multiple regions are specified for Linode")
	}
	if !strings.Contains(err.Error(), "one region only") {
		t.Fatalf("unexpected error: %v", err)
	}
}

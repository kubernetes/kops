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

package cloudup

import (
	"reflect"
	"strings"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestKarpenterNodePoolStaticCapacity(t *testing.T) {
	tf := &TemplateFunctions{}
	tf.Cluster = &kops.Cluster{}

	grid := []struct {
		desc        string
		minSize     *int32
		maxSize     *int32
		hasReplicas bool
		limitsNodes string
	}{
		{
			desc: "dynamic",
		},
		{
			desc:        "static",
			minSize:     new(int32(4)),
			hasReplicas: true,
		},
		{
			desc:        "dynamic with maxSize",
			maxSize:     new(int32(10)),
			limitsNodes: "10",
		},
		{
			desc:        "static with maxSize",
			minSize:     new(int32(4)),
			maxSize:     new(int32(10)),
			hasReplicas: true,
			limitsNodes: "10",
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleNode,
					MinSize: g.minSize,
					MaxSize: g.maxSize,
				},
			}
			ig.Name = "nodes"

			rendered, err := tf.KarpenterNodePool(ig)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			hasReplicas := strings.Contains(rendered, "\n  replicas: 4\n")
			if hasReplicas != g.hasReplicas {
				t.Errorf("expected replicas present=%t, got:\n%s", g.hasReplicas, rendered)
			}
			hasLimits := strings.Contains(rendered, "\n  limits:\n    nodes: \""+g.limitsNodes+"\"\n")
			if hasLimits != (g.limitsNodes != "") {
				t.Errorf("expected limits.nodes=%q present=%t, got:\n%s", g.limitsNodes, g.limitsNodes != "", rendered)
			}
		})
	}
}

func TestBuildKarpenterAMITerms(t *testing.T) {
	grid := []struct {
		image    string
		expected []karpenterAMITerm
		error    bool
	}{
		{
			image:    "ami-0123456789abcdef0",
			expected: []karpenterAMITerm{{ID: "ami-0123456789abcdef0"}},
		},
		{
			image:    "ssm:/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id",
			expected: []karpenterAMITerm{{SSMParameter: "/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id"}},
		},
		{
			image:    "kops-node-image",
			expected: []karpenterAMITerm{{Name: "kops-node-image", Owner: "self"}},
		},
		{
			image:    "ubuntu/images/hvm-ssd/ubuntu-noble-24.04-amd64-server-*",
			expected: []karpenterAMITerm{{Owner: "099720109477", Name: "images/hvm-ssd/ubuntu-noble-24.04-amd64-server-*"}},
		},
		{
			image:    "rocky/Rocky-9-EC2-Base-*",
			expected: []karpenterAMITerm{{Owner: "792107900819", Name: "Rocky-9-EC2-Base-*"}},
		},
		{
			image:    "rockylinux/Rocky-9-EC2-Base-*",
			expected: []karpenterAMITerm{{Owner: "792107900819", Name: "Rocky-9-EC2-Base-*"}},
		},
		{
			image:    "debian/debian-12-amd64-*",
			expected: []karpenterAMITerm{{Owner: "136693071363", Name: "debian-12-amd64-*"}},
		},
		{
			image:    "flatcar/Flatcar-stable-*",
			expected: []karpenterAMITerm{{Owner: "075585003325", Name: "Flatcar-stable-*"}},
		},
		{
			image:    "redhat/RHEL-9.4-*",
			expected: []karpenterAMITerm{{Owner: "309956199498", Name: "RHEL-9.4-*"}},
		},
		{
			image:    "amazon/al2023-ami-*",
			expected: []karpenterAMITerm{{Owner: "137112412989", Name: "al2023-ami-*"}},
		},
		{
			image:    "123456789012/my-custom-image-*",
			expected: []karpenterAMITerm{{Owner: "123456789012", Name: "my-custom-image-*"}},
		},
		{
			image: "ssm:",
			error: true,
		},
		{
			image: "https://example.com/image",
			error: true,
		},
		{
			image: "/missing-owner",
			error: true,
		},
		{
			image: "missing-name/",
			error: true,
		},
	}

	for _, g := range grid {
		t.Run(g.image, func(t *testing.T) {
			actual, err := buildKarpenterAMITerms(g.image)
			if g.error {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, g.expected) {
				t.Errorf("expected %#v, got %#v", g.expected, actual)
			}
		})
	}
}

func TestKarpenterInstanceTypes(t *testing.T) {
	ig := &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{
			MachineType: "m6i.large, c6i.large, m6i.large",
			MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
				Instances: []string{"r6i.large", "c6i.large"},
			},
		},
	}
	expected := []string{"c6i.large", "m6i.large", "r6i.large"}

	actual := karpenterInstanceTypes(ig)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}

func TestKarpenterCapacityTypes(t *testing.T) {
	grid := []struct {
		desc     string
		ig       *kops.InstanceGroup
		expected []string
	}{
		{
			desc:     "default",
			ig:       &kops.InstanceGroup{},
			expected: []string{"on-demand"},
		},
		{
			desc: "max price",
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					MaxPrice: new("0.10"),
				},
			},
			expected: []string{"spot"},
		},
		{
			desc: "mixed",
			ig: &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
						OnDemandAboveBase: new(int64(50)),
					},
				},
			},
			expected: []string{"on-demand", "spot"},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			actual := karpenterCapacityTypes(g.ig)
			if !reflect.DeepEqual(actual, g.expected) {
				t.Errorf("expected %#v, got %#v", g.expected, actual)
			}
		})
	}
}

func TestBuildKarpenterKubeletConfiguration(t *testing.T) {
	grid := []struct {
		desc     string
		kubelet  *kops.KubeletConfigSpec
		expected *karpenterKubeletConfiguration
	}{
		{
			desc: "no kubelet config",
		},
		{
			desc:    "kubelet config without maxPods",
			kubelet: &kops.KubeletConfigSpec{},
		},
		{
			desc:     "maxPods set",
			kubelet:  &kops.KubeletConfigSpec{MaxPods: new(int32(50))},
			expected: &karpenterKubeletConfiguration{MaxPods: new(int32(50))},
		},
		{
			desc: "systemReserved and kubeReserved set",
			kubelet: &kops.KubeletConfigSpec{
				SystemReserved: map[string]string{"cpu": "500m", "memory": "1G"},
				KubeReserved:   map[string]string{"cpu": "500m", "memory": "1G"},
			},
			expected: &karpenterKubeletConfiguration{
				SystemReserved: map[string]string{"cpu": "500m", "memory": "1G"},
				KubeReserved:   map[string]string{"cpu": "500m", "memory": "1G"},
			},
		},
		{
			desc: "all supported fields set",
			kubelet: &kops.KubeletConfigSpec{
				MaxPods:        new(int32(50)),
				SystemReserved: map[string]string{"cpu": "500m"},
				KubeReserved:   map[string]string{"memory": "1G"},
			},
			expected: &karpenterKubeletConfiguration{
				MaxPods:        new(int32(50)),
				SystemReserved: map[string]string{"cpu": "500m"},
				KubeReserved:   map[string]string{"memory": "1G"},
			},
		},
		{
			desc:    "unsupported field only is ignored",
			kubelet: &kops.KubeletConfigSpec{KubeReservedCgroup: "/kube"},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleNode,
					Kubelet: g.kubelet,
				},
			}
			actual := buildKarpenterKubeletConfiguration(ig)
			if !reflect.DeepEqual(actual, g.expected) {
				t.Errorf("expected %#v, got %#v", g.expected, actual)
			}
		})
	}
}

func TestKarpenterAssociatePublicIP(t *testing.T) {
	tf := &TemplateFunctions{}
	tf.Cluster = &kops.Cluster{
		Spec: kops.ClusterSpec{
			Networking: kops.NetworkingSpec{
				Subnets: []kops.ClusterSubnetSpec{
					{Name: "public", Type: kops.SubnetTypePublic},
					{Name: "utility", Type: kops.SubnetTypeUtility},
					{Name: "private", Type: kops.SubnetTypePrivate},
					{Name: "dualstack", Type: kops.SubnetTypeDualStack},
					{Name: "bogus", Type: kops.SubnetType("Bogus")},
				},
			},
		},
	}

	grid := []struct {
		desc              string
		subnets           []string
		associatePublicIP *bool
		expected          *bool
		expectError       bool
	}{
		{
			desc:     "public subnet defaults to true",
			subnets:  []string{"public"},
			expected: new(true),
		},
		{
			desc:     "utility subnet defaults to true",
			subnets:  []string{"utility"},
			expected: new(true),
		},
		{
			desc:              "public subnet honors explicit false",
			subnets:           []string{"public"},
			associatePublicIP: new(false),
			expected:          new(false),
		},
		{
			desc:              "public subnet honors explicit true",
			subnets:           []string{"public"},
			associatePublicIP: new(true),
			expected:          new(true),
		},
		{
			desc:     "private subnet is false",
			subnets:  []string{"private"},
			expected: new(false),
		},
		{
			desc:     "dualstack subnet is false",
			subnets:  []string{"dualstack"},
			expected: new(false),
		},
		{
			desc:              "private subnet ignores explicit true",
			subnets:           []string{"private"},
			associatePublicIP: new(true),
			expected:          new(false),
		},
		{
			desc:        "no subnets is an error",
			subnets:     nil,
			expectError: true,
		},
		{
			desc:        "unknown subnet name is an error",
			subnets:     []string{"missing"},
			expectError: true,
		},
		{
			desc:        "unknown subnet type is an error",
			subnets:     []string{"bogus"},
			expectError: true,
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:              kops.InstanceGroupRoleNode,
					Subnets:           g.subnets,
					AssociatePublicIP: g.associatePublicIP,
				},
			}
			ig.Name = "nodes"

			actual, err := tf.karpenterAssociatePublicIP(ig)
			if g.expectError {
				if err == nil {
					t.Fatalf("expected error, got %v", fi.ValueOf(actual))
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(actual, g.expected) {
				t.Errorf("expected %v, got %v", fi.ValueOf(g.expected), fi.ValueOf(actual))
			}
		})
	}
}

func TestKarpenterEC2NodeClassTags(t *testing.T) {
	tags := map[string]string{
		"":                                                 "empty",
		"KubernetesCluster":                                "example.com",
		"k8s.io/role/node":                                 "1",
		"kops.k8s.io/instancegroup":                        "nodes",
		"kubernetes.io/cluster/example.com":                "owned",
		"eks:eks-cluster-name":                             "example",
		"karpenter.sh/nodepool":                            "nodes",
		"karpenter.sh/nodeclaim":                           "claim",
		"karpenter.k8s.aws/ec2nodeclass":                   "nodes",
		"node-template/label/karpenter.sh/nodepool":        "nodes",
		"node-template/label/node-role.kubernetes.io/node": "",
	}
	expected := map[string]string{
		"KubernetesCluster":                                "example.com",
		"k8s.io/role/node":                                 "1",
		"kops.k8s.io/instancegroup":                        "nodes",
		"node-template/label/karpenter.sh/nodepool":        "nodes",
		"node-template/label/node-role.kubernetes.io/node": "",
	}

	actual := karpenterEC2NodeClassTags(tags)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}

func TestKarpenterNodePoolTemplateLabels(t *testing.T) {
	labels := map[string]string{
		"kops.k8s.io/instancegroup":             "nodes",
		"karpenter.sh/nodepool":                 "nodes",
		"karpenter.k8s.aws/ec2nodeclass":        "nodes",
		"kubernetes.io/hostname":                "ip-10-0-0-1",
		"node-role.kubernetes.io/node":          "",
		"node-restriction.kubernetes.io/worker": "true",
		"example.com/team":                      "platform",
		"team.kubernetes.io/owner":              "platform",
	}
	expected := map[string]string{
		"kops.k8s.io/instancegroup":             "nodes",
		"node-role.kubernetes.io/node":          "",
		"node-restriction.kubernetes.io/worker": "true",
		"example.com/team":                      "platform",
		"team.kubernetes.io/owner":              "platform",
	}

	actual := karpenterNodePoolTemplateLabels(labels)
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected %#v, got %#v", expected, actual)
	}
}

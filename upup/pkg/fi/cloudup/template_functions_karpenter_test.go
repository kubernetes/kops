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

	"github.com/google/go-cmp/cmp"

	"k8s.io/apimachinery/pkg/api/resource"
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

func TestKarpenterRequirements(t *testing.T) {
	osRequirement := karpenterRequirement{Key: "kubernetes.io/os", Operator: "In", Values: []string{"linux"}}
	onDemand := karpenterRequirement{Key: "karpenter.sh/capacity-type", Operator: "In", Values: []string{"on-demand"}}
	quantity := func(s string) *resource.Quantity {
		q := resource.MustParse(s)
		return &q
	}

	grid := []struct {
		desc        string
		spec        kops.InstanceGroupSpec
		expected    []karpenterRequirement
		expectError bool
	}{
		{
			// Regression guard: without instanceRequirements the output must not move.
			desc: "no instanceRequirements keeps the instance type list",
			spec: kops.InstanceGroupSpec{MachineType: "t3.medium"},
			expected: []karpenterRequirement{
				osRequirement,
				{Key: "node.kubernetes.io/instance-type", Operator: "In", Values: []string{"t3.medium"}},
				onDemand,
			},
		},
		{
			desc: "cpu bounds",
			spec: kops.InstanceGroupSpec{
				MachineType: "t3.medium",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						CPU: &kops.MinMaxSpec{Min: quantity("2"), Max: quantity("16")},
					},
				},
			},
			expected: []karpenterRequirement{
				osRequirement,
				{Key: "karpenter.k8s.aws/instance-cpu", Operator: "Gte", Values: []string{"2"}},
				{Key: "karpenter.k8s.aws/instance-cpu", Operator: "Lte", Values: []string{"16"}},
				onDemand,
			},
		},
		{
			desc: "memory bounds",
			spec: kops.InstanceGroupSpec{
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						Memory: &kops.MinMaxSpec{Min: quantity("2Gi"), Max: quantity("64Gi")},
					},
				},
			},
			expected: []karpenterRequirement{
				osRequirement,
				{Key: "karpenter.k8s.aws/instance-memory", Operator: "Gte", Values: []string{"2048"}},
				{Key: "karpenter.k8s.aws/instance-memory", Operator: "Lte", Values: []string{"65536"}},
				onDemand,
			},
		},
		{
			// instanceRequirements describes a capacity envelope, so an instance type
			// requirement would defeat it by pinning the pool to spec.machineType.
			desc: "instanceRequirements suppress the instance type list",
			spec: kops.InstanceGroupSpec{
				MachineType: "t3.medium",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					Instances: []string{"m6i.large"},
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						CPU: &kops.MinMaxSpec{Min: quantity("2")},
					},
				},
			},
			expected: []karpenterRequirement{
				osRequirement,
				{Key: "karpenter.k8s.aws/instance-cpu", Operator: "Gte", Values: []string{"2"}},
				onDemand,
			},
		},
		{
			desc: "excluded instance types and families",
			spec: kops.InstanceGroupSpec{
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						ExcludedInstanceTypes: []string{"m3.*", "t2.nano", "c3.*", "t2.micro"},
					},
				},
			},
			expected: []karpenterRequirement{
				osRequirement,
				{Key: "karpenter.k8s.aws/instance-family", Operator: "NotIn", Values: []string{"c3", "m3"}},
				{Key: "node.kubernetes.io/instance-type", Operator: "NotIn", Values: []string{"t2.micro", "t2.nano"}},
				onDemand,
			},
		},
		{
			desc: "unsupported wildcard is rejected",
			spec: kops.InstanceGroupSpec{
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						ExcludedInstanceTypes: []string{"m5.*large"},
					},
				},
			},
			expectError: true,
		},
		{
			desc: "negative quantity is rejected",
			spec: kops.InstanceGroupSpec{
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						CPU: &kops.MinMaxSpec{Min: quantity("-1")},
					},
				},
			},
			expectError: true,
		},
		{
			desc: "overflowing quantity is rejected",
			spec: kops.InstanceGroupSpec{
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					InstanceRequirements: &kops.InstanceRequirementsSpec{
						CPU: &kops.MinMaxSpec{Max: quantity("5000000000")},
					},
				},
			},
			expectError: true,
		},
	}

	tf := &TemplateFunctions{}
	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			actual, err := tf.karpenterRequirements(&kops.InstanceGroup{Spec: g.spec})
			if g.expectError {
				if err == nil {
					t.Fatalf("expected an error, got %#v", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(g.expected, actual); diff != "" {
				t.Errorf("unexpected requirements, diff=%v", diff)
			}
		})
	}
}

func TestKarpenterQuantityConversion(t *testing.T) {
	grid := []struct {
		desc        string
		quantity    string
		convert     func(*resource.Quantity) (int64, error)
		expected    int64
		expectError bool
	}{
		{desc: "cpu whole number", quantity: "4", convert: quantityToCount, expected: 4},
		{desc: "cpu fractional is rejected", quantity: "500m", convert: quantityToCount, expectError: true},
		// Binary suffixes are exact; decimal ones are not, so the rounding direction matters.
		{desc: "2Gi rounds to 2048", quantity: "2Gi", convert: quantityToMiBRoundUp, expected: 2048},
		{desc: "2G rounds up to 1908", quantity: "2G", convert: quantityToMiBRoundUp, expected: 1908},
		{desc: "2G rounds down to 1907", quantity: "2G", convert: quantityToMiBRoundDown, expected: 1907},
		{desc: "64Gi rounds to 65536", quantity: "64Gi", convert: quantityToMiBRoundDown, expected: 65536},
		{desc: "1000Mi is exact", quantity: "1000Mi", convert: quantityToMiBRoundUp, expected: 1000},
		{desc: "negative memory is rejected", quantity: "-1Gi", convert: quantityToMiBRoundUp, expectError: true},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			q := resource.MustParse(g.quantity)
			actual, err := g.convert(&q)
			if g.expectError {
				if err == nil {
					t.Fatalf("expected an error, got %d", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != g.expected {
				t.Errorf("expected %d, got %d", g.expected, actual)
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

func TestBuildKarpenterBlockDeviceMappings(t *testing.T) {
	// rootEBS returns the defaults every case starts from, so that each case below
	// only has to state what it changes.
	rootEBS := func(mutators ...func(*karpenterBlockDeviceEBS)) *karpenterBlockDeviceEBS {
		ebs := &karpenterBlockDeviceEBS{
			DeleteOnTermination: new(true),
			Encrypted:           new(true),
			IOPS:                new(int64(3000)),
			Throughput:          new(int64(125)),
			VolumeSize:          "128Gi",
			VolumeType:          "gp3",
		}
		for _, mutate := range mutators {
			mutate(ebs)
		}
		return ebs
	}

	grid := []struct {
		desc        string
		role        kops.InstanceGroupRole
		rootVolume  *kops.InstanceRootVolumeSpec
		expected    *karpenterBlockDeviceEBS
		expectError bool
	}{
		{
			desc:     "no rootVolume uses the kOps defaults",
			expected: rootEBS(),
		},
		{
			desc:       "empty rootVolume uses the kOps defaults",
			rootVolume: &kops.InstanceRootVolumeSpec{},
			expected:   rootEBS(),
		},
		{
			desc:       "explicit size",
			rootVolume: &kops.InstanceRootVolumeSpec{Size: new(int32(200))},
			expected:   rootEBS(func(e *karpenterBlockDeviceEBS) { e.VolumeSize = "200Gi" }),
		},
		{
			desc:       "zero size falls back to the role default",
			rootVolume: &kops.InstanceRootVolumeSpec{Size: new(int32(0))},
			expected:   rootEBS(),
		},
		{
			desc:     "control plane role default size",
			role:     kops.InstanceGroupRoleControlPlane,
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) { e.VolumeSize = "64Gi" }),
		},
		{
			desc:       "gp2 drops iops and throughput",
			rootVolume: &kops.InstanceRootVolumeSpec{Type: new("gp2"), IOPS: new(int32(4000)), Throughput: new(int32(200))},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) {
				e.VolumeType = "gp2"
				e.IOPS = nil
				e.Throughput = nil
			}),
		},
		{
			desc:       "standard drops iops and throughput",
			rootVolume: &kops.InstanceRootVolumeSpec{Type: new("standard")},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) {
				e.VolumeType = "standard"
				e.IOPS = nil
				e.Throughput = nil
			}),
		},
		{
			desc:       "io2 below the iops floor is raised",
			rootVolume: &kops.InstanceRootVolumeSpec{Type: new("io2"), IOPS: new(int32(50))},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) {
				e.VolumeType = "io2"
				e.IOPS = new(int64(100))
				e.Throughput = nil
			}),
		},
		{
			desc:       "io1 above the iops floor is kept",
			rootVolume: &kops.InstanceRootVolumeSpec{Type: new("io1"), IOPS: new(int32(5000)), Throughput: new(int32(200))},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) {
				e.VolumeType = "io1"
				e.IOPS = new(int64(5000))
				e.Throughput = nil
			}),
		},
		{
			desc:       "gp3 below the floors is raised",
			rootVolume: &kops.InstanceRootVolumeSpec{Type: new("gp3"), IOPS: new(int32(100)), Throughput: new(int32(50))},
			expected:   rootEBS(),
		},
		{
			desc:       "gp3 above the floors is kept",
			rootVolume: &kops.InstanceRootVolumeSpec{Type: new("gp3"), IOPS: new(int32(6000)), Throughput: new(int32(250))},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) {
				e.IOPS = new(int64(6000))
				e.Throughput = new(int64(250))
			}),
		},
		{
			desc: "encryption disabled drops the key",
			rootVolume: &kops.InstanceRootVolumeSpec{
				Encryption:    new(false),
				EncryptionKey: new("arn:aws:kms:us-test-1:123456789012:key/test"),
			},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) { e.Encrypted = new(false) }),
		},
		{
			desc: "encryption enabled keeps the key",
			rootVolume: &kops.InstanceRootVolumeSpec{
				Encryption:    new(true),
				EncryptionKey: new("arn:aws:kms:us-test-1:123456789012:key/test"),
			},
			expected: rootEBS(func(e *karpenterBlockDeviceEBS) {
				e.KMSKeyID = new("arn:aws:kms:us-test-1:123456789012:key/test")
			}),
		},
		{
			// Matches the ASG launch template path, which only honours the key when
			// encryption is set explicitly rather than defaulted.
			desc:       "encryption key without explicit encryption is ignored",
			rootVolume: &kops.InstanceRootVolumeSpec{EncryptionKey: new("arn:aws:kms:us-test-1:123456789012:key/test")},
			expected:   rootEBS(),
		},
		{
			// The Karpenter IAM policy grants KMS access on the same condition, so an
			// empty key must be dropped here rather than emitted as kmsKeyID: "".
			desc: "empty encryption key is dropped",
			rootVolume: &kops.InstanceRootVolumeSpec{
				Encryption:    new(true),
				EncryptionKey: new(""),
			},
			expected: rootEBS(),
		},
		{
			desc:        "unknown role",
			role:        kops.InstanceGroupRole("Bogus"),
			expectError: true,
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			role := g.role
			if role == "" {
				role = kops.InstanceGroupRoleNode
			}
			ig := &kops.InstanceGroup{
				Spec: kops.InstanceGroupSpec{
					Role:       role,
					RootVolume: g.rootVolume,
				},
			}
			actual, err := buildKarpenterBlockDeviceMappings(ig, "/dev/xvda")
			if g.expectError {
				if err == nil {
					t.Fatalf("expected an error, got %#v", actual)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			expected := []karpenterBlockDeviceMapping{
				{
					DeviceName: "/dev/xvda",
					EBS:        g.expected,
					RootVolume: new(true),
				},
			}
			if diff := cmp.Diff(expected, actual); diff != "" {
				t.Errorf("unexpected blockDeviceMappings, diff=%v", diff)
			}
		})
	}
}

func TestKarpenterRootDeviceNameWithoutAWSCloud(t *testing.T) {
	tf := &TemplateFunctions{}
	if _, err := tf.karpenterRootDeviceName("ami-12345678"); err == nil {
		t.Errorf("expected an error when the cloud is not an AWS cloud")
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

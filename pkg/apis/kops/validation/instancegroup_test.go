/*
Copyright 2017 The Kubernetes Authors.

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

package validation

import (
	"testing"

	"k8s.io/kops/pkg/nodeidentity/aws"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func s(v string) *string {
	return fi.PtrTo(v)
}

func TestValidateInstanceProfile(t *testing.T) {
	grid := []struct {
		Input          *kops.IAMProfileSpec
		ExpectedErrors []string
		ExpectedDetail string
	}{
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws:iam::123456789012:instance-profile/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws:iam::123456789012:instance-profile/has/path/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws-cn:iam::123456789012:instance-profile/has/path/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws-us-gov:iam::123456789012:instance-profile/has/path/S3Access"),
			},
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("42"),
			},
			ExpectedErrors: []string{"Invalid value::iam.profile"},
			ExpectedDetail: "Instance Group IAM Instance Profile must be a valid aws arn such as arn:aws:iam::123456789012:instance-profile/KopsExampleRole",
		},
		{
			Input: &kops.IAMProfileSpec{
				Profile: s("arn:aws:iam::123456789012:group/division_abc/subdivision_xyz/product_A/Developers"),
			},
			ExpectedErrors: []string{"Invalid value::iam.profile"},
			ExpectedDetail: "Instance Group IAM Instance Profile must be a valid aws arn such as arn:aws:iam::123456789012:instance-profile/KopsExampleRole",
		},
	}

	for _, g := range grid {
		allErrs := validateInstanceProfile(g.Input, field.NewPath("iam"))
		testErrors(t, g.Input, allErrs, g.ExpectedErrors)

		if g.ExpectedDetail != "" {
			found := false
			for _, err := range allErrs {
				if err.Detail == g.ExpectedDetail {
					found = true
				}
			}
			if !found {
				for _, err := range allErrs {
					t.Logf("found detail: %q", err.Detail)
				}

				t.Errorf("did not find expected error %q", g.ExpectedDetail)
			}
		}
	}
}

func TestValidMasterInstanceGroup(t *testing.T) {
	grid := []struct {
		Cluster        *kops.Cluster
		IG             *kops.InstanceGroup
		ExpectedErrors int
		Description    string
	}{
		{
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Name: "main",
							Members: []kops.EtcdMemberSpec{
								{
									Name:          "a",
									InstanceGroup: fi.PtrTo("eu-central-1a"),
								},
								{
									Name:          "b",
									InstanceGroup: fi.PtrTo("eu-central-1b"),
								},
								{
									Name:          "c",
									InstanceGroup: fi.PtrTo("eu-central-1c"),
								},
							},
						},
					},
				},
			},
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleControlPlane,
				},
			},
			ExpectedErrors: 0,
			Description:    "Valid instance group failed to validate",
		},
		{
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					EtcdClusters: []kops.EtcdClusterSpec{
						{
							Name: "main",
							Members: []kops.EtcdMemberSpec{
								{
									Name:          "a",
									InstanceGroup: fi.PtrTo("eu-central-1a"),
								},
								{
									Name:          "b",
									InstanceGroup: fi.PtrTo("eu-central-1b"),
								},
								{
									Name:          "c",
									InstanceGroup: fi.PtrTo("eu-central-1c"),
								},
							},
						},
					},
				},
			},
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1d",
				},
				Spec: kops.InstanceGroupSpec{
					Role: kops.InstanceGroupRoleControlPlane,
				},
			},
			ExpectedErrors: 1,
			Description:    "Master IG without etcd member validated",
		},
	}

	for _, g := range grid {
		errList := ValidateControlPlaneInstanceGroup(g.IG, g.Cluster)
		if len(errList) != g.ExpectedErrors {
			t.Error(g.Description)
		}
	}
}

func TestValidBootDevice(t *testing.T) {
	cluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
		},
	}
	grid := []struct {
		volumeType string
		expected   []string
	}{
		{
			volumeType: "standard",
		},
		{
			volumeType: "gp3",
		},
		{
			volumeType: "gp2",
		},
		{
			volumeType: "io1",
		},
		{
			volumeType: "io2",
		},
		{
			volumeType: "st1",
			expected:   []string{"Unsupported value::spec.rootVolume.type"},
		},
		{
			volumeType: "sc1",
			expected:   []string{"Unsupported value::spec.rootVolume.type"},
		},
	}

	for _, g := range grid {
		ig := createMinimalInstanceGroup()
		ig.Spec.RootVolume = &kops.InstanceRootVolumeSpec{
			Type: fi.PtrTo(g.volumeType),
		}
		errs := CrossValidateInstanceGroup(ig, cluster, nil, true)
		testErrors(t, g.volumeType, errs, g.expected)
	}
}

func TestValidNodeLabels(t *testing.T) {
	grid := []struct {
		label    string
		expected []string
	}{
		{
			label: "foo",
		},
		{
			label: "subdomain.domain.tld",
		},
		{
			label: "subdomain.domain.tld/foo",
		},
		{
			label:    "subdomain.domain.tld/foo/bar",
			expected: []string{"Invalid value::spec.nodeLabels"},
		},
	}

	for _, g := range grid {

		ig := createMinimalInstanceGroup()
		ig.Spec.NodeLabels = make(map[string]string)
		ig.Spec.NodeLabels[g.label] = "placeholder"
		errs := ValidateInstanceGroup(ig, nil, true)
		testErrors(t, g.label, errs, g.expected)
	}
}

func TestCrossValidateKarpenterInstanceGroup(t *testing.T) {
	awsCluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				AWS: &kops.AWSSpec{},
			},
		},
	}
	gceCluster := &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider: kops.CloudProviderSpec{
				GCE: &kops.GCESpec{},
			},
		},
	}

	grid := []struct {
		desc     string
		cluster  *kops.Cluster
		role     kops.InstanceGroupRole
		image    string
		expected []string
	}{
		{
			desc:    "ami id",
			cluster: awsCluster,
			role:    kops.InstanceGroupRoleNode,
			image:   "ami-0123456789abcdef0",
		},
		{
			desc:    "ssm parameter",
			cluster: awsCluster,
			role:    kops.InstanceGroupRoleNode,
			image:   "ssm:/aws/service/canonical/ubuntu/server/24.04/stable/current/amd64/hvm/ebs-gp3/ami-id",
		},
		{
			desc:    "name",
			cluster: awsCluster,
			role:    kops.InstanceGroupRoleNode,
			image:   "kops-node-image",
		},
		{
			desc:    "owner and name",
			cluster: awsCluster,
			role:    kops.InstanceGroupRoleNode,
			image:   "ubuntu/images/hvm-ssd/ubuntu-noble-24.04-amd64-server-*",
		},
		{
			desc:     "not aws",
			cluster:  gceCluster,
			role:     kops.InstanceGroupRoleNode,
			image:    "ami-0123456789abcdef0",
			expected: []string{"Forbidden::spec.manager"},
		},
		{
			desc:     "not node",
			cluster:  awsCluster,
			role:     kops.InstanceGroupRoleAPIServer,
			image:    "ami-0123456789abcdef0",
			expected: []string{"Forbidden::spec.role"},
		},
		{
			desc:     "url image",
			cluster:  awsCluster,
			role:     kops.InstanceGroupRoleNode,
			image:    "https://example.com/image",
			expected: []string{"Invalid value::spec.image"},
		},
		{
			desc:     "empty ssm parameter",
			cluster:  awsCluster,
			role:     kops.InstanceGroupRoleNode,
			image:    "ssm:",
			expected: []string{"Invalid value::spec.image"},
		},
		{
			desc:     "missing owner",
			cluster:  awsCluster,
			role:     kops.InstanceGroupRoleNode,
			image:    "/missing-owner",
			expected: []string{"Invalid value::spec.image"},
		},
		{
			desc:     "missing name",
			cluster:  awsCluster,
			role:     kops.InstanceGroupRoleNode,
			image:    "missing-name/",
			expected: []string{"Invalid value::spec.image"},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			ig := createMinimalInstanceGroup()
			ig.Spec.Manager = kops.InstanceManagerKarpenter
			ig.Spec.Role = g.role
			ig.Spec.Image = g.image

			errs := CrossValidateInstanceGroup(ig, g.cluster, nil, true)
			testErrors(t, g.desc, errs, g.expected)
		})
	}
}

func TestValidateKarpenterStaticCapacity(t *testing.T) {
	grid := []struct {
		desc         string
		minSize      *int32
		maxSize      *int32
		featureGates string
		expected     []string
	}{
		{
			desc:         "dynamic",
			featureGates: "StaticCapacity=false",
		},
		{
			desc:    "static with default feature gates",
			minSize: fi.PtrTo(int32(4)),
		},
		{
			desc:         "static with custom feature gates",
			minSize:      fi.PtrTo(int32(4)),
			featureGates: "NodeRepair=false,StaticCapacity=true",
		},
		{
			desc:         "static with final feature gate enabled",
			minSize:      fi.PtrTo(int32(4)),
			featureGates: "StaticCapacity=false,StaticCapacity=true",
		},
		{
			desc:     "zero minSize",
			minSize:  fi.PtrTo(int32(0)),
			expected: []string{"Invalid value::spec.minSize"},
		},
		{
			desc:     "negative minSize",
			minSize:  fi.PtrTo(int32(-1)),
			expected: []string{"Invalid value::spec.minSize"},
		},
		{
			desc:         "custom feature gates omit static capacity",
			minSize:      fi.PtrTo(int32(4)),
			featureGates: "NodeRepair=true",
			expected:     []string{"Forbidden::spec.minSize"},
		},
		{
			desc:         "whitespace feature gates",
			minSize:      fi.PtrTo(int32(4)),
			featureGates: " ",
			expected:     []string{"Forbidden::spec.minSize"},
		},
		{
			desc:         "static capacity disabled",
			minSize:      fi.PtrTo(int32(4)),
			featureGates: "StaticCapacity=false",
			expected:     []string{"Forbidden::spec.minSize"},
		},
		{
			desc:         "final feature gate disabled",
			minSize:      fi.PtrTo(int32(4)),
			featureGates: "StaticCapacity=true,StaticCapacity=false",
			expected:     []string{"Forbidden::spec.minSize"},
		},
		{
			desc:    "maxSize for dynamic",
			maxSize: fi.PtrTo(int32(4)),
		},
		{
			desc:    "maxSize for static",
			minSize: fi.PtrTo(int32(4)),
			maxSize: fi.PtrTo(int32(4)),
		},
		{
			desc:     "zero maxSize",
			maxSize:  fi.PtrTo(int32(0)),
			expected: []string{"Invalid value::spec.maxSize"},
		},
		{
			desc:     "negative maxSize",
			maxSize:  fi.PtrTo(int32(-1)),
			expected: []string{"Invalid value::spec.maxSize"},
		},
	}

	for _, g := range grid {
		t.Run(g.desc, func(t *testing.T) {
			cluster := &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
					Karpenter: &kops.KarpenterConfig{
						Enabled:      true,
						FeatureGates: g.featureGates,
					},
				},
			}
			ig := &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-ig",
				},
				Spec: kops.InstanceGroupSpec{
					Manager: kops.InstanceManagerKarpenter,
					Role:    kops.InstanceGroupRoleNode,
					Image:   "my-image",
					MinSize: g.minSize,
					MaxSize: g.maxSize,
				},
			}

			errs := CrossValidateInstanceGroup(ig, cluster, nil, true)
			testErrors(t, g.desc, errs, g.expected)
		})
	}
}

func TestValidateIGCloudLabels(t *testing.T) {
	grid := []struct {
		label    string
		expected []string
	}{
		{
			label: "k8s.io/cluster-autoscaler/test.example.com",
		},
		{
			label:    "KubernetesCluster",
			expected: []string{"Forbidden::spec.cloudLabels.KubernetesCluster"},
		},
		{
			label: "MyBillingLabel",
		},
		{
			label: "subdomain.domain.tld/foo/bar",
		},
	}

	for _, g := range grid {
		ig := createMinimalInstanceGroup()

		ig.Spec.CloudLabels[g.label] = "placeholder"
		errs := ValidateInstanceGroup(ig, nil, true)
		testErrors(t, g.label, errs, g.expected)
	}
}

func TestIGCloudLabelIsIGName(t *testing.T) {
	grid := []struct {
		label    string
		expected []string
	}{
		{
			label: "some-ig",
		},
		{
			label:    "not-some-ig",
			expected: []string{"Invalid value::spec.cloudLabels.kops.k8s.io/instancegroup"},
		},
	}

	for _, g := range grid {
		ig := createMinimalInstanceGroup()

		ig.Spec.CloudLabels[aws.CloudTagInstanceGroupName] = g.label
		errs := ValidateInstanceGroup(ig, nil, true)
		testErrors(t, g.label, errs, g.expected)
	}
}

func TestValidTaints(t *testing.T) {
	grid := []struct {
		taints   []string
		expected []string
	}{
		{
			taints: []string{
				"nvidia.com/gpu:NoSchedule",
			},
		},
		{
			taints: []string{
				"nvidia.com/gpu:NoSchedule",
				"nvidia.com/gpu:NoExecute",
			},
		},
		{
			taints: []string{
				"nvidia.com/gpu:NoSchedule",
				"nvidia.com/gpu:NoSchedule",
			},
			expected: []string{"Duplicate value::spec.taints[1]"},
		},
	}

	for _, g := range grid {
		ig := createMinimalInstanceGroup()

		ig.Spec.Taints = g.taints
		errs := ValidateInstanceGroup(ig, nil, true)
		testErrors(t, g.taints, errs, g.expected)
	}
}

func TestIGUpdatePolicy(t *testing.T) {
	const unsupportedValueError = "Unsupported value::spec.updatePolicy"
	for _, test := range []struct {
		label    string
		policy   *string
		expected []string
	}{
		{
			label: "missing",
		},
		{
			label:  "automatic",
			policy: fi.PtrTo(kops.UpdatePolicyAutomatic),
		},
		{
			label:  "external",
			policy: fi.PtrTo(kops.UpdatePolicyExternal),
		},
		{
			label:    "empty",
			policy:   fi.PtrTo(""),
			expected: []string{unsupportedValueError},
		},
		{
			label:    "unknown",
			policy:   fi.PtrTo("something-else"),
			expected: []string{unsupportedValueError},
		},
	} {
		ig := createMinimalInstanceGroup()

		t.Run(test.label, func(t *testing.T) {
			ig.Spec.UpdatePolicy = test.policy
			errs := ValidateInstanceGroup(ig, nil, true)
			testErrors(t, test.label, errs, test.expected)
		})
	}
}

func TestValidateInstanceGroupGVisorWorkerOnly(t *testing.T) {
	for _, test := range []struct {
		name     string
		role     kops.InstanceGroupRole
		enabled  *bool
		expected []string
	}{
		{
			name:    "enabled on worker",
			role:    kops.InstanceGroupRoleNode,
			enabled: fi.PtrTo(true),
		},
		{
			name:     "enabled on control plane",
			role:     kops.InstanceGroupRoleControlPlane,
			enabled:  fi.PtrTo(true),
			expected: []string{"Forbidden::spec.containerd.gvisor"},
		},
		{
			name:     "enabled on apiserver",
			role:     kops.InstanceGroupRoleAPIServer,
			enabled:  fi.PtrTo(true),
			expected: []string{"Forbidden::spec.containerd.gvisor"},
		},
		{
			name:     "enabled on bastion",
			role:     kops.InstanceGroupRoleBastion,
			enabled:  fi.PtrTo(true),
			expected: []string{"Forbidden::spec.containerd.gvisor"},
		},
		{
			name:    "disabled on apiserver",
			role:    kops.InstanceGroupRoleAPIServer,
			enabled: fi.PtrTo(false),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			ig := createMinimalInstanceGroup()
			ig.Spec.Role = test.role
			ig.Spec.Subnets = []string{"eu-central-1a"}
			ig.Spec.Containerd = &kops.ContainerdConfig{
				GVisor: &kops.GVisorConfig{
					Enabled: test.enabled,
				},
			}

			errs := ValidateInstanceGroup(ig, nil, true)
			testErrors(t, test.name, errs, test.expected)
		})
	}
}

func TestValidInstanceGroup(t *testing.T) {
	grid := []struct {
		IG             *kops.InstanceGroup
		ExpectedErrors []string
		Description    string
	}{
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleControlPlane,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(1)),
					MinSize: fi.PtrTo(int32(1)),
					Image:   "my-image",
				},
			},
			ExpectedErrors: []string{},
			Description:    "Valid control-plane instance group failed to validate",
		},
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleAPIServer,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(1)),
					MinSize: fi.PtrTo(int32(1)),
					Image:   "my-image",
				},
			},
			ExpectedErrors: []string{},
			Description:    "Valid API Server instance group failed to validate",
		},
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleNode,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(1)),
					MinSize: fi.PtrTo(int32(1)),
					Image:   "my-image",
				},
			},
			ExpectedErrors: []string{},
			Description:    "Valid node instance group failed to validate",
		},
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleBastion,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(1)),
					MinSize: fi.PtrTo(int32(1)),
					Image:   "my-image",
				},
			},
			ExpectedErrors: []string{},
			Description:    "Valid bastion instance group failed to validate",
		},
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleBastion,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(1)),
					MinSize: fi.PtrTo(int32(1)),
				},
			},
			ExpectedErrors: []string{"Forbidden::spec.image"},
			Description:    "Valid instance group must have image set",
		},
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleControlPlane,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(2)),
					MinSize: fi.PtrTo(int32(2)),
					Image:   "my-image",
				},
			},
			ExpectedErrors: []string{
				"Invalid value::spec.minSize",
				"Invalid value::spec.maxSize",
			},
			Description: "Invalid control-plane instance group sizes",
		},
	}
	for _, g := range grid {
		errList := ValidateInstanceGroup(g.IG, nil, true)
		testErrors(t, g.Description, errList, g.ExpectedErrors)
	}
}

func createMinimalInstanceGroup() *kops.InstanceGroup {
	ig := &kops.InstanceGroup{
		ObjectMeta: v1.ObjectMeta{
			Name: "some-ig",
		},
		Spec: kops.InstanceGroupSpec{
			CloudLabels: make(map[string]string),
			Role:        "Node",
			MaxSize:     fi.PtrTo(int32(1)),
			MinSize:     fi.PtrTo(int32(1)),
			Image:       "my-image",
		},
	}
	return ig
}

func TestCrossValidateAPIServerRole(t *testing.T) {
	noneDNSTopology := &kops.TopologySpec{DNS: kops.DNSTypeNone}
	grid := []struct {
		Description    string
		Cluster        *kops.Cluster
		ExpectedErrors int
	}{
		{
			Description: "APIServer role allowed on AWS",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
				},
			},
			ExpectedErrors: 0,
		},
		{
			Description: "APIServer role allowed on GCE",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						GCE: &kops.GCESpec{},
					},
				},
			},
			ExpectedErrors: 0,
		},
		{
			Description: "APIServer role forbidden on GCE with dns=None",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						GCE: &kops.GCESpec{},
					},
					Networking: kops.NetworkingSpec{Topology: noneDNSTopology},
				},
			},
			ExpectedErrors: 1,
		},
		{
			Description: "APIServer role forbidden on AWS with dns=None",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						AWS: &kops.AWSSpec{},
					},
					Networking: kops.NetworkingSpec{Topology: noneDNSTopology},
				},
			},
			ExpectedErrors: 1,
		},
		{
			Description: "APIServer role forbidden on DO",
			Cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						DO: &kops.DOSpec{},
					},
				},
			},
			ExpectedErrors: 1,
		},
	}

	for _, g := range grid {
		t.Run(g.Description, func(t *testing.T) {
			ig := &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "apiserver",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleAPIServer,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.PtrTo(int32(1)),
					MinSize: fi.PtrTo(int32(1)),
					Image:   "my-image",
				},
			}
			g.Cluster.Spec.Networking.Subnets = []kops.ClusterSubnetSpec{
				{Name: "eu-central-1a", Region: "eu-central-1"},
			}
			errs := CrossValidateInstanceGroup(ig, g.Cluster, nil, true)
			if len(errs) != g.ExpectedErrors {
				t.Errorf("expected %d errors, got %d: %v", g.ExpectedErrors, len(errs), errs)
			}
		})
	}
}

func TestValidateInstanceGroupName(t *testing.T) {
	cases := []struct {
		name      string
		input     string
		wantError bool
	}{
		// Empty is rejected; the validator does not have a special-case bypass.
		// Callers that allow an empty name in a specific branch (e.g. the
		// CAPI synthesis path in kops-controller) must skip the call.
		{name: "empty", input: "", wantError: true},

		// Valid DNS1123 subdomain names.
		{name: "simple", input: "nodes", wantError: false},
		{name: "with hyphen", input: "nodes-us-east-1a", wantError: false},
		{name: "control plane", input: "control-plane-us-east-1a", wantError: false},
		{name: "fqdn style", input: "nodes.example.k8s.local", wantError: false},
		{name: "numeric", input: "nodes1", wantError: false},

		// Path-traversal payloads.
		{name: "parent traversal", input: "..", wantError: true},
		{name: "parent traversal with target", input: "../master-foo", wantError: true},
		{name: "embedded traversal", input: "nodes/../master-foo", wantError: true},
		{name: "forward slash", input: "nodes/foo", wantError: true},
		{name: "backslash", input: "nodes\\foo", wantError: true},
		{name: "absolute path", input: "/etc/passwd", wantError: true},

		// Other invalid DNS1123 subdomain inputs.
		{name: "leading dot", input: ".nodes", wantError: true},
		{name: "trailing dot", input: "nodes.", wantError: true},
		{name: "uppercase", input: "Nodes", wantError: true},
		{name: "whitespace", input: "nodes foo", wantError: true},
		{name: "null byte", input: "nodes\x00", wantError: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			errs := ValidateInstanceGroupName(tc.input, field.NewPath("name"))
			if tc.wantError {
				if len(errs) == 0 {
					t.Fatalf("expected errors for input %q, got none", tc.input)
				}
			} else if len(errs) > 0 {
				t.Fatalf("unexpected errors for input %q: %v", tc.input, errs.ToAggregate())
			}
		})
	}
}

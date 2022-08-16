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
	return fi.String(v)
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
									InstanceGroup: fi.String("eu-central-1a"),
								},
								{
									Name:          "b",
									InstanceGroup: fi.String("eu-central-1b"),
								},
								{
									Name:          "c",
									InstanceGroup: fi.String("eu-central-1c"),
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
					Role: kops.InstanceGroupRoleMaster,
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
									InstanceGroup: fi.String("eu-central-1a"),
								},
								{
									Name:          "b",
									InstanceGroup: fi.String("eu-central-1b"),
								},
								{
									Name:          "c",
									InstanceGroup: fi.String("eu-central-1c"),
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
					Role: kops.InstanceGroupRoleMaster,
				},
			},
			ExpectedErrors: 1,
			Description:    "Master IG without etcd member validated",
		},
	}

	for _, g := range grid {
		errList := ValidateMasterInstanceGroup(g.IG, g.Cluster)
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
			expected:   []string{"Unsupported value::spec.rootVolumeType"},
		},
		{
			volumeType: "sc1",
			expected:   []string{"Unsupported value::spec.rootVolumeType"},
		},
	}

	for _, g := range grid {
		ig := createMinimalInstanceGroup()
		ig.Spec.RootVolumeType = fi.String(g.volumeType)
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
			policy: fi.String(kops.UpdatePolicyAutomatic),
		},
		{
			label:  "external",
			policy: fi.String(kops.UpdatePolicyExternal),
		},
		{
			label:    "empty",
			policy:   fi.String(""),
			expected: []string{unsupportedValueError},
		},
		{
			label:    "unknown",
			policy:   fi.String("something-else"),
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
					Role:    kops.InstanceGroupRoleMaster,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.Int32(1),
					MinSize: fi.Int32(1),
					Image:   "my-image",
				},
			},
			ExpectedErrors: []string{},
			Description:    "Valid master instance group failed to validate",
		},
		{
			IG: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "eu-central-1a",
				},
				Spec: kops.InstanceGroupSpec{
					Role:    kops.InstanceGroupRoleAPIServer,
					Subnets: []string{"eu-central-1a"},
					MaxSize: fi.Int32(1),
					MinSize: fi.Int32(1),
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
					MaxSize: fi.Int32(1),
					MinSize: fi.Int32(1),
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
					MaxSize: fi.Int32(1),
					MinSize: fi.Int32(1),
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
					MaxSize: fi.Int32(1),
					MinSize: fi.Int32(1),
				},
			},
			ExpectedErrors: []string{"Forbidden::spec.image"},
			Description:    "Valid instance group must have image set",
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
			MaxSize:     fi.Int32(1),
			MinSize:     fi.Int32(1),
			Image:       "my-image",
		},
	}
	return ig
}

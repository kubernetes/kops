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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
)

func TestAWSValidateExternalCloudConfig(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
				CloudConfig: &kops.CloudConfiguration{
					AWSEBSCSIDriver: &kops.AWSEBSCSIDriver{
						Enabled: fi.Bool(false),
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::spec.externalCloudControllerManager"},
		},
		{
			Input: kops.ClusterSpec{
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
				CloudConfig: &kops.CloudConfiguration{
					AWSEBSCSIDriver: &kops.AWSEBSCSIDriver{
						Enabled: fi.Bool(true),
					},
				},
			},
		},
		{
			Input: kops.ClusterSpec{
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
				KubeControllerManager: &kops.KubeControllerManagerConfig{
					ExternalCloudVolumePlugin: "aws",
				},
			},
		},
	}
	for _, g := range grid {
		g.Input.KubernetesVersion = "1.21.0"
		cluster := &kops.Cluster{
			Spec: g.Input,
		}
		errs := awsValidateExternalCloudControllerManager(cluster)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func TestValidateInstanceGroupSpec(t *testing.T) {
	grid := []struct {
		Input          kops.InstanceGroupSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"sg-1234abcd"},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"sg-1234abcd", ""},
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalSecurityGroups[1]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{" ", ""},
			},
			ExpectedErrors: []string{
				"Invalid value::spec.additionalSecurityGroups[0]",
				"Invalid value::spec.additionalSecurityGroups[1]",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				AdditionalSecurityGroups: []string{"--invalid"},
			},
			ExpectedErrors: []string{"Invalid value::spec.additionalSecurityGroups[0]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "t2.micro",
				Image:       "ami-073c8c0760395aab8",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "t2.invalidType",
				Image:       "ami-073c8c0760395aab8",
			},
			ExpectedErrors: []string{"Invalid value::test-nodes.spec.machineType"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "m4.large",
				Image:       "ami-073c8c0760395aab8",
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "c5.large",
				Image:       "ami-073c8c0760395aab8",
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "a1.large",
				Image:       "ami-073c8c0760395aab8",
			},
			ExpectedErrors: []string{
				"Invalid value::test-nodes.spec.machineType",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(55),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(380),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(125),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.spotDurationInMinutes",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				SpotDurationInMinutes: fi.Int64(120),
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("invalidValue"),
			},
			ExpectedErrors: []string{
				"Unsupported value::test-nodes.spec.instanceInterruptionBehavior",
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("terminate"),
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("hibernate"),
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.InstanceGroupSpec{
				InstanceInterruptionBehavior: fi.String("stop"),
			},
			ExpectedErrors: []string{},
		},
	}
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockEC2 := &mockec2.MockEC2{}
	cloud.MockEC2 = mockEC2

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2016-10-21T20:07:19.000Z"),
		ImageId:        aws.String("ami-073c8c0760395aab8"),
		Name:           aws.String("focal"),
		OwnerId:        aws.String(awsup.WellKnownAccountUbuntu),
		RootDeviceName: aws.String("/dev/xvda"),
		Architecture:   aws.String("x86_64"),
	})

	for _, g := range grid {
		ig := &kops.InstanceGroup{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-nodes",
			},
			Spec: g.Input,
		}
		errs := awsValidateInstanceGroup(ig, cloud)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

func TestMixedInstancePolicies(t *testing.T) {
	grid := []struct {
		Input          kops.InstanceGroupSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "m4.large",
				Image:       "ami-073c8c0760395aab8",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					Instances: []string{
						"m4.large",
						"t3.medium",
						"c5.large",
					},
				},
			},
			ExpectedErrors: nil,
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "m4.large",
				Image:       "ami-073c8c0760395aab8",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					Instances: []string{
						"a1.large",
						"c4.large",
						"c5.large",
					},
				},
			},
			ExpectedErrors: []string{"Invalid value::spec.mixedInstancesPolicy.instances[0]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "g4dn.xlarge",
				Image:       "ami-073c8c0760395aab8",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					Instances: []string{
						"g4dn.xlarge",
						"g4ad.16xlarge",
					},
				},
			},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "g4dn.xlarge",
				Image:       "ami-073c8c0760395aab8",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					Instances: []string{
						"g4dn.xlarge",
						"g4ad.16xlarge",
						"c4.xlarge",
					},
				},
			},
			ExpectedErrors: []string{"Forbidden::spec.mixedInstancesPolicy.instances[2]"},
		},
		{
			Input: kops.InstanceGroupSpec{
				MachineType: "m4.large",
				Image:       "ami-073c8c0760395aab8",
				MixedInstancesPolicy: &kops.MixedInstancesPolicySpec{
					Instances: []string{
						"t3.medium",
						"c4.large",
						"c5.large",
					},
					OnDemandAboveBase: fi.Int64(231),
				},
			},
			ExpectedErrors: []string{"Invalid value::spec.mixedInstancesPolicy.onDemandAboveBase"},
		},
	}
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")
	mockEC2 := &mockec2.MockEC2{}
	cloud.MockEC2 = mockEC2

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2016-10-21T20:07:19.000Z"),
		ImageId:        aws.String("ami-073c8c0760395aab8"),
		Name:           aws.String("focal"),
		OwnerId:        aws.String(awsup.WellKnownAccountUbuntu),
		RootDeviceName: aws.String("/dev/xvda"),
		Architecture:   aws.String("x86_64"),
	})

	for _, g := range grid {
		ig := &kops.InstanceGroup{
			ObjectMeta: v1.ObjectMeta{
				Name: "test-nodes",
			},
			Spec: g.Input,
		}
		errs := awsValidateInstanceGroup(ig, cloud)

		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}

}

func TestInstanceMetadataOptions(t *testing.T) {
	cloud := awsup.BuildMockAWSCloud("us-east-1", "abc")

	mockEC2 := &mockec2.MockEC2{}
	cloud.MockEC2 = mockEC2

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2016-10-21T20:07:19.000Z"),
		ImageId:        aws.String("ami-073c8c0760395aab8"),
		Name:           aws.String("focal"),
		OwnerId:        aws.String(awsup.WellKnownAccountUbuntu),
		RootDeviceName: aws.String("/dev/xvda"),
		Architecture:   aws.String("x86_64"),
	})

	tests := []struct {
		ig       *kops.InstanceGroup
		expected []string
	}{
		{
			ig: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-ig",
				},
				Spec: kops.InstanceGroupSpec{
					Role: "Node",
					InstanceMetadata: &kops.InstanceMetadataOptions{
						HTTPPutResponseHopLimit: fi.Int64(1),
						HTTPTokens:              fi.String("abc"),
					},
					MachineType: "t3.medium",
				},
			},
			expected: []string{"Unsupported value::spec.instanceMetadata.httpTokens"},
		},
		{
			ig: &kops.InstanceGroup{
				ObjectMeta: v1.ObjectMeta{
					Name: "some-ig",
				},
				Spec: kops.InstanceGroupSpec{
					Role: "Node",
					InstanceMetadata: &kops.InstanceMetadataOptions{
						HTTPPutResponseHopLimit: fi.Int64(-1),
						HTTPTokens:              fi.String("required"),
					},
					MachineType: "t3.medium",
				},
			},
			expected: []string{"Invalid value::spec.instanceMetadata.httpPutResponseHopLimit"},
		},
	}

	for _, test := range tests {
		errs := ValidateInstanceGroup(test.ig, cloud)
		testErrors(t, test.ig.ObjectMeta.Name, errs, test.expected)
	}
}

func TestLoadBalancerSubnets(t *testing.T) {
	cidr := "10.0.0.0/24"
	tests := []struct {
		lbType         *string
		class          *string
		clusterSubnets []string
		lbSubnets      []kops.LoadBalancerSubnetSpec
		expected       []string
	}{
		{ // valid (no privateIPv4Address, no allocationID)
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: nil,
					AllocationID:       nil,
				},
				{
					Name:               "b",
					PrivateIPv4Address: nil,
					AllocationID:       nil,
				},
			},
		},
		{ // valid (with privateIPv4Address)
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: fi.String("10.0.0.10"),
					AllocationID:       nil,
				},
				{
					Name:               "b",
					PrivateIPv4Address: nil,
					AllocationID:       nil,
				},
			},
		},
		{ // empty subnet name
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "",
					PrivateIPv4Address: nil,
					AllocationID:       nil,
				},
			},
			expected: []string{"Required value::spec.api.loadBalancer.subnets[0].name"},
		},
		{ // subnet not found
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "d",
					PrivateIPv4Address: nil,
					AllocationID:       nil,
				},
			},
			expected: []string{"Not found::spec.api.loadBalancer.subnets[0].name"},
		},
		{ // empty privateIPv4Address, no allocationID
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: fi.String(""),
					AllocationID:       nil,
				},
			},
			expected: []string{"Required value::spec.api.loadBalancer.subnets[0].privateIPv4Address"},
		},
		{ // empty no privateIPv4Address, with allocationID
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: nil,
					AllocationID:       fi.String(""),
				},
			},
			expected: []string{"Required value::spec.api.loadBalancer.subnets[0].allocationID"},
		},
		{ // invalid privateIPv4Address, no allocationID
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: fi.String("invalidip"),
					AllocationID:       nil,
				},
			},
			expected: []string{"Invalid value::spec.api.loadBalancer.subnets[0].privateIPv4Address"},
		},
		{ // privateIPv4Address not matching subnet cidr, no allocationID
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: fi.String("11.0.0.10"),
					AllocationID:       nil,
				},
			},
			expected: []string{"Invalid value::spec.api.loadBalancer.subnets[0].privateIPv4Address"},
		},
		{ // invalid class - with privateIPv4Address, no allocationID
			class:          fi.String(string(kops.LoadBalancerClassClassic)),
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: fi.String("10.0.0.10"),
					AllocationID:       nil,
				},
			},
			expected: []string{"Forbidden::spec.api.loadBalancer.subnets[0].privateIPv4Address"},
		},
		{ // invalid class - no privateIPv4Address, with allocationID
			class:          fi.String(string(kops.LoadBalancerClassClassic)),
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: nil,
					AllocationID:       fi.String("eipalloc-222ghi789"),
				},
			},
			expected: []string{"Forbidden::spec.api.loadBalancer.subnets[0].allocationID"},
		},
		{ // invalid type external for private IP
			lbType:         fi.String(string(kops.LoadBalancerTypePublic)),
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: fi.String("10.0.0.10"),
					AllocationID:       nil,
				},
			},
			expected: []string{"Forbidden::spec.api.loadBalancer.subnets[0].privateIPv4Address"},
		},
		{ // invalid type Internal for public IP
			lbType:         fi.String(string(kops.LoadBalancerTypeInternal)),
			clusterSubnets: []string{"a", "b", "c"},
			lbSubnets: []kops.LoadBalancerSubnetSpec{
				{
					Name:               "a",
					PrivateIPv4Address: nil,
					AllocationID:       fi.String("eipalloc-222ghi789"),
				},
			},
			expected: []string{"Forbidden::spec.api.loadBalancer.subnets[0].allocationID"},
		},
	}

	for _, test := range tests {
		cluster := kops.Cluster{
			Spec: kops.ClusterSpec{
				API: &kops.AccessSpec{
					LoadBalancer: &kops.LoadBalancerAccessSpec{
						Class: kops.LoadBalancerClassNetwork,
						Type:  kops.LoadBalancerTypeInternal,
					},
				},
			},
		}
		if test.class != nil {
			cluster.Spec.API.LoadBalancer.Class = kops.LoadBalancerClass(*test.class)
		}
		if test.lbType != nil {
			cluster.Spec.API.LoadBalancer.Type = kops.LoadBalancerType(*test.lbType)
		}
		for _, s := range test.clusterSubnets {
			cluster.Spec.Subnets = append(cluster.Spec.Subnets, kops.ClusterSubnetSpec{
				Name: s,
				CIDR: cidr,
			})
		}
		cluster.Spec.API.LoadBalancer.Subnets = test.lbSubnets
		errs := awsValidateCluster(&cluster)
		testErrors(t, test, errs, test.expected)
	}
}

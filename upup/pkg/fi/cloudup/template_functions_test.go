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
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
)

func Test_TemplateFunctions_CloudControllerConfigArgv(t *testing.T) {
	tests := []struct {
		desc          string
		cluster       *kops.Cluster
		expectedArgv  []string
		expectedError error
	}{
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{},
			}},
			expectedArgv: []string{
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Log Level Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{},
					},
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						LogLevel: 3,
					},
				},
			},
			expectedArgv: []string{
				"--v=3",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "ExternalCloudControllerManager CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						CloudProvider: string(kops.CloudProviderOpenstack),
						LogLevel:      3,
					},
				},
			},
			expectedArgv: []string{
				"--cloud-provider=openstack",
				"--v=3",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "No CloudProvider Configuration",
			cluster: &kops.Cluster{
				Spec: kops.ClusterSpec{
					ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
						LogLevel: 3,
					},
				},
			},
			expectedError: fmt.Errorf("Cloud Provider is not set"),
		},
		{
			desc: "k8s cluster name",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterName: "k8s",
				},
			}},
			expectedArgv: []string{
				"--cluster-name=k8s",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Default Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					Master: "127.0.0.1",
				},
			}},
			expectedArgv: []string{
				"--master=127.0.0.1",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Cluster-cidr Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ClusterCIDR: "10.0.0.0/24",
				},
			}},
			expectedArgv: []string{
				"--cluster-cidr=10.0.0.0/24",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "AllocateNodeCIDRs Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					AllocateNodeCIDRs: fi.PtrTo(true),
				},
			}},
			expectedArgv: []string{
				"--allocate-node-cidrs=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "ConfigureCloudRoutes Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					ConfigureCloudRoutes: fi.PtrTo(true),
				},
			}},
			expectedArgv: []string{
				"--configure-cloud-routes=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "CIDRAllocatorType Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					CIDRAllocatorType: fi.PtrTo("RangeAllocator"),
				},
			}},
			expectedArgv: []string{
				"--cidr-allocator-type=RangeAllocator",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "CIDRAllocatorType Configuration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					UseServiceAccountCredentials: fi.PtrTo(false),
				},
			}},
			expectedArgv: []string{
				"--use-service-account-credentials=false",
				"--v=2",
				"--cloud-provider=openstack",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Leader Election",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					LeaderElection: &kops.LeaderElectionConfiguration{LeaderElect: fi.PtrTo(true)},
				},
			}},
			expectedArgv: []string{
				"--leader-elect=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Leader Migration",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					LeaderElection:        &kops.LeaderElectionConfiguration{LeaderElect: fi.PtrTo(true)},
					EnableLeaderMigration: fi.PtrTo(true),
				},
			}},
			expectedArgv: []string{
				"--enable-leader-migration=true",
				"--leader-elect=true",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
		{
			desc: "Disable Controller",
			cluster: &kops.Cluster{Spec: kops.ClusterSpec{
				CloudProvider: kops.CloudProviderSpec{
					Openstack: &kops.OpenstackSpec{},
				},
				ExternalCloudControllerManager: &kops.CloudControllerManagerConfig{
					Controllers: []string{"*", "-nodeipam"},
				},
			}},
			expectedArgv: []string{
				"--controllers=*,-nodeipam",
				"--v=2",
				"--cloud-provider=openstack",
				"--use-service-account-credentials=true",
				"--cloud-config=/etc/kubernetes/cloud.config",
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.desc, func(t *testing.T) {
			tf := &TemplateFunctions{}
			tf.Cluster = testCase.cluster

			actual, error := tf.CloudControllerConfigArgv()
			if !reflect.DeepEqual(error, testCase.expectedError) {
				t.Errorf("Error differs: %+v instead of %+v", error, testCase.expectedError)
			}
			if !reflect.DeepEqual(actual, testCase.expectedArgv) {
				t.Errorf("Argv differs: %+v instead of %+v", actual, testCase.expectedArgv)
			}
		})
	}
}

func Test_KarpenterInstanceTypes(t *testing.T) {
	amiId := "ami-073c8c0760395aab8"
	ec2Client := &mockec2.MockEC2{}
	ec2Client.Images = append(ec2Client.Images, &ec2.Image{
		CreationDate:   aws.String("2016-10-21T20:07:19.000Z"),
		ImageId:        &amiId,
		Name:           aws.String("focal"),
		OwnerId:        aws.String(awsup.WellKnownAccountUbuntu),
		RootDeviceName: aws.String("/dev/xvda"),
		Architecture:   aws.String("x86_64"),
	})
	ig := kops.InstanceGroupSpec{
		Image: amiId,
	}
	cloud := &awsup.MockAWSCloud{MockCloud: awsup.MockCloud{
		MockEC2: ec2Client,
	}}
	_, err := karpenterInstanceTypes(cloud, ig)
	if err != nil {
		t.Errorf("failed to fetch instance types: %v", err)
	}
}

func TestKopsFeatureEnabled(t *testing.T) {
	tests := []struct {
		name          string
		featureFlags  string
		featureName   string
		expectedValue bool
		expectError   bool
	}{
		{
			name:         "Missing feature",
			featureFlags: "",
			featureName:  "NonExistingFeature",
			expectError:  true,
		},
		{
			name:          "Existing feature",
			featureFlags:  "+Scaleway",
			featureName:   "Scaleway",
			expectError:   false,
			expectedValue: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			featureflag.ParseFlags(tc.featureFlags)
			tf := &TemplateFunctions{}
			value, err := tf.kopsFeatureEnabled(tc.featureName)
			if err != nil && !tc.expectError {
				t.Errorf("unexpected error: %s", err)
			}
			if err == nil && tc.expectError {
				t.Errorf("expected error, got nil")
			}
			if value != tc.expectedValue {
				t.Errorf("expected value %t, got %t", tc.expectedValue, value)
			}
		})
	}
}

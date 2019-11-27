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

package testutils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"k8s.io/klog"
	kopsroot "k8s.io/kops"
	"k8s.io/kops/cloudmock/aws/mockautoscaling"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cloudmock/aws/mockelb"
	"k8s.io/kops/cloudmock/aws/mockelbv2"
	"k8s.io/kops/cloudmock/aws/mockiam"
	"k8s.io/kops/cloudmock/aws/mockroute53"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/pki"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/util/pkg/vfs"
)

type IntegrationTestHarness struct {
	TempDir string
	T       *testing.T

	// The original kops DefaultChannelBase value, restored on Close
	originalDefaultChannelBase string

	// originalKopsVersion is the original kops.Version value, restored on Close
	originalKopsVersion string

	// originalPKIDefaultPrivateKeySize is the saved pki.DefaultPrivateKeySize value, restored on Close
	originalPKIDefaultPrivateKeySize int
}

func NewIntegrationTestHarness(t *testing.T) *IntegrationTestHarness {
	h := &IntegrationTestHarness{}
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	h.TempDir = tempDir

	vfs.Context.ResetMemfsContext(true)

	// Generate much smaller keys, as this is often the bottleneck for tests
	h.originalPKIDefaultPrivateKeySize = pki.DefaultPrivateKeySize
	pki.DefaultPrivateKeySize = 512

	// Replace the default channel path with a local filesystem path, so we don't try to retrieve it from a server
	{
		channelPath, err := filepath.Abs(path.Join("../../channels/"))
		if err != nil {
			t.Fatalf("error resolving stable channel path: %v", err)
		}
		channelPath += "/"
		h.originalDefaultChannelBase = kops.DefaultChannelBase

		// Make sure any platform-specific separators that aren't /, are converted to / for use in a file: protocol URL
		kops.DefaultChannelBase = "file://" + filepath.ToSlash(channelPath)
	}

	return h
}

func (h *IntegrationTestHarness) Close() {
	if h.TempDir != "" {
		if os.Getenv("KEEP_TEMP_DIR") != "" {
			klog.Infof("NOT removing temp directory, because KEEP_TEMP_DIR is set: %s", h.TempDir)
		} else {
			err := os.RemoveAll(h.TempDir)
			if err != nil {
				h.T.Fatalf("failed to remove temp dir %q: %v", h.TempDir, err)
			}
		}
	}

	if h.originalKopsVersion != "" {
		kopsroot.Version = h.originalKopsVersion
	}

	if h.originalDefaultChannelBase != "" {
		kops.DefaultChannelBase = h.originalDefaultChannelBase
	}

	if h.originalPKIDefaultPrivateKeySize != 0 {
		pki.DefaultPrivateKeySize = h.originalPKIDefaultPrivateKeySize
	}
}

func (h *IntegrationTestHarness) SetupMockAWS() *awsup.MockAWSCloud {
	cloud := awsup.InstallMockAWSCloud("us-test-1", "abc")
	mockEC2 := &mockec2.MockEC2{}
	cloud.MockEC2 = mockEC2
	mockRoute53 := &mockroute53.MockRoute53{}
	cloud.MockRoute53 = mockRoute53
	mockELB := &mockelb.MockELB{}
	cloud.MockELB = mockELB
	mockELBV2 := &mockelbv2.MockELBV2{}
	cloud.MockELBV2 = mockELBV2
	mockIAM := &mockiam.MockIAM{}
	cloud.MockIAM = mockIAM
	mockAutoscaling := &mockautoscaling.MockAutoscaling{}
	cloud.MockAutoscaling = mockAutoscaling

	mockRoute53.MockCreateZone(&route53.HostedZone{
		Id:   aws.String("/hostedzone/Z1AFAKE1ZON3YO"),
		Name: aws.String("example.com."),
		Config: &route53.HostedZoneConfig{
			PrivateZone: aws.Bool(false),
		},
	}, nil)
	mockRoute53.MockCreateZone(&route53.HostedZone{
		Id:   aws.String("/hostedzone/Z2AFAKE1ZON3NO"),
		Name: aws.String("internal.example.com."),
		Config: &route53.HostedZoneConfig{
			PrivateZone: aws.Bool(true),
		},
	}, []*route53.VPC{{
		VPCId: aws.String("vpc-23456789"),
	}})
	mockRoute53.MockCreateZone(&route53.HostedZone{
		Id:   aws.String("/hostedzone/Z3AFAKE1ZOMORE"),
		Name: aws.String("private.example.com."),
		Config: &route53.HostedZoneConfig{
			PrivateZone: aws.Bool(true),
		},
	}, []*route53.VPC{{
		VPCId: aws.String("vpc-12345678"),
	}})

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2016-10-21T20:07:19.000Z"),
		ImageId:        aws.String("ami-12345678"),
		Name:           aws.String("k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-10-21"),
		OwnerId:        aws.String(awsup.WellKnownAccountKopeio),
		RootDeviceName: aws.String("/dev/xvda"),
	})

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2017-01-09T17:08:27.000Z"),
		ImageId:        aws.String("ami-15000000"),
		Name:           aws.String("k8s-1.5-debian-jessie-amd64-hvm-ebs-2017-01-09"),
		OwnerId:        aws.String(awsup.WellKnownAccountKopeio),
		RootDeviceName: aws.String("/dev/xvda"),
	})

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		CreationDate:   aws.String("2019-08-06T00:00:00.000Z"),
		ImageId:        aws.String("ami-11400000"),
		Name:           aws.String("k8s-1.14-debian-stretch-amd64-hvm-ebs-2019-08-16"),
		OwnerId:        aws.String(awsup.WellKnownAccountKopeio),
		RootDeviceName: aws.String("/dev/xvda"),
	})

	mockEC2.CreateVpcWithId(&ec2.CreateVpcInput{
		CidrBlock: aws.String("172.20.0.0/16"),
	}, "vpc-12345678")
	mockEC2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	mockEC2.AttachInternetGateway(&ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String("igw-1"),
		VpcId:             aws.String("vpc-12345678"),
	})

	mockEC2.CreateRouteTableWithId(&ec2.CreateRouteTableInput{
		VpcId: aws.String("vpc-12345678"),
	}, "rtb-12345678")

	mockEC2.CreateSubnetWithId(&ec2.CreateSubnetInput{
		VpcId:            aws.String("vpc-12345678"),
		AvailabilityZone: aws.String("us-test-1a"),
		CidrBlock:        aws.String("172.20.32.0/19"),
	}, "subnet-12345678")
	mockEC2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: aws.String("rtb-12345678"),
		SubnetId:     aws.String("subnet-12345678"),
	})
	mockEC2.CreateSubnetWithId(&ec2.CreateSubnetInput{
		VpcId:            aws.String("vpc-12345678"),
		AvailabilityZone: aws.String("us-test-1a"),
		CidrBlock:        aws.String("172.20.4.0/22"),
	}, "subnet-abcdef")
	mockEC2.CreateSubnetWithId(&ec2.CreateSubnetInput{
		VpcId:            aws.String("vpc-12345678"),
		AvailabilityZone: aws.String("us-test-1b"),
		CidrBlock:        aws.String("172.20.8.0/22"),
	}, "subnet-b2345678")

	mockEC2.AssociateRouteTable(&ec2.AssociateRouteTableInput{
		RouteTableId: aws.String("rtb-12345678"),
		SubnetId:     aws.String("subnet-abcdef"),
	})

	mockEC2.AllocateAddressWithId(&ec2.AllocateAddressInput{
		Address: aws.String("123.45.67.8"),
	}, "eipalloc-12345678")

	mockEC2.CreateNatGatewayWithId(&ec2.CreateNatGatewayInput{
		SubnetId:     aws.String("subnet-12345678"),
		AllocationId: aws.String("eipalloc-12345678"),
	}, "nat-a2345678")

	mockEC2.AllocateAddressWithId(&ec2.AllocateAddressInput{
		Address: aws.String("2.22.22.22"),
	}, "eipalloc-b2345678")

	mockEC2.CreateNatGatewayWithId(&ec2.CreateNatGatewayInput{
		SubnetId:     aws.String("subnet-b2345678"),
		AllocationId: aws.String("eipalloc-b2345678"),
	}, "nat-b2345678")

	return cloud
}

// SetupMockGCE configures a mock GCE cloud provider
func (h *IntegrationTestHarness) SetupMockGCE() {
	gce.InstallMockGCECloud("us-test1", "testproject")
}

// MockKopsVersion will set the kops version to the specified value, until Close is called
func (h *IntegrationTestHarness) MockKopsVersion(version string) {
	if h.originalKopsVersion != "" {
		h.T.Fatalf("MockKopsVersion called twice (%s and %s)", version, h.originalKopsVersion)
	}

	h.originalKopsVersion = kopsroot.Version
	kopsroot.Version = version
}

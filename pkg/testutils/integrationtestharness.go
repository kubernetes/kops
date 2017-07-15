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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/golang/glog"
	"io/ioutil"
	"k8s.io/kops/cloudmock/aws/mockec2"
	"k8s.io/kops/cloudmock/aws/mockroute53"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/util/pkg/vfs"
	"os"
	"testing"
)

type IntegrationTestHarness struct {
	TempDir string
	T       *testing.T
}

func NewIntegrationTestHarness(t *testing.T) *IntegrationTestHarness {
	h := &IntegrationTestHarness{}
	tempDir, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	h.TempDir = tempDir

	vfs.Context.ResetMemfsContext(true)

	return h
}

func (h *IntegrationTestHarness) Close() {
	if h.TempDir != "" {
		if os.Getenv("KEEP_TEMP_DIR") != "" {
			glog.Infof("NOT removing temp directory, because KEEP_TEMP_DIR is set: %s", h.TempDir)
		} else {
			err := os.RemoveAll(h.TempDir)
			if err != nil {
				h.T.Fatalf("failed to remove temp dir %q: %v", h.TempDir, err)
			}
		}
	}
}

func (h *IntegrationTestHarness) SetupMockAWS() {
	cloud := awsup.InstallMockAWSCloud("us-test-1", "abc")
	mockEC2 := &mockec2.MockEC2{}
	cloud.MockEC2 = mockEC2
	mockRoute53 := &mockroute53.MockRoute53{}
	cloud.MockRoute53 = mockRoute53

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
		VPCId: aws.String("vpc-234"),
	}})
	mockRoute53.MockCreateZone(&route53.HostedZone{
		Id:   aws.String("/hostedzone/Z3AFAKE1ZOMORE"),
		Name: aws.String("private.example.com."),
		Config: &route53.HostedZoneConfig{
			PrivateZone: aws.Bool(true),
		},
	}, []*route53.VPC{{
		VPCId: aws.String("vpc-123"),
	}})

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		ImageId:        aws.String("ami-12345678"),
		Name:           aws.String("k8s-1.4-debian-jessie-amd64-hvm-ebs-2016-10-21"),
		OwnerId:        aws.String(awsup.WellKnownAccountKopeio),
		RootDeviceName: aws.String("/dev/xvda"),
	})

	mockEC2.Images = append(mockEC2.Images, &ec2.Image{
		ImageId:        aws.String("ami-15000000"),
		Name:           aws.String("k8s-1.5-debian-jessie-amd64-hvm-ebs-2017-01-09"),
		OwnerId:        aws.String(awsup.WellKnownAccountKopeio),
		RootDeviceName: aws.String("/dev/xvda"),
	})
}

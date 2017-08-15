/*
Copyright 2016 The Kubernetes Authors.

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

package protokube

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Test_Find_ETCD_Volumes(t *testing.T) {
	awsVolumes := &AWSVolumes{
		instanceId: "i-0dc7301acf2dfbc0c",
	}

	p := &ec2.DescribeVolumesOutput{
		Volumes: []*ec2.Volume{
			{
				VolumeId: aws.String("vol-0c681fe311525b927"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("KubernetesCluster"),
						Value: aws.String("k8s.example.com"),
					},
					{
						Key:   aws.String("Name"),
						Value: aws.String("d.etcd-main.example.com"),
					},
					{
						Key:   aws.String("k8s.io/etcd/main"),
						Value: aws.String("d/d"),
					},
					{
						Key:   aws.String("k8s.io/role/master"),
						Value: aws.String("1"),
					},
				},
				Attachments: []*ec2.VolumeAttachment{
					{
						InstanceId: aws.String("i-0dc7301acf2dfbc0c"),
					},
				},
				VolumeType: aws.String("gp2"),
			},
			{
				VolumeId: aws.String("vol-0c681fe311525b926"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("KubernetesCluster"),
						Value: aws.String("k8s.example.com"),
					},
					{
						Key:   aws.String("Name"),
						Value: aws.String("master-us-east-1d.masters.k8s.example.com"),
					},
					{
						Key:   aws.String("k8s.io/role/master"),
						Value: aws.String("1"),
					},
				},
				Attachments: []*ec2.VolumeAttachment{
					{
						InstanceId: aws.String("i-0dc7301acf2dfbc0c"),
					},
				},
				VolumeType: aws.String("gp2"),
			},
			{
				VolumeId: aws.String("vol-0c681fe311525b928"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("KubernetesClusters"),
						Value: aws.String("k8s.example.com"),
					},
					{
						Key:   aws.String("NameFOO"),
						Value: aws.String("master-us-east-1d.masters.k8s.example.com"),
					},
					{
						Key:   aws.String("k8s.io/role/bar"),
						Value: aws.String("1"),
					},
				},
				Attachments: []*ec2.VolumeAttachment{
					{
						InstanceId: aws.String("i-0dc7301acf2dfbc0c"),
					},
				},
				VolumeType: aws.String("gp2"),
			},
			{
				VolumeId: aws.String("vol-0c681fe311525b926"),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("KubernetesCluster"),
						Value: aws.String("k8s.example.com"),
					},
					{
						Key:   aws.String("Name"),
						Value: aws.String("d.etcd-events.example.com"),
					},
					{
						Key:   aws.String("k8s.io/etcd/events"),
						Value: aws.String("d/d"),
					},
					{
						Key:   aws.String("k8s.io/role/master"),
						Value: aws.String("1"),
					},
				},
				Attachments: []*ec2.VolumeAttachment{
					{
						InstanceId: aws.String("i-0dc7301acf2dfbc0c"),
					},
				},
				VolumeType: aws.String("gp2"),
			},
		},
	}
	var volumes []*Volume
	volumes = awsVolumes.findEctdVolumes(p, volumes)

	if len(volumes) == 0 {
		t.Fatalf("volumes len should not be zero")
	}
	if len(volumes) != 2 {
		t.Fatalf("volumes len should be two")
	}

}

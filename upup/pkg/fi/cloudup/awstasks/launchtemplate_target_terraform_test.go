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

package awstasks

import (
	"testing"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"k8s.io/kops/upup/pkg/fi"
)

func TestLaunchTemplateTerraformRender(t *testing.T) {
	cases := []*renderTest{
		{
			Resource: &LaunchTemplate{
				Name:              new("test"),
				AssociatePublicIP: new(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: new("nodes"),
				},
				ID:                           new("test-11"),
				InstanceMonitoring:           new(true),
				InstanceType:                 new(ec2types.InstanceTypeT2Medium),
				SpotPrice:                    new("0.1"),
				SpotDurationInMinutes:        new(int32(60)),
				InstanceInterruptionBehavior: new(ec2types.InstanceInterruptionBehaviorHibernate),
				RootVolumeOptimization:       new(true),
				RootVolumeIops:               new(int32(100)),
				RootVolumeSize:               new(int32(64)),
				SSHKey: &SSHKey{
					Name:      new("newkey"),
					PublicKey: fi.NewStringResource("newkey"),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: new("nodes-1"), ID: new("1111")},
					{Name: new("nodes-2"), ID: new("2222")},
				},
				Tenancy:                 new(ec2types.TenancyDedicated),
				HTTPTokens:              new(ec2types.LaunchTemplateHttpTokensStateOptional),
				HTTPPutResponseHopLimit: new(int32(1)),
			},
			Expected: `provider "aws" {
  region = "eu-west-2"
}

resource "aws_launch_template" "test" {
  ebs_optimized = true
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes.id
  }
  instance_market_options {
    market_type = "spot"
    spot_options {
      block_duration_minutes         = 60
      instance_interruption_behavior = "hibernate"
      max_price                      = "0.1"
    }
  }
  instance_type = "t2.medium"
  key_name      = aws_key_pair.newkey.id
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_put_response_hop_limit = 1
    http_tokens                 = "optional"
  }
  monitoring {
    enabled = true
  }
  name = "test"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.nodes-1.id, aws_security_group.nodes-2.id]
  }
  placement {
    tenancy = "dedicated"
  }
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "source"  = "hashicorp/aws"
      "version" = ">= 5.0.0"
    }
  }
}
`,
		},
		{
			Resource: &LaunchTemplate{
				Name:              new("test"),
				AssociatePublicIP: new(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: new("nodes"),
				},
				BlockDeviceMappings: []*BlockDeviceMapping{
					{
						DeviceName:             new("/dev/xvdd"),
						EbsVolumeType:          ec2types.VolumeTypeGp2,
						EbsVolumeSize:          new(int32(100)),
						EbsDeleteOnTermination: new(true),
						EbsEncrypted:           new(true),
					},
				},
				ID:                     new("test-11"),
				InstanceMonitoring:     new(true),
				InstanceType:           new(ec2types.InstanceTypeT2Medium),
				RootVolumeOptimization: new(true),
				RootVolumeIops:         new(int32(100)),
				RootVolumeSize:         new(int32(64)),
				SSHKey: &SSHKey{
					Name: new("mykey"),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: new("nodes-1"), ID: new("1111")},
					{Name: new("nodes-2"), ID: new("2222")},
				},
				Tenancy:                 new(ec2types.TenancyDedicated),
				HTTPTokens:              new(ec2types.LaunchTemplateHttpTokensStateRequired),
				HTTPPutResponseHopLimit: new(int32(5)),
			},
			Expected: `provider "aws" {
  region = "eu-west-2"
}

resource "aws_launch_template" "test" {
  block_device_mappings {
    device_name = "/dev/xvdd"
    ebs {
      delete_on_termination = true
      encrypted             = true
      volume_size           = 100
      volume_type           = "gp2"
    }
  }
  ebs_optimized = true
  iam_instance_profile {
    name = aws_iam_instance_profile.nodes.id
  }
  instance_type = "t2.medium"
  key_name      = "mykey"
  lifecycle {
    create_before_destroy = true
  }
  metadata_options {
    http_endpoint               = "enabled"
    http_put_response_hop_limit = 5
    http_tokens                 = "required"
  }
  monitoring {
    enabled = true
  }
  name = "test"
  network_interfaces {
    associate_public_ip_address = true
    delete_on_termination       = true
    security_groups             = [aws_security_group.nodes-1.id, aws_security_group.nodes-2.id]
  }
  placement {
    tenancy = "dedicated"
  }
}

terraform {
  required_version = ">= 0.15.0"
  required_providers {
    aws = {
      "source"  = "hashicorp/aws"
      "version" = ">= 5.0.0"
    }
  }
}
`,
		},
	}
	doRenderTests(t, "RenderTerraform", cases)
}

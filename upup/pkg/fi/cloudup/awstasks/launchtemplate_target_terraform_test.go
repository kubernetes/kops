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

	"k8s.io/kops/pkg/testutils/testcontext"
	"k8s.io/kops/upup/pkg/fi"
)

func TestLaunchTemplateTerraformRender(t *testing.T) {
	ctx := testcontext.ForTest(t)

	cases := []*renderTest{
		{
			Resource: &LaunchTemplate{
				Name:              fi.PtrTo("test"),
				AssociatePublicIP: fi.PtrTo(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.PtrTo("nodes"),
				},
				ID:                           fi.PtrTo("test-11"),
				InstanceMonitoring:           fi.PtrTo(true),
				InstanceType:                 fi.PtrTo("t2.medium"),
				SpotPrice:                    fi.PtrTo("0.1"),
				SpotDurationInMinutes:        fi.PtrTo(int64(60)),
				InstanceInterruptionBehavior: fi.PtrTo("hibernate"),
				RootVolumeOptimization:       fi.PtrTo(true),
				RootVolumeIops:               fi.PtrTo(int64(100)),
				RootVolumeSize:               fi.PtrTo(int64(64)),
				SSHKey: &SSHKey{
					Name:      fi.PtrTo("newkey"),
					PublicKey: fi.NewStringResource("newkey"),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: fi.PtrTo("nodes-1"), ID: fi.PtrTo("1111")},
					{Name: fi.PtrTo("nodes-2"), ID: fi.PtrTo("2222")},
				},
				Tenancy:                 fi.PtrTo("dedicated"),
				HTTPTokens:              fi.PtrTo("optional"),
				HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
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
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 4.0.0"
    }
  }
}
`,
		},
		{
			Resource: &LaunchTemplate{
				Name:              fi.PtrTo("test"),
				AssociatePublicIP: fi.PtrTo(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.PtrTo("nodes"),
				},
				BlockDeviceMappings: []*BlockDeviceMapping{
					{
						DeviceName:             fi.PtrTo("/dev/xvdd"),
						EbsVolumeType:          fi.PtrTo("gp2"),
						EbsVolumeSize:          fi.PtrTo(int64(100)),
						EbsDeleteOnTermination: fi.PtrTo(true),
						EbsEncrypted:           fi.PtrTo(true),
					},
				},
				ID:                     fi.PtrTo("test-11"),
				InstanceMonitoring:     fi.PtrTo(true),
				InstanceType:           fi.PtrTo("t2.medium"),
				RootVolumeOptimization: fi.PtrTo(true),
				RootVolumeIops:         fi.PtrTo(int64(100)),
				RootVolumeSize:         fi.PtrTo(int64(64)),
				SSHKey: &SSHKey{
					Name: fi.PtrTo("mykey"),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: fi.PtrTo("nodes-1"), ID: fi.PtrTo("1111")},
					{Name: fi.PtrTo("nodes-2"), ID: fi.PtrTo("2222")},
				},
				Tenancy:                 fi.PtrTo("dedicated"),
				HTTPTokens:              fi.PtrTo("required"),
				HTTPPutResponseHopLimit: fi.PtrTo(int64(5)),
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
      "configuration_aliases" = [aws.files]
      "source"                = "hashicorp/aws"
      "version"               = ">= 4.0.0"
    }
  }
}
`,
		},
	}
	doRenderTests(ctx, t, "RenderTerraform", cases)
}

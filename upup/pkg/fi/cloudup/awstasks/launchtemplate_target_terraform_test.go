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

	"k8s.io/kops/upup/pkg/fi"
)

func TestLaunchTemplateTerraformRender(t *testing.T) {
	cases := []*renderTest{
		{
			Resource: &LaunchTemplate{
				Name:              fi.String("test"),
				AssociatePublicIP: fi.Bool(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.String("nodes"),
				},
				ID:                           fi.String("test-11"),
				InstanceMonitoring:           fi.Bool(true),
				InstanceType:                 fi.String("t2.medium"),
				SpotPrice:                    "0.1",
				SpotDurationInMinutes:        fi.Int64(60),
				InstanceInterruptionBehavior: fi.String("hibernate"),
				RootVolumeOptimization:       fi.Bool(true),
				RootVolumeIops:               fi.Int64(100),
				RootVolumeSize:               fi.Int64(64),
				SSHKey: &SSHKey{
					Name:      fi.String("newkey"),
					PublicKey: fi.WrapResource(fi.NewStringResource("newkey")),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: fi.String("nodes-1"), ID: fi.String("1111")},
					{Name: fi.String("nodes-2"), ID: fi.String("2222")},
				},
				Tenancy: fi.String("dedicated"),
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
  name_prefix = "test-"
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
  required_version = ">= 0.12.0"
}
`,
		},
		{
			Resource: &LaunchTemplate{
				Name:              fi.String("test"),
				AssociatePublicIP: fi.Bool(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.String("nodes"),
				},
				BlockDeviceMappings: []*BlockDeviceMapping{
					{
						DeviceName:             fi.String("/dev/xvdd"),
						EbsVolumeType:          fi.String("gp2"),
						EbsVolumeSize:          fi.Int64(100),
						EbsDeleteOnTermination: fi.Bool(true),
						EbsEncrypted:           fi.Bool(true),
					},
				},
				ID:                     fi.String("test-11"),
				InstanceMonitoring:     fi.Bool(true),
				InstanceType:           fi.String("t2.medium"),
				RootVolumeOptimization: fi.Bool(true),
				RootVolumeIops:         fi.Int64(100),
				RootVolumeSize:         fi.Int64(64),
				SSHKey: &SSHKey{
					Name: fi.String("mykey"),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: fi.String("nodes-1"), ID: fi.String("1111")},
					{Name: fi.String("nodes-2"), ID: fi.String("2222")},
				},
				Tenancy: fi.String("dedicated"),
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
  name_prefix = "test-"
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
  required_version = ">= 0.12.0"
}
`,
		},
	}
	doRenderTests(t, "RenderTerraform", cases)
}

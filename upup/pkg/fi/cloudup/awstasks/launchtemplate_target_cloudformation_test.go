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

func TestLaunchTemplateCloudformationRender(t *testing.T) {
	cases := []*renderTest{
		{
			Resource: &LaunchTemplate{
				Name:              fi.String("test"),
				AssociatePublicIP: fi.Bool(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.String("nodes"),
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
			Expected: `{
  "Resources": {
    "AWSEC2LaunchTemplatetest": {
      "Type": "AWS::EC2::LaunchTemplate",
      "Properties": {
        "LaunchTemplateName": "test",
        "LaunchTemplateData": {
          "EbsOptimized": true,
          "IamInstanceProfile": {
            "Name": {
              "Ref": "AWSIAMInstanceProfilenodes"
            }
          },
          "InstanceType": "t2.medium",
          "KeyName": "mykey",
          "NetworkInterfaces": [
            {
              "AssociatePublicIpAddress": true,
              "DeleteOnTermination": true,
              "DeviceIndex": 0,
              "Groups": [
                {
                  "Ref": "AWSEC2SecurityGroupnodes1"
                },
                {
                  "Ref": "AWSEC2SecurityGroupnodes2"
                }
              ]
            }
          ],
          "Placement": [
            {
              "Tenancy": "dedicated"
            }
          ]
        }
      }
    }
  }
}`,
		},
		{
			Resource: &LaunchTemplate{
				Name:              fi.String("test"),
				AssociatePublicIP: fi.Bool(true),
				BlockDeviceMappings: []*BlockDeviceMapping{
					{
						DeviceName:             fi.String("/dev/xvdd"),
						EbsVolumeType:          fi.String("gp2"),
						EbsVolumeSize:          fi.Int64(100),
						EbsDeleteOnTermination: fi.Bool(true),
						EbsEncrypted:           fi.Bool(true),
					},
				},
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.String("nodes"),
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
			Expected: `{
  "Resources": {
    "AWSEC2LaunchTemplatetest": {
      "Type": "AWS::EC2::LaunchTemplate",
      "Properties": {
        "LaunchTemplateName": "test",
        "LaunchTemplateData": {
          "BlockDeviceMappings": [
            {
              "DeviceName": "/dev/xvdd",
              "Ebs": {
                "VolumeType": "gp2",
                "VolumeSize": 100,
                "DeleteOnTermination": true,
                "Encrypted": true
              }
            }
          ],
          "EbsOptimized": true,
          "IamInstanceProfile": {
            "Name": {
              "Ref": "AWSIAMInstanceProfilenodes"
            }
          },
          "InstanceType": "t2.medium",
          "KeyName": "mykey",
          "NetworkInterfaces": [
            {
              "AssociatePublicIpAddress": true,
              "DeleteOnTermination": true,
              "DeviceIndex": 0,
              "Groups": [
                {
                  "Ref": "AWSEC2SecurityGroupnodes1"
                },
                {
                  "Ref": "AWSEC2SecurityGroupnodes2"
                }
              ]
            }
          ],
          "Placement": [
            {
              "Tenancy": "dedicated"
            }
          ]
        }
      }
    }
  }
}`,
		},
	}
	doRenderTests(t, "RenderCloudformation", cases)
}

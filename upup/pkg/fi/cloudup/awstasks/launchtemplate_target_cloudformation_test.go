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
				Name:              fi.PtrTo("test"),
				AssociatePublicIP: fi.PtrTo(true),
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.PtrTo("nodes"),
				},
				ID:                           fi.PtrTo("test-11"),
				InstanceMonitoring:           fi.PtrTo(true),
				InstanceType:                 fi.PtrTo("t2.medium"),
				RootVolumeOptimization:       fi.PtrTo(true),
				RootVolumeIops:               fi.PtrTo(int64(100)),
				RootVolumeSize:               fi.PtrTo(int64(64)),
				SpotPrice:                    fi.PtrTo("10"),
				SpotDurationInMinutes:        fi.PtrTo(int64(120)),
				InstanceInterruptionBehavior: fi.PtrTo("hibernate"),
				SSHKey: &SSHKey{
					Name: fi.PtrTo("mykey"),
				},
				SecurityGroups: []*SecurityGroup{
					{Name: fi.PtrTo("nodes-1"), ID: fi.PtrTo("1111")},
					{Name: fi.PtrTo("nodes-2"), ID: fi.PtrTo("2222")},
				},
				Tenancy:                 fi.PtrTo("dedicated"),
				HTTPTokens:              fi.PtrTo("required"),
				HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
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
          "InstanceMarketOptions": {
            "MarketType": "spot",
            "SpotOptions": {
              "BlockDurationMinutes": 120,
              "InstanceInterruptionBehavior": "hibernate",
              "MaxPrice": "10"
            }
          },
          "MetadataOptions": {
            "HttpPutResponseHopLimit": 1,
            "HttpTokens": "required"
          },
          "Monitoring": {
            "Enabled": true
          },
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
				Name:              fi.PtrTo("test"),
				AssociatePublicIP: fi.PtrTo(true),
				BlockDeviceMappings: []*BlockDeviceMapping{
					{
						DeviceName:             fi.PtrTo("/dev/xvdd"),
						EbsVolumeType:          fi.PtrTo("gp2"),
						EbsVolumeSize:          fi.PtrTo(int64(100)),
						EbsDeleteOnTermination: fi.PtrTo(true),
						EbsEncrypted:           fi.PtrTo(true),
					},
				},
				IAMInstanceProfile: &IAMInstanceProfile{
					Name: fi.PtrTo("nodes"),
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
				HTTPTokens:              fi.PtrTo("optional"),
				HTTPPutResponseHopLimit: fi.PtrTo(int64(1)),
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
          "MetadataOptions": {
            "HttpPutResponseHopLimit": 1,
            "HttpTokens": "optional"
          },
          "Monitoring": {
            "Enabled": true
          },
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

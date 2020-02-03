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
	"sort"
	"testing"

	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/ghodss/yaml"
)

func TestGetASGTagsToDelete(t *testing.T) {
	asg := &AutoscalingGroup{
		Name: aws.String("MyASGName"),
		Tags: map[string]string{
			"KubernetesCluster": "MyCluster",
			"Name":              "nodes.cluster.k8s.local",
		},
	}

	cases := []struct {
		CurrentTags          map[string]string
		ExpectedTagsToDelete []*autoscaling.Tag
	}{
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.local",
			},
			ExpectedTagsToDelete: []*autoscaling.Tag{},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.locall",
			},
			ExpectedTagsToDelete: []*autoscaling.Tag{},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
			},
			ExpectedTagsToDelete: []*autoscaling.Tag{},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.local",
				"OldTag":            "OldValue",
			},
			ExpectedTagsToDelete: []*autoscaling.Tag{
				{
					Key:          aws.String("OldTag"),
					Value:        aws.String("OldValue"),
					ResourceId:   asg.Name,
					ResourceType: aws.String("auto-scaling-group"),
				},
			},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.local",
				"MyCustomTag":       "MyCustomValue",
				"k8s.io/cluster-autoscaler/node-template/taint/sometaint": "somevalue:NoSchedule",
			},
			ExpectedTagsToDelete: []*autoscaling.Tag{
				{
					Key:          aws.String("MyCustomTag"),
					Value:        aws.String("MyCustomValue"),
					ResourceId:   asg.Name,
					ResourceType: aws.String("auto-scaling-group"),
				},
				{
					Key:          aws.String("k8s.io/cluster-autoscaler/node-template/taint/sometaint"),
					Value:        aws.String("somevalue:NoSchedule"),
					ResourceId:   asg.Name,
					ResourceType: aws.String("auto-scaling-group"),
				},
			},
		},
	}

	for i, x := range cases {
		tagsToDelete := asg.getASGTagsToDelete(x.CurrentTags)

		// Sort both lists to ensure comparisons don't show a false negative
		sort.Slice(tagsToDelete, func(i, j int) bool {
			return *tagsToDelete[i].Key < *tagsToDelete[j].Key
		})
		sort.Slice(x.ExpectedTagsToDelete, func(i, j int) bool {
			return *x.ExpectedTagsToDelete[i].Key < *x.ExpectedTagsToDelete[j].Key
		})

		expected, err := yaml.Marshal(x.ExpectedTagsToDelete)
		if err != nil {
			t.Errorf("case %d, unexpected error converting expected tags to yaml: %v", i, err)
		}

		actual, err := yaml.Marshal(tagsToDelete)
		if err != nil {
			t.Errorf("case %d, unexpected error converting actual tags to yaml: %v", i, err)
		}

		if string(expected) != string(actual) {
			diffString := diff.FormatDiff(string(expected), string(actual))
			t.Errorf("case %d failed, actual output differed from expected.", i)
			t.Logf("diff:\n%s\n", diffString)
		}
	}
}

func TestProcessCompare(t *testing.T) {
	rebalance := "AZRebalance"
	healthcheck := "HealthCheck"

	a := []string{}
	b := []string{
		rebalance,
	}
	c := []string{
		rebalance,
		healthcheck,
	}

	cases := []struct {
		A                 *[]string
		B                 *[]string
		ExpectedProcesses []*string
	}{
		{
			A:                 &a,
			B:                 &b,
			ExpectedProcesses: []*string{},
		},
		{
			A: &b,
			B: &a,
			ExpectedProcesses: []*string{
				&rebalance,
			},
		},
		{
			A: &c,
			B: &b,
			ExpectedProcesses: []*string{
				&healthcheck,
			},
		},
		{
			A: &c,
			B: &a,
			ExpectedProcesses: []*string{
				&rebalance,
				&healthcheck,
			},
		},
	}

	for i, x := range cases {
		result := processCompare(x.A, x.B)

		expected, err := yaml.Marshal(x.ExpectedProcesses)
		if err != nil {
			t.Errorf("case %d, unexpected error converting expected processes to yaml: %v", i, err)
		}

		actual, err := yaml.Marshal(result)
		if err != nil {
			t.Errorf("case %d, unexpected error converting actual result to yaml: %v", i, err)
		}

		if string(expected) != string(actual) {
			diffString := diff.FormatDiff(string(expected), string(actual))
			t.Errorf("case %d failed, actual output differed from expected.", i)
			t.Logf("diff:\n%s\n", diffString)
		}
	}
}

func TestAutoscalingGroupTerraformRender(t *testing.T) {
	cases := []*renderTest{
		{
			Resource: &AutoscalingGroup{
				Name:                fi.String("test"),
				Granularity:         fi.String("5min"),
				LaunchConfiguration: &LaunchConfiguration{Name: fi.String("test_lc")},
				MaxSize:             fi.Int64(10),
				Metrics:             []string{"test"},
				MinSize:             fi.Int64(1),
				Subnets: []*Subnet{
					{
						Name: fi.String("test-sg"),
						ID:   fi.String("sg-1111"),
					},
				},
				Tags: map[string]string{
					"test":    "tag",
					"cluster": "test",
				},
			},
			Expected: `provider "aws" {
  region = "eu-west-2"
}

resource "aws_autoscaling_group" "test" {
  name                 = "test"
  launch_configuration = "${aws_launch_configuration.test_lc.id}"
  max_size             = 10
  min_size             = 1
  vpc_zone_identifier  = ["${aws_subnet.test-sg.id}"]

  tag = {
    key                 = "cluster"
    value               = "test"
    propagate_at_launch = true
  }

  tag = {
    key                 = "test"
    value               = "tag"
    propagate_at_launch = true
  }

  metrics_granularity = "5min"
  enabled_metrics     = ["test"]
}

terraform = {
  required_version = ">= 0.9.3"
}
`,
		},
		{
			Resource: &AutoscalingGroup{
				Name:                        fi.String("test1"),
				LaunchTemplate:              &LaunchTemplate{Name: fi.String("test_lt")},
				MaxSize:                     fi.Int64(10),
				Metrics:                     []string{"test"},
				MinSize:                     fi.Int64(5),
				MixedInstanceOverrides:      []string{"t2.medium", "t2.large"},
				MixedOnDemandBase:           fi.Int64(4),
				MixedOnDemandAboveBase:      fi.Int64(30),
				MixedSpotAllocationStrategy: fi.String("capacity-optimized"),
				Subnets: []*Subnet{
					{
						Name: fi.String("test-sg"),
						ID:   fi.String("sg-1111"),
					},
				},
				Tags: map[string]string{
					"test":    "tag",
					"cluster": "test",
				},
			},
			Expected: `provider "aws" {
  region = "eu-west-2"
}

resource "aws_autoscaling_group" "test1" {
  name     = "test1"
  max_size = 10
  min_size = 5

  mixed_instances_policy = {
    launch_template = {
      launch_template_specification = {
        launch_template_id = "${aws_launch_template.test_lt.id}"
        version            = "${aws_launch_template.test_lt.latest_version}"
      }

      override = {
        instance_type = "t2.medium"
      }

      override = {
        instance_type = "t2.large"
      }
    }

    instances_distribution = {
      on_demand_base_capacity                  = 4
      on_demand_percentage_above_base_capacity = 30
      spot_allocation_strategy                 = "capacity-optimized"
    }
  }

  vpc_zone_identifier = ["${aws_subnet.test-sg.id}"]

  tag = {
    key                 = "cluster"
    value               = "test"
    propagate_at_launch = true
  }

  tag = {
    key                 = "test"
    value               = "tag"
    propagate_at_launch = true
  }

  enabled_metrics = ["test"]
}

terraform = {
  required_version = ">= 0.9.3"
}
`,
		},
	}

	doRenderTests(t, "RenderTerraform", cases)
}

func TestAutoscalingGroupCloudformationRender(t *testing.T) {
	cases := []*renderTest{
		{
			Resource: &AutoscalingGroup{
				Name:                   fi.String("test1"),
				LaunchTemplate:         &LaunchTemplate{Name: fi.String("test_lt")},
				MaxSize:                fi.Int64(10),
				Metrics:                []string{"test"},
				MinSize:                fi.Int64(5),
				MixedInstanceOverrides: []string{"t2.medium", "t2.large"},
				MixedOnDemandBase:      fi.Int64(4),
				MixedOnDemandAboveBase: fi.Int64(30),
				Subnets: []*Subnet{
					{
						Name: fi.String("test-sg"),
						ID:   fi.String("sg-1111"),
					},
				},
				Tags: map[string]string{
					"test":    "tag",
					"cluster": "test",
				},
			},
			Expected: `{
  "Resources": {
    "AWSAutoScalingAutoScalingGrouptest1": {
      "Type": "AWS::AutoScaling::AutoScalingGroup",
      "Properties": {
        "AutoScalingGroupName": "test1",
        "MaxSize": 10,
        "MinSize": 5,
        "VPCZoneIdentifier": [
          {
            "Ref": "AWSEC2Subnettestsg"
          }
        ],
        "Tags": [
          {
            "Key": "cluster",
            "Value": "test",
            "PropagateAtLaunch": true
          },
          {
            "Key": "test",
            "Value": "tag",
            "PropagateAtLaunch": true
          }
        ],
        "MetricsCollection": [
          {
            "Granularity": null,
            "Metrics": [
              "test"
            ]
          }
        ],
        "MixedInstancesPolicy": {
          "LaunchTemplate": {
            "LaunchTemplateSpecification": {
              "LaunchTemplateId": {
                "Ref": "AWSEC2LaunchTemplatetest_lt"
              },
              "Version": {
                "Fn::GetAtt": [
                  "AWSEC2LaunchTemplatetest_lt",
                  "LatestVersionNumber"
                ]
              }
            },
            "Overrides": [
              {
                "InstanceType": "t2.medium"
              },
              {
                "InstanceType": "t2.large"
              }
            ]
          },
          "InstancesDistribution": {
            "OnDemandBaseCapacity": 4,
            "OnDemandPercentageAboveBaseCapacity": 30
          }
        }
      }
    }
  }
}`,
		},
	}

	doRenderTests(t, "RenderCloudformation", cases)
}

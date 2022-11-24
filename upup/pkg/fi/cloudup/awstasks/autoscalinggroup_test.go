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
	"fmt"
	"sort"
	"testing"

	"k8s.io/kops/pkg/diff"
	"k8s.io/kops/upup/pkg/fi"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"sigs.k8s.io/yaml"
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
				Name:           fi.String("test"),
				Granularity:    fi.String("5min"),
				LaunchTemplate: &LaunchTemplate{Name: fi.String("test_lc")},
				MaxSize:        fi.Int64(10),
				Metrics:        []string{"test"},
				MinSize:        fi.Int64(1),
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
  enabled_metrics = ["test"]
  launch_template {
    id      = aws_launch_template.test_lc.id
    version = aws_launch_template.test_lc.latest_version
  }
  max_size            = 10
  metrics_granularity = "5min"
  min_size            = 1
  name                = "test"
  tag {
    key                 = "cluster"
    propagate_at_launch = true
    value               = "test"
  }
  tag {
    key                 = "test"
    propagate_at_launch = true
    value               = "tag"
  }
  vpc_zone_identifier = [aws_subnet.test-sg.id]
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
  enabled_metrics = ["test"]
  max_size        = 10
  min_size        = 5
  mixed_instances_policy {
    instances_distribution {
      on_demand_base_capacity                  = 4
      on_demand_percentage_above_base_capacity = 30
      spot_allocation_strategy                 = "capacity-optimized"
    }
    launch_template {
      launch_template_specification {
        launch_template_id = aws_launch_template.test_lt.id
        version            = aws_launch_template.test_lt.latest_version
      }
      override {
        instance_type = "t2.medium"
      }
      override {
        instance_type = "t2.large"
      }
    }
  }
  name = "test1"
  tag {
    key                 = "cluster"
    propagate_at_launch = true
    value               = "test"
  }
  tag {
    key                 = "test"
    propagate_at_launch = true
    value               = "tag"
  }
  vpc_zone_identifier = [aws_subnet.test-sg.id]
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

	doRenderTests(t, "RenderTerraform", cases)
}

func TestTGsARNsChunks(t *testing.T) {
	var tgsARNs []*string
	for i := 0; i < 30; i++ {
		tgsARNs = append(tgsARNs, fi.PtrTo(fmt.Sprintf("arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/my-targets/00000000000000%02d", i)))
	}

	tests := []struct {
		tgsARNs   []*string
		chunkSize int
		expected  [][]*string
	}{
		{
			tgsARNs:   tgsARNs[0:1],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:1]},
		},
		{
			tgsARNs:   tgsARNs[0:5],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:5]},
		},
		{
			tgsARNs:   tgsARNs[0:10],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10]},
		},
		{
			tgsARNs:   tgsARNs[0:11],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10], tgsARNs[10:11]},
		},
		{
			tgsARNs:   tgsARNs[0:15],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10], tgsARNs[10:15]},
		},
		{
			tgsARNs:   tgsARNs[0:20],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10], tgsARNs[10:20]},
		},
		{
			tgsARNs:   tgsARNs[0:21],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10], tgsARNs[10:20], tgsARNs[20:21]},
		},
		{
			tgsARNs:   tgsARNs[0:25],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10], tgsARNs[10:20], tgsARNs[20:25]},
		},
		{
			tgsARNs:   tgsARNs[0:30],
			chunkSize: 10,
			expected:  [][]*string{tgsARNs[0:10], tgsARNs[10:20], tgsARNs[20:30]},
		},
	}

	for i, test := range tests {
		result := sliceChunks(test.tgsARNs, test.chunkSize)

		expected, err := yaml.Marshal(test.expected)
		if err != nil {
			t.Errorf("case %d: failed to convert expected to yaml: %v", i, err)
			continue
		}

		actual, err := yaml.Marshal(result)
		if err != nil {
			t.Errorf("case %d: failed to convert actual to yaml: %v", i, err)
			continue
		}

		if string(expected) != string(actual) {
			diffString := diff.FormatDiff(string(expected), string(actual))
			t.Errorf("case %d: actual output differed from expected", i)
			t.Logf("diff:\n%s\n", diffString)
		}
	}
}

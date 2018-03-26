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

package awstasks

import (
	"sort"
	"testing"

	"k8s.io/kops/pkg/diff"

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

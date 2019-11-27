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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/ghodss/yaml"

	"k8s.io/kops/pkg/diff"
)

func TestGetEBSVolumeTagsToDelete(t *testing.T) {
	ebsv := &EBSVolume{
		ID: aws.String("ebs-1234567"),
		Tags: map[string]string{
			"KubernetesCluster": "MyCluster",
			"Name":              "nodes.cluster.k8s.local",
		},
	}

	cases := []struct {
		CurrentTags          map[string]string
		ExpectedTagsToDelete map[string]string
	}{
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.local",
			},
			ExpectedTagsToDelete: map[string]string{},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.locall",
			},
			ExpectedTagsToDelete: map[string]string{},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
			},
			ExpectedTagsToDelete: map[string]string{},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.local",
				"OldTag":            "OldValue",
			},
			ExpectedTagsToDelete: map[string]string{
				"OldTag": "OldValue",
			},
		},
		{
			CurrentTags: map[string]string{
				"KubernetesCluster": "MyCluster",
				"Name":              "nodes.cluster.k8s.local",
				"MyCustomTag":       "MyCustomValue",
				"k8s.io/cluster-autoscaler/node-template/taint/sometaint": "somevalue:NoSchedule",
			},
			ExpectedTagsToDelete: map[string]string{
				"MyCustomTag": "MyCustomValue",
				"k8s.io/cluster-autoscaler/node-template/taint/sometaint": "somevalue:NoSchedule",
			},
		},
	}

	for i, x := range cases {
		tagsToDelete := ebsv.getEBSVolumeTagsToDelete(x.CurrentTags)

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

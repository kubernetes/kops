/*
Copyright 2020 The Kubernetes Authors.

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

package azuremodel

import (
	"reflect"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestCloudTagsForInstanceGroup(t *testing.T) {
	c := newTestAzureModelContext()
	c.Cluster.Spec.CloudLabels = map[string]string{
		"cluster_label_key": "cluster_label_value",
		"test_label":        "from_cluster",
	}
	ig := c.InstanceGroups[0]
	ig.Spec.CloudLabels = map[string]string{
		"ig_label_key": "ig_label_value",
		"test_label":   "from_ig",
	}
	ig.Spec.NodeLabels = map[string]string{
		"node_label/key": "node_label_value",
	}
	ig.Spec.Taints = []string{
		"taint_key=taint_value",
	}

	actual := c.CloudTagsForInstanceGroup(c.InstanceGroups[0])
	expected := map[string]*string{
		"cluster_label_key": fi.String("cluster_label_value"),
		"ig_label_key":      fi.String("ig_label_value"),
		"test_label":        fi.String("from_ig"),
		"k8s.io_cluster_node-template_label_node_label_key": fi.String("node_label_value"),
		"k8s.io_cluster_node-template_taint_taint_key":      fi.String("taint_value"),
		"k8s.io_role_node":          fi.String("1"),
		"kops.k8s.io_instancegroup": fi.String("nodes"),
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected tags %+v, but got %+v", expected, actual)
	}

}

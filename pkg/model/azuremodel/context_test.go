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
	"strings"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestCloudTagsForInstanceGroupResource(t *testing.T) {
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

	actual := c.CloudTagsForInstanceGroupResource(c.InstanceGroups[0])
	expected := map[string]*string{
		"cluster_label_key":                            fi.PtrTo("cluster_label_value"),
		"ig_label_key":                                 fi.PtrTo("ig_label_value"),
		"test_label":                                   fi.PtrTo("from_ig"),
		"k8s.io_cluster_node-template_label_0":         fi.PtrTo("node_label/key=node_label_value"),
		"k8s.io_cluster_node-template_taint_taint_key": fi.PtrTo("taint_value"),
		"k8s.io_role_node":                             fi.PtrTo("1"),
		"kops.k8s.io_instancegroup":                    fi.PtrTo("nodes"),
		"KubernetesCluster":                            fi.PtrTo("testcluster.test.com"),
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected tags %+v, but got %+v", expected, actual)
	}
}

func TestCloudTagsForClusterResource(t *testing.T) {
	c := newTestAzureModelContext()
	c.Cluster.ObjectMeta.Name = "my.k8s"
	c.Cluster.Spec.CloudLabels = map[string]string{
		"cluster_label_key": "cluster_label_value",
		"node_label/key":    "node_label_value",
	}

	actual := c.CloudTagsForClusterResource()
	expected := map[string]*string{
		"cluster_label_key": fi.PtrTo("cluster_label_value"),
		"node_label_key":    fi.PtrTo("node_label_value"),
		"KubernetesCluster": fi.PtrTo("my.k8s"),
	}

	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected tags %+v, but got %+v", expected, actual)
	}
}

func TestSanitizeUserAssignedManagedIdentityName(t *testing.T) {
	t.Run("preserves valid names", func(t *testing.T) {
		const name = "nodepool-1"
		if got := sanitizeUserAssignedManagedIdentityName(name); got != name {
			t.Fatalf("expected %q, but got %q", name, got)
		}
	})

	t.Run("replaces invalid characters", func(t *testing.T) {
		got := sanitizeUserAssignedManagedIdentityName("my.cluster.example.com")
		if strings.Contains(got, ".") {
			t.Fatalf("expected dots to be replaced, got %q", got)
		}
		if got != "my-cluster-example-com" {
			t.Fatalf("expected %q, but got %q", "my-cluster-example-com", got)
		}
	})

	t.Run("truncates long names", func(t *testing.T) {
		name := strings.Repeat("a", maxUserAssignedManagedIdentityNameLength+1)
		got := sanitizeUserAssignedManagedIdentityName(name)
		if len(got) > maxUserAssignedManagedIdentityNameLength {
			t.Fatalf("expected name length <= %d, got %d (%q)", maxUserAssignedManagedIdentityNameLength, len(got), got)
		}
	})
}

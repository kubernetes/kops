/*
Copyright 2017 The Kubernetes Authors.

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

package cloudup

import (
	"reflect"
	"sort"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestPrecreateDNSNames(t *testing.T) {
	cluster := &kops.Cluster{}
	cluster.ObjectMeta.Name = "cluster1.example.com"
	cluster.Spec.MasterPublicName = "api." + cluster.ObjectMeta.Name
	cluster.Spec.MasterInternalName = "api.internal." + cluster.ObjectMeta.Name
	cluster.Spec.EtcdClusters = []*kops.EtcdClusterSpec{
		{
			Name: "main",
			Members: []*kops.EtcdMemberSpec{
				{Name: "zone1"},
				{Name: "zone2"},
				{Name: "zone3"},
			},
		},
		{
			Name: "events",
			Members: []*kops.EtcdMemberSpec{
				{Name: "zonea"},
				{Name: "zoneb"},
				{Name: "zonec"},
			},
		},
	}

	actual := buildPrecreateDNSHostnames(cluster)

	expected := []string{
		"api.cluster1.example.com",
		"api.internal.cluster1.example.com",
		"etcd-events-zonea.internal.cluster1.example.com",
		"etcd-events-zoneb.internal.cluster1.example.com",
		"etcd-events-zonec.internal.cluster1.example.com",
		"etcd-zone1.internal.cluster1.example.com",
		"etcd-zone2.internal.cluster1.example.com",
		"etcd-zone3.internal.cluster1.example.com",
	}

	sort.Strings(actual)
	sort.Strings(expected)

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("unexpected records.  expected=%v actual=%v", expected, actual)
	}
}

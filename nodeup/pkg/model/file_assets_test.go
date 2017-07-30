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

package model

import (
	"io/ioutil"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
)

func TestFileAssetRenderOK(t *testing.T) {
	cs := []struct {
		Content  string
		Expected string
	}{
		{},
		{
			Content:  "{{ .Cluster.NetworkCIDR }}",
			Expected: "10.79.0.0/24",
		},
		{
			Content: `
*filter
:INPUT ACCEPT [0:0]
-A INPUT -i docker0 -p tcp -m tcp --dport 22 -m state --state NEW -j REJECT
-A INPUT -i docker0 -p tcp -m tcp --dport 10250 -m state --state NEW -j REJECT
-A INPUT -p tcp -m state -s {{ .Cluster.NonMasqueradeCIDR }} --dport 22 --state NEW -j REJECT
:FORWARD ACCEPT [0:0]
-A FORWARD -i docker0 -p tcp -m tcp -d 169.254.169.254/32 --dport 80 -m state --state NEW -j REJECT
-A FORWARD -i docker0 -p tcp -m tcp --dport 2379 -m state --state NEW -j REJECT
-A FORWARD -i docker0 -p tcp -m tcp -d {{ .Cluster.NetworkCIDR }} --dport 22 -m state --state NEW -j REJECT
-A FORWARD -i docker0 -p tcp -m tcp -d {{ .Cluster.NetworkCIDR }} --dport 10250 -m state --state NEW -j REJECT
:OUTPUT ACCEPT [0:0]
COMMIT`,
			Expected: `
*filter
:INPUT ACCEPT [0:0]
-A INPUT -i docker0 -p tcp -m tcp --dport 22 -m state --state NEW -j REJECT
-A INPUT -i docker0 -p tcp -m tcp --dport 10250 -m state --state NEW -j REJECT
-A INPUT -p tcp -m state -s 10.100.0.0/16 --dport 22 --state NEW -j REJECT
:FORWARD ACCEPT [0:0]
-A FORWARD -i docker0 -p tcp -m tcp -d 169.254.169.254/32 --dport 80 -m state --state NEW -j REJECT
-A FORWARD -i docker0 -p tcp -m tcp --dport 2379 -m state --state NEW -j REJECT
-A FORWARD -i docker0 -p tcp -m tcp -d 10.79.0.0/24 --dport 22 -m state --state NEW -j REJECT
-A FORWARD -i docker0 -p tcp -m tcp -d 10.79.0.0/24 --dport 10250 -m state --state NEW -j REJECT
:OUTPUT ACCEPT [0:0]
COMMIT`,
		},
	}
	cluster := makeTestCluster()
	group := makeTestInstanceGroup()

	for i, x := range cs {
		fb := &FileAssetsBuilder{
			NodeupModelContext: &NodeupModelContext{Cluster: cluster, InstanceGroup: group},
		}
		resource, err := fb.getRenderedResource(x.Content)
		if err != nil {
			t.Errorf("case %d failed to create resource. error: %s", i, err)
			continue
		}
		reader, err := resource.Open()
		if err != nil {
			t.Errorf("case %d failed to render resource, error: %s", i, err)
			continue
		}
		rendered, err := ioutil.ReadAll(reader)
		if err != nil {
			t.Errorf("case %d failed to read reander, error: %s", i, err)
			continue
		}

		if string(rendered) != x.Expected {
			t.Errorf("case %d, expected: %s. got: %s", i, x.Expected, rendered)
		}
	}
}

func makeTestCluster() *kops.Cluster {
	return &kops.Cluster{
		Spec: kops.ClusterSpec{
			CloudProvider:     "aws",
			KubernetesVersion: "1.7.0",
			Subnets: []kops.ClusterSubnetSpec{
				{Name: "test", Zone: "eu-west-1a"},
			},
			NonMasqueradeCIDR: "10.100.0.0/16",
			EtcdClusters: []*kops.EtcdClusterSpec{
				{
					Name: "main",
					Members: []*kops.EtcdMemberSpec{
						{
							Name:          "test",
							InstanceGroup: s("master-1"),
						},
					},
				},
			},
			NetworkCIDR: "10.79.0.0/24",
		},
	}
}

func makeTestInstanceGroup() *kops.InstanceGroup {
	return &kops.InstanceGroup{
		Spec: kops.InstanceGroupSpec{},
	}
}

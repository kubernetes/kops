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

package model

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
)

func Test_SharedGroups(t *testing.T) {
	grid := []struct {
		Prefix      string
		ClusterName string
		Expected    string
	}{
		{
			"bastion", "mycluster",
			"bastion-mycluster-vnrjie",
		},
		{
			"bastion", "mycluster.example.com",
			"bastion-mycluster-example-o8elkm",
		},
		{
			"api", "this.is.a.very.long.cluster.example.com",
			"api-this-is-a-very-long-c-q4ukp4",
		},
		{
			"bastion", "this.is.a.very.long.cluster.example.com",
			"bastion-this-is-a-very-lo-4ggpa2",
		},
	}
	for _, g := range grid {
		c := &KopsModelContext{
			Cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: g.ClusterName,
				},
			},
		}
		actual := c.GetELBName32(g.Prefix)
		if actual != g.Expected {
			t.Errorf("unexpected result from %q+%q.  expected %q, got %q", g.Prefix, g.ClusterName, g.Expected, actual)
		}
	}
}

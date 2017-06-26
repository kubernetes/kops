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

package instancegroups

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/client/simple/fake"
)

func TestGetPrefix(t *testing.T) {
	timeStamp := time.Now()
	timeStampLayout := timeStamp.Format(IG_TS_LAYOUT)
	grid := []struct {
		Input  string
		Output string
	}{
		{
			Input:  "foo",
			Output: fmt.Sprintf("foo%s%s", IG_PREFIX, timeStampLayout),
		},
		{
			Input:  "node01-rolling-update-2017-01-02-15-04-05",
			Output: fmt.Sprintf("node01%s%s", IG_PREFIX, timeStampLayout),
		},
	}

	for _, g := range grid {
		result := getSuffixWithTime(g.Input, timeStamp)
		if result != g.Output {
			t.Fatalf("testing %q failed, expected %q, got %q", g.Input, g.Output, result)
		}
	}
}

func pointerInt32(v int) *int32 {
	i := int32(v)
	return &i
}

func pointerString(v string) *string {
	return &v
}

func pointerbool(v bool) *bool {
	return &v
}

func TestDuplicateIG(t *testing.T) {

	cluster := &api.Cluster{
		ObjectMeta: v1meta.ObjectMeta{
			Name: "test-cluster.example.com",
		},
		Spec: api.ClusterSpec{
			KubernetesVersion: "1.6.4",
		},
	}

	tm := time.Now()
	ts := tm.Format(IG_TS_LAYOUT)

	grid := []struct {
		Input        *CloudInstanceGroup
		ExpectedName string
	}{
		{
			ExpectedName: "nodes-01" + IG_PREFIX + ts,
			Input: &CloudInstanceGroup{
				InstanceGroup: &api.InstanceGroup{
					ObjectMeta: v1meta.ObjectMeta{
						Name:        "nodes-01",
						ClusterName: "test-cluster.example.com",
						Namespace:   "test-cluster.example.com",
					},
					Spec: api.InstanceGroupSpec{
						Role:           api.InstanceGroupRoleNode,
						Image:          "foo",
						MinSize:        pointerInt32(42),
						MaxSize:        pointerInt32(42),
						MachineType:    "m4.10xlarge",
						RootVolumeSize: pointerInt32(248),
						RootVolumeType: pointerString("gp2"),
						Subnets: []string{
							"us-east-2a",
							"us-east-2c",
						},
						MaxPrice:          pointerString("0.42"),
						AssociatePublicIP: pointerbool(true),
						AdditionalSecurityGroups: []string{
							"i-123455",
							"i-232425",
						},
						CloudLabels: map[string]string{
							"foo": "bar",
							"car": "ferarri",
						},

						NodeLabels: map[string]string{
							"1": "2",
							"4": "5",
						},
						Tenancy: "default",
						Kubelet: &api.KubeletConfigSpec{
							NvidiaGPUs:    1,
							CloudProvider: "aws",
							Taints:        []string{"foo", "bar", "baz"},
							NodeLabels: map[string]string{
								"1": "2",
								"4": "5",
							},
						},

						Taints: []string{"foo", "bar", "baz"},
					},
				},
			},
		},
		{
			ExpectedName: "node02" + IG_PREFIX + ts,
			Input: &CloudInstanceGroup{
				InstanceGroup: &api.InstanceGroup{
					ObjectMeta: v1meta.ObjectMeta{
						Name:        "node02-rolling-update-2017-01-02-15-04-05",
						ClusterName: "test-cluster.example.com",
						Namespace:   "test-cluster.example.com",
					},
					Spec: api.InstanceGroupSpec{
						Role:        api.InstanceGroupRoleNode,
						Image:       "foo",
						MinSize:     pointerInt32(42),
						MaxSize:     pointerInt32(42),
						MachineType: "m4.xlarge",
						Subnets: []string{
							"us-east-1a",
							"us-east-1b",
						},
					},
				},
			},
		},
	}

	for _, g := range grid {
		ig := g.Input.InstanceGroup
		fakeCS := fake.NewSimpleClientset(cluster, ig)
		newName := getSuffixWithTime(g.Input.InstanceGroup.ObjectMeta.Name, tm)

		duplicate, err := g.Input.DuplicateClusterInstanceGroup(cluster, fakeCS, newName)
		if duplicate == nil {
			t.Fatalf("testing failed nil returned")
		}

		if err != nil {
			t.Fatalf("testing failed: %v", err)
		}

		if duplicate.ObjectMeta.Name != g.ExpectedName {
			t.Fatalf("testing failed as ig was not named correctly: expected %s, got %s", g.ExpectedName, duplicate.ObjectMeta.Name)
		}

		if _, ok := duplicate.Annotations[KOPS_IG_PARENT]; !ok {
			t.Fatalf("parent annotation not found")
		}

		// FIXME: Need to do integration testing on this.  Not certain this is the
		// FIXME: fake or actually is a bug.

		origIG, err := fakeCS.InstanceGroupsFor(cluster).Get(ig.ObjectMeta.Name, v1meta.GetOptions{})

		if err != nil {
			t.Fatalf("unable to retrieve original instance group: %v", err)
		}

		if _, ok := origIG.Annotations[KOPS_IG_CHILD]; !ok {
			t.Fatalf("child annotation not found")
		}

		duplicate.ObjectMeta.Name = ig.ObjectMeta.Name
		duplicate.Annotations = nil
		g.Input.InstanceGroup.Annotations = nil

		if !reflect.DeepEqual(duplicate, ig) {
			y, err := api.ToVersionedYaml(duplicate)
			if err != nil {
				t.Fatalf("unable to marshal yaml")
			}

			t.Fatalf("duplicate is not equal\n\n %s", y)
		}

	}
}

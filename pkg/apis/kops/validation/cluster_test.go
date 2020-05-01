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

package validation

import (
	"testing"

	"k8s.io/kops/upup/pkg/fi"

	"k8s.io/kops/pkg/apis/kops"
)

func TestValidEtcdChanges(t *testing.T) {
	grid := []struct {
		OldSpec *kops.EtcdClusterSpec
		NewSpec *kops.EtcdClusterSpec
		Status  *kops.ClusterStatus
		Details string
	}{
		{
			OldSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
					{
						Name:          "b",
						InstanceGroup: fi.String("eu-central-1b"),
					},
					{
						Name:          "c",
						InstanceGroup: fi.String("eu-central-1c"),
					},
				},
			},

			NewSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
					{
						Name:          "b",
						InstanceGroup: fi.String("eu-central-1b"),
					},
					{
						Name:          "d",
						InstanceGroup: fi.String("eu-central-1d"),
					},
				},
			},

			Status: &kops.ClusterStatus{
				EtcdClusters: []kops.EtcdClusterStatus{
					{
						Name: "main",
					},
				},
			},

			Details: "Could not move master from one zone to another",
		},

		{
			OldSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
				},
			},

			NewSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
					{
						Name:          "b",
						InstanceGroup: fi.String("eu-central-1b"),
					},
					{
						Name:          "c",
						InstanceGroup: fi.String("eu-central-1c"),
					},
				},
			},

			Status: &kops.ClusterStatus{
				EtcdClusters: []kops.EtcdClusterStatus{
					{
						Name: "main",
					},
				},
			},

			Details: "Could not update from single to multi-master",
		},

		{
			OldSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
				},
			},

			NewSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
				},
			},

			Status: &kops.ClusterStatus{
				EtcdClusters: []kops.EtcdClusterStatus{
					{
						Name: "main",
					},
				},
			},

			Details: "Could not update identical specs",
		},
	}

	for _, g := range grid {
		errorList := validateEtcdClusterUpdate(nil, g.NewSpec, g.Status, g.OldSpec)
		if len(errorList) != 0 {
			t.Errorf("%v: %v", g.Details, errorList)
		}
	}
}

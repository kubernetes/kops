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

	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/kops/pkg/apis/kops"
)

func TestValidEtcdChanges(t *testing.T) {
	grid := []struct {
		OldSpec kops.EtcdClusterSpec
		NewSpec kops.EtcdClusterSpec
		Status  *kops.ClusterStatus
		Details string
	}{
		{
			OldSpec: kops.EtcdClusterSpec{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: new("eu-central-1a"),
					},
					{
						Name:          "b",
						InstanceGroup: new("eu-central-1b"),
					},
					{
						Name:          "c",
						InstanceGroup: new("eu-central-1c"),
					},
				},
			},

			NewSpec: kops.EtcdClusterSpec{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: new("eu-central-1a"),
					},
					{
						Name:          "b",
						InstanceGroup: new("eu-central-1b"),
					},
					{
						Name:          "d",
						InstanceGroup: new("eu-central-1d"),
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
			OldSpec: kops.EtcdClusterSpec{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: new("eu-central-1a"),
					},
				},
			},

			NewSpec: kops.EtcdClusterSpec{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: new("eu-central-1a"),
					},
					{
						Name:          "b",
						InstanceGroup: new("eu-central-1b"),
					},
					{
						Name:          "c",
						InstanceGroup: new("eu-central-1c"),
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
			OldSpec: kops.EtcdClusterSpec{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: new("eu-central-1a"),
					},
				},
			},

			NewSpec: kops.EtcdClusterSpec{
				Name: "main",
				Members: []kops.EtcdMemberSpec{
					{
						Name:          "a",
						InstanceGroup: new("eu-central-1a"),
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

func TestEtcdImageChanges(t *testing.T) {
	createdStatus := &kops.ClusterStatus{
		EtcdClusters: []kops.EtcdClusterStatus{
			{
				Name: "main",
			},
		},
	}

	grid := []struct {
		Details        string
		OldSpec        kops.EtcdClusterSpec
		NewSpec        kops.EtcdClusterSpec
		Status         *kops.ClusterStatus
		ExpectedErrors []string
	}{
		{
			Details: "image cannot be added to an existing etcd cluster",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			Status:  createdStatus,
			ExpectedErrors: []string{
				"Forbidden::spec.etcdClusters[main].image",
			},
		},
		{
			Details: "image cannot be changed on an existing etcd cluster",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "example.com/etcd:v3.6.99"},
			Status:  createdStatus,
			ExpectedErrors: []string{
				"Forbidden::spec.etcdClusters[main].image",
			},
		},
		{
			Details: "image cannot be removed from an existing etcd cluster",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99"},
			Status:  createdStatus,
			ExpectedErrors: []string{
				"Forbidden::spec.etcdClusters[main].image",
			},
		},
		{
			Details: "version cannot be changed when image is set",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.100", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			Status:  createdStatus,
			ExpectedErrors: []string{
				"Forbidden::spec.etcdClusters[main].version",
			},
		},
		{
			Details: "unchanged image and version are allowed",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			Status:  createdStatus,
		},
		{
			Details: "version changes without image are allowed",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.11"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.12"},
			Status:  createdStatus,
		},
		{
			Details: "image can be set before the etcd cluster is created",
			OldSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99"},
			NewSpec: kops.EtcdClusterSpec{Name: "main", Version: "3.6.99", Image: "gcr.io/etcd-development/etcd:v3.6.99"},
			Status:  &kops.ClusterStatus{},
		},
	}

	for _, g := range grid {
		fp := field.NewPath("spec", "etcdClusters").Key(g.NewSpec.Name)
		errorList := validateEtcdClusterUpdate(fp, g.NewSpec, g.Status, g.OldSpec)
		testErrors(t, g.Details, errorList, g.ExpectedErrors)
	}
}

func TestEtcdVersionRequiredWithImage(t *testing.T) {
	grid := []struct {
		Details        string
		Spec           kops.EtcdClusterSpec
		ExpectedErrors []string
	}{
		{
			Details: "image requires version",
			Spec: kops.EtcdClusterSpec{
				Name:    "main",
				Members: []kops.EtcdMemberSpec{{Name: "a", InstanceGroup: new("eu-central-1a")}},
				Image:   "gcr.io/etcd-development/etcd:v3.6.99",
			},
			ExpectedErrors: []string{
				"Required value::spec.etcdClusters[0].version",
			},
		},
		{
			Details: "image with version is valid",
			Spec: kops.EtcdClusterSpec{
				Name:    "main",
				Members: []kops.EtcdMemberSpec{{Name: "a", InstanceGroup: new("eu-central-1a")}},
				Version: "3.6.99",
				Image:   "gcr.io/etcd-development/etcd:v3.6.99",
			},
		},
		{
			Details: "neither image nor version is valid",
			Spec: kops.EtcdClusterSpec{
				Name:    "main",
				Members: []kops.EtcdMemberSpec{{Name: "a", InstanceGroup: new("eu-central-1a")}},
			},
		},
	}

	for _, g := range grid {
		fp := field.NewPath("spec", "etcdClusters").Index(0)
		errorList := validateEtcdClusterSpec(g.Spec, nil, fp)
		testErrors(t, g.Details, errorList, g.ExpectedErrors)
	}
}

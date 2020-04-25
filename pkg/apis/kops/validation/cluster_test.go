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
					&kops.EtcdMemberSpec{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
				},
			},

			NewSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					&kops.EtcdMemberSpec{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
					&kops.EtcdMemberSpec{
						Name:          "b",
						InstanceGroup: fi.String("eu-central-1b"),
					},
					&kops.EtcdMemberSpec{
						Name:          "c",
						InstanceGroup: fi.String("eu-central-1c"),
					},
				},
			},

			Status: &kops.ClusterStatus{
				EtcdClusters: []kops.EtcdClusterStatus{
					kops.EtcdClusterStatus{
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
					&kops.EtcdMemberSpec{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
				},
			},

			NewSpec: &kops.EtcdClusterSpec{
				Name: "main",
				Members: []*kops.EtcdMemberSpec{
					&kops.EtcdMemberSpec{
						Name:          "a",
						InstanceGroup: fi.String("eu-central-1a"),
					},
				},
			},

			Status: &kops.ClusterStatus{
				EtcdClusters: []kops.EtcdClusterStatus{
					kops.EtcdClusterStatus{
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
			t.Error(g.Details)
		}
	}
}

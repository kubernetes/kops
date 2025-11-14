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

package gcemodel

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
)

func TestCloudTagsForInstanceGroup(t *testing.T) {
	c := newTestGCEModelContext()
	c.Cluster.Spec.CloudLabels = map[string]string{
		"cluster_label_key": "cluster_label_value",
		"test_label":        "from_cluster",
	}
	igCP := c.InstanceGroups[0]
	igCP.Spec.CloudLabels = map[string]string{
		"ig_cp_label_key": "ig_cp_label_value",
		"test_label":      "from_ig_cp",
	}
	igWorkers := c.InstanceGroups[1]
	igWorkers.Spec.CloudLabels = map[string]string{
		"ig_workers_label_key": "ig_workers_label_value",
		"test_label":           "from_ig_workers",
	}

	t.Run("control-plane", func(t *testing.T) {
		actual := c.CloudTagsForInstanceGroup(igCP)
		expected := map[string]string{
			"cluster_label_key": "cluster_label_value",
			"ig_cp_label_key":   "ig_cp_label_value",
			"test_label":        "from_ig_cp",

			"k8s-io-cluster-name":       "testcluster-test-com",
			"k8s-io-role-control-plane": "control-plane",
			"k8s-io-instance-group":     "cp-nodes",
			"k8s-io-role-master":        "master",
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected control-plane tags %+v, but got %+v", expected, actual)
		}
	})
	t.Run("workers", func(t *testing.T) {
		actual := c.CloudTagsForInstanceGroup(igWorkers)
		expected := map[string]string{
			"cluster_label_key":    "cluster_label_value",
			"ig_workers_label_key": "ig_workers_label_value",
			"test_label":           "from_ig_workers",

			"k8s-io-cluster-name":   "testcluster-test-com",
			"k8s-io-role-node":      "node",
			"k8s-io-instance-group": "worker-nodes",
		}

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("expected worker tags %+v, but got %+v", expected, actual)
		}
	})
}

func newTestGCEModelContext() *GCEModelContext {
	cluster := newTestCluster()
	igCP := newTestControlPlaneInstanceGroup()
	igWorkers := newTestWorkersInstanceGroup()
	return &GCEModelContext{
		KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{
				Cluster: cluster,
			},
			AllInstanceGroups: []*kops.InstanceGroup{igCP, igWorkers},
			InstanceGroups:    []*kops.InstanceGroup{igCP, igWorkers},
			SSHPublicKeys:     [][]byte{[]byte("ssh-rsa ...")},
		},
	}
}

func newTestCluster() *kops.Cluster {
	return &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testcluster.test.com",
		},
		Spec: kops.ClusterSpec{
			API: kops.APISpec{
				LoadBalancer: &kops.LoadBalancerAccessSpec{
					Type: kops.LoadBalancerTypeInternal,
				},
			},
			CloudProvider: kops.CloudProviderSpec{
				GCE: &kops.GCESpec{
					Project: "test-project",
				},
			},
			Networking: kops.NetworkingSpec{
				NetworkID:   "test-virtual-network",
				NetworkCIDR: "10.0.0.0/8",
				Subnets: []kops.ClusterSubnetSpec{
					{
						Name: "test-subnet",
						CIDR: "10.0.1.0/24",
						Type: kops.SubnetTypePrivate,
					},
				},
			},
		},
	}
}

func newTestControlPlaneInstanceGroup() *kops.InstanceGroup {
	return &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cp-nodes",
		},
		Spec: kops.InstanceGroupSpec{
			Role:  kops.InstanceGroupRoleControlPlane,
			Image: "Canonical:UbuntuServer:18.04-LTS:latest",
			RootVolume: &kops.InstanceRootVolumeSpec{
				Size: fi.PtrTo(int32(32)),
			},
			Subnets: []string{"test-subnet"},
		},
	}
}
func newTestWorkersInstanceGroup() *kops.InstanceGroup {
	return &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-nodes",
		},
		Spec: kops.InstanceGroupSpec{
			Role:  kops.InstanceGroupRoleNode,
			Image: "Canonical:UbuntuServer:18.04-LTS:latest",
			RootVolume: &kops.InstanceRootVolumeSpec{
				Size: fi.PtrTo(int32(32)),
			},
			Subnets: []string{"test-subnet"},
		},
	}
}

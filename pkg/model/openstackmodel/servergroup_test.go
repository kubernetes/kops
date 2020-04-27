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

package openstackmodel

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

func Test_ServerGroupModelBuilder(t *testing.T) {
	tests := []struct {
		desc                 string
		cluster              *kops.Cluster
		instanceGroups       []*kops.InstanceGroup
		clusterLifecycle     *fi.Lifecycle
		expectedTasksBuilder func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task
		expectedError        error
	}{
		{
			desc: "one master one node",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image-master",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image-node",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.2-4",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master"),
					ClusterName: s("cluster"),
					IGName:      s("master"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterPort := &openstacktasks.Port{
					Name:    s("port-master-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterInstance := &openstacktasks.Instance{
					Name:        s("master-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image-master"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-1-cluster"),
					Server:    masterInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-1-cluster"),
					Server:    nodeInstance,
					Lifecycle: &clusterLifecycle,
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master":      masterServerGroup,
					"Instance/master-1-cluster":       masterInstance,
					"Port/port-master-1-cluster":      masterPort,
					"FloatingIP/fip-master-1-cluster": masterFloatingIP,
					"ServerGroup/cluster-node":        nodeServerGroup,
					"Instance/node-1-cluster":         nodeInstance,
					"Port/port-node-1-cluster":        nodePort,
					"FloatingIP/fip-node-1-cluster":   nodeFloatingIP,
				}
			},
		},
		{
			desc: "one master one node one bastion",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
						{
							Name:   "utility-subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bastion",
					},
					Spec: kops.InstanceGroupSpec{
						AdditionalUserData: []kops.UserData{
							{
								Name:    "x",
								Type:    "shell",
								Content: "echo 'hello'",
							},
						},
						Role:        kops.InstanceGroupRoleBastion,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"utility-subnet"},
						Zones:       []string{"zone-1"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master"),
					ClusterName: s("cluster"),
					IGName:      s("master"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterPort := &openstacktasks.Port{
					Name:    s("port-master-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterInstance := &openstacktasks.Instance{
					Name:        s("master-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-1-cluster"),
					Server:    masterInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				bastionServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-bastion"),
					ClusterName: s("cluster"),
					IGName:      s("bastion"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				bastionPort := &openstacktasks.Port{
					Name:    s("port-bastion-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("bastion.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("utility-subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				bastionInstance := &openstacktasks.Instance{
					Name:        s("bastion-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: bastionServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Bastion"),
					Port:        bastionPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[2])),
					Metadata: map[string]string{
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "bastion",
						"KopsNetwork":               "cluster",
						"KopsRole":                  "Bastion",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/bastion":       "1",
						"kops.k8s.io/instancegroup": "bastion",
						"KubernetesCluster":         "cluster",
						"Name":                      "bastion.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				bastionFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-bastion-1-cluster"),
					Server:    bastionInstance,
					Lifecycle: &clusterLifecycle,
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master":       masterServerGroup,
					"Instance/master-1-cluster":        masterInstance,
					"Port/port-master-1-cluster":       masterPort,
					"FloatingIP/fip-master-1-cluster":  masterFloatingIP,
					"ServerGroup/cluster-node":         nodeServerGroup,
					"Instance/node-1-cluster":          nodeInstance,
					"Port/port-node-1-cluster":         nodePort,
					"ServerGroup/cluster-bastion":      bastionServerGroup,
					"Instance/bastion-1-cluster":       bastionInstance,
					"Port/port-bastion-1-cluster":      bastionPort,
					"FloatingIP/fip-bastion-1-cluster": bastionFloatingIP,
				}
			},
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet-a",
							Region: "region",
						},
						{
							Name:   "subnet-b",
							Region: "region",
						},
						{
							Name:   "subnet-c",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-a"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-a"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-b"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-b"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-c"},
						Zones:       []string{"zone-3"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-c"},
						Zones:       []string{"zone-3"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterAServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-a"),
					ClusterName: s("cluster"),
					IGName:      s("master-a"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterAPort := &openstacktasks.Port{
					Name:    s("port-master-a-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterAInstance := &openstacktasks.Instance{
					Name:        s("master-a-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterAServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-a",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-a",
						"Name":                      "master-a.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-a-1-cluster"),
					Server:    masterAInstance,
					Lifecycle: &clusterLifecycle,
				}
				masterBServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-b"),
					ClusterName: s("cluster"),
					IGName:      s("master-b"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterBPort := &openstacktasks.Port{
					Name:    s("port-master-b-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterBInstance := &openstacktasks.Instance{
					Name:        s("master-b-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterBServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-b",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-b",
						"Name":                      "master-b.masters.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				masterBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-b-1-cluster"),
					Server:    masterBInstance,
					Lifecycle: &clusterLifecycle,
				}
				masterCServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-c"),
					ClusterName: s("cluster"),
					IGName:      s("master-c"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterCPort := &openstacktasks.Port{
					Name:    s("port-master-c-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterCInstance := &openstacktasks.Instance{
					Name:        s("master-c-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterCServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-c",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-c",
						"Name":                      "master-c.masters.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				masterCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-c-1-cluster"),
					Server:    masterCInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeAServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-a"),
					ClusterName: s("cluster"),
					IGName:      s("node-a"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeAPort := &openstacktasks.Port{
					Name:    s("port-node-a-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeAInstance := &openstacktasks.Instance{
					Name:        s("node-a-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeAServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-a",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-a",
						"Name":                      "node-a.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-a-1-cluster"),
					Server:    nodeAInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeBServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-b"),
					ClusterName: s("cluster"),
					IGName:      s("node-b"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeBPort := &openstacktasks.Port{
					Name:    s("port-node-b-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeBInstance := &openstacktasks.Instance{
					Name:        s("node-b-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeBServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-b",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-b",
						"Name":                      "node-b.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-b-1-cluster"),
					Server:    nodeBInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeCServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-c"),
					ClusterName: s("cluster"),
					IGName:      s("node-c"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeCPort := &openstacktasks.Port{
					Name:    s("port-node-c-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeCInstance := &openstacktasks.Instance{
					Name:        s("node-c-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeCServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-c",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-c",
						"Name":                      "node-c.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-c-1-cluster"),
					Server:    nodeCInstance,
					Lifecycle: &clusterLifecycle,
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master-a":      masterAServerGroup,
					"Instance/master-a-1-cluster":       masterAInstance,
					"Port/port-master-a-1-cluster":      masterAPort,
					"FloatingIP/fip-master-a-1-cluster": masterAFloatingIP,
					"ServerGroup/cluster-master-b":      masterBServerGroup,
					"Instance/master-b-1-cluster":       masterBInstance,
					"Port/port-master-b-1-cluster":      masterBPort,
					"ServerGroup/cluster-master-c":      masterCServerGroup,
					"Instance/master-c-1-cluster":       masterCInstance,
					"Port/port-master-c-1-cluster":      masterCPort,
					"FloatingIP/fip-master-c-1-cluster": masterCFloatingIP,
					"FloatingIP/fip-master-b-1-cluster": masterBFloatingIP,
					"ServerGroup/cluster-node-a":        nodeAServerGroup,
					"Instance/node-a-1-cluster":         nodeAInstance,
					"Port/port-node-a-1-cluster":        nodeAPort,
					"FloatingIP/fip-node-a-1-cluster":   nodeAFloatingIP,
					"ServerGroup/cluster-node-b":        nodeBServerGroup,
					"Instance/node-b-1-cluster":         nodeBInstance,
					"Port/port-node-b-1-cluster":        nodeBPort,
					"FloatingIP/fip-node-b-1-cluster":   nodeBFloatingIP,
					"ServerGroup/cluster-node-c":        nodeCServerGroup,
					"Instance/node-c-1-cluster":         nodeCInstance,
					"Port/port-node-c-1-cluster":        nodeCPort,
					"FloatingIP/fip-node-c-1-cluster":   nodeCFloatingIP,
				}
			},
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion with API loadbalancer",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Loadbalancer: &kops.OpenstackLoadbalancerConfig{},
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet-a",
							Region: "region",
							Type:   kops.SubnetTypePrivate,
						},
						{
							Name:   "subnet-b",
							Region: "region",
							Type:   kops.SubnetTypePrivate,
						},
						{
							Name:   "subnet-c",
							Region: "region",
							Type:   kops.SubnetTypePrivate,
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-a"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-a"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-b"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-b"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-c"},
						Zones:       []string{"zone-3"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-c"},
						Zones:       []string{"zone-3"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterAServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-a"),
					ClusterName: s("cluster"),
					IGName:      s("master-a"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterAPort := &openstacktasks.Port{
					Name:    s("port-master-a-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterAInstance := &openstacktasks.Instance{
					Name:        s("master-a-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterAServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-a",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-a",
						"Name":                      "master-a.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterBServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-b"),
					ClusterName: s("cluster"),
					IGName:      s("master-b"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterBPort := &openstacktasks.Port{
					Name:    s("port-master-b-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterBInstance := &openstacktasks.Instance{
					Name:        s("master-b-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterBServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-b",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-b",
						"Name":                      "master-b.masters.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				masterCServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-c"),
					ClusterName: s("cluster"),
					IGName:      s("master-c"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterCPort := &openstacktasks.Port{
					Name:    s("port-master-c-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterCInstance := &openstacktasks.Instance{
					Name:        s("master-c-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterCServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-c",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-c",
						"Name":                      "master-c.masters.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeAServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-a"),
					ClusterName: s("cluster"),
					IGName:      s("node-a"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeAPort := &openstacktasks.Port{
					Name:    s("port-node-a-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeAInstance := &openstacktasks.Instance{
					Name:        s("node-a-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeAServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-a",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-a",
						"Name":                      "node-a.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-a-1-cluster"),
					Server:    nodeAInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeBServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-b"),
					ClusterName: s("cluster"),
					IGName:      s("node-b"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeBPort := &openstacktasks.Port{
					Name:    s("port-node-b-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeBInstance := &openstacktasks.Instance{
					Name:        s("node-b-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeBServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-b",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-b",
						"Name":                      "node-b.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-b-1-cluster"),
					Server:    nodeBInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeCServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-c"),
					ClusterName: s("cluster"),
					IGName:      s("node-c"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeCPort := &openstacktasks.Port{
					Name:    s("port-node-c-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeCInstance := &openstacktasks.Instance{
					Name:        s("node-c-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeCServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-c",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-c",
						"Name":                      "node-c.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-c-1-cluster"),
					Server:    nodeCInstance,
					Lifecycle: &clusterLifecycle,
				}
				loadbalancer := &openstacktasks.LB{
					Name:      s("master-public-name"),
					Subnet:    s("subnet-a.cluster"),
					Lifecycle: &clusterLifecycle,
					SecurityGroup: &openstacktasks.SecurityGroup{
						Name: s("master-public-name"),
					},
				}
				loadbalancerFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-public-name"),
					LB:        loadbalancer,
					Lifecycle: &clusterLifecycle,
				}
				lbPool := &openstacktasks.LBPool{
					Name:         s("master-public-name-https"),
					Loadbalancer: loadbalancer,
					Lifecycle:    &clusterLifecycle,
				}
				lbListener := &openstacktasks.LBListener{
					Name:      s("master-public-name"),
					Pool:      lbPool,
					Lifecycle: &clusterLifecycle,
				}
				masterAPoolAssoc := &openstacktasks.PoolAssociation{
					Name:          s("cluster-master-a"),
					Pool:          lbPool,
					ServerGroup:   masterAServerGroup,
					InterfaceName: s("cluster"),
					ProtocolPort:  fi.Int(443),
					Lifecycle:     &clusterLifecycle,
				}
				masterBPoolAssoc := &openstacktasks.PoolAssociation{
					Name:          s("cluster-master-b"),
					Pool:          lbPool,
					ServerGroup:   masterBServerGroup,
					InterfaceName: s("cluster"),
					ProtocolPort:  fi.Int(443),
					Lifecycle:     &clusterLifecycle,
				}
				masterCPoolAssoc := &openstacktasks.PoolAssociation{
					Name:          s("cluster-master-c"),
					Pool:          lbPool,
					ServerGroup:   masterCServerGroup,
					InterfaceName: s("cluster"),
					ProtocolPort:  fi.Int(443),
					Lifecycle:     &clusterLifecycle,
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master-a":      masterAServerGroup,
					"Instance/master-a-1-cluster":       masterAInstance,
					"Port/port-master-a-1-cluster":      masterAPort,
					"ServerGroup/cluster-master-b":      masterBServerGroup,
					"Instance/master-b-1-cluster":       masterBInstance,
					"Port/port-master-b-1-cluster":      masterBPort,
					"ServerGroup/cluster-master-c":      masterCServerGroup,
					"Instance/master-c-1-cluster":       masterCInstance,
					"Port/port-master-c-1-cluster":      masterCPort,
					"ServerGroup/cluster-node-a":        nodeAServerGroup,
					"Instance/node-a-1-cluster":         nodeAInstance,
					"Port/port-node-a-1-cluster":        nodeAPort,
					"FloatingIP/fip-node-a-1-cluster":   nodeAFloatingIP,
					"ServerGroup/cluster-node-b":        nodeBServerGroup,
					"Instance/node-b-1-cluster":         nodeBInstance,
					"Port/port-node-b-1-cluster":        nodeBPort,
					"FloatingIP/fip-node-b-1-cluster":   nodeBFloatingIP,
					"ServerGroup/cluster-node-c":        nodeCServerGroup,
					"Instance/node-c-1-cluster":         nodeCInstance,
					"Port/port-node-c-1-cluster":        nodeCPort,
					"FloatingIP/fip-node-c-1-cluster":   nodeCFloatingIP,
					"LB/master-public-name":             loadbalancer,
					"FloatingIP/fip-master-public-name": loadbalancerFloatingIP,
					"LBListener/master-public-name":     lbListener,
					"LBPool/master-public-name-https":   lbPool,
					"PoolAssociation/cluster-master-a":  masterAPoolAssoc,
					"PoolAssociation/cluster-master-b":  masterBPoolAssoc,
					"PoolAssociation/cluster-master-c":  masterCPoolAssoc,
				}
			},
		},
		{
			desc: "multizone setup 3 masters 3 nodes without external router",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet-a",
							Region: "region",
						},
						{
							Name:   "subnet-b",
							Region: "region",
						},
						{
							Name:   "subnet-c",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-a"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-a"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-b"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-b"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-c"},
						Zones:       []string{"zone-3"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-c"},
						Zones:       []string{"zone-3"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterAServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-a"),
					ClusterName: s("cluster"),
					IGName:      s("master-a"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterAPort := &openstacktasks.Port{
					Name:    s("port-master-a-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterAInstance := &openstacktasks.Instance{
					Name:        s("master-a-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterAServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-a",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-a",
						"Name":                      "master-a.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterBServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-b"),
					ClusterName: s("cluster"),
					IGName:      s("master-b"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterBPort := &openstacktasks.Port{
					Name:    s("port-master-b-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterBInstance := &openstacktasks.Instance{
					Name:        s("master-b-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterBServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-b",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-b",
						"Name":                      "master-b.masters.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				masterCServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master-c"),
					ClusterName: s("cluster"),
					IGName:      s("master-c"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterCPort := &openstacktasks.Port{
					Name:    s("port-master-c-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterCInstance := &openstacktasks.Instance{
					Name:        s("master-c-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterCServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-c",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master-c",
						"Name":                      "master-c.masters.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeAServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-a"),
					ClusterName: s("cluster"),
					IGName:      s("node-a"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeAPort := &openstacktasks.Port{
					Name:    s("port-node-a-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeAInstance := &openstacktasks.Instance{
					Name:        s("node-a-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeAServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-a",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-a",
						"Name":                      "node-a.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeBServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-b"),
					ClusterName: s("cluster"),
					IGName:      s("node-b"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeBPort := &openstacktasks.Port{
					Name:    s("port-node-b-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeBInstance := &openstacktasks.Instance{
					Name:        s("node-b-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeBServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-b",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-b",
						"Name":                      "node-b.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeCServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node-c"),
					ClusterName: s("cluster"),
					IGName:      s("node-c"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodeCPort := &openstacktasks.Port{
					Name:    s("port-node-c-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeCInstance := &openstacktasks.Instance{
					Name:        s("node-c-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeCServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node-c",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node-c",
						"Name":                      "node-c.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master-a": masterAServerGroup,
					"Instance/master-a-1-cluster":  masterAInstance,
					"Port/port-master-a-1-cluster": masterAPort,
					"ServerGroup/cluster-master-b": masterBServerGroup,
					"Instance/master-b-1-cluster":  masterBInstance,
					"Port/port-master-b-1-cluster": masterBPort,
					"ServerGroup/cluster-master-c": masterCServerGroup,
					"Instance/master-c-1-cluster":  masterCInstance,
					"Port/port-master-c-1-cluster": masterCPort,
					"ServerGroup/cluster-node-a":   nodeAServerGroup,
					"Instance/node-a-1-cluster":    nodeAInstance,
					"Port/port-node-a-1-cluster":   nodeAPort,
					"ServerGroup/cluster-node-b":   nodeBServerGroup,
					"Instance/node-b-1-cluster":    nodeBInstance,
					"Port/port-node-b-1-cluster":   nodeBPort,
					"ServerGroup/cluster-node-c":   nodeCServerGroup,
					"Instance/node-c-1-cluster":    nodeCInstance,
					"Port/port-node-c-1-cluster":   nodeCPort,
				}
			},
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion auto zone distribution",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet-a",
							Region: "region",
						},
						{
							Name:   "subnet-b",
							Region: "region",
						},
						{
							Name:   "subnet-c",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleMaster,
						Image:       "image",
						MinSize:     i32(3),
						MaxSize:     i32(3),
						MachineType: "blc.1-2",
						Subnets: []string{
							"subnet-a",
							"subnet-b",
							"subnet-c",
						},
						Zones: []string{
							"zone-1",
							"zone-2",
							"zone-3",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image",
						MinSize:     i32(3),
						MaxSize:     i32(3),
						MachineType: "blc.1-2",
						Subnets: []string{
							"subnet-a",
							"subnet-b",
							"subnet-c",
						},
						Zones: []string{
							"zone-1",
							"zone-2",
							"zone-3",
						},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master"),
					ClusterName: s("cluster"),
					IGName:      s("master"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(3),
				}
				masterAPort := &openstacktasks.Port{
					Name:    s("port-master-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterAInstance := &openstacktasks.Instance{
					Name:        s("master-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-1-cluster"),
					Server:    masterAInstance,
					Lifecycle: &clusterLifecycle,
				}
				masterBPort := &openstacktasks.Port{
					Name:    s("port-master-2-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterBInstance := &openstacktasks.Instance{
					Name:        s("master-2-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				masterBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-2-cluster"),
					Server:    masterBInstance,
					Lifecycle: &clusterLifecycle,
				}
				masterCPort := &openstacktasks.Port{
					Name:    s("port-master-3-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterCInstance := &openstacktasks.Instance{
					Name:        s("master-3-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				masterCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-3-cluster"),
					Server:    masterCInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(3),
				}
				nodeAPort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-a.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeAInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-1-cluster"),
					Server:    nodeAInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeBPort := &openstacktasks.Port{
					Name:    s("port-node-2-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-b.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeBInstance := &openstacktasks.Instance{
					Name:        s("node-2-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-2-cluster"),
					Server:    nodeBInstance,
					Lifecycle: &clusterLifecycle,
				}
				nodeCPort := &openstacktasks.Port{
					Name:    s("port-node-3-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet-c.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeCInstance := &openstacktasks.Instance{
					Name:        s("node-3-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-3-cluster"),
					Server:    nodeCInstance,
					Lifecycle: &clusterLifecycle,
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master":      masterServerGroup,
					"Instance/master-1-cluster":       masterAInstance,
					"Port/port-master-1-cluster":      masterAPort,
					"FloatingIP/fip-master-1-cluster": masterAFloatingIP,
					"Instance/master-2-cluster":       masterBInstance,
					"Port/port-master-2-cluster":      masterBPort,
					"FloatingIP/fip-master-2-cluster": masterBFloatingIP,
					"Instance/master-3-cluster":       masterCInstance,
					"Port/port-master-3-cluster":      masterCPort,
					"FloatingIP/fip-master-3-cluster": masterCFloatingIP,
					"ServerGroup/cluster-node":        nodeServerGroup,
					"Instance/node-1-cluster":         nodeAInstance,
					"Port/port-node-1-cluster":        nodeAPort,
					"FloatingIP/fip-node-1-cluster":   nodeAFloatingIP,
					"Instance/node-2-cluster":         nodeBInstance,
					"Port/port-node-2-cluster":        nodeBPort,
					"FloatingIP/fip-node-2-cluster":   nodeBFloatingIP,
					"Instance/node-3-cluster":         nodeCInstance,
					"Port/port-node-3-cluster":        nodeCPort,
					"FloatingIP/fip-node-3-cluster":   nodeCFloatingIP,
				}
			},
		},
		{
			desc: "one master one node without bastion no public ip association",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:              kops.InstanceGroupRoleMaster,
						Image:             "image-master",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.1-2",
						Subnets:           []string{"subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.Bool(false),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:              kops.InstanceGroupRoleNode,
						Image:             "image-node",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.2-4",
						Subnets:           []string{"subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.Bool(false),
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master"),
					ClusterName: s("cluster"),
					IGName:      s("master"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterPort := &openstacktasks.Port{
					Name:    s("port-master-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterInstance := &openstacktasks.Instance{
					Name:        s("master-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image-master"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master": masterServerGroup,
					"Instance/master-1-cluster":  masterInstance,
					"Port/port-master-1-cluster": masterPort,
					"ServerGroup/cluster-node":   nodeServerGroup,
					"Instance/node-1-cluster":    nodeInstance,
					"Port/port-node-1-cluster":   nodePort,
				}
			},
		},
		{
			desc: "one master one node one bastion",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.String("test"),
							},
						},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
						{
							Name:   "utility-subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:              kops.InstanceGroupRoleMaster,
						Image:             "image",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.1-2",
						Subnets:           []string{"subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.Bool(false),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:              kops.InstanceGroupRoleNode,
						Image:             "image",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.1-2",
						Subnets:           []string{"subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.Bool(false),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "bastion",
					},
					Spec: kops.InstanceGroupSpec{
						AdditionalUserData: []kops.UserData{
							{
								Name:    "x",
								Type:    "shell",
								Content: "echo 'hello'",
							},
						},
						Role:              kops.InstanceGroupRoleBastion,
						Image:             "image",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.1-2",
						Subnets:           []string{"utility-subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.Bool(false),
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				masterServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-master"),
					ClusterName: s("cluster"),
					IGName:      s("master"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				masterPort := &openstacktasks.Port{
					Name:    s("port-master-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("master-public-name")},
						{Name: s("masters.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				masterInstance := &openstacktasks.Instance{
					Name:        s("master-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: masterServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io/role/master":        "1",
						"kops.k8s.io/instancegroup": "master",
						"Name":                      "master.masters.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[1])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				bastionServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-bastion"),
					ClusterName: s("cluster"),
					IGName:      s("bastion"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				bastionPort := &openstacktasks.Port{
					Name:    s("port-bastion-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("bastion.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("utility-subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				bastionInstance := &openstacktasks.Instance{
					Name:        s("bastion-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.1-2"),
					Image:       s("image"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: bastionServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Bastion"),
					Port:        bastionPort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[2])),
					Metadata: map[string]string{
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "bastion",
						"KopsNetwork":               "cluster",
						"KopsRole":                  "Bastion",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/bastion":       "1",
						"kops.k8s.io/instancegroup": "bastion",
						"KubernetesCluster":         "cluster",
						"Name":                      "bastion.cluster",
					},
					AvailabilityZone: s("zone-1"),
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-master":  masterServerGroup,
					"Instance/master-1-cluster":   masterInstance,
					"Port/port-master-1-cluster":  masterPort,
					"ServerGroup/cluster-node":    nodeServerGroup,
					"Instance/node-1-cluster":     nodeInstance,
					"Port/port-node-1-cluster":    nodePort,
					"ServerGroup/cluster-bastion": bastionServerGroup,
					"Instance/bastion-1-cluster":  bastionInstance,
					"Port/port-bastion-1-cluster": bastionPort,
				}
			},
		},
		{
			desc: "adds additional security groups",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image-node",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.2-4",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
						AdditionalSecurityGroups: []string{
							"additional-sg",
						},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					AdditionalSecurityGroups: []string{
						"additional-sg",
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-1"),
					SecurityGroups: []string{
						"additional-sg",
					},
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-node": nodeServerGroup,
					"Instance/node-1-cluster":  nodeInstance,
					"Port/port-node-1-cluster": nodePort,
				}
			},
		},
		{
			desc: "uses instance group zones as availability zones",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image-node",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.2-4",
						Subnets:     []string{"subnet"},
						Zones: []string{
							"zone-a",
						},
						AdditionalSecurityGroups: []string{
							"additional-sg",
						},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					AdditionalSecurityGroups: []string{
						"additional-sg",
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("zone-a"),
					SecurityGroups: []string{
						"additional-sg",
					},
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-node": nodeServerGroup,
					"Instance/node-1-cluster":  nodeInstance,
					"Port/port-node-1-cluster": nodePort,
				}
			},
		},
		{
			desc: "uses instance group subnet as availability zones fallback",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image-node",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.2-4",
						Subnets:     []string{"subnet"},
						Zones:       []string{},
						AdditionalSecurityGroups: []string{
							"additional-sg",
						},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					AdditionalSecurityGroups: []string{
						"additional-sg",
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
					},
					AvailabilityZone: s("subnet"),
					SecurityGroups: []string{
						"additional-sg",
					},
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-node": nodeServerGroup,
					"Instance/node-1-cluster":  nodeInstance,
					"Port/port-node-1-cluster": nodePort,
				}
			},
		},
		{
			desc: "adds cloud labels from ClusterSpec",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
					CloudLabels: map[string]string{
						"some": "label",
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleNode,
						Image:       "image-node",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.2-4",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
						"some":                      "label",
					},
					AvailabilityZone: s("zone-1"),
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-node": nodeServerGroup,
					"Instance/node-1-cluster":  nodeInstance,
					"Port/port-node-1-cluster": nodePort,
				}
			},
		},
		{
			desc: "adds cloud labels from InstanceGroupSpec",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					MasterPublicName: "master-public-name",
					CloudConfig: &kops.CloudConfiguration{
						Openstack: &kops.OpenstackConfiguration{},
					},
					Subnets: []kops.ClusterSubnetSpec{
						{
							Name:   "subnet",
							Region: "region",
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Spec: kops.InstanceGroupSpec{
						Role: kops.InstanceGroupRoleNode,
						CloudLabels: map[string]string{
							"some": "label",
						},
						Image:       "image-node",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.2-4",
						Subnets:     []string{"subnet"},
						Zones:       []string{"zone-1"},
					},
				},
			},
			expectedTasksBuilder: func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task {
				clusterLifecycle := fi.LifecycleSync
				nodeServerGroup := &openstacktasks.ServerGroup{
					Name:        s("cluster-node"),
					ClusterName: s("cluster"),
					IGName:      s("node"),
					Policies:    []string{"anti-affinity"},
					Lifecycle:   &clusterLifecycle,
					MaxSize:     i32(1),
				}
				nodePort := &openstacktasks.Port{
					Name:    s("port-node-1-cluster"),
					Network: &openstacktasks.Network{Name: s("cluster")},
					SecurityGroups: []*openstacktasks.SecurityGroup{
						{Name: s("nodes.cluster")},
					},
					Subnets: []*openstacktasks.Subnet{
						{Name: s("subnet.cluster")},
					},
					Lifecycle: &clusterLifecycle,
				}
				nodeInstance := &openstacktasks.Instance{
					Name:        s("node-1-cluster"),
					Region:      s("region"),
					Flavor:      s("blc.2-4"),
					Image:       s("image-node"),
					SSHKey:      s("kubernetes.cluster-ba_d8_85_a0_5b_50_b0_01_e0_b2_b0_ae_5d_f6_7a_d1"),
					ServerGroup: nodeServerGroup,
					Tags:        []string{"KubernetesCluster:cluster"},
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    s(mustUserdataForClusterInstance(cluster, instanceGroups[0])),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io/role/node":          "1",
						"kops.k8s.io/instancegroup": "node",
						"Name":                      "node.cluster",
						"some":                      "label",
					},
					AvailabilityZone: s("zone-1"),
				}
				return map[string]fi.Task{
					"ServerGroup/cluster-node": nodeServerGroup,
					"Instance/node-1-cluster":  nodeInstance,
					"Port/port-node-1-cluster": nodePort,
				}
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.desc, func(t *testing.T) {
			clusterLifecycle := fi.LifecycleSync
			bootstrapScriptBuilder := &model.BootstrapScript{
				NodeUpConfigBuilder: func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
					return &nodeup.Config{}, nil
				},
				NodeUpSource:     "source",
				NodeUpSourceHash: "source-hash",
			}

			builder := createBuilderForCluster(testCase.cluster, testCase.instanceGroups, clusterLifecycle, bootstrapScriptBuilder)

			context := &fi.ModelBuilderContext{
				Tasks:              make(map[string]fi.Task),
				LifecycleOverrides: map[string]fi.Lifecycle{},
			}

			err := builder.Build(context)

			compareErrors(t, err, testCase.expectedError)

			expectedTasks := testCase.expectedTasksBuilder(testCase.cluster, testCase.instanceGroups)

			if len(expectedTasks) != len(context.Tasks) {
				t.Errorf("expected %d tasks, got %d tasks", len(expectedTasks), len(context.Tasks))
			}

			for taskName, task := range expectedTasks {
				actual, ok := context.Tasks[taskName]
				if !ok {
					t.Errorf("did not find a task for key %q", taskName)
					continue
				}
				switch expected := task.(type) {
				case *openstacktasks.ServerGroup:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareServerGroups(t, actual, expected)
					})
				case *openstacktasks.Port:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						comparePorts(t, actual, expected)
					})
				case *openstacktasks.FloatingIP:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareFloatingIPs(t, actual, expected)
					})
				case *openstacktasks.Instance:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareInstances(t, actual, expected)
					})
				case *openstacktasks.LB:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareLoadbalancers(t, actual, expected)
					})
				case *openstacktasks.LBPool:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareLBPools(t, actual, expected)
					})
				case *openstacktasks.PoolAssociation:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						comparePoolAssociations(t, actual, expected)
					})
				case *openstacktasks.LBListener:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareLBListeners(t, actual, expected)
					})
				case *openstacktasks.SecurityGroup:
					t.Run("creates a task for "+taskName, func(t *testing.T) {
						compareSecurityGroups(t, actual, expected)
					})
				default:
					t.Errorf("found a task with name %q and type %T", taskName, expected)
				}
			}
			if t.Failed() {
				t.Logf("created tasks:")
				for k := range context.Tasks {
					t.Logf("- %v", k)
				}
			}
		})
	}
}

func createBuilderForCluster(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup, clusterLifecycle fi.Lifecycle, bootstrapScript *model.BootstrapScript) *ServerGroupModelBuilder {
	sshPublicKey := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDF2sghZsClUBXJB4mBMIw8rb0hJWjg1Vz4eUeXwYmTdi92Gf1zNc5xISSip9Y+PWX/jJokPB7tgPnMD/2JOAKhG1bi4ZqB15pYRmbbBekVpM4o4E0dx+czbqjiAm6wlccTrINK5LYenbucAAQt19eH+D0gJwzYUK9SYz1hWnlGS+qurt2bz7rrsG73lN8E2eiNvGtIXqv3GabW/Hea3acOBgCUJQWUDTRu0OmmwxzKbFN/UpNKeRaHlCqwZWjVAsmqA8TX8LIocq7Np7MmIBwt7EpEeZJxThcmC8DEJs9ClAjD+jlLIvMPXKC3JWCPgwCLGxHjy7ckSGFCSzbyPduh")

	modelContext := &model.KopsModelContext{
		Cluster:        cluster,
		InstanceGroups: instanceGroups,
		SSHPublicKeys:  [][]byte{sshPublicKey},
	}
	openstackModelContext := &OpenstackModelContext{
		KopsModelContext: modelContext,
	}

	return &ServerGroupModelBuilder{
		OpenstackModelContext: openstackModelContext,
		BootstrapScript:       bootstrapScript,
		Lifecycle:             &clusterLifecycle,
	}
}

func comparePorts(t *testing.T, actualTask fi.Task, expected *openstacktasks.Port) {
	t.Helper()
	if pointersAreBothNil(t, "Port", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.Port)
	if !ok {
		t.Fatalf("task is not a port task, got %T", actualTask)
	}

	compareStrings(t, "Name", actual.Name, expected.Name)
	compareSecurityGroupLists(t, actual.SecurityGroups, expected.SecurityGroups)
	sort.Strings(actual.AdditionalSecurityGroups)
	sort.Strings(expected.AdditionalSecurityGroups)
	actualSgs := strings.Join(actual.AdditionalSecurityGroups, " ")
	expectedSgs := strings.Join(expected.AdditionalSecurityGroups, " ")
	if actualSgs != expectedSgs {
		t.Errorf("AdditionalSecurityGroups differ: %q instead of %q", actualSgs, expectedSgs)
	}
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	if actual.Network == nil {
		t.Fatal("Network is nil")
	}
	compareStrings(t, "Network name", actual.Network.Name, expected.Network.Name)
	if len(actual.Subnets) == len(expected.Subnets) {
		for i, subnet := range expected.Subnets {
			compareSubnets(t, actual.Subnets[i], subnet)
		}
	} else {
		compareNamedTasks(t, "Subnets", asHasName(actual.Subnets), asHasName(expected.Subnets))
	}
}

func asHasName(tasks interface{}) []fi.HasName {
	var namedTasks []fi.HasName
	rType := reflect.TypeOf(tasks)
	if rType.Kind() != reflect.Array && rType.Kind() != reflect.Slice {
		fmt.Printf("type is not an array or slice: %v\n", rType.Kind())
		return namedTasks
	}
	rVal := reflect.ValueOf(tasks)
	for i := 0; i < rVal.Len(); i++ {
		elem := rVal.Index(i)
		if named, ok := elem.Interface().(fi.HasName); ok {
			namedTasks = append(namedTasks, named)
		}
	}
	return namedTasks
}

func compareNamedTasks(t *testing.T, name string, actual, expected []fi.HasName) {
	actualTaskNames := make([]string, len(actual))
	for i, task := range actual {
		actualTaskNames[i] = *task.GetName()
	}
	sort.Strings(actualTaskNames)
	expectedTaskNames := make([]string, len(expected))
	for i, task := range expected {
		if task.GetName() == nil {
			expectedTaskNames[i] = ""
		} else {
			expectedTaskNames[i] = *task.GetName()
		}
	}
	sort.Strings(expectedTaskNames)
	if !reflect.DeepEqual(expectedTaskNames, actualTaskNames) {
		t.Errorf("%s differ: %v instead of %v", name, actualTaskNames, expectedTaskNames)
	}
}

func compareSubnets(t *testing.T, actualTask fi.Task, expected *openstacktasks.Subnet) {
	t.Helper()
	if pointersAreBothNil(t, "Subnet", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.Subnet)
	if !ok {
		t.Fatalf("task is not an Subnet task, got %T", actualTask)
	}

	compareStrings(t, "Name", actual.Name, expected.Name)
}

func compareInstances(t *testing.T, actualTask fi.Task, expected *openstacktasks.Instance) {
	t.Helper()
	if pointersAreBothNil(t, "Instance", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.Instance)
	if !ok {
		t.Fatalf("task is not an instance task, got %T", actualTask)
	}

	compareStrings(t, "Name", actual.Name, expected.Name)
	compareStrings(t, "Region", actual.Region, expected.Region)
	compareStrings(t, "Flavor", actual.Flavor, expected.Flavor)
	compareStrings(t, "Image", actual.Image, expected.Image)
	compareStrings(t, "SSHKey", actual.SSHKey, expected.SSHKey)
	compareStrings(t, "Role", actual.Role, expected.Role)
	compareUserData(t, actual.UserData, expected.UserData)
	compareStrings(t, "AvailabilityZone", actual.AvailabilityZone, expected.AvailabilityZone)
	comparePorts(t, actual.Port, expected.Port)
	compareServerGroups(t, actual.ServerGroup, expected.ServerGroup)
	if !reflect.DeepEqual(actual.Tags, expected.Tags) {
		t.Errorf("Tags differ:\n%v\n\tinstead of\n%v", actual.Tags, expected.Tags)
	}
	if !reflect.DeepEqual(actual.Metadata, expected.Metadata) {
		t.Errorf("Metadata differ:\n%v\n\tinstead of\n%v", actual.Metadata, expected.Metadata)
	}
	sort.Strings(actual.SecurityGroups)
	sort.Strings(expected.SecurityGroups)
	actualSgs := strings.Join(actual.SecurityGroups, " ")
	expectedSgs := strings.Join(expected.SecurityGroups, " ")
	if actualSgs != expectedSgs {
		t.Errorf("SecurityGroups differ: %q instead of %q", actualSgs, expectedSgs)
	}
}

func compareLoadbalancers(t *testing.T, actualTask fi.Task, expected *openstacktasks.LB) {
	t.Helper()
	if pointersAreBothNil(t, "Loadbalancer", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.LB)
	if !ok {
		t.Fatalf("task is not a loadbalancer task, got %T", actualTask)
	}
	compareStrings(t, "Name", actual.Name, expected.Name)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	compareStrings(t, "Subnet", actual.Subnet, expected.Subnet)
	compareSecurityGroupLists(t, []*openstacktasks.SecurityGroup{actual.SecurityGroup}, []*openstacktasks.SecurityGroup{expected.SecurityGroup})
}

func compareLBPools(t *testing.T, actualTask fi.Task, expected *openstacktasks.LBPool) {
	t.Helper()
	if pointersAreBothNil(t, "LBPool", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.LBPool)
	if !ok {
		t.Fatalf("task is not a LBPool task, got %T", actualTask)
	}
	compareStrings(t, "Name", actual.Name, expected.Name)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	compareLoadbalancers(t, actual.Loadbalancer, expected.Loadbalancer)
}

func compareLBListeners(t *testing.T, actualTask fi.Task, expected *openstacktasks.LBListener) {
	t.Helper()
	if pointersAreBothNil(t, "LBListener", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.LBListener)
	if !ok {
		t.Fatalf("task is not a LBListener task, got %T", actualTask)
	}
	compareStrings(t, "Name", actual.Name, expected.Name)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	compareLBPools(t, actual.Pool, expected.Pool)
}

func comparePoolAssociations(t *testing.T, actualTask fi.Task, expected *openstacktasks.PoolAssociation) {
	t.Helper()
	if pointersAreBothNil(t, "PoolAssociation", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.PoolAssociation)
	if !ok {
		t.Fatalf("task is not a PoolAssociation task, got %T", actualTask)
	}
	compareStrings(t, "Name", actual.Name, expected.Name)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	compareLBPools(t, actual.Pool, expected.Pool)
	compareInts(t, "ProtocolPort", actual.ProtocolPort, expected.ProtocolPort)
	compareStrings(t, "InterfaceName", actual.InterfaceName, expected.InterfaceName)
	compareServerGroups(t, actual.ServerGroup, expected.ServerGroup)
}

func compareServerGroups(t *testing.T, actualTask fi.Task, expected *openstacktasks.ServerGroup) {
	t.Helper()
	if pointersAreBothNil(t, "ServerGroup", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.ServerGroup)
	if !ok {
		t.Fatalf("task is not a server group task, got %T", actualTask)
	}
	compareStrings(t, "Name", actual.Name, expected.Name)
	compareStrings(t, "ClusterName", actual.ClusterName, expected.ClusterName)
	compareStrings(t, "IGName", actual.IGName, expected.IGName)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	if !reflect.DeepEqual(actual.Policies, expected.Policies) {
		t.Errorf("Policies differ:\n%v\n\tinstead of\n%v", actual.Policies, expected.Policies)
	}
	compareInt32s(t, "MaxSize", actual.MaxSize, expected.MaxSize)
}

func compareFloatingIPs(t *testing.T, actualTask fi.Task, expected *openstacktasks.FloatingIP) {
	t.Helper()
	if pointersAreBothNil(t, "FloatingIP", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.FloatingIP)
	if !ok {
		t.Fatalf("task is not a floating ip task, got %T", actualTask)
	}

	compareStrings(t, "Name", actual.Name, expected.Name)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	if pointersAreBothNil(t, "Server", actual.Server, expected.Server) {
		compareLoadbalancers(t, actual.LB, expected.LB)
	} else {
		compareInstances(t, actual.Server, expected.Server)
	}
}

func compareLifecycles(t *testing.T, actual, expected *fi.Lifecycle) {
	t.Helper()
	if pointersAreBothNil(t, "Lifecycle", actual, expected) {
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		var a, e string
		if actual != nil {
			a = string(*actual)
		}
		if expected != nil {
			e = string(*expected)
		}
		t.Errorf("Lifecycle differs: %+v instead of %+v", a, e)
	}
}

func compareSecurityGroups(t *testing.T, actualTask fi.Task, expected *openstacktasks.SecurityGroup) {
	t.Helper()
	if pointersAreBothNil(t, "SecurityGroup", actualTask, expected) {
		return
	}
	actual, ok := actualTask.(*openstacktasks.SecurityGroup)
	if !ok {
		t.Fatalf("task is not a security group task, got %T", actualTask)
	}

	compareStrings(t, "Name", actual.Name, expected.Name)
	compareLifecycles(t, actual.Lifecycle, expected.Lifecycle)
	compareStrings(t, "Description", actual.Description, expected.Description)
}

func compareSecurityGroupLists(t *testing.T, actual, expected []*openstacktasks.SecurityGroup) {
	sgs := make([]string, len(actual))
	for i, sg := range actual {
		sgs[i] = *sg.Name
	}
	sort.Strings(sgs)
	expectedSgs := make([]string, len(expected))
	for i, sg := range expected {
		if sg.Name == nil {
			expectedSgs[i] = ""
		} else {
			expectedSgs[i] = *sg.Name
		}
	}
	sort.Strings(expectedSgs)
	if !reflect.DeepEqual(expectedSgs, sgs) {
		t.Errorf("SecurityGroups differ: %v instead of %v", sgs, expectedSgs)
	}
}

func compareStrings(t *testing.T, name string, actual, expected *string) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		var a, e string
		if actual != nil {
			a = *actual
		}
		if expected != nil {
			e = *expected
		}
		t.Errorf("%s differs: %+v instead of %+v", name, a, e)
	}
}

func compareUserData(t *testing.T, actual, expected *string) {
	t.Helper()
	if pointersAreBothNil(t, "UserData", actual, expected) {
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		var a, e string
		if actual != nil {
			a = *actual
		}
		if expected != nil {
			e = *expected
		}
		aLines := strings.Split(a, "\n")
		eLines := strings.Split(e, "\n")
		sort.Strings(aLines)
		sort.Strings(eLines)
		if !reflect.DeepEqual(aLines, eLines) {
			t.Errorf("UserData differ: %+v instead of %+v", a, e)
		}
	}
}

func compareInts(t *testing.T, name string, actual, expected *int) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		var a, e int
		if actual != nil {
			a = *actual
		}
		if expected != nil {
			e = *expected
		}
		t.Errorf("%s differs: %+v instead of %+v", name, a, e)
	}
}

func compareInt32s(t *testing.T, name string, actual, expected *int32) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		var a, e int32
		if actual != nil {
			a = *actual
		}
		if expected != nil {
			e = *expected
		}
		t.Errorf("%s differs: %+v instead of %+v", name, a, e)
	}
}

func pointersAreBothNil(t *testing.T, name string, actual, expected interface{}) bool {
	t.Helper()
	if actual == nil && expected == nil {
		return true
	}
	if reflect.ValueOf(actual).IsNil() && reflect.ValueOf(expected).IsNil() {
		return true
	}
	if actual == nil && expected != nil {
		t.Fatalf("%s differ: actual is nil, expected is not", name)
	}
	if actual != nil && expected == nil {
		t.Fatalf("%s differ: expected is nil, actual is not", name)
	}
	return false
}

func compareErrors(t *testing.T, actual, expected error) {
	t.Helper()
	if pointersAreBothNil(t, "errors", actual, expected) {
		return
	}
	a := fmt.Sprintf("%v", actual)
	e := fmt.Sprintf("%v", expected)
	if a != e {
		t.Errorf("error differs: %+v instead of %+v", actual, expected)
	}
}

func mustUserdataForClusterInstance(cluster *kops.Cluster, ig *kops.InstanceGroup) string {
	bootstrapScriptBuilder := &model.BootstrapScript{
		NodeUpConfigBuilder: func(ig *kops.InstanceGroup) (*nodeup.Config, error) {
			return &nodeup.Config{}, nil
		},
		NodeUpSource:     "source",
		NodeUpSourceHash: "source-hash",
	}
	startupResources, err := bootstrapScriptBuilder.ResourceNodeUp(ig, cluster)
	if err != nil {
		panic(fmt.Errorf("error getting userdata: %v", err))
	}
	userdata, err := startupResources.AsString()
	if err != nil {
		panic(fmt.Errorf("error converting userdata to string: %v", err))
	}
	return userdata
}

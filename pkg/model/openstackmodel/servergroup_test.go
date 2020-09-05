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
	"path/filepath"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
	"k8s.io/kops/util/pkg/architectures"
)

type serverGroupModelBuilderTestInput struct {
	desc                 string
	cluster              *kops.Cluster
	instanceGroups       []*kops.InstanceGroup
	expectedTasksBuilder func(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) map[string]fi.Task
}

func getServerGroupModelBuilderTestInput() []serverGroupModelBuilderTestInput {
	return []serverGroupModelBuilderTestInput{
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
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-1-cluster"),
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
					Topology: &kops.TopologySpec{
						Nodes: "private",
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
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
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
					Role:        s("Bastion"),
					Port:        bastionPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[2]),
					Metadata: map[string]string{
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "bastion",
						"KopsNetwork":               "cluster",
						"KopsRole":                  "Bastion",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_bastion":       "1",
						"kops.k8s.io_instancegroup": "bastion",
					},
					AvailabilityZone: s("zone-1"),
				}
				bastionFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-bastion-1-cluster"),
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
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-a",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-a",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-a-1-cluster"),
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
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-b",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-b",
					},
					AvailabilityZone: s("zone-2"),
				}
				masterBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-b-1-cluster"),
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
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-c",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-c",
					},
					AvailabilityZone: s("zone-3"),
				}
				masterCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-c-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-a",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-a",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-a-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-b",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-b",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-b-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-c",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-c",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-c-1-cluster"),
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
					Topology: &kops.TopologySpec{
						Masters: kops.TopologyPrivate,
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
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-a",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-a",
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
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-b",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-b",
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
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-c",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-c",
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
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-a",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-a",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-a-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-b",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-b",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-b-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-c",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-c",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-c-1-cluster"),
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
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-a",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-a",
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
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-b",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-b",
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
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master-c",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master-c",
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
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-a",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-a",
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
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node-b",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-b",
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
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node-c",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node-c",
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
					Role:        s("Master"),
					Port:        masterAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
					},
					AvailabilityZone: s("zone-1"),
				}
				masterAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-1-cluster"),
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
					Role:        s("Master"),
					Port:        masterBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
					},
					AvailabilityZone: s("zone-2"),
				}
				masterBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-2-cluster"),
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
					Role:        s("Master"),
					Port:        masterCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
					},
					AvailabilityZone: s("zone-3"),
				}
				masterCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-master-3-cluster"),
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
					Role:        s("Node"),
					Port:        nodeAPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
					},
					AvailabilityZone: s("zone-1"),
				}
				nodeAFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-1-cluster"),
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
					Role:        s("Node"),
					Port:        nodeBPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
					},
					AvailabilityZone: s("zone-2"),
				}
				nodeBFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-2-cluster"),
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
					Role:        s("Node"),
					Port:        nodeCPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
					},
					AvailabilityZone: s("zone-3"),
				}
				nodeCFloatingIP := &openstacktasks.FloatingIP{
					Name:      s("fip-node-3-cluster"),
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
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
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
			desc: "one master one node one bastion 2",
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
					Role:        s("Master"),
					Port:        masterPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "master",
						"KopsRole":                  "Master",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io_role_master":        "1",
						"kops.k8s.io_instancegroup": "master",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[1]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
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
					Role:        s("Bastion"),
					Port:        bastionPort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[2]),
					Metadata: map[string]string{
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "bastion",
						"KopsNetwork":               "cluster",
						"KopsRole":                  "Bastion",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_bastion":       "1",
						"kops.k8s.io_instancegroup": "bastion",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"KopsNetwork":               "cluster",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
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
						"some%$/:X": "label",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
						"some___:x":                 "label",
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
							"some%$/:X": "label",
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
					Role:        s("Node"),
					Port:        nodePort,
					UserData:    mustUserdataForClusterInstance(cluster, instanceGroups[0]),
					Metadata: map[string]string{
						"KubernetesCluster":         "cluster",
						"k8s":                       "cluster",
						"KopsNetwork":               "cluster",
						"KopsInstanceGroup":         "node",
						"KopsRole":                  "Node",
						"ig_generation":             "0",
						"cluster_generation":        "0",
						"k8s.io_role_node":          "1",
						"kops.k8s.io_instancegroup": "node",
						"some___:x":                 "label",
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
}

func createBuilderForCluster(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup, clusterLifecycle fi.Lifecycle, bootstrapScriptBuilder *model.BootstrapScriptBuilder) *ServerGroupModelBuilder {
	sshPublicKey := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDF2sghZsClUBXJB4mBMIw8rb0hJWjg1Vz4eUeXwYmTdi92Gf1zNc5xISSip9Y+PWX/jJokPB7tgPnMD/2JOAKhG1bi4ZqB15pYRmbbBekVpM4o4E0dx+czbqjiAm6wlccTrINK5LYenbucAAQt19eH+D0gJwzYUK9SYz1hWnlGS+qurt2bz7rrsG73lN8E2eiNvGtIXqv3GabW/Hea3acOBgCUJQWUDTRu0OmmwxzKbFN/UpNKeRaHlCqwZWjVAsmqA8TX8LIocq7Np7MmIBwt7EpEeZJxThcmC8DEJs9ClAjD+jlLIvMPXKC3JWCPgwCLGxHjy7ckSGFCSzbyPduh")

	modelContext := &model.KopsModelContext{
		IAMModelContext: iam.IAMModelContext{Cluster: cluster},
		InstanceGroups:  instanceGroups,
		SSHPublicKeys:   [][]byte{sshPublicKey},
	}
	openstackModelContext := &OpenstackModelContext{
		KopsModelContext: modelContext,
	}

	return &ServerGroupModelBuilder{
		OpenstackModelContext:  openstackModelContext,
		BootstrapScriptBuilder: bootstrapScriptBuilder,
		Lifecycle:              &clusterLifecycle,
	}
}

type nodeupConfigBuilder struct {
}

func (n *nodeupConfigBuilder) BuildConfig(ig *kops.InstanceGroup, apiserverAdditionalIPs []string) (*nodeup.Config, error) {
	return &nodeup.Config{}, nil
}

func mustUserdataForClusterInstance(cluster *kops.Cluster, ig *kops.InstanceGroup) *fi.ResourceHolder {
	bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
		NodeUpConfigBuilder: &nodeupConfigBuilder{},
		NodeUpSource: map[architectures.Architecture]string{
			architectures.ArchitectureAmd64: "source-amd64",
			architectures.ArchitectureArm64: "source-arm64",
		},
		NodeUpSourceHash: map[architectures.Architecture]string{
			architectures.ArchitectureAmd64: "source-hash-amd64",
			architectures.ArchitectureArm64: "source-hash-arm64",
		},
	}
	c := &fi.ModelBuilderContext{
		Tasks: make(map[string]fi.Task),
	}

	startupResources, err := bootstrapScriptBuilder.ResourceNodeUp(c, ig)
	if err != nil {
		panic(fmt.Errorf("error getting userdata: %v", err))
	}

	err = c.Tasks["BootstrapScript/"+ig.Name].Run(&fi.Context{Cluster: cluster})
	if err != nil {
		panic(fmt.Errorf("error running BootstrapScript task: %v", err))

	}
	userdata, err := startupResources.AsString()
	if err != nil {
		panic(fmt.Errorf("error converting userdata to string: %v", err))
	}
	return fi.WrapResource(fi.NewStringResource(userdata))
}

func TestServerGroupBuilder(t *testing.T) {
	tests := getServerGroupModelBuilderTestInput()
	for _, testCase := range tests {
		RunGoldenTest(t, "tests/servergroup", testCase)
	}
}

func RunGoldenTest(t *testing.T, basedir string, testCase serverGroupModelBuilderTestInput) {
	h := testutils.NewIntegrationTestHarness(t)
	defer h.Close()

	h.MockKopsVersion("1.18.0")
	h.SetupMockOpenstack()

	clusterLifecycle := fi.LifecycleSync
	bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
		NodeUpConfigBuilder: &nodeupConfigBuilder{},
		NodeUpSource: map[architectures.Architecture]string{
			architectures.ArchitectureAmd64: "source-amd64",
			architectures.ArchitectureArm64: "source-arm64",
		},
		NodeUpSourceHash: map[architectures.Architecture]string{
			architectures.ArchitectureAmd64: "source-hash-amd64",
			architectures.ArchitectureArm64: "source-hash-arm64",
		},
	}

	builder := createBuilderForCluster(testCase.cluster, testCase.instanceGroups, clusterLifecycle, bootstrapScriptBuilder)

	context := &fi.ModelBuilderContext{
		Tasks:              make(map[string]fi.Task),
		LifecycleOverrides: map[string]fi.Lifecycle{},
	}

	builder.Build(context)

	file := filepath.Join(basedir, strings.ReplaceAll(testCase.desc, " ", "-")+".yaml")

	testutils.ValidateTasks(t, file, context)
}

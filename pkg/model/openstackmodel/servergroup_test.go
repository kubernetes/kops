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
	"path/filepath"
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/testutils"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/hashing"
)

type serverGroupModelBuilderTestInput struct {
	desc           string
	cluster        *kops.Cluster
	instanceGroups []*kops.InstanceGroup
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
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "one master one node one bastion",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePrivate,
								Region: "region",
							},
							{
								Name:   "utility-subnet",
								Type:   kops.SubnetTypeUtility,
								Region: "region",
							},
						},
						Topology: &kops.TopologySpec{},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet-a",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
							{
								Name:   "subnet-b",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
							{
								Name:   "subnet-c",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "truncate cluster names to 42 characters",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "tom-software-dev-playground-real33-k8s-local",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
						},
						Topology: &kops.TopologySpec{},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion with API loadbalancer",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Loadbalancer: &kops.OpenstackLoadbalancerConfig{},
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
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
						Topology: &kops.TopologySpec{},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-a",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion with API loadbalancer dns none",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						LoadBalancer: &kops.LoadBalancerAccessSpec{
							Type: kops.LoadBalancerTypePublic,
						},
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							BlockStorage: &kops.OpenstackBlockStorageConfig{
								Version:            fi.PtrTo("v3"),
								IgnoreAZ:           fi.PtrTo(false),
								CreateStorageClass: fi.PtrTo(false),
								CSITopologySupport: fi.PtrTo(true),
							},
							Loadbalancer: &kops.OpenstackLoadbalancerConfig{
								FloatingNetwork: fi.PtrTo("test"),
								FloatingSubnet:  fi.PtrTo("test-lb-subnet"),
								Method:          fi.PtrTo("ROUND_ROBIN"),
								Provider:        fi.PtrTo("amphora"),
								UseOctavia:      fi.PtrTo(true),
							},
							Monitor: &kops.OpenstackMonitor{
								Delay:      fi.PtrTo("1m"),
								MaxRetries: fi.PtrTo(3),
								Timeout:    fi.PtrTo("30s"),
							},
							Network: &kops.OpenstackNetwork{
								AvailabilityZoneHints: []*string{fi.PtrTo("zone-1"), fi.PtrTo("zone-2"), fi.PtrTo("zone-3")},
							},
							Router: &kops.OpenstackRouter{
								DNSServers:            fi.PtrTo("8.8.8.8,8.8.4.4"),
								ExternalSubnet:        fi.PtrTo("test-router-subnet"),
								ExternalNetwork:       fi.PtrTo("test"),
								AvailabilityZoneHints: []*string{fi.PtrTo("ha-zone")},
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.25.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name: "subnet-1",
								Zone: "zone-1",
								Type: kops.SubnetTypePrivate,
							},
							{
								Name: "subnet-2",
								Zone: "zone-2",
								Type: kops.SubnetTypePrivate,
							},
							{
								Name: "subnet-3",
								Zone: "zone-3",
								Type: kops.SubnetTypePrivate,
							},
						},
						Topology: &kops.TopologySpec{
							DNS: kops.DNSTypeNone,
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
						Role:        kops.InstanceGroupRoleControlPlane,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-1"},
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
						Subnets:     []string{"subnet-1"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-b",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-2"},
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
						Subnets:     []string{"subnet-2"},
						Zones:       []string{"zone-2"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-c",
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-3"},
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
						Subnets:     []string{"subnet-3"},
						Zones:       []string{"zone-3"},
					},
				},
			},
		},
		{
			desc: "single-zone setup 3 masters 1 node without bastion with API loadbalancer dns none",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						LoadBalancer: &kops.LoadBalancerAccessSpec{
							Type: kops.LoadBalancerTypePublic,
						},
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							BlockStorage: &kops.OpenstackBlockStorageConfig{
								Version:            fi.PtrTo("v3"),
								IgnoreAZ:           fi.PtrTo(false),
								CreateStorageClass: fi.PtrTo(false),
								CSITopologySupport: fi.PtrTo(true),
							},
							Loadbalancer: &kops.OpenstackLoadbalancerConfig{
								FloatingNetwork: fi.PtrTo("test"),
								FloatingSubnet:  fi.PtrTo("test-lb-subnet"),
								Method:          fi.PtrTo("ROUND_ROBIN"),
								Provider:        fi.PtrTo("amphora"),
								UseOctavia:      fi.PtrTo(true),
							},
							Monitor: &kops.OpenstackMonitor{
								Delay:      fi.PtrTo("1m"),
								MaxRetries: fi.PtrTo(3),
								Timeout:    fi.PtrTo("30s"),
							},
							Network: &kops.OpenstackNetwork{
								AvailabilityZoneHints: []*string{fi.PtrTo("zone-1")},
							},
							Router: &kops.OpenstackRouter{
								DNSServers:            fi.PtrTo("8.8.8.8,8.8.4.4"),
								ExternalSubnet:        fi.PtrTo("test-router-subnet"),
								ExternalNetwork:       fi.PtrTo("test"),
								AvailabilityZoneHints: []*string{fi.PtrTo("zone-1")},
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.25.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name: "subnet-1",
								Zone: "zone-1",
								Type: kops.SubnetTypePrivate,
							},
						},
						Topology: &kops.TopologySpec{
							DNS: kops.DNSTypeNone,
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-a",
						Annotations: map[string]string{
							"openstack.kops.io/serverGroupName": "control-plane",
						},
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-1"},
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
						Subnets:     []string{"subnet-1"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-b",
						Annotations: map[string]string{
							"openstack.kops.io/serverGroupName": "control-plane",
						},
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-1"},
						Zones:       []string{"zone-1"},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "master-c",
						Annotations: map[string]string{
							"openstack.kops.io/serverGroupName": "control-plane",
						},
					},
					Spec: kops.InstanceGroupSpec{
						Role:        kops.InstanceGroupRoleControlPlane,
						Image:       "image",
						MinSize:     i32(1),
						MaxSize:     i32(1),
						MachineType: "blc.1-2",
						Subnets:     []string{"subnet-1"},
						Zones:       []string{"zone-1"},
					},
				},
			},
		},
		{
			desc: "multizone setup 3 masters 3 nodes without external router",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet-a",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
							{
								Name:   "subnet-b",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
							{
								Name:   "subnet-c",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "multizone setup 3 masters 3 nodes without bastion auto zone distribution",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet-a",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
							{
								Name:   "subnet-b",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
							{
								Name:   "subnet-c",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
						Role:        kops.InstanceGroupRoleControlPlane,
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
		},
		{
			desc: "one master one node without bastion no public ip association",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
						Role:              kops.InstanceGroupRoleControlPlane,
						Image:             "image-master",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.1-2",
						Subnets:           []string{"subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.PtrTo(false),
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
						AssociatePublicIP: fi.PtrTo(false),
					},
				},
			},
		},
		{
			desc: "one master one node one bastion 2",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Router: &kops.OpenstackRouter{
								ExternalNetwork: fi.PtrTo("test"),
							},
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePrivate,
								Region: "region",
							},
							{
								Name:   "utility-subnet",
								Type:   kops.SubnetTypeUtility,
								Region: "region",
							},
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
						Role:              kops.InstanceGroupRoleControlPlane,
						Image:             "image",
						MinSize:           i32(1),
						MaxSize:           i32(1),
						MachineType:       "blc.1-2",
						Subnets:           []string{"subnet"},
						Zones:             []string{"zone-1"},
						AssociatePublicIP: fi.PtrTo(false),
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
						AssociatePublicIP: fi.PtrTo(false),
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
						AssociatePublicIP: fi.PtrTo(false),
					},
				},
			},
		},
		{
			desc: "adds additional security groups",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
		},
		{
			desc: "uses instance group zones as availability zones",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
		},
		{
			desc: "uses instance group subnet as availability zones fallback",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
		},
		{
			desc: "adds cloud labels from ClusterSpec",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
		},
		{
			desc: "adds cloud labels from InstanceGroupSpec",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
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
		},
		{
			desc: "configures server group affinity with annotations",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
						Annotations: map[string]string{
							"openstack.kops.io/serverGroupAffinity": "soft-anti-affinity",
						},
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
		},
		{
			desc: "configures allowed address pairs with annotations",
			cluster: &kops.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster",
				},
				Spec: kops.ClusterSpec{
					API: kops.APISpec{
						PublicName: "master-public-name",
					},
					CloudProvider: kops.CloudProviderSpec{
						Openstack: &kops.OpenstackSpec{
							Metadata: &kops.OpenstackMetadata{
								ConfigDrive: fi.PtrTo(false),
							},
						},
					},
					KubernetesVersion: "1.30.0",
					Networking: kops.NetworkingSpec{
						Subnets: []kops.ClusterSubnetSpec{
							{
								Name:   "subnet",
								Type:   kops.SubnetTypePublic,
								Region: "region",
							},
						},
					},
				},
			},
			instanceGroups: []*kops.InstanceGroup{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
						Annotations: map[string]string{
							"openstack.kops.io/allowedAddressPair/0": "192.168.0.0/16",
							"openstack.kops.io/allowedAddressPair/1": "10.123.0.1,12:34:56:78:90:AB",
						},
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
		},
	}
}

func createBuilderForCluster(cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup, clusterLifecycle fi.Lifecycle, bootstrapScriptBuilder *model.BootstrapScriptBuilder) *ServerGroupModelBuilder {
	sshPublicKey := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDF2sghZsClUBXJB4mBMIw8rb0hJWjg1Vz4eUeXwYmTdi92Gf1zNc5xISSip9Y+PWX/jJokPB7tgPnMD/2JOAKhG1bi4ZqB15pYRmbbBekVpM4o4E0dx+czbqjiAm6wlccTrINK5LYenbucAAQt19eH+D0gJwzYUK9SYz1hWnlGS+qurt2bz7rrsG73lN8E2eiNvGtIXqv3GabW/Hea3acOBgCUJQWUDTRu0OmmwxzKbFN/UpNKeRaHlCqwZWjVAsmqA8TX8LIocq7Np7MmIBwt7EpEeZJxThcmC8DEJs9ClAjD+jlLIvMPXKC3JWCPgwCLGxHjy7ckSGFCSzbyPduh")

	modelContext := &model.KopsModelContext{
		IAMModelContext:   iam.IAMModelContext{Cluster: cluster},
		AllInstanceGroups: instanceGroups,
		InstanceGroups:    instanceGroups,
		SSHPublicKeys:     [][]byte{sshPublicKey},
	}
	openstackModelContext := &OpenstackModelContext{
		KopsModelContext: modelContext,
	}

	return &ServerGroupModelBuilder{
		OpenstackModelContext:  openstackModelContext,
		BootstrapScriptBuilder: bootstrapScriptBuilder,
		Lifecycle:              clusterLifecycle,
	}
}

type nodeupConfigBuilder struct{}

func (n *nodeupConfigBuilder) BuildConfig(ig *kops.InstanceGroup, wellKnownAddresses model.WellKnownAddresses, keysets map[string]*fi.Keyset) (*nodeup.Config, *nodeup.BootConfig, error) {
	return &nodeup.Config{}, &nodeup.BootConfig{}, nil
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
	testutils.SetupMockOpenstack()

	clusterLifecycle := fi.LifecycleSync
	bootstrapScriptBuilder := &model.BootstrapScriptBuilder{
		KopsModelContext: &model.KopsModelContext{
			IAMModelContext:   iam.IAMModelContext{Cluster: testCase.cluster},
			AllInstanceGroups: testCase.instanceGroups,
			InstanceGroups:    testCase.instanceGroups,
		},
		NodeUpConfigBuilder: &nodeupConfigBuilder{},
		NodeUpAssets: map[architectures.Architecture]*assets.MirroredAsset{
			architectures.ArchitectureAmd64: {
				Locations: []string{"nodeup-amd64-1", "nodeup-amd64-2"},
				Hash:      hashing.MustFromString("833723369ad345a88dd85d61b1e77336d56e61b864557ded71b92b6e34158e6a"),
			},
			architectures.ArchitectureArm64: {
				Locations: []string{"nodeup-arm64-1", "nodeup-arm64-2"},
				Hash:      hashing.MustFromString("e525c28a65ff0ce4f95f9e730195b4e67fdcb15ceb1f36b5ad6921a8a4490c71"),
			},
		},
	}

	builder := createBuilderForCluster(testCase.cluster, testCase.instanceGroups, clusterLifecycle, bootstrapScriptBuilder)

	context := &fi.CloudupModelBuilderContext{
		Tasks:              make(map[string]fi.CloudupTask),
		LifecycleOverrides: map[string]fi.Lifecycle{},
	}

	// We need the CA and service-account for the bootstrap script
	caTask := &fitasks.Keypair{
		Name:    fi.PtrTo(fi.CertificateIDCA),
		Subject: "cn=kubernetes",
		Type:    "ca",
	}
	context.AddTask(caTask)
	for _, keypair := range []string{
		"apiserver-aggregator-ca",
		"etcd-clients-ca",
		"etcd-manager-ca-events",
		"etcd-manager-ca-main",
		"etcd-peers-ca-events",
		"etcd-peers-ca-main",
		"service-account",
	} {
		task := &fitasks.Keypair{
			Name:    fi.PtrTo(keypair),
			Subject: "cn=" + keypair,
			Type:    "ca",
		}
		context.AddTask(task)
	}
	for _, keypair := range []string{
		"kubelet",
		"kube-proxy",
	} {
		task := &fitasks.Keypair{
			Name:    fi.PtrTo(keypair),
			Subject: "cn=" + keypair,
			Signer:  caTask,
			Type:    "client",
		}
		context.AddTask(task)
	}

	if err := builder.Build(context); err != nil {
		t.Fatalf("error from Build: %v", err)
	}

	file := filepath.Join(basedir, strings.ReplaceAll(testCase.desc, " ", "-")+".yaml")

	testutils.ValidateTasks(t, file, context)
}

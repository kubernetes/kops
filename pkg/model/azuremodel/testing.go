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

package azuremodel

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
)

func newTestAzureModelContext() *AzureModelContext {
	cluster := newTestCluster()
	ig := newTestInstanceGroup()
	return &AzureModelContext{
		KopsModelContext: &model.KopsModelContext{
			IAMModelContext: iam.IAMModelContext{
				Cluster: cluster,
			},
			InstanceGroups: []*kops.InstanceGroup{ig},
			SSHPublicKeys:  [][]byte{[]byte("ssh-rsa ...")},
		},
	}
}

func newTestCluster() *kops.Cluster {
	return &kops.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "testcluster.test.com",
		},
		Spec: kops.ClusterSpec{
			API: &kops.AccessSpec{
				LoadBalancer: &kops.LoadBalancerAccessSpec{
					Type: kops.LoadBalancerTypeInternal,
				},
			},
			CloudProvider: kops.CloudProviderSpec{
				Azure: &kops.AzureSpec{
					ResourceGroupName: "test-resource-group",
					RouteTableName:    "test-route-table",
				},
			},
			Networking:  &kops.NetworkingSpec{},
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
	}
}

func newTestInstanceGroup() *kops.InstanceGroup {
	return &kops.InstanceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nodes",
		},
		Spec: kops.InstanceGroupSpec{
			Role:           kops.InstanceGroupRoleNode,
			Image:          "Canonical:UbuntuServer:18.04-LTS:latest",
			RootVolumeSize: fi.Int32(32),
			Subnets:        []string{"test-subnet"},
		},
	}
}

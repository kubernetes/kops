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
	"reflect"
	"testing"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

func TestAPILoadBalancerModelBuilder_Build(t *testing.T) {
	b := APILoadBalancerModelBuilder{
		AzureModelContext: newTestAzureModelContext(),
	}
	b.InstanceGroups[0].Spec.Role = kops.InstanceGroupRoleControlPlane
	c := &fi.CloudupModelBuilderContext{
		Tasks: make(map[string]fi.CloudupTask),
	}
	err := b.Build(c)
	if err != nil {
		t.Errorf("unexpected error %s", err)
	}
}

func TestSubnetForLoadbalancer(t *testing.T) {
	b := APILoadBalancerModelBuilder{
		AzureModelContext: newTestAzureModelContext(),
	}
	b.Cluster.Spec.Networking.Subnets = []kops.ClusterSubnetSpec{
		{
			Name: "master",
			Type: kops.SubnetTypePrivate,
		},
		{
			Name: "node",
			Type: kops.SubnetTypePrivate,
		},
		{
			Name: "utility",
			Type: kops.SubnetTypeUtility,
		},
	}
	b.InstanceGroups[0].Spec.Role = kops.InstanceGroupRoleControlPlane
	b.InstanceGroups[0].Spec.Subnets = []string{
		"master",
	}

	actual, err := b.subnetForLoadBalancer()
	if err != nil {
		t.Error(err)
	}
	expected := &kops.ClusterSubnetSpec{
		Name: "master",
		Type: kops.SubnetTypePrivate,
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected subnet %+v, but got %+v", expected, actual)
	}
}

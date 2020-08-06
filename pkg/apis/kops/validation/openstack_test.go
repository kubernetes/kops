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

func Test_ValidateTopology(t *testing.T) {
	grid := []struct {
		Input          kops.ClusterSpec
		ExpectedErrors []string
	}{
		{
			Input: kops.ClusterSpec{
				CloudConfig: &kops.CloudConfiguration{
					Openstack: &kops.OpenstackConfiguration{},
				},
			},
			ExpectedErrors: []string{
				"Forbidden::spec.topology.nodes",
				"Forbidden::spec.topology.masters",
			},
		},
		{
			Input: kops.ClusterSpec{
				CloudConfig: &kops.CloudConfiguration{
					Openstack: &kops.OpenstackConfiguration{
						Router: &kops.OpenstackRouter{},
					},
				},
			},
			ExpectedErrors: []string{
				"Forbidden::spec.topology.nodes",
				"Forbidden::spec.topology.masters",
			},
		},
		{

			Input: kops.ClusterSpec{
				CloudConfig: &kops.CloudConfiguration{
					Openstack: &kops.OpenstackConfiguration{},
				},
				Topology: &kops.TopologySpec{
					Masters: kops.TopologyPrivate,
					Nodes:   kops.TopologyPrivate,
				},
			},
			ExpectedErrors: []string{},
		},
		{
			Input: kops.ClusterSpec{
				CloudConfig: &kops.CloudConfiguration{
					Openstack: &kops.OpenstackConfiguration{
						Router: &kops.OpenstackRouter{
							ExternalNetwork: fi.String("foo"),
						},
					},
				},
			},
			ExpectedErrors: []string{},
		},
	}

	for _, g := range grid {
		cluster := &kops.Cluster{
			Spec: g.Input,
		}
		errs := openstackValidateCluster(cluster)
		testErrors(t, g.Input, errs, g.ExpectedErrors)
	}
}

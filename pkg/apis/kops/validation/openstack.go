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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
)

func openstackValidateCluster(c *kops.Cluster) (errList field.ErrorList) {
	if c.Spec.CloudConfig.Openstack.Router == nil || c.Spec.CloudConfig.Openstack.Router.ExternalNetwork == nil {
		topology := c.Spec.Topology
		if topology == nil || topology.Nodes == kops.TopologyPublic {
			errList = append(errList, field.Forbidden(field.NewPath("spec", "topology", "nodes"), "Public topology requires an external network"))
		}
		if topology == nil || topology.Masters == kops.TopologyPublic {
			errList = append(errList, field.Forbidden(field.NewPath("spec", "topology", "masters"), "Public topology requires an external network"))
		}
	}
	return errList
}

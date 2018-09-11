/*
Copyright 2018 The Kubernetes Authors.

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
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstacktasks"
)

// APILBModelBuilder configures loadbalancer objects
type APILBModelBuilder struct {
	*OpenstackModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &APILBModelBuilder{}

func (b *APILBModelBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.UseLoadBalancerForAPI() {
		return nil
	}

	lbSpec := b.Cluster.Spec.API.LoadBalancer
	if lbSpec == nil {
		// Skipping API LB creation; not requested in Spec
		return nil
	}

	switch lbSpec.Type {
	case kops.LoadBalancerTypePublic:
	// OK

	case kops.LoadBalancerTypeInternal:
		return fmt.Errorf("internal loadbalancers are not yet supported by kops on openstack")

	default:
		return fmt.Errorf("unhandled loadbalancers type %q", lbSpec.Type)
	}

	clusterName := b.ClusterName()
	lbName := "api-" + strings.Replace(clusterName, ".", "-", -1)

	{
		t := &openstacktasks.LB{
			Name:      s(lbName),
			Subnet:    b.LinkToSubnet(s(b.Cluster.Spec.Subnets[0].Name)),
			Lifecycle: b.Lifecycle,
		}

		c.AddTask(t)
	}

	return nil
}

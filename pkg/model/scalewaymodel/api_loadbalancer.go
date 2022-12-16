/*
Copyright 2022 The Kubernetes Authors.

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

package scalewaymodel

import (
	"fmt"
	"os"

	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/scalewaytasks"
)

// APILoadBalancerModelBuilder builds a load-balancer for accessing the API
type APILoadBalancerModelBuilder struct {
	*ScwModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &APILoadBalancerModelBuilder{}

func (b *APILoadBalancerModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	// Configuration where a load balancer fronts the API
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
		klog.V(8).Infof("Using public load-balancer")
	case kops.LoadBalancerTypeInternal:
		return fmt.Errorf("internal load-balancers are not yet supported by Scaleway on kops")
	default:
		return fmt.Errorf("unhandled load-balancer type %q", lbSpec.Type)
	}

	loadBalancerName := "api." + b.ClusterName()
	region, err := scw.ParseRegion(os.Getenv("SCW_DEFAULT_REGION"))
	if err != nil {
		return fmt.Errorf("error building load-balancer task for %q: %w", loadBalancerName, err)
	}

	loadBalancer := &scalewaytasks.LoadBalancer{
		Name:      fi.PtrTo(loadBalancerName),
		Region:    fi.PtrTo(string(region)),
		Lifecycle: b.Lifecycle,
		Tags: []string{
			scaleway.TagClusterName + "=" + b.ClusterName(),
			scaleway.TagNameRolePrefix + "=" + scaleway.TagRoleMaster,
		},
	}

	c.AddTask(loadBalancer)

	if dns.IsGossipClusterName(b.Cluster.Name) || b.Cluster.UsesPrivateDNS() || b.Cluster.UsesNoneDNS() {
		// Ensure the LB hostname is included in the TLS certificate,
		// if we're not going to use an alias for it
		loadBalancer.ForAPIServer = true
	}

	return nil
}

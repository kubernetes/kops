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

package yandexmodel

import (
	"fmt"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/loadbalancer/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/yandextasks"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
)

// APILoadBalancerModelBuilder builds a LoadBalancer for accessing the API
type APILoadBalancerModelBuilder struct {
	*YandexModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &APILoadBalancerModelBuilder{}

// createPublicLB validates the existence of a target pool with the given name,
// and creates an IP address and forwarding rule pointing to that target pool.
func createPublicLB(b *APILoadBalancerModelBuilder, c *fi.ModelBuilderContext) error {
	targetPool := &yandextasks.TargetGroup{
		Name:      fi.String("api"), //TODO(YuraBeznos): add check if it exists already and make configuration for it
		FolderId:  b.Cluster.Spec.Project,
		Lifecycle: b.Lifecycle,
	}
	c.AddTask(targetPool)

	loadBalancer := &yandextasks.LoadBalancer{
		FolderId:    b.Cluster.Spec.Project,
		Name:        fi.String("api"),
		Lifecycle:   b.Lifecycle,
		Description: b.ClusterName(),
		//Labels: map[string]string{
		//	yandex.TagKubernetesClusterName: b.ClusterName(),
		//},
		Type: loadbalancer.NetworkLoadBalancer_EXTERNAL,
	}

	c.AddTask(loadBalancer)
	return nil

}

func (b *APILoadBalancerModelBuilder) Build(c *fi.ModelBuilderContext) error {
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
		return createPublicLB(b, c)

	default:
		return fmt.Errorf("unhandled LoadBalancer type %q", lbSpec.Type)
	}
}

// subnetNotSpecified returns true if the given LB subnet is not listed in the list of cluster subnets.
func subnetNotSpecified(sn kops.LoadBalancerSubnetSpec, subnets []kops.ClusterSubnetSpec) bool {
	for _, csn := range subnets {
		if csn.Name == sn.Name || csn.ProviderID == sn.Name {
			return false
		}
	}
	return true
}

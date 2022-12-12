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

package components

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeProxyOptionsBuilder adds options for kube-proxy to the model
type KubeProxyOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeProxyOptionsBuilder{}

func (b *KubeProxyOptionsBuilder) BuildOptions(ctx context.Context, o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	if clusterSpec.KubeProxy == nil {
		clusterSpec.KubeProxy = &kops.KubeProxyConfig{}
	}

	config := clusterSpec.KubeProxy

	if config.LogLevel == 0 {
		// TODO: No way to set to 0?
		config.LogLevel = 2
	}

	// Any change here should be accompanied by a proportional change in CPU
	// requests of other per-node add-ons (e.g. fluentd).
	if config.CPURequest == nil {
		config.CPURequest = resource.NewScaledQuantity(100, resource.Milli)
	}

	image, err := Image(ctx, "kube-proxy", clusterSpec, b.Context.AssetBuilder)
	if err != nil {
		return err
	}
	config.Image = image

	// We set the master URL during node configuration, because if we use the internal dns name,
	// then we get a circular dependency:
	// * kube-proxy uses DNS for resolution
	// * dns is set up by dns-controller
	// * dns-controller talks to the API using the kube-proxy configured kubernetes service

	if config.ClusterCIDR == nil {
		if b.needsClusterCIDR(clusterSpec) {
			config.ClusterCIDR = fi.PtrTo(clusterSpec.KubeControllerManager.ClusterCIDR)
		}
	}

	return nil
}

func (*KubeProxyOptionsBuilder) needsClusterCIDR(clusterSpec *kops.ClusterSpec) bool {
	// If we use podCIDR from cloud, we should not set cluster cidr.
	if clusterSpec.IsKopsControllerIPAM() {
		return false
	}

	// If we're using the AmazonVPC networking, we should omit the ClusterCIDR
	// because pod IPs are real, routable IPs in the VPC, and they are not in a specific
	// CIDR range that allows us to distinguish them from other IPs.  Omitting the ClusterCIDR
	// causes kube-proxy never to SNAT when proxying clusterIPs, which is the behavior
	// we want for pods.
	// If we're not using the AmazonVPC networking, and the KubeControllerMananger has
	// a ClusterCIDR, use that because most networking plug ins draw pod IPs from this range.
	if clusterSpec.Networking.AmazonVPC != nil {
		return false
	}

	// If KCM doesn't have a ClusterCIDR, KubeProxy should not either.
	if clusterSpec.KubeControllerManager == nil || clusterSpec.KubeControllerManager.ClusterCIDR == "" {
		return false
	}

	return true
}

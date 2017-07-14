/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KubeProxyOptionsBuilder adds options for kube-proxy to the model
type KubeProxyOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &KubeProxyOptionsBuilder{}

func (b *KubeProxyOptionsBuilder) BuildOptions(o interface{}) error {
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
	if config.CPURequest == "" {
		config.CPURequest = "100m"
	}

	image, err := Image("kube-proxy", clusterSpec)
	if err != nil {
		return err
	}
	config.Image = image

	// We set the master URL during node configuration, because if we use the internal dns name,
	// then we get a circular dependency:
	// * kube-proxy uses DNS for resolution
	// * dns is set up by dns-controller
	// * dns-controller talks to the API using the kube-proxy configured kubernetes service

	if config.ClusterCIDR == "" {
		if clusterSpec.KubeControllerManager != nil {
			config.ClusterCIDR = clusterSpec.KubeControllerManager.ClusterCIDR
		}
	}

	// Set the kube-proxy hostname-override (actually the NodeName), to avoid #2915 et al
	cloudProvider := kops.CloudProviderID(clusterSpec.CloudProvider)
	if cloudProvider == kops.CloudProviderAWS {
		// Use the hostname from the AWS metadata service
		config.HostnameOverride = "@aws"
	}

	return nil
}

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

package components

import (
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// CiliumOptionsBuilder adds options for the cilium to the model
type CiliumOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.OptionsBuilder = &CiliumOptionsBuilder{}

func (b *CiliumOptionsBuilder) BuildOptions(o interface{}) error {
	clusterSpec := o.(*kops.ClusterSpec)
	c := clusterSpec.Networking.Cilium
	if c == nil {
		return nil
	}

	if c.Version == "" {
		if b.Context.IsKubernetesLT("1.12.0") {
			c.Version = "v1.6.9"
		} else {
			c.Version = "v1.7.4"
		}
	}

	if c.BPFCTGlobalAnyMax == 0 {
		c.BPFCTGlobalAnyMax = 262144

	}
	if c.BPFCTGlobalTCPMax == 0 {
		c.BPFCTGlobalTCPMax = 524288
	}

	if c.ClusterName == "" {
		c.ClusterName = "default"
	}

	if c.MonitorAggregation == "" {
		c.MonitorAggregation = "medium"
	}

	if c.SidecarIstioProxyImage == "" {
		c.SidecarIstioProxyImage = "cilium/istio_proxy"
	}

	if c.Tunnel == "" {
		c.Tunnel = "vxlan"
	}

	if c.ToFqdnsDNSRejectResponseCode == "" {
		c.ToFqdnsDNSRejectResponseCode = "refused"
	}

	if c.ContainerRuntimeLabels == "" {
		c.ContainerRuntimeLabels = "none"
	}

	if c.AgentPrometheusPort == 0 {
		c.AgentPrometheusPort = 9090
	}

	return nil

}

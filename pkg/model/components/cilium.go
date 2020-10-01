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
	"github.com/blang/semver/v4"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
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
			c.Version = "v1.6.12"
		} else {
			c.Version = "v1.8.4"
		}
	}

	version, _ := semver.ParseTolerant(c.Version)

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
		c.AgentPrometheusPort = wellknownports.CiliumPrometheusPort
	}

	if c.Ipam == "" {
		if version.Minor >= 8 {
			c.Ipam = "kubernetes"
		} else {
			c.Ipam = "hostscope"
		}
	}

	if c.EnableRemoteNodeIdentity == nil {
		c.EnableRemoteNodeIdentity = fi.Bool(true)
	}

	hubble := c.Hubble
	if hubble != nil {
		if hubble.Enabled == nil {
			hubble.Enabled = fi.Bool(true)
		}
	} else {
		c.Hubble = &kops.HubbleSpec{
			Enabled: fi.Bool(false),
		}
	}

	return nil

}

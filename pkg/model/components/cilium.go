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
	"k8s.io/apimachinery/pkg/api/resource"
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
		c.Version = "v1.10.0"
	}

	if c.EnableEndpointHealthChecking == nil {
		c.EnableEndpointHealthChecking = fi.Bool(true)
	}

	if c.IdentityAllocationMode == "" {
		c.IdentityAllocationMode = "crd"
	}

	if c.IdentityChangeGracePeriod == "" {
		c.IdentityChangeGracePeriod = "5s"
	}

	if c.BPFCTGlobalAnyMax == 0 {
		c.BPFCTGlobalAnyMax = 262144

	}

	if c.BPFCTGlobalTCPMax == 0 {
		c.BPFCTGlobalTCPMax = 524288
	}

	if c.BPFLBAlgorithm == "" {
		c.BPFLBAlgorithm = "random"
	}

	if c.BPFLBMaglevTableSize == "" {
		c.BPFLBMaglevTableSize = "16381"
	}

	if c.BPFNATGlobalMax == 0 {
		c.BPFNATGlobalMax = 524288
	}

	if c.BPFNeighGlobalMax == 0 {
		c.BPFNeighGlobalMax = 524288
	}

	if c.BPFPolicyMapMax == 0 {
		c.BPFPolicyMapMax = 16384
	}

	if c.BPFLBMapMax == 0 {
		c.BPFLBMapMax = 65536
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
		c.Ipam = "kubernetes"
	}

	if c.DisableMasquerade == nil {
		c.DisableMasquerade = fi.Bool(c.Ipam == "eni")
	}

	if c.Tunnel == "" {
		if c.Ipam == "eni" {
			c.Tunnel = "disabled"
		} else {
			c.Tunnel = "vxlan"
		}
	}

	if c.EnableRemoteNodeIdentity == nil {
		c.EnableRemoteNodeIdentity = fi.Bool(true)
	}

	if c.EnableBPFMasquerade == nil {
		c.EnableBPFMasquerade = fi.Bool(false)
	}

	if c.EnableL7Proxy == nil {
		c.EnableL7Proxy = fi.Bool(true)
	}

	if c.CPURequest == nil {
		defaultCPURequest := resource.MustParse("25m")
		c.CPURequest = &defaultCPURequest
	}

	if c.MemoryRequest == nil {
		defaultMemoryRequest := resource.MustParse("128Mi")
		c.MemoryRequest = &defaultMemoryRequest
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

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

var _ loader.ClusterOptionsBuilder = &CiliumOptionsBuilder{}

func (b *CiliumOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	clusterSpec := &o.Spec
	c := clusterSpec.Networking.Cilium
	if c == nil {
		return nil
	}

	if c.Version == "" {
		c.Version = "v1.16.2"
	}

	if c.EnableEndpointHealthChecking == nil {
		c.EnableEndpointHealthChecking = fi.PtrTo(true)
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

	if c.ToFQDNsDNSRejectResponseCode == "" {
		c.ToFQDNsDNSRejectResponseCode = "refused"
	}

	if c.AgentPrometheusPort == 0 {
		c.AgentPrometheusPort = wellknownports.CiliumPrometheusPort
	}

	if c.IPAM == "" {
		c.IPAM = "kubernetes"
	}

	if c.Masquerade == nil {
		c.Masquerade = fi.PtrTo(!clusterSpec.IsIPv6Only())
	}

	if c.Tunnel == "" {
		if c.IPAM == "eni" || clusterSpec.IsIPv6Only() {
			c.Tunnel = "disabled"
		} else {
			c.Tunnel = "vxlan"
		}
	}

	if c.EnableRemoteNodeIdentity == nil {
		c.EnableRemoteNodeIdentity = fi.PtrTo(true)
	}

	if c.EnableUnreachableRoutes == nil {
		c.EnableUnreachableRoutes = fi.PtrTo(false)
	}

	if c.EnableBPFMasquerade == nil {
		c.EnableBPFMasquerade = fi.PtrTo(c.IPAM == "eni")
	}

	if c.EnableL7Proxy == nil {
		c.EnableL7Proxy = fi.PtrTo(true)
	}

	if c.DisableCNPStatusUpdates == nil {
		c.DisableCNPStatusUpdates = fi.PtrTo(true)
	}

	if c.CPURequest == nil {
		defaultCPURequest := resource.MustParse("25m")
		c.CPURequest = &defaultCPURequest
	}

	if c.MemoryRequest == nil {
		defaultMemoryRequest := resource.MustParse("128Mi")
		c.MemoryRequest = &defaultMemoryRequest
	}

	if c.EnableEncryption && c.EncryptionType == "" {
		c.EncryptionType = kops.CiliumEncryptionTypeIPSec
	}

	hubble := c.Hubble
	if hubble != nil {
		if hubble.Enabled == nil {
			hubble.Enabled = fi.PtrTo(true)
		}
	} else {
		c.Hubble = &kops.HubbleSpec{
			Enabled: fi.PtrTo(false),
		}
	}

	ingress := c.Ingress
	if ingress != nil {
		if ingress.Enabled == nil {
			ingress.Enabled = fi.PtrTo(true)
		}
	} else {
		c.Ingress = &kops.CiliumIngressSpec{
			Enabled: fi.PtrTo(false),
		}
	}

	return nil
}

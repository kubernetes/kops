/*
Copyright 2024 The Kubernetes Authors.

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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/loader"
)

// KindnetOptionsBuilder adds options for kindnet to the model
type KindnetOptionsBuilder struct {
	Context *OptionsContext
}

var _ loader.ClusterOptionsBuilder = &KindnetOptionsBuilder{}

func (b *KindnetOptionsBuilder) BuildOptions(o *kops.Cluster) error {
	clusterSpec := &o.Spec
	c := clusterSpec.Networking.Kindnet
	if c == nil {
		return nil
	}

	if c.Version == "" {
		c.Version = "v1.8.0"
	}

	if c.Masquerade == nil {
		c.Masquerade = &kops.KindnetMasqueradeSpec{}
		if clusterSpec.IsIPv6Only() {
			// Kindnet should NOT masquerade when IPv6 is used
			c.Masquerade.Enabled = fi.PtrTo(false)
			if o.GetCloudProvider() != kops.CloudProviderAWS {
				c.NAT64 = fi.PtrTo(true)
			}
		} else {
			// Kindnet should masquerade well known ranges if kops is not doing it
			c.Masquerade.Enabled = fi.PtrTo(true)
			if clusterSpec.Networking.NetworkCIDR != "" {
				c.Masquerade.NonMasqueradeCIDRs = append(c.Masquerade.NonMasqueradeCIDRs, clusterSpec.Networking.NetworkCIDR)
			}
			if clusterSpec.Networking.PodCIDR != "" {
				c.Masquerade.NonMasqueradeCIDRs = append(c.Masquerade.NonMasqueradeCIDRs, clusterSpec.Networking.PodCIDR)
			}
			if clusterSpec.Networking.ServiceClusterIPRange != "" {
				c.Masquerade.NonMasqueradeCIDRs = append(c.Masquerade.NonMasqueradeCIDRs, clusterSpec.Networking.ServiceClusterIPRange)
			}
		}
	}

	if c.FastPathThreshold == nil {
		c.FastPathThreshold = fi.PtrTo(int32(0))
	}

	if c.LogLevel == nil {
		c.LogLevel = fi.PtrTo(int32(2))
	}

	return nil
}

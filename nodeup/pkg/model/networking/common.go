/*
Copyright 2017 The Kubernetes Authors.

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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
)

// CommonBuilder runs common tasks
type CommonBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &CommonBuilder{}

// Build is responsible for copying the common CNI binaries
func (b *CommonBuilder) Build(c *fi.ModelBuilderContext) error {
	// Based on https://github.com/containernetworking/plugins/releases/tag/v0.7.5
	assets := []string{
		"bridge",
		"dhcp",
		"flannel",
		"host-device",
		"host-local",
		"ipvlan",
		"loopback",
		"macvlan",
		"portmap",
		"ptp",
		"tuning",
		"vlan",
	}

	// Additions in https://github.com/containernetworking/plugins/releases/tag/v0.8.6
	if b.IsKubernetesGTE("1.15") {
		assets = append(assets, "bandwidth")
		assets = append(assets, "firewall")
		assets = append(assets, "sbr")
		assets = append(assets, "static")
	}

	if err := b.AddCNIBinAssets(c, assets); err != nil {
		return err
	}

	return nil
}

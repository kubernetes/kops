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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

// FlannelBuilder writes the Amazon VPC CNI configuration
type FlannelBuilder struct {
	*model.NodeupModelContext
}

var _ fi.NodeupModelBuilder = &FlannelBuilder{}

// Build is responsible for configuring the network cni
func (b *FlannelBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.NodeupConfig.Networking.Flannel == nil {
		return nil
	}

	if b.Distribution == distributions.DistributionUbuntu2404 {
		// https://github.com/flannel-io/flannel/blob/master/Documentation/troubleshooting.md#nat
		c.AddTask(&nodetasks.File{
			Path: "/etc/udev/rules.d/90-flannel.rules",
			Contents: fi.NewStringResource(
				`SUBSYSTEM=="net", ACTION=="add|change|move", ENV{INTERFACE}=="flannel.1", RUN+="/usr/sbin/ethtool -K flannel.1 tx-checksum-ip-generic off"`,
			),
			Type: nodetasks.FileType_File,
			OnChangeExecute: [][]string{
				{"udevadm", "control", "--reload-rules"},
				{"udevadm", "trigger"},
			},
		})
	}

	return nil
}

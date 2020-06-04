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

package networking

import (
	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/pkg/systemd"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// FlannelBuilder writes Flannel's assets
type FlannelBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &FlannelBuilder{}

// Build is responsible for configuring the network cni
func (b *FlannelBuilder) Build(c *fi.ModelBuilderContext) error {
	networking := b.Cluster.Spec.Networking

	if networking.Flannel != nil {
		b.AddCNIBinAssets(c, []string{"flannel", "portmap", "bridge", "host-local", "loopback"})
	}

	if networking.Canal != nil || networking.Flannel != nil && networking.Flannel.Backend == "vxlan" {
		return b.buildFlannelTxChecksumOffloadDisableService(c)
	}
	return nil
}

// Tx checksum offloading is buggy for NAT-ed VXLAN endpoints, leading to an invalid checksum sent and causing
// Flannel to stop to working as the traffic is being discarded by the receiver.
// https://github.com/coreos/flannel/issues/1279
func (b *FlannelBuilder) buildFlannelTxChecksumOffloadDisableService(c *fi.ModelBuilderContext) error {
	const serviceName = "flannel-tx-checksum-offload-disable.service"

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Disable TX checksum offload on flannel.1")

	manifest.Set("Unit", "After", "sys-devices-virtual-net-flannel.1.device")
	manifest.Set("Install", "WantedBy", "sys-devices-virtual-net-flannel.1.device")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "ExecStart", "/sbin/ethtool -K flannel.1 tx-checksum-ip-generic off")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", serviceName, manifestString)

	service := &nodetasks.Service{
		Name:       serviceName,
		Definition: fi.String(manifestString),
	}

	c.AddTask(service)

	return nil
}

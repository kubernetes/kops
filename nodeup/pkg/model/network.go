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

package model

import (
	"fmt"
	"path/filepath"

	"k8s.io/klog"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// NetworkBuilder writes CNI assets
type NetworkBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &NetworkBuilder{}

// Build is responsible for configuring the network cni
func (b *NetworkBuilder) Build(c *fi.ModelBuilderContext) error {
	var assetNames []string

	// @TODO need to clean up this code, it isn't the easiest to read
	networking := b.Cluster.Spec.Networking
	if networking == nil || networking.Classic != nil {
	} else if networking.Kubenet != nil || networking.GCE != nil {
		assetNames = append(assetNames, "bridge", "host-local", "loopback")
	} else if networking.External != nil {
		// external is based on kubenet
		assetNames = append(assetNames, "bridge", "host-local", "loopback")

	} else if networking.CNI != nil || networking.Weave != nil || networking.Flannel != nil || networking.Calico != nil || networking.Canal != nil || networking.Kuberouter != nil || networking.Romana != nil || networking.AmazonVPC != nil || networking.Cilium != nil {
		assetNames = append(assetNames, "bridge", "host-local", "loopback", "ptp", "portmap")
		// Do we need tuning?

		// TODO: Only when using flannel ?
		assetNames = append(assetNames, "flannel")
	} else if networking.Kopeio != nil {
		// TODO combine with External
		// Kopeio is based on kubenet / external
		assetNames = append(assetNames, "bridge", "host-local", "loopback")
	} else if networking.LyftVPC != nil {
		assetNames = append(assetNames, "cni-ipvlan-vpc-k8s-ipam", "cni-ipvlan-vpc-k8s-ipvlan", "cni-ipvlan-vpc-k8s-tool", "cni-ipvlan-vpc-k8s-unnumbered-ptp", "loopback")
	} else {
		return fmt.Errorf("no networking mode set")
	}

	for _, assetName := range assetNames {
		if err := b.addCNIBinAsset(c, assetName); err != nil {
			return err
		}
	}

	// Tx checksum offloading is buggy for NAT-ed VXLAN endpoints, leading to an invalid checksum sent and causing
	// Flannel to stop to working as the traffic is being discarded by the receiver.
	// https://github.com/coreos/flannel/issues/1279
	if networking != nil && (networking.Canal != nil || (networking.Flannel != nil && networking.Flannel.Backend == "vxlan")) {
		c.AddTask(b.buildFlannelTxChecksumOffloadDisableService())
	}

	return nil
}

func (b *NetworkBuilder) addCNIBinAsset(c *fi.ModelBuilderContext, assetName string) error {
	assetPath := ""
	asset, err := b.Assets.Find(assetName, assetPath)
	if err != nil {
		return fmt.Errorf("error trying to locate asset %q: %v", assetName, err)
	}
	if asset == nil {
		return fmt.Errorf("unable to locate asset %q", assetName)
	}

	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(b.CNIBinDir(), assetName),
		Contents: asset,
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	})

	return nil
}

func (b *NetworkBuilder) buildFlannelTxChecksumOffloadDisableService() *nodetasks.Service {
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
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

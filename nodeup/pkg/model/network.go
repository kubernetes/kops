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

package model

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/unix"
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
		assetNames = append(assetNames, "bridge", "host-local", "loopback", "ptp")
		// Do we need tuning?

		if b.IsKubernetesGTE("1.9") {
			assetNames = append(assetNames, "portmap")
		}

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

	if networking.Cilium != nil {
		// systemd v238 includes the bpffs mount by default; and gives an error "has a bad unit file setting" if we try to mount it again (see mount_point_is_api)
		var alreadyMounted bool
		// bpffs magic number
		magic := uint32(0xCAFE4A11)
		var fsdata unix.Statfs_t
		err := unix.Statfs("/sys/fs/bpf", &fsdata)

		if err != nil {
			alreadyMounted = false
		} else {
			alreadyMounted = int32(magic) == int32(fsdata.Type)
		}

		if !alreadyMounted {
			unit := s(`
[Unit]
Description=Cilium BPF mounts
Documentation=http://docs.cilium.io/
DefaultDependencies=no
Before=local-fs.target umount.target kubelet.service

[Mount]
What=bpffs
Where=/sys/fs/bpf
Type=bpf

[Install]
WantedBy=multi-user.target
`)

			service := &nodetasks.Service{
				Name:       "sys-fs-bpf.mount",
				Definition: unit,
			}
			service.InitDefaults()
			c.AddTask(service)
		}
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

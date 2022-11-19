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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"

	"k8s.io/kops/nodeup/pkg/model"
	apiModel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// CiliumBuilder writes Cilium's assets
type CiliumBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &CiliumBuilder{}

// Build is responsible for configuring the network cni
func (b *CiliumBuilder) Build(c *fi.ModelBuilderContext) error {
	cilium := b.Cluster.Spec.Networking.Cilium

	// As long as the Cilium Etcd cluster exists, we should do this
	if apiModel.UseCiliumEtcd(b.Cluster) {
		if err := b.buildCiliumEtcdSecrets(c); err != nil {
			return err
		}
	}

	if cilium == nil {
		return nil
	}

	if err := b.buildBPFMount(c); err != nil {
		return fmt.Errorf("failed to create bpf mount unit: %w", err)
	}

	if err := b.buildCgroup2Mount(c); err != nil {
		return fmt.Errorf("failed to create cgroupv2 mount unit: %w", err)
	}

	return nil
}

func (b *CiliumBuilder) buildBPFMount(c *fi.ModelBuilderContext) error {
	var fsdata unix.Statfs_t
	err := unix.Statfs("/sys/fs/bpf", &fsdata)
	if err != nil {
		return fmt.Errorf("error checking for /sys/fs/bpf: %v", err)
	}

	// equivalent to unix.BPF_FS_MAGIC in golang.org/x/sys/unix
	BPF_FS_MAGIC := uint32(0xcafe4a11)

	// systemd v238 includes the bpffs mount by default; and gives an error "has a bad unit file setting" if we try to mount it again (see mount_point_is_api)
	alreadyMounted := uint32(fsdata.Type) == BPF_FS_MAGIC

	if !alreadyMounted {
		unit := `
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
`

		service := &nodetasks.Service{
			Name:       "sys-fs-bpf.mount",
			Definition: fi.PtrTo(unit),
		}
		service.InitDefaults()
		c.AddTask(service)
	}

	return nil
}

func (b *CiliumBuilder) buildCgroup2Mount(c *fi.ModelBuilderContext) error {
	cgroupPath := "/run/cilium/cgroupv2"

	var fsdata unix.Statfs_t
	err := unix.Statfs(cgroupPath, &fsdata)

	// If the path does not exist, systemd will create it
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("error checking for /run/cilium/cgroupv2: %v", err)
	}

	CGROUP_FS_MAGIC := uint32(0x63677270)

	alreadyMounted := uint32(fsdata.Type) == CGROUP_FS_MAGIC

	if !alreadyMounted {
		unit := `
[Unit]
Description=Cilium Cgroup2 mounts
Documentation=http://docs.cilium.io/
DefaultDependencies=no
Before=local-fs.target umount.target kubelet.service

[Mount]
What=cgroup2
Where=/run/cilium/cgroupv2
Type=cgroup2

[Install]
WantedBy=multi-user.target
`

		service := &nodetasks.Service{
			Name:         "run-cilium-cgroupv2.mount",
			Definition:   fi.PtrTo(unit),
			SmartRestart: fi.PtrTo(false),
		}
		service.InitDefaults()
		c.AddTask(service)
	}

	return nil
}

func (b *CiliumBuilder) buildCiliumEtcdSecrets(c *fi.ModelBuilderContext) error {
	name := "etcd-client-cilium"
	dir := "/etc/kubernetes/pki/cilium"
	signer := "etcd-clients-ca-cilium"
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(dir, "etcd-ca.crt"),
		Contents: fi.NewStringResource(b.NodeupConfig.CAs[signer]),
		Type:     nodetasks.FileType_File,
		Mode:     fi.PtrTo("0600"),
	})
	if b.HasAPIServer {
		issueCert := &nodetasks.IssueCert{
			Name:      name,
			Signer:    signer,
			KeypairID: b.NodeupConfig.KeypairIDs[signer],
			Type:      "client",
			Subject: nodetasks.PKIXName{
				CommonName: "cilium",
			},
		}
		c.AddTask(issueCert)
		return issueCert.AddFileTasks(c, dir, name, "", nil)
	} else {
		if b.UseKopsControllerForNodeBootstrap() {
			cert, key, err := b.GetBootstrapCert(name, signer)
			if err != nil {
				return err
			}

			c.AddTask(&nodetasks.File{
				Path:           filepath.Join(dir, name+".crt"),
				Contents:       cert,
				Type:           nodetasks.FileType_File,
				Mode:           fi.PtrTo("0644"),
				BeforeServices: []string{"kubelet.service"},
			})

			c.AddTask(&nodetasks.File{
				Path:           filepath.Join(dir, name+".key"),
				Contents:       key,
				Type:           nodetasks.FileType_File,
				Mode:           fi.PtrTo("0400"),
				BeforeServices: []string{"kubelet.service"},
			})

			return nil
		} else {
			return b.BuildCertificatePairTask(c, name, dir, name, nil, []string{"kubelet.service"})
		}
	}
}

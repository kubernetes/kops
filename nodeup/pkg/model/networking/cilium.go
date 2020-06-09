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
	"fmt"

	"golang.org/x/sys/unix"

	"k8s.io/kops/nodeup/pkg/model"
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
	networking := b.Cluster.Spec.Networking

	if networking.Cilium == nil {
		return nil
	}

	if err := b.buildBPFMount(c); err != nil {
		return err
	}

	if networking.Cilium.EtcdManaged {
		if err := b.buildCiliumEtcdSecrets(c); err != nil {
			return err
		}
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
			Definition: fi.String(unit),
		}
		c.AddTask(service)
	}

	return nil
}

func (b *CiliumBuilder) buildCiliumEtcdSecrets(c *fi.ModelBuilderContext) error {

	if b.IsMaster {
		d := "/etc/kubernetes/pki/etcd-manager-cilium"

		keys := make(map[string]string)
		keys["etcd-manager-ca"] = "etcd-manager-ca-cilium"
		keys["etcd-peers-ca"] = "etcd-peers-ca-cilium"
		keys["etcd-clients-ca"] = "etcd-clients-ca-cilium"

		for fileName, keystoreName := range keys {
			_, err := b.KeyStore.FindCert(keystoreName)
			if err != nil {
				return err
			}

			if err := b.BuildCertificateTask(c, keystoreName, d+"/"+fileName+".crt"); err != nil {
				return err
			}
			if err := b.BuildPrivateKeyTask(c, keystoreName, d+"/"+fileName+".key"); err != nil {
				return err
			}
		}
	}

	name := "etcd-client-cilium"
	issueCert := &nodetasks.IssueCert{
		Name:   name,
		Signer: "etcd-clients-ca-cilium",
		Type:   "client",
		Subject: nodetasks.PKIXName{
			CommonName: "cilium",
		},
	}
	c.AddTask(issueCert)
	issueCert.AddFileTasks(c, "/etc/kubernetes/pki/cilium", name, "etcd-ca", nil)

	return nil
}

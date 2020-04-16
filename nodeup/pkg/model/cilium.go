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

package model

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
	certutil "k8s.io/client-go/util/cert"
	"k8s.io/klog"
	"k8s.io/kops/pkg/pkiutil"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// CiliumBuilder writes Cilium's assets
type CiliumBuilder struct {
	*NodeupModelContext
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

	etcdClientsCACertificate, err := b.KeyStore.FindCert("etcd-clients-ca-cilium")
	if err != nil {
		return err
	}

	etcdClientsCAPrivateKey, err := b.KeyStore.FindPrivateKey("etcd-clients-ca-cilium")
	if err != nil {
		return err
	}

	dir := "/etc/kubernetes/pki/cilium"

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directories %q: %v", dir, err)
	}

	{
		p := filepath.Join(dir, "etcd-ca.crt")
		certBytes := pkiutil.EncodeCertPEM(etcdClientsCACertificate.Certificate)
		if err := ioutil.WriteFile(p, certBytes, 0644); err != nil {
			return fmt.Errorf("error writing certificate key file %q: %v", p, err)
		}
	}

	name := "etcd-client"

	humanName := dir + "/" + name
	privateKey, err := pkiutil.NewPrivateKey()
	if err != nil {
		return fmt.Errorf("unable to create private key %q: %v", humanName, err)
	}
	privateKeyBytes := pkiutil.EncodePrivateKeyPEM(privateKey)

	certConfig := &certutil.Config{
		CommonName: "cilium",
		Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	signingKey, ok := etcdClientsCAPrivateKey.Key.(*rsa.PrivateKey)
	if !ok {
		return fmt.Errorf("etcd-clients-ca private key had unexpected type %T", etcdClientsCAPrivateKey.Key)
	}

	klog.Infof("signing certificate for %q", humanName)
	cert, err := pkiutil.NewSignedCert(certConfig, privateKey, etcdClientsCACertificate.Certificate, signingKey)
	if err != nil {
		return fmt.Errorf("error signing certificate for %q: %v", humanName, err)
	}

	certBytes := pkiutil.EncodeCertPEM(cert)

	p := filepath.Join(dir, name)
	{
		if err := ioutil.WriteFile(p+".crt", certBytes, 0644); err != nil {
			return fmt.Errorf("error writing certificate key file %q: %v", p+".crt", err)
		}
	}

	{
		if err := ioutil.WriteFile(p+".key", privateKeyBytes, 0600); err != nil {
			return fmt.Errorf("error writing private key file %q: %v", p+".key", err)
		}
	}

	return nil
}

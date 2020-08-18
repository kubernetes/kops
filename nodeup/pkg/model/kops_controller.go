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
	"path/filepath"

	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KopsControllerBuilder installs the keys for a kops-controller.
type KopsControllerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KopsControllerBuilder{}

// Build is responsible for configuring keys that will be used by kops-controller (via hostPath)
func (b *KopsControllerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	// Create the directory, even if we aren't going to populate it
	pkiDir := "/etc/kubernetes/kops-controller"
	c.AddTask(&nodetasks.File{
		Path: pkiDir,
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	})

	if !b.UseKopsControllerForNodeBootstrap() {
		return nil
	}

	// We run kops-controller under an unprivileged user (wellknownusers.KopsControllerID), and then grant specific permissions
	c.AddTask(&nodetasks.UserTask{
		Name:  wellknownusers.KopsControllerName,
		UID:   wellknownusers.KopsControllerID,
		Shell: "/sbin/nologin",
	})

	issueCert := &nodetasks.IssueCert{
		Name:           "kops-controller",
		Signer:         fi.CertificateIDCA,
		Type:           "server",
		Subject:        nodetasks.PKIXName{CommonName: "kops-controller"},
		AlternateNames: []string{b.Cluster.Spec.MasterInternalName},
	}
	c.AddTask(issueCert)

	certResource, keyResource, _ := issueCert.GetResources()
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(pkiDir, "kops-controller.crt"),
		Contents: certResource,
		Type:     nodetasks.FileType_File,
		Mode:     s("0644"),
		Owner:    s(wellknownusers.KopsControllerName),
	})
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(pkiDir, "kops-controller.key"),
		Contents: keyResource,
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
		Owner:    s(wellknownusers.KopsControllerName),
	})

	caList := []string{fi.CertificateIDCA}
	if model.UseCiliumEtcd(b.Cluster) {
		caList = append(caList, "etcd-clients-ca-cilium")
	}
	for _, cert := range caList {
		owner := wellknownusers.KopsControllerName
		err := b.BuildCertificatePairTask(c, cert, pkiDir, cert, &owner)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Copyright 2023 The Kubernetes Authors.

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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// MachineCertificateBuilder ensures the machine has a PKI certificate.
type MachineCertificateBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &MachineCertificateBuilder{}

// Build is responsible for building the tasks for the machine-certificate.
func (b *MachineCertificateBuilder) Build(ctx *fi.NodeupModelBuilderContext) error {
	switch b.CloudProvider() {
	case kops.CloudProviderGCE:

	default:
		return nil
	}

	nodeName, err := b.NodeName()
	if err != nil {
		return err
	}
	name := "machine-key"

	// TODO: Include cluster name as suffix?

	var certResource fi.Resource
	var keyResource fi.Resource

	if !b.IsMaster && b.UseKopsControllerForNodeBootstrap() {
		cert, key, err := b.GetBootstrapCert(name, fi.CertificateIDCA)
		if err != nil {
			return err
		}

		certResource = cert
		keyResource = key
	} else {

		issueCert := &nodetasks.IssueCert{
			Name:      name,
			Signer:    fi.CertificateIDCA,
			KeypairID: b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
			Type:      "client",
			Subject: nodetasks.PKIXName{
				CommonName:   fmt.Sprintf("kops:machine:%s", nodeName),
				Organization: []string{"kops:machines"},
			},
			// TODO? AlternateNames: []string{},
		}
		ctx.AddTask(issueCert)

		certResource, keyResource, _ = issueCert.GetResources()
	}

	dir := b.MachineKeyDir()
	ctx.AddTask(&nodetasks.File{
		Path:     filepath.Join(dir, "machine.crt"),
		Contents: certResource,
		Type:     nodetasks.FileType_File,
		Mode:     fi.PtrTo("0644"),
	})
	ctx.AddTask(&nodetasks.File{
		Path:     filepath.Join(dir, "machine.key"),
		Contents: keyResource,
		Type:     nodetasks.FileType_File,
		Mode:     fi.PtrTo("0600"),
	})

	return nil
}

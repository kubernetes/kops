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
	"path/filepath"

	"k8s.io/kops/upup/pkg/fi"
)

// EtcdTLSBuilder configures the etcd TLS support
type EtcdTLSBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcdTLSBuilder{}

// Build is responsible for performing setup for CNIs that need etcd TLS support
func (b *EtcdTLSBuilder) Build(c *fi.ModelBuilderContext) error {
	// @check if tls is enabled and if so, we need to download the client certificates
	if !b.UseEtcdManager() && b.UseEtcdTLS() {
		name := "calico-client"
		dirname := "calico"
		ca := filepath.Join(dirname, "ca.pem")
		certificate := filepath.Join(dirname, name+".pem")
		key := filepath.Join(dirname, name+"-key.pem")

		if err := b.BuildCertificateTask(c, name, certificate); err != nil {
			return err
		}
		if err := b.BuildPrivateKeyTask(c, name, key); err != nil {
			return err
		}
		if err := b.BuildCertificateTask(c, fi.CertificateId_CA, ca); err != nil {
			return err
		}
	}

	return nil
}

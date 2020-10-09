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

package networking

import (
	"path/filepath"

	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
)

// CalicoBuilder configures the etcd TLS support for Calico
type CalicoBuilder struct {
	*model.NodeupModelContext
}

var _ fi.ModelBuilder = &CalicoBuilder{}

// Build is responsible for performing setup for CNIs that need etcd TLS support
func (b *CalicoBuilder) Build(c *fi.ModelBuilderContext) error {
	networking := b.Cluster.Spec.Networking

	if networking.Calico == nil {
		return nil
	}

	// @check if tls is enabled and if so, we need to download the client certificates
	if b.IsKubernetesLT("1.12") && !b.UseEtcdManager() && b.UseEtcdTLS() {
		name := "calico-client"
		dirname := "calico"
		ca := filepath.Join(dirname, "ca.pem")
		certificate := filepath.Join(dirname, name+".pem")
		key := filepath.Join(dirname, name+"-key.pem")

		if err := b.BuildCertificateTask(c, name, certificate, nil); err != nil {
			return err
		}
		if err := b.BuildPrivateKeyTask(c, name, key, nil); err != nil {
			return err
		}
		if err := b.BuildCertificateTask(c, fi.CertificateIDCA, ca, nil); err != nil {
			return err
		}
	}

	return nil
}

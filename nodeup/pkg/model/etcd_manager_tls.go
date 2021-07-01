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

	"k8s.io/kops/upup/pkg/fi"
)

// EtcdManagerTLSBuilder configures TLS support for etcd-manager
type EtcdManagerTLSBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcdManagerTLSBuilder{}

// Build is responsible for TLS configuration for etcd-manager
func (b *EtcdManagerTLSBuilder) Build(ctx *fi.ModelBuilderContext) error {
	if !b.HasAPIServer || !b.UseEtcdManager() {
		return nil
	}

	for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
		k := etcdCluster.Name

		// The certs for cilium etcd are managed by CiliumBuilder
		if k == "cilium" {
			continue
		}

		d := "/etc/kubernetes/pki/etcd-manager-" + k

		keys := make(map[string]string)

		// Only nodes running etcd need the peers CA
		if b.IsMaster {
			keys["etcd-manager-ca"] = "etcd-manager-ca-" + k
			keys["etcd-peers-ca"] = "etcd-peers-ca-" + k
		}
		// Because API server can only have a single client certificate for etcd, we need to share a client CA
		keys["etcd-clients-ca"] = "etcd-clients-ca"

		for fileName, keystoreName := range keys {
			cert, err := b.KeyStore.FindCert(keystoreName)
			if err != nil {
				return err
			}
			if cert == nil {
				return fmt.Errorf("keypair %q not found", keystoreName)
			}

			if err := b.BuildCertificateTask(ctx, keystoreName, d+"/"+fileName+".crt", nil); err != nil {
				return err
			}
			if err := b.BuildPrivateKeyTask(ctx, keystoreName, d+"/"+fileName+".key", nil); err != nil {
				return err
			}
		}
	}

	return nil
}

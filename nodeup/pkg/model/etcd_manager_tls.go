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
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// EtcdManagerTLSBuilder configures TLS support for etcd-manager
type EtcdManagerTLSBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &EtcdManagerTLSBuilder{}

// Build is responsible for TLS configuration for etcd-manager
func (b *EtcdManagerTLSBuilder) Build(ctx *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
		k := etcdCluster.Name

		d := "/etc/kubernetes/pki/etcd-manager-" + k

		keys := make(map[string]string)

		keys["etcd-manager-ca"] = "etcd-manager-ca-" + k
		keys["etcd-peers-ca"] = "etcd-peers-ca-" + k
		keys["etcd-clients-ca"] = "etcd-clients-ca-" + k

		// Because API server can only have a single client certificate for etcd, we need to share a client CA
		if k == "main" || k == "events" {
			keys["etcd-clients-ca"] = "etcd-clients-ca"
		}

		for fileName, keystoreName := range keys {
			if err := b.buildCertificatePairTask(ctx, keystoreName, d, fileName, nil, nil, false); err != nil {
				return err
			}
			ctx.AddTask(&nodetasks.File{
				Path:     filepath.Join(d, fileName+".crt"),
				Contents: fi.NewStringResource(b.NodeupConfig.CAs[keystoreName]),
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("0644"),
			})

		}
	}

	return nil
}

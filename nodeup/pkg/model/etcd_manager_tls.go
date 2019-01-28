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
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
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

	for _, k := range []string{"main", "events"} {
		d := "/etc/kubernetes/pki/etcd-manager-" + k

		keys := make(map[string]string)
		keys["etcd-manager-ca"] = "etcd-manager-ca-" + k

		for fileName, keystoreName := range keys {
			cert, err := b.KeyStore.FindCert(keystoreName)
			if err != nil {
				return err
			}
			if cert == nil {
				glog.Warningf("keypair %q not found, won't configure", keystoreName)
				continue
			}

			if err := b.BuildCertificateTask(ctx, keystoreName, d+"/"+fileName+".crt"); err != nil {
				return err
			}
			if err := b.BuildPrivateKeyTask(ctx, keystoreName, d+"/"+fileName+".key"); err != nil {
				return err
			}
		}
	}

	return nil
}

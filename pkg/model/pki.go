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
	"k8s.io/kops/pkg/tokens"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/vfs"
)

// PKIModelBuilder configures PKI keypairs, as well as tokens
type PKIModelBuilder struct {
	*KopsModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &PKIModelBuilder{}

// Build is responsible for generating the various pki assets.
func (b *PKIModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	// TODO: Only create the CA via this task
	defaultCA := &fitasks.Keypair{
		Name:      fi.PtrTo(fi.CertificateIDCA),
		Lifecycle: b.Lifecycle,
		Subject:   "cn=kubernetes-ca",
		Type:      "ca",
	}
	c.AddTask(defaultCA)

	{
		aggregatorCA := &fitasks.Keypair{
			Name:      fi.PtrTo("apiserver-aggregator-ca"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=apiserver-aggregator-ca",
			Type:      "ca",
		}
		c.AddTask(aggregatorCA)
	}

	{
		serviceAccount := &fitasks.Keypair{
			// We only need the private key, but it's easier to create a certificate as well.
			Name:      fi.PtrTo("service-account"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=service-account",
			Type:      "ca",
		}
		c.AddTask(serviceAccount)
	}

	// Create auth tokens (though this is deprecated)
	for _, x := range tokens.GetKubernetesAuthTokens_Deprecated() {
		c.AddTask(&fitasks.Secret{Name: fi.PtrTo(x), Lifecycle: b.Lifecycle})
	}

	{
		mirrorPath, err := vfs.Context.BuildVfsPath(b.Cluster.Spec.SecretStore)
		if err != nil {
			return err
		}

		t := &fitasks.MirrorSecrets{
			Name:       fi.PtrTo("mirror-secrets"),
			Lifecycle:  b.Lifecycle,
			MirrorPath: mirrorPath,
		}
		c.AddTask(t)
	}

	{
		mirrorPath, err := vfs.Context.BuildVfsPath(b.Cluster.Spec.KeyStore)
		if err != nil {
			return err
		}

		// Keypair used by the kubelet
		t := &fitasks.MirrorKeystore{
			Name:       fi.PtrTo("mirror-keystore"),
			Lifecycle:  b.Lifecycle,
			MirrorPath: mirrorPath,
		}
		c.AddTask(t)
	}

	return nil
}

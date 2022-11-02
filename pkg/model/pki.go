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
	"strings"

	"k8s.io/kops/pkg/rbac"
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

var _ fi.ModelBuilder = &PKIModelBuilder{}

// Build is responsible for generating the various pki assets.
func (b *PKIModelBuilder) Build(c *fi.ModelBuilderContext) error {
	// TODO: Only create the CA via this task
	defaultCA := &fitasks.Keypair{
		Name:      fi.String(fi.CertificateIDCA),
		Lifecycle: b.Lifecycle,
		Subject:   "cn=kubernetes-ca",
		Type:      "ca",
	}
	c.AddTask(defaultCA)

	{
		// @check if kops-controller bootstrap or bootstrap tokens are enabled. If so, disable the creation of the kubelet certificate - we also
		// block at the IAM level for AWS cluster for pre-existing clusters.
		if !b.UseKopsControllerForNodeBootstrap() && !b.UseBootstrapTokens() {
			c.AddTask(&fitasks.Keypair{
				Name:      fi.String("kubelet"),
				Lifecycle: b.Lifecycle,
				Subject:   "o=" + rbac.NodesGroup + ",cn=kubelet",
				Type:      "client",
				Signer:    defaultCA,
			})
		}
	}

	if !b.UseKopsControllerForNodeBootstrap() {
		t := &fitasks.Keypair{
			Name:      fi.String("kube-proxy"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + rbac.KubeProxy,
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	if b.KopsModelContext.Cluster.Spec.Networking.Kuberouter != nil && !b.UseKopsControllerForNodeBootstrap() {
		t := &fitasks.Keypair{
			Name:      fi.String("kube-router"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + rbac.KubeRouter,
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	{
		aggregatorCA := &fitasks.Keypair{
			Name:      fi.String("apiserver-aggregator-ca"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=apiserver-aggregator-ca",
			Type:      "ca",
		}
		c.AddTask(aggregatorCA)
	}

	{
		serviceAccount := &fitasks.Keypair{
			// We only need the private key, but it's easier to create a certificate as well.
			Name:      fi.String("service-account"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=service-account",
			Type:      "ca",
		}
		c.AddTask(serviceAccount)
	}

	// @TODO this is VERY presumptuous, i'm going on the basis we can make it configurable in the future.
	// But I'm conscious not to do too much work on bootstrap tokens as it might overlay further down the
	// line with the machines api
	if b.UseBootstrapTokens() {
		serviceName := "node-authorizer-internal"

		alternateNames := []string{
			"127.0.0.1",
			"localhost",
			serviceName,
			strings.Join([]string{serviceName, b.Cluster.Name}, "."),
		}
		if b.Cluster.Spec.DNSZone != "" {
			alternateNames = append(alternateNames, strings.Join([]string{serviceName, b.Cluster.Spec.DNSZone}, "."))
		}

		// @note: the certificate used by the node authorizers
		c.AddTask(&fitasks.Keypair{
			Name:           fi.String("node-authorizer"),
			Lifecycle:      b.Lifecycle,
			Subject:        "cn=node-authorizaer",
			Type:           "server",
			AlternateNames: alternateNames,
			Signer:         defaultCA,
		})

		// @note: we use this for mutual tls between node and authorizer
		c.AddTask(&fitasks.Keypair{
			Name:      fi.String("node-authorizer-client"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=node-authorizer-client",
			Type:      "client",
			Signer:    defaultCA,
		})
	}

	// Create auth tokens (though this is deprecated)
	for _, x := range tokens.GetKubernetesAuthTokens_Deprecated() {
		c.AddTask(&fitasks.Secret{Name: fi.String(x), Lifecycle: b.Lifecycle})
	}

	{
		mirrorPath, err := vfs.Context.BuildVfsPath(b.Cluster.Spec.SecretStore)
		if err != nil {
			return err
		}

		t := &fitasks.MirrorSecrets{
			Name:       fi.String("mirror-secrets"),
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
			Name:       fi.String("mirror-keystore"),
			Lifecycle:  b.Lifecycle,
			MirrorPath: mirrorPath,
		}
		c.AddTask(t)
	}

	return nil
}

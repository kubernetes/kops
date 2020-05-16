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
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &PKIModelBuilder{}

// Build is responsible for generating the various pki assets.
func (b *PKIModelBuilder) Build(c *fi.ModelBuilderContext) error {

	// TODO: Only create the CA via this task
	defaultCA := &fitasks.Keypair{
		Name:      fi.String(fi.CertificateId_CA),
		Lifecycle: b.Lifecycle,
		Subject:   "cn=kubernetes",
		Type:      "ca",
	}
	c.AddTask(defaultCA)

	{
		// @check of bootstrap tokens are enable if so, disable the creation of the kubelet certificate - we also
		// block at the IAM level for AWS cluster for pre-existing clusters.
		if !b.UseBootstrapTokens() {
			c.AddTask(&fitasks.Keypair{
				Name:      fi.String("kubelet"),
				Lifecycle: b.Lifecycle,
				Subject:   "o=" + rbac.NodesGroup + ",cn=kubelet",
				Type:      "client",
				Signer:    defaultCA,
			})
		}
	}
	{
		// Generate a kubelet client certificate for api to speak securely to kubelets. This change was first
		// introduced in https://github.com/kubernetes/kops/pull/2831 where server.cert/key were used. With kubernetes >= 1.7
		// the certificate usage is being checked (obviously the above was server not client certificate) and so now fails
		c.AddTask(&fitasks.Keypair{
			Name:      fi.String("kubelet-api"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=kubelet-api",
			Type:      "client",
			Signer:    defaultCA,
		})
	}
	{
		t := &fitasks.Keypair{
			Name:      fi.String("kube-scheduler"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + rbac.KubeScheduler,
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kube-proxy"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + rbac.KubeProxy,
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kube-controller-manager"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + rbac.KubeControllerManager,
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	if b.UseEtcdManager() {
		// We generate keypairs in the etcdmanager task itself
	} else if b.UseEtcdTLS() {
		// check if we need to generate certificates for etcd peers certificates from a different CA?
		// @question i think we should use another KeyStore for this, perhaps registering a EtcdKeyStore given
		// that mutual tls used to verify between the peers we don't want certificates for kubernetes able to act as a peer.
		// For clients assuming we are using etcdv3 is can switch on user authentication and map the common names for auth.
		servingNames := []string{fmt.Sprintf("*.internal.%s", b.ClusterName()), "localhost", "127.0.0.1"}
		// @question should wildcard's be here instead of generating per node. If we ever provide the
		// ability to resize the master, this will become a blocker
		c.AddTask(&fitasks.Keypair{
			AlternateNames: servingNames,
			Lifecycle:      b.Lifecycle,
			Name:           fi.String("etcd"),
			Subject:        "cn=etcd",
			// TODO: Can this be "server" now that we're not using it for peer connectivity?
			Type:   "clientServer",
			Signer: defaultCA,
		})

		// For peer authentication, the same cert is used both as a client
		// cert and as a server cert (which is unusual).  Moreover, etcd
		// 3.2 introduces some breaking changes to certificate validation
		// where it tries to match any IP or DNS names to the client IP
		// (including reverse DNS lookups!)  We _could_ include a wildcard
		// reverse DNS name e.g. *.ec2.internal for EC2, but it seems
		// better just to list the names that we expect peer connectivity
		// to happen on.
		var peerNames []string
		for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
			prefix := "etcd-" + etcdCluster.Name + "-"
			if prefix == "etcd-main-" {
				prefix = "etcd-"
			}
			for _, m := range etcdCluster.Members {
				peerNames = append(peerNames, prefix+m.Name+".internal."+b.ClusterName())
			}
		}
		c.AddTask(&fitasks.Keypair{
			AlternateNames: peerNames,

			Lifecycle: b.Lifecycle,
			Name:      fi.String("etcd-peer"),
			Subject:   "cn=etcd-peer",
			Type:      "clientServer",
			Signer:    defaultCA,
		})

		c.AddTask(&fitasks.Keypair{
			Name:      fi.String("etcd-client"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=etcd-client",
			Type:      "client",
			Signer:    defaultCA,
		})

		// @check if calico is enabled as the CNI provider
		if b.KopsModelContext.Cluster.Spec.Networking.Calico != nil {
			c.AddTask(&fitasks.Keypair{
				Name:      fi.String("calico-client"),
				Lifecycle: b.Lifecycle,
				Subject:   "cn=calico-client",
				Type:      "client",
				Signer:    defaultCA,
			})
		}
	}

	if b.KopsModelContext.Cluster.Spec.Networking.Kuberouter != nil {
		t := &fitasks.Keypair{
			Name:    fi.String("kube-router"),
			Subject: "cn=" + "system:kube-router",
			Type:    "client",
			Signer:  defaultCA,
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kubecfg"),
			Lifecycle: b.Lifecycle,
			Subject:   "o=" + rbac.SystemPrivilegedGroup + ",cn=kubecfg",
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("apiserver-proxy-client"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=apiserver-proxy-client",
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

		aggregator := &fitasks.Keypair{
			Name:      fi.String("apiserver-aggregator"),
			Lifecycle: b.Lifecycle,
			// Must match RequestheaderAllowedNames
			Subject: "cn=aggregator",
			Type:    "client",
			Signer:  aggregatorCA,
		}
		c.AddTask(aggregator)
	}

	{
		// Used by e.g. protokube
		t := &fitasks.Keypair{
			Name:      fi.String("kops"),
			Lifecycle: b.Lifecycle,
			Subject:   "o=" + rbac.SystemPrivilegedGroup + ",cn=kops",
			Type:      "client",
			Signer:    defaultCA,
		}
		c.AddTask(t)
	}

	{
		// A few names used from inside the cluster, which all resolve the same based on our default suffixes
		alternateNames := []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc." + b.Cluster.Spec.ClusterDNSDomain,
		}

		// Names specified in the cluster spec
		alternateNames = append(alternateNames, b.Cluster.Spec.MasterPublicName)
		alternateNames = append(alternateNames, b.Cluster.Spec.MasterInternalName)
		alternateNames = append(alternateNames, b.Cluster.Spec.AdditionalSANs...)

		// Referencing it by internal IP should work also
		{
			ip, err := b.WellKnownServiceIP(1)
			if err != nil {
				return err
			}
			alternateNames = append(alternateNames, ip.String())
		}

		// We also want to be able to reference it locally via https://127.0.0.1
		alternateNames = append(alternateNames, "127.0.0.1")

		t := &fitasks.Keypair{
			Name:           fi.String("master"),
			Lifecycle:      b.Lifecycle,
			Subject:        "cn=kubernetes-master",
			Type:           "server",
			AlternateNames: alternateNames,
			Signer:         defaultCA,
		}
		c.AddTask(t)
	}

	if b.Cluster.Spec.Authentication != nil {
		if b.KopsModelContext.Cluster.Spec.Authentication.Aws != nil {
			alternateNames := []string{
				"localhost",
				"127.0.0.1",
			}

			t := &fitasks.Keypair{
				Name:           fi.String("aws-iam-authenticator"),
				Subject:        "cn=aws-iam-authenticator",
				Type:           "server",
				AlternateNames: alternateNames,
				Signer:         defaultCA,
			}
			c.AddTask(t)
		}
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
			strings.Join([]string{serviceName, b.Cluster.Spec.DNSZone}, "."),
		}

		// @note: the certificate used by the node authorizers
		c.AddTask(&fitasks.Keypair{
			Name:           fi.String("node-authorizer"),
			Subject:        "cn=node-authorizaer",
			Type:           "server",
			AlternateNames: alternateNames,
			Signer:         defaultCA,
		})

		// @note: we use this for mutual tls between node and authorizer
		c.AddTask(&fitasks.Keypair{
			Name:    fi.String("node-authorizer-client"),
			Subject: "cn=node-authorizer-client",
			Type:    "client",
			Signer:  defaultCA,
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
			MirrorPath: mirrorPath,
		}
		c.AddTask(t)
	}

	return nil
}

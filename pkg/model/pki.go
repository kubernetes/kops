/*
Copyright 2016 The Kubernetes Authors.

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

	"k8s.io/apiserver/pkg/authentication/user"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// PKIModelBuilder configures PKI keypairs, as well as tokens
type PKIModelBuilder struct {
	*KopsModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &PKIModelBuilder{}

// Build is responsible for generating the pki assets for the cluster
func (b *PKIModelBuilder) Build(c *fi.ModelBuilderContext) error {
	{
		t := &fitasks.Keypair{
			Name:      fi.String("kubelet"),
			Lifecycle: b.Lifecycle,
			Subject:   "o=" + user.NodesGroup + ",cn=kubelet",
			Type:      "client",
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kube-scheduler"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + user.KubeScheduler,
			Type:      "client",
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kube-proxy"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + user.KubeProxy,
			Type:      "client",
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kube-controller-manager"),
			Lifecycle: b.Lifecycle,
			Subject:   "cn=" + user.KubeControllerManager,
			Type:      "client",
		}
		c.AddTask(t)
	}

	// check if we need to generate certificates for etcd peers certificates from a different CA?
	// @question i think we should use another KeyStore for this, perhaps registering a EtcdKeyStore given
	// that mutual tls used to verify between the peers we don't want certificates for kubernetes able to act as a peer.
	// For clients assuming we are using etcdv3 is can switch on user authentication and map the common names for auth.
	if b.UseEtcdTLS() {
		alternativeNames := []string{fmt.Sprintf("*.internal.%s", b.ClusterName()), "localhost", "127.0.0.1"}
		{
			// @question should wildcard's be here instead of generating per node. If we ever provide the
			// ability to resize the master, this will become a blocker
			c.AddTask(&fitasks.Keypair{
				AlternateNames: alternativeNames,
				Lifecycle:      b.Lifecycle,
				Name:           fi.String("etcd"),
				Subject:        "cn=etcd",
				Type:           "server",
			})
		}
		{
			c.AddTask(&fitasks.Keypair{
				Name:      fi.String("etcd-client"),
				Lifecycle: b.Lifecycle,
				Subject:   "cn=etcd-client",
				Type:      "client",
			})
		}
	}

	if b.KopsModelContext.Cluster.Spec.Networking.Kuberouter != nil {
		t := &fitasks.Keypair{
			Name:    fi.String("kube-router"),
			Subject: "cn=" + "system:kube-router",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kubecfg"),
			Lifecycle: b.Lifecycle,
			Subject:   "o=" + user.SystemPrivilegedGroup + ",cn=kubecfg",
			Type:      "client",
		}
		c.AddTask(t)
	}

	{
		t := &fitasks.Keypair{
			Name:      fi.String("kops"),
			Lifecycle: b.Lifecycle,
			Subject:   "o=" + user.SystemPrivilegedGroup + ",cn=kops",
			Type:      "client",
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
		}
		c.AddTask(t)
	}

	// @@ The following are deprecated for > 1.6 and should be dropped at the appropreciate time
	deprecated := []string{
		"kubelet", "kube-proxy", "system:scheduler", "system:controller_manager",
		"system:logging", "system:monitoring", "system:dns", "kube", "admin"}

	for _, x := range deprecated {
		t := &fitasks.Secret{Name: fi.String(x), Lifecycle: b.Lifecycle}
		c.AddTask(t)
	}

	return nil
}

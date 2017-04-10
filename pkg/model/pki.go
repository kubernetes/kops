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

func (b *PKIModelBuilder) Build(c *fi.ModelBuilderContext) error {
	{
		// Keypair used by the kubelet
		t := &fitasks.Keypair{
			Name:      fi.String("kubelet"),
			Lifecycle: b.Lifecycle,

			Subject: "o=" + user.NodesGroup + ",cn=kubelet",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Secret used by the kubelet
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("kubelet"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Keypair used by the kube-scheduler
		t := &fitasks.Keypair{
			Name:      fi.String("kube-scheduler"),
			Lifecycle: b.Lifecycle,

			Subject: "cn=" + user.KubeScheduler,
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Secret used by the kube-scheduler
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("system:scheduler"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Keypair used by the kube-proxy
		t := &fitasks.Keypair{
			Name:      fi.String("kube-proxy"),
			Lifecycle: b.Lifecycle,

			Subject: "cn=" + user.KubeProxy,
			Type:    "client",
		}
		c.AddTask(t)
	}

	if b.KopsModelContext.Cluster.Spec.Networking.Kuberouter != nil {
		// Keypair used by the kube-router
		t := &fitasks.Keypair{
			Name:    fi.String("kube-router"),
			Subject: "cn=" + "system:kube-router",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Secret used by the kube-proxy
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("kube-proxy"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Keypair used by the kube-controller-manager
		t := &fitasks.Keypair{
			Name:      fi.String("kube-controller-manager"),
			Lifecycle: b.Lifecycle,

			Subject: "cn=" + user.KubeControllerManager,
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Secret used by the kube-controller-manager
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("system:controller_manager"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Keypair used for admin kubecfg
		t := &fitasks.Keypair{
			Name:      fi.String("kubecfg"),
			Lifecycle: b.Lifecycle,

			Subject: "o=" + user.SystemPrivilegedGroup + ",cn=kubecfg",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Keypair used by kops / protokube
		t := &fitasks.Keypair{
			Name:      fi.String("kops"),
			Lifecycle: b.Lifecycle,

			Subject: "o=" + user.SystemPrivilegedGroup + ",cn=kops",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// TLS certificate used for apiserver

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
			Name:      fi.String("master"),
			Lifecycle: b.Lifecycle,

			Subject:        "cn=kubernetes-master",
			Type:           "server",
			AlternateNames: alternateNames,
		}
		c.AddTask(t)
	}

	{
		// Secret used by logging (?)
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("system:logging"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Secret used by monitoring (?)
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("system:monitoring"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Secret used by dns (?)
		// TODO: Can this be removed... at least from 1.6 on?
		t := &fitasks.Secret{
			Name:      fi.String("system:dns"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Secret used by kube (?)
		// TODO: Can this be removed... at least from 1.6 on? Although one of kube/admin is the primary token auth
		t := &fitasks.Secret{
			Name:      fi.String("kube"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	{
		// Secret used by admin (?)
		// TODO: Can this be removed... at least from 1.6 on? Although one of kube/admin is the primary token auth
		t := &fitasks.Secret{
			Name:      fi.String("admin"),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(t)
	}

	return nil
}

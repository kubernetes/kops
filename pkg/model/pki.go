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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
)

// PKIModelBuilder configures PKI keypairs
type PKIModelBuilder struct {
	*KopsModelContext
}

var _ fi.ModelBuilder = &PKIModelBuilder{}

func (b *PKIModelBuilder) Build(c *fi.ModelBuilderContext) error {
	{
		// Keypair used by the kubelet
		t := &fitasks.Keypair{
			Name:    fi.String("kubelet"),
			Subject: "cn=kubelet",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Keypair used for admin kubecfg
		t := &fitasks.Keypair{
			Name:    fi.String("kubecfg"),
			Subject: "cn=kubecfg",
			Type:    "client",
		}
		c.AddTask(t)
	}

	{
		// Keypair used for apiserver

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

		t := &fitasks.Keypair{
			Name:           fi.String("master"),
			Subject:        "cn=kubernetes-master",
			Type:           "server",
			AlternateNames: alternateNames,
		}
		c.AddTask(t)
	}

	return nil
}

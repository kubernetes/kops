/*
Copyright 2026 The Kubernetes Authors.

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

package elementomodel

import (
	"strings"

	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/elementotasks"
)

const (
	placeholderIP                      = "203.0.113.123"
	kopsControllerInternalRecordPrefix = "kops-controller.internal."
	defaultTTL                         = int64(60)
)

// DNSModelBuilder is the provider-native integration point for Elemento-managed
// DNS records that must exist before nodeup starts.
type DNSModelBuilder struct {
	*ElementoModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.Cluster.PublishesDNSRecords() {
		return nil
	}

	// This builder mirrors the role of other provider-specific DNS builders in kOps:
	// it declares the DNS records that must exist before nodeup starts. The actual
	// Elemento DNS API calls are delegated to elementotasks.DNSRecord.

	if !b.UseLoadBalancerForAPI() {
		recordName := trimZoneSuffix(b.Cluster.Spec.API.PublicName, b.Cluster.Spec.DNSZone)
		c.AddTask(&elementotasks.DNSRecord{
			Name:      fi.PtrTo(recordName),
			DNSZone:   fi.PtrTo(b.Cluster.Spec.DNSZone),
			Type:      fi.PtrTo("A"),
			Data:      fi.PtrTo(placeholderIP),
			TTL:       fi.PtrTo(defaultTTL),
			Lifecycle: b.Lifecycle,
			Comment: fi.PtrTo(
				"Bootstrap record for the public Kubernetes API endpoint. " +
					"Replace the placeholder target with the final Elemento-managed VIP or public API address.",
			),
		})
	}

	if !b.UseLoadBalancerForInternalAPI() {
		recordName := trimZoneSuffix(b.Cluster.APIInternalName(), b.Cluster.Spec.DNSZone)
		c.AddTask(&elementotasks.DNSRecord{
			Name:      fi.PtrTo(recordName),
			DNSZone:   fi.PtrTo(b.Cluster.Spec.DNSZone),
			Type:      fi.PtrTo("A"),
			Data:      fi.PtrTo(placeholderIP),
			TTL:       fi.PtrTo(defaultTTL),
			Lifecycle: b.Lifecycle,
			Comment: fi.PtrTo(
				"Bootstrap record for api.internal. This must resolve before kubeconfig, " +
					"service-account issuer discovery, and early control-plane traffic start using it.",
			),
		})
	}

	recordName := kopsControllerInternalRecordPrefix + strings.TrimSuffix(b.Cluster.ObjectMeta.Name, "."+b.Cluster.Spec.DNSZone)
	c.AddTask(&elementotasks.DNSRecord{
		Name:      fi.PtrTo(recordName),
		DNSZone:   fi.PtrTo(b.Cluster.Spec.DNSZone),
		Type:      fi.PtrTo("A"),
		Data:      fi.PtrTo(placeholderIP),
		TTL:       fi.PtrTo(defaultTTL),
		Lifecycle: b.Lifecycle,
		Comment: fi.PtrTo(
			"Bootstrap record for kops-controller.internal. Worker nodeup may use this very early " +
				"to fetch configuration from the config server.",
		),
	})

	return nil
}

func trimZoneSuffix(name string, zone string) string {
	return strings.TrimSuffix(name, "."+zone)
}

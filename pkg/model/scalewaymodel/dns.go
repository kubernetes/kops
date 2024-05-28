/*
Copyright 2023 The Kubernetes Authors.

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

package scalewaymodel

import (
	"strings"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scalewaytasks"
)

const (
	kopsControllerInternalRecordPrefix = "kops-controller.internal."
	defaultTTL                         = uint32(60)
)

type DNSModelBuilder struct {
	*ScwModelContext
	Lifecycle fi.Lifecycle
}

var _ fi.CloudupModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	if !b.Cluster.PublishesDNSRecords() {
		return nil
	}

	if !b.UseLoadBalancerForAPI() {
		recordShortName := strings.TrimSuffix(b.Cluster.Spec.API.PublicName, "."+b.Cluster.Spec.DNSZone)
		dnsAPIExternal := &scalewaytasks.DNSRecord{
			Name:        fi.PtrTo(recordShortName),
			Data:        fi.PtrTo(scalewaytasks.PlaceholderIP),
			DNSZone:     fi.PtrTo(b.Cluster.Spec.DNSZone),
			Type:        fi.PtrTo(domain.RecordTypeA.String()),
			TTL:         fi.PtrTo(defaultTTL),
			Lifecycle:   b.Lifecycle,
			IsInternal:  fi.PtrTo(false),
			ClusterName: fi.PtrTo(b.Cluster.Name),
		}
		c.AddTask(dnsAPIExternal)
	}

	if !b.UseLoadBalancerForInternalAPI() {
		recordShortName := strings.TrimSuffix(b.Cluster.APIInternalName(), "."+b.Cluster.Spec.DNSZone)
		dnsAPIInternal := &scalewaytasks.DNSRecord{
			Name:        fi.PtrTo(recordShortName),
			Data:        fi.PtrTo(scalewaytasks.PlaceholderIP),
			DNSZone:     fi.PtrTo(b.Cluster.Spec.DNSZone),
			Type:        fi.PtrTo(domain.RecordTypeA.String()),
			TTL:         fi.PtrTo(defaultTTL),
			IsInternal:  fi.PtrTo(true),
			ClusterName: fi.PtrTo(b.Cluster.Name),
			Lifecycle:   b.Lifecycle,
		}
		c.AddTask(dnsAPIInternal)
	}

	recordSuffix := strings.TrimSuffix(b.Cluster.ObjectMeta.Name, "."+b.Cluster.Spec.DNSZone)
	recordShortName := kopsControllerInternalRecordPrefix + recordSuffix
	kopsControllerInternal := &scalewaytasks.DNSRecord{
		Name:        fi.PtrTo(recordShortName),
		Data:        fi.PtrTo(scalewaytasks.PlaceholderIP),
		DNSZone:     fi.PtrTo(b.Cluster.Spec.DNSZone),
		Type:        fi.PtrTo(domain.RecordTypeA.String()),
		TTL:         fi.PtrTo(defaultTTL),
		IsInternal:  fi.PtrTo(true),
		ClusterName: fi.PtrTo(b.Cluster.Name),
		Lifecycle:   b.Lifecycle,
	}
	c.AddTask(kopsControllerInternal)

	return nil
}

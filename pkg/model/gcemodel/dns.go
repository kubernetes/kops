/*
Copyright 2022 The Kubernetes Authors.

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

package gcemodel

import (
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// DNSModelBuilder builds DNS related model objects
type DNSModelBuilder struct {
	*GCEModelContext

	Cloud     gce.GCECloud
	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &DNSModelBuilder{}

func (b *DNSModelBuilder) Build(c *fi.ModelBuilderContext) error {
	name := b.Cluster.Spec.DNSZone
	// Just in case our user forgot to end their name with a '.'
	// This matches the behavior in the GCP console (which is indifferent to trailing dots)
	if !strings.HasSuffix(name, ".") {
		name = name + "."
	}
	visibility := s("private")
	if b.Cluster.Spec.Topology == nil || b.Cluster.Spec.Topology.DNS == nil || b.Cluster.Spec.Topology.DNS.Type != kops.DNSTypePrivate {
		klog.Info("DNS topology is non-private, skipping the creation / editing of the zone.")
		return nil
	}
	mz := &gcetasks.ManagedZone{
		DNSName:    s(name),
		Name:       s(b.SafeObjectName("zone")),
		Visibility: visibility,
		Lifecycle:  b.Lifecycle,
		Labels:     map[string]string{"created-by": "kops", "cluster": b.Cluster.Name},
	}
	c.AddTask(mz)

	return nil
}

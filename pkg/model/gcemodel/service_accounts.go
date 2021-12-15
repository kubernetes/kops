/*
Copyright 2021 The Kubernetes Authors.

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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gcetasks"
)

// ServiceAccountsBuilder configures service accounts and grants project permissions
type ServiceAccountsBuilder struct {
	*GCEModelContext

	Lifecycle fi.Lifecycle
}

var _ fi.ModelBuilder = &ServiceAccountsBuilder{}

func (b *ServiceAccountsBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Cluster.Spec.CloudConfig.GCEServiceAccount != "" {
		serviceAccount := &gcetasks.ServiceAccount{
			Name:      s("shared"),
			Email:     &b.Cluster.Spec.CloudConfig.GCEServiceAccount,
			Shared:    fi.Bool(true),
			Lifecycle: b.Lifecycle,
		}
		c.AddTask(serviceAccount)

		return nil
	}

	return nil
}

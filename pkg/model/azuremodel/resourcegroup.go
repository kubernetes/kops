/*
Copyright 2020 The Kubernetes Authors.

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

package azuremodel

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/azuretasks"
)

// ResourceGroupModelBuilder configures a Resource Group.
type ResourceGroupModelBuilder struct {
	*AzureModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &ResourceGroupModelBuilder{}

// Build builds a task for creating a Resource Group.
func (b *ResourceGroupModelBuilder) Build(c *fi.ModelBuilderContext) error {
	t := &azuretasks.ResourceGroup{
		Name:      fi.String(b.NameForResourceGroup()),
		Lifecycle: b.Lifecycle,
		Tags:      map[string]*string{},
		Shared:    fi.Bool(b.Cluster.IsSharedAzureResourceGroup()),
	}
	c.AddTask(t)
	return nil
}

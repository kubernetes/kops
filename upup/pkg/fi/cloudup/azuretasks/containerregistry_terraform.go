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

package azuretasks

import (
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type terraformContainerRegistry struct {
	Name              *string                  `cty:"name"`
	Location          *string                  `cty:"location"`
	ResourceGroupName *terraformWriter.Literal `cty:"resource_group_name"`
	SKU               *string                  `cty:"sku"`
	Tags              map[string]string        `cty:"tags"`
}

func (*ContainerRegistry) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ContainerRegistry) error {
	tf := &terraformContainerRegistry{
		Name:              e.Name,
		Location:          to.Ptr(t.Cloud.Region()),
		ResourceGroupName: e.ResourceGroup.terraformName(),
		SKU:               to.Ptr("Basic"),
		Tags:              stringMap(e.Tags),
	}
	return t.RenderResource("azurerm_container_registry", fi.ValueOf(e.Name), tf)
}

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
	"fmt"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type terraformAzureManagedDisk struct {
	Name               *string                  `cty:"name"`
	Location           *string                  `cty:"location"`
	ResourceGroupName  *terraformWriter.Literal `cty:"resource_group_name"`
	StorageAccountType *string                  `cty:"storage_account_type"`
	CreateOption       *string                  `cty:"create_option"`
	DiskSizeGB         *int32                   `cty:"disk_size_gb"`
	Zone               *string                  `cty:"zone"`
	Tags               map[string]string        `cty:"tags"`
}

func (*Disk) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Disk) error {
	if len(e.Zones) > 1 {
		return fmt.Errorf("expected at most one zone for disk %q, got %d", fi.ValueOf(e.Name), len(e.Zones))
	}

	createOption := string(compute.DiskCreateOptionEmpty)
	tf := &terraformAzureManagedDisk{
		Name:               e.Name,
		Location:           new(t.Cloud.Region()),
		ResourceGroupName:  e.ResourceGroup.terraformName(),
		StorageAccountType: stringPtr(e.VolumeType),
		CreateOption:       &createOption,
		DiskSizeGB:         e.SizeGB,
		Tags:               stringMap(e.Tags),
	}
	if len(e.Zones) == 1 {
		tf.Zone = e.Zones[0]
	}

	return t.RenderResource("azurerm_managed_disk", fi.ValueOf(e.Name), tf)
}

/*
Copyright 2025 The Kubernetes Authors.

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
	_ "embed"
	"errors"
	"os"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// AzureBuilder writes the Azure-specific configuration
type AzureBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &AzureBuilder{}

//go:embed resources/80-azure-disk.rules
var azureUdevDiskRules string

func (b *AzureBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.CloudProvider() != kops.CloudProviderAzure {
		return nil
	}

	// Create symlinks in /dev/disk/azure for local, data, and os disks, if not already present:
	// https://github.com/Azure/azure-vm-utils/blob/main/udev/80-azure-disk.rules
	azureUdevDiskRulesPath := "/etc/udev/rules.d/80-azure-disk.rules"
	if _, err := os.Stat(azureUdevDiskRulesPath); errors.Is(err, os.ErrNotExist) {
		c.AddTask(&nodetasks.File{
			Path:     azureUdevDiskRulesPath,
			Contents: fi.NewStringResource(azureUdevDiskRules),
			Type:     nodetasks.FileType_File,
			OnChangeExecute: [][]string{
				{"udevadm", "control", "--reload"},
				{"udevadm", "trigger", "--subsystem-match=block"},
			},
		})
	}

	return nil
}

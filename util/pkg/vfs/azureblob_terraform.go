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

package vfs

import (
	"fmt"
	"io"

	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

type terraformAzureBlobFile struct {
	Name               string                   `cty:"name"`
	StorageContainerID string                   `cty:"storage_container_id"`
	Type               string                   `cty:"type"`
	Source             *terraformWriter.Literal `cty:"source"`
	Provider           *terraformWriter.Literal `cty:"provider"`
}

func (p *AzureBlobPath) RenderTerraform(w *terraformWriter.TerraformWriter, name string, data io.Reader, acl ACL) error {
	bytes, err := io.ReadAll(data)
	if err != nil {
		return fmt.Errorf("reading data: %w", err)
	}

	w.EnsureTerraformProvider("azurerm", map[string]string{})

	if p.account == "" {
		return fmt.Errorf("Azure storage account is not set on path %q", p.Path())
	}
	if w.AzureStorageAccountID == "" {
		return fmt.Errorf("Azure storage account ID is not set; it is required to render blob %q", p.Path())
	}

	source, err := w.AddFilePath("azurerm_storage_blob", name, "source", bytes, false)
	if err != nil {
		return fmt.Errorf("rendering Azure Blob file: %w", err)
	}

	tf := &terraformAzureBlobFile{
		Name:               p.key,
		StorageContainerID: w.AzureStorageAccountID + "/blobServices/default/containers/" + p.container,
		Type:               "Block",
		Source:             source,
		Provider:           terraformWriter.LiteralTokens("azurerm", "files"),
	}
	return w.RenderResource("azurerm_storage_blob", name, tf)
}

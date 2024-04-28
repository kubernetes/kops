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

package vfs

import (
	"context"
	"fmt"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"k8s.io/klog/v2"
)

func newAzureClient(ctx context.Context) (*azblob.Client, error) {
	klog.Infof("New Azure Blob client")
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	if accountName == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_ACCOUNT must be set")
	}

	url := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)

	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}

	client, err := azblob.NewClient(url, credential, nil)
	if err != nil {
		return nil, err
	}

	return client, nil
}

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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

const (
	storageResourceID = "https://storage.azure.com/"
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

// getAccessTokenFromInstanceMetadataService obtains the access token from Instance Metadata Service.
func getAccessTokenFromInstanceMetadataService(ctx context.Context) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/metadata/identity/oauth2/token", nil)
	if err != nil {
		return "", fmt.Errorf("error creating a new request: %s", err)
	}
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("resource", storageResourceID)
	q.Add("api-version", "2020-06-01")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request to the metadata server: %s", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading a response from the metadata server: %s", err)
	}

	token := &oauth2.Token{}
	if err := json.Unmarshal(body, token); err != nil {
		return "", fmt.Errorf("error unmarsharlling token: %s", err)
	}
	return token.AccessToken, nil
}

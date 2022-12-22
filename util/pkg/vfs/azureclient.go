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
	"net/url"
	"os"

	"github.com/Azure/azure-pipeline-go/pipeline"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"golang.org/x/oauth2"
	"k8s.io/klog/v2"
)

const (
	storageResourceID = "https://storage.azure.com/"
)

type azureClient struct {
	p           pipeline.Pipeline
	accountName string
}

func (c *azureClient) newContainerURL(containerName string) (*azblob.ContainerURL, error) {
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net/%s", c.accountName, containerName))
	if err != nil {
		return nil, err
	}
	// Create a ContainerURL object that wraps the container URL and a request
	// pipeline to make requests.
	cURL := azblob.NewContainerURL(*u, c.p)
	return &cURL, nil
}

func newAzureClient(ctx context.Context) (*azureClient, error) {
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	if accountName == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_ACCOUNT must be set")
	}
	credential, err := newAzureCredential(ctx, accountName)
	if err != nil {
		return nil, err
	}
	return &azureClient{
		p:           azblob.NewPipeline(credential, azblob.PipelineOptions{}),
		accountName: accountName,
	}, nil
}

// newAzureCredential returns a Azure credential. When env var
// AZURE_STORAGE_KEY is set, obtain a credential from the env
// var. When the env var is not set, obtain a credential from Instance
// Metadata Service.
//
// Please note that Instance Metadata Service is available only within a VM
// running in Azure (and when necessary role is attached to the VM).
func newAzureCredential(ctx context.Context, accountName string) (azblob.Credential, error) {
	accountKey := os.Getenv("AZURE_STORAGE_KEY")
	if accountKey != "" {
		klog.V(2).Infof("Creating a shared key credential")
		return azblob.NewSharedKeyCredential(accountName, accountKey)
	}

	klog.V(2).Infof("Creating a token credential from Instance Metadata Service.")
	token, err := getAccessTokenFromInstanceMetadataService(ctx)
	if err != nil {
		return nil, err
	}
	return azblob.NewTokenCredential(token, nil), nil
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

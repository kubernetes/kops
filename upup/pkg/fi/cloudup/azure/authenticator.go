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

package azure

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
)

const AzureAuthenticationTokenPrefix = "x-azure-id "

type azureAuthenticator struct {
}

var _ bootstrap.Authenticator = &azureAuthenticator{}

func NewAzureAuthenticator() (bootstrap.Authenticator, error) {
	return &azureAuthenticator{}, nil
}

func (h *azureAuthenticator) CreateToken(body []byte) (string, error) {
	m, err := queryInstanceMetadata()
	if err != nil {
		return "", fmt.Errorf("querying instance metadata: %w", err)
	}
	if m.Compute == nil || m.Compute.VMID == "" {
		return "", fmt.Errorf("missing virtual machine ID")
	}

	token := m.Compute.VMID

	// This should be a fully qualified VM resource ID.
	r := strings.Split(m.Compute.ResourceID, "/")
	switch len(r) {
	case 11:
		if r[7] == "virtualMachineScaleSets" && r[9] == "virtualMachines" {
			token += " " + r[8] + " " + r[10]
		}
	case 9:
		if r[7] == "virtualMachines" {
			token += " " + r[8]
		}
	}

	klog.Info(token)
	return AzureAuthenticationTokenPrefix + token, nil
}

type instanceComputeMetadata struct {
	ResourceGroupName string `json:"resourceGroupName"`
	ResourceID        string `json:"resourceId"`
	SubscriptionID    string `json:"subscriptionId"`
	VMID              string `json:"vmId"`
}

type instanceMetadata struct {
	Compute *instanceComputeMetadata `json:"compute"`
}

// queryInstanceMetadata queries Azure Instance Metadata Service (IMDS)
// https://learn.microsoft.com/en-us/azure/virtual-machines/instance-metadata-service?tabs=linux
func queryInstanceMetadata() (*instanceMetadata, error) {
	transport := &http.Transport{Proxy: nil}

	client := http.Client{Transport: transport}

	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance", nil)
	if err != nil {
		return nil, fmt.Errorf("creating a new request: %w", err)
	}
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api-version", "2021-02-01")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request to the instance metadata server: %w", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading a response from the metadata server: %w", err)
	}
	metadata := &instanceMetadata{}
	err = json.Unmarshal(body, metadata)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling instance metadata: %w", err)
	}

	return metadata, nil
}

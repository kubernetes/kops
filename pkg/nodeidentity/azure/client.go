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

package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2020-06-01/compute"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

type instanceComputeMetadata struct {
	ResourceGroupName string `json:"resourceGroupName"`
	SubscriptionID    string `json:"subscriptionId"`
}

type instanceMetadata struct {
	Compute *instanceComputeMetadata `json:"compute"`
}

// client is an Azure client.
type client struct {
	metadata     *instanceMetadata
	vmssesClient *compute.VirtualMachineScaleSetsClient
}

// newClient returns a new Client.
func newClient() (*client, error) {
	m, err := queryInstanceMetadata()
	if err != nil {
		return nil, fmt.Errorf("error querying instance metadata: %s", err)
	}
	if m.Compute.SubscriptionID == "" {
		return nil, fmt.Errorf("empty subscription name")
	}
	if m.Compute.ResourceGroupName == "" {
		return nil, fmt.Errorf("empty resource group name")
	}

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("error creating an authorizer: %s", err)
	}

	vmssesClient := compute.NewVirtualMachineScaleSetsClient(m.Compute.SubscriptionID)
	vmssesClient.Authorizer = authorizer

	return &client{
		metadata:     m,
		vmssesClient: &vmssesClient,
	}, nil
}

// getVMScaleSet returns the specified VM ScaleSet.
func (c *client) getVMScaleSet(ctx context.Context, vmssName string) (compute.VirtualMachineScaleSet, error) {
	return c.vmssesClient.Get(ctx, c.metadata.Compute.ResourceGroupName, vmssName)
}

// queryInstanceMetadata queries Azure Instance Metadata documented in
// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service.
func queryInstanceMetadata() (*instanceMetadata, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating a new request: %s", err)
	}
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("format", "json")
	q.Add("api-version", "2020-06-01")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to the metadata server: %s", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading a response from the metadata server: %s", err)
	}
	metadata, err := unmarshalInstanceMetadata(body)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %s", err)
	}
	return metadata, nil
}

func unmarshalInstanceMetadata(data []byte) (*instanceMetadata, error) {
	m := &instanceMetadata{}
	if err := json.Unmarshal(data, m); err != nil {
		return nil, err
	}
	return m, nil
}

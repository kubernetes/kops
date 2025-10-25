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
	"io"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

// client is an Azure client.
type client struct {
	subscriptionID string
	vmClient       *compute.VirtualMachinesClient
	vmssClient     *compute.VirtualMachineScaleSetVMsClient
}

// newClient returns a new Client.
func newClient() (*client, error) {
	metadata, err := queryComputeInstanceMetadata()
	if err != nil {
		return nil, fmt.Errorf("error querying instance metadata: %s", err)
	}
	if metadata.SubscriptionID == "" {
		return nil, fmt.Errorf("empty subscription ID")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("creating identity: %w", err)
	}

	vmClient, err := compute.NewVirtualMachinesClient(metadata.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMs client: %w", err)
	}

	vmssClient, err := compute.NewVirtualMachineScaleSetVMsClient(metadata.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSS VMs client: %w", err)
	}

	return &client{
		vmClient:   vmClient,
		vmssClient: vmssClient,
	}, nil
}

func (c *client) getVMTags(ctx context.Context, providerID string) (map[string]*string, error) {
	if !strings.HasPrefix(providerID, "azure://") {
		return nil, fmt.Errorf("unknown providerID : %s", providerID)
	}

	res, err := arm.ParseResourceID(strings.TrimPrefix(providerID, "azure://"))
	if err != nil {
		return nil, fmt.Errorf("error parsing providerID: %v", err)
	}

	switch res.ResourceType.String() {
	case "Microsoft.Compute/virtualMachines":
		resp, err := c.vmClient.Get(ctx, res.ResourceGroupName, res.Name, nil)
		if err != nil {
			return nil, fmt.Errorf("getting VM: %w", err)
		}
		return resp.VirtualMachine.Tags, nil
	case "Microsoft.Compute/virtualMachineScaleSets/virtualMachines":
		resp, err := c.vmssClient.Get(ctx, res.ResourceGroupName, res.Parent.Name, res.Name, nil)
		if err != nil {
			return nil, fmt.Errorf("getting VMSS VM: %w", err)
		}
		return resp.VirtualMachineScaleSetVM.Tags, nil
	default:
		return nil, fmt.Errorf("unsupported resource type %q for %q", res.ResourceType, providerID)
	}
}

type instanceMetadata struct {
	SubscriptionID    string `json:"subscriptionId"`
	ResourceGroupName string `json:"resourceGroupName"`
}

// queryComputeInstanceMetadata queries Azure Instance Metadata.
// https://docs.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service
func queryComputeInstanceMetadata() (*instanceMetadata, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://169.254.169.254/metadata/instance/compute", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating a new request: %s", err)
	}
	req.Header.Add("Metadata", "True")

	q := req.URL.Query()
	q.Add("api-version", "2025-04-07")
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request to the metadata server: %s", err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading a response from the metadata server: %s", err)
	}
	metadata := &instanceMetadata{}
	if err := json.Unmarshal(body, metadata); err != nil {
		return nil, fmt.Errorf("error unmarshalling metadata: %s", err)
	}
	return metadata, nil
}

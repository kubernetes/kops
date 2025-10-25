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
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
)

const (
	// InstanceGroupNameTag is the key of the tag used to identify an instance group that VM belongs to.
	InstanceGroupNameTag = "kops.k8s.io_instancegroup"
)

type AzureVerifierOptions struct {
	ClusterName string `json:"clusterName,omitempty"`
}

type azureVerifier struct {
	client      *client
	clusterName string
}

var _ bootstrap.Verifier = &azureVerifier{}

func NewAzureVerifier(ctx context.Context, opt *AzureVerifierOptions) (bootstrap.Verifier, error) {
	azureClient, err := newClient()
	if err != nil {
		return nil, err
	}

	if opt == nil || opt.ClusterName == "" {
		return nil, fmt.Errorf("determining cluster name")
	}

	return &azureVerifier{
		client:      azureClient,
		clusterName: opt.ClusterName,
	}, nil
}

func (a azureVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, AzureAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}

	v := strings.Split(strings.TrimPrefix(token, AzureAuthenticationTokenPrefix), " ")
	if len(v) != 2 {
		return nil, fmt.Errorf("incorrect token format")
	}
	resourceID := v[0]
	vmID := v[1]

	res, err := arm.ParseResourceID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %v", err)
	}

	var nodeName, igName string
	var addrs, challengeEndpoints []string

	switch res.ResourceType.String() {
	case "Microsoft.Compute/virtualMachines":
		vmName := res.Name

		vm, err := a.client.vmsClient.Get(ctx, a.client.resourceGroup, vmName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VM %q: %w", vmName, err)
		}
		if vm.Properties == nil || vm.Properties.VMID == nil {
			return nil, fmt.Errorf("determining VMID for VM %q", vmName)
		}
		if vmID != *vm.Properties.VMID {
			return nil, fmt.Errorf("matching VMID %q to VM %q", vmID, vmName)
		}
		if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil || *vm.Properties.OSProfile.ComputerName == "" {
			return nil, fmt.Errorf("determining ComputerName for VM %q", vmName)
		}

		nodeName = strings.ToLower(*vm.Properties.OSProfile.ComputerName)
		if igNameTag, ok := vm.Tags[InstanceGroupNameTag]; ok && igNameTag != nil {
			igName = *igNameTag
		} else {
			return nil, fmt.Errorf("determining IG name for VM %q", vmName)
		}

		ni, err := a.client.nisClient.Get(ctx, a.client.resourceGroup, nodeName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VM network interface %q: %w", vmName, err)
		}

		for _, ipc := range ni.Properties.IPConfigurations {
			if ipc.Properties != nil && ipc.Properties.PrivateIPAddress != nil {
				addrs = append(addrs, *ipc.Properties.PrivateIPAddress)
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(*ipc.Properties.PrivateIPAddress, strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}

	case "Microsoft.Compute/virtualMachineScaleSets/virtualMachines":
		vmssName := res.Parent.Name
		vmssIndex := res.Name

		if !strings.HasSuffix(vmssName, "."+a.clusterName) {
			return nil, fmt.Errorf("matching cluster name %q to VMSS %q", a.clusterName, vmssName)
		}

		vm, err := a.client.vmssVMsClient.Get(ctx, a.client.resourceGroup, vmssName, vmssIndex, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VMSS VM %q #%s: %w", vmssName, vmssIndex, err)
		}
		if vm.Properties == nil || vm.Properties.VMID == nil {
			return nil, fmt.Errorf("determining VMID for VMSS %q VM #%s", vmssName, vmssIndex)
		}
		if vmID != *vm.Properties.VMID {
			return nil, fmt.Errorf("matching VMID %q to VMSS %q VM #%s", vmID, vmssName, vmssIndex)
		}
		if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil || *vm.Properties.OSProfile.ComputerName == "" {
			return nil, fmt.Errorf("determining ComputerName for VMSS %q VM #%s", vmssName, vmssIndex)
		}

		nodeName = strings.ToLower(*vm.Properties.OSProfile.ComputerName)
		if igNameTag, ok := vm.Tags[InstanceGroupNameTag]; ok && igNameTag != nil {
			igName = *igNameTag
		} else {
			return nil, fmt.Errorf("determining IG name for VM %q", vmssName)
		}

		ni, err := a.client.nisClient.GetVirtualMachineScaleSetNetworkInterface(ctx, a.client.resourceGroup, vmssName, vmssIndex, vmssName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VMSS VM network interface %q #%s: %w", vmssName, vmssIndex, err)
		}

		for _, ipc := range ni.Properties.IPConfigurations {
			if ipc.Properties != nil && ipc.Properties.PrivateIPAddress != nil {
				addrs = append(addrs, *ipc.Properties.PrivateIPAddress)
				challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(*ipc.Properties.PrivateIPAddress, strconv.Itoa(wellknownports.NodeupChallenge)))
			}
		}

	default:
		return nil, fmt.Errorf("unsupported resource type %q", res.ResourceType)
	}

	if len(addrs) == 0 {
		return nil, fmt.Errorf("determining certificate alternate names for node %q", nodeName)
	}
	if len(challengeEndpoints) == 0 {
		return nil, fmt.Errorf("determining challenge endpoint for node %q", nodeName)
	}

	result := &bootstrap.VerifyResult{
		NodeName:          nodeName,
		InstanceGroupName: igName,
		CertificateNames:  addrs,
		ChallengeEndpoint: challengeEndpoints[0],
	}

	return result, nil
}

// client is an Azure client.
type client struct {
	resourceGroup string
	nisClient     *network.InterfacesClient
	vmsClient     *compute.VirtualMachinesClient
	vmssVMsClient *compute.VirtualMachineScaleSetVMsClient
}

// newClient returns a new Client.
func newClient() (*client, error) {
	metadata, err := queryComputeInstanceMetadata()
	if err != nil || metadata == nil {
		return nil, fmt.Errorf("getting instance metadata: %w", err)
	}
	if metadata.ResourceGroupName == "" {
		return nil, fmt.Errorf("empty resource group name")
	}
	if metadata.SubscriptionID == "" {
		return nil, fmt.Errorf("empty subscription ID")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("creating an identity: %w", err)
	}

	nisClient, err := network.NewInterfacesClient(metadata.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating interfaces client: %w", err)
	}
	vmsClient, err := compute.NewVirtualMachinesClient(metadata.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMs client: %w", err)
	}
	vmssVMsClient, err := compute.NewVirtualMachineScaleSetVMsClient(metadata.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSSVMs client: %w", err)
	}

	return &client{
		resourceGroup: metadata.ResourceGroupName,
		nisClient:     nisClient,
		vmsClient:     vmsClient,
		vmssVMsClient: vmssVMsClient,
	}, nil
}

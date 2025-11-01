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

	var nodeName, igName string
	var addrs, challengeEndpoints []string

	v := strings.Split(strings.TrimPrefix(token, AzureAuthenticationTokenPrefix), " ")
	switch len(v) {
	case 2:
		vmId := v[0]
		vmName := v[1]

		vm, err := a.client.vmsClient.Get(ctx, a.client.resourceGroup, vmName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VM %q: %w", vmName, err)
		}
		if vm.Properties == nil || vm.Properties.VMID == nil {
			return nil, fmt.Errorf("determining VMID for VM %q", vmName)
		}
		if vmId != *vm.Properties.VMID {
			return nil, fmt.Errorf("matching VMID %q to VM %q", vmId, vmName)
		}
		if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil || *vm.Properties.OSProfile.ComputerName == "" {
			return nil, fmt.Errorf("determining ComputerName for VM %q", vmName)
		}

		nodeName = strings.ToLower(*vm.Properties.OSProfile.ComputerName)

		if v, ok := vm.Tags[InstanceGroupNameTag]; !ok || v == nil {
			return nil, fmt.Errorf("determining IG name for VM %q", vmName)
		}
		igName = *vm.Tags[InstanceGroupNameTag]

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

	case 3:
		vmId := v[0]
		vmssName := v[1]
		vmssIndex := v[2]

		if !strings.HasSuffix(vmssName, "."+a.clusterName) {
			return nil, fmt.Errorf("matching cluster name %q to VMSS %q", a.clusterName, vmssName)
		}
		igName = strings.TrimSuffix(vmssName, "."+a.clusterName)

		vm, err := a.client.vmssVMsClient.Get(ctx, a.client.resourceGroup, vmssName, vmssIndex, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VMSS VM %q #%s: %w", vmssName, vmssIndex, err)
		}
		if vm.Properties == nil || vm.Properties.VMID == nil {
			return nil, fmt.Errorf("determining VMID for VMSS %q VM #%s", vmssName, vmssIndex)
		}
		if vmId != *vm.Properties.VMID {
			return nil, fmt.Errorf("matching VMID %q to VMSS %q VM #%s", vmId, vmssName, vmssIndex)
		}
		if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil || *vm.Properties.OSProfile.ComputerName == "" {
			return nil, fmt.Errorf("determining ComputerName for VMSS %q VM #%s", vmssName, vmssIndex)
		}
		nodeName = strings.ToLower(*vm.Properties.OSProfile.ComputerName)

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
		return nil, fmt.Errorf("incorrect token format")
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
	m, err := queryInstanceMetadata()
	if err != nil || m == nil {
		return nil, fmt.Errorf("getting instance metadata: %w", err)
	}
	if m.Compute == nil || m.Compute.ResourceGroupName == "" {
		return nil, fmt.Errorf("empty resource group name")
	}
	if m.Compute == nil || m.Compute.SubscriptionID == "" {
		return nil, fmt.Errorf("empty subscription name")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("creating an identity: %w", err)
	}

	nisClient, err := network.NewInterfacesClient(m.Compute.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating interfaces client: %w", err)
	}
	vmsClient, err := compute.NewVirtualMachinesClient(m.Compute.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMs client: %w", err)
	}
	vmssVMsClient, err := compute.NewVirtualMachineScaleSetVMsClient(m.Compute.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("creating VMSSVMs client: %w", err)
	}

	return &client{
		resourceGroup: m.Compute.ResourceGroupName,
		nisClient:     nisClient,
		vmsClient:     vmsClient,
		vmssVMsClient: vmssVMsClient,
	}, nil
}

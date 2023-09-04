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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2022-08-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2022-05-01/network"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
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
		return nil, fmt.Errorf("incorrect authorization type")
	}

	v := strings.Split(strings.TrimPrefix(token, AzureAuthenticationTokenPrefix), " ")
	if len(v) != 3 {
		return nil, fmt.Errorf("incorrect token format")
	}
	vmId := v[0]
	vmssName := v[1]
	vmssIndex := v[2]

	if !strings.HasSuffix(vmssName, "."+a.clusterName) {
		return nil, fmt.Errorf("matching cluster name %q to VMSS %q", a.clusterName, vmssName)
	}
	igName := strings.TrimSuffix(vmssName, "."+a.clusterName)

	vm, err := a.client.vmsClient.Get(ctx, a.client.resourceGroup, vmssName, vmssIndex, "")
	if err != nil {
		return nil, fmt.Errorf("getting info for VMSS virtual machine %q #%s: %w", vmssName, vmssIndex, err)
	}
	if vm.VMID == nil {
		return nil, fmt.Errorf("determining VMID for VMSS %q virtual machine #%s", vmssName, vmssIndex)
	}
	if vmId != *vm.VMID {
		return nil, fmt.Errorf("matching VMID %q to VMSS %q virtual machine #%s", vmId, vmssName, vmssIndex)
	}
	if vm.OsProfile == nil || vm.OsProfile.ComputerName == nil || *vm.OsProfile.ComputerName == "" {
		return nil, fmt.Errorf("determining ComputerName for VMSS %q virtual machine #%s", vmssName, vmssIndex)
	}
	nodeName := *vm.OsProfile.ComputerName

	ni, err := a.client.nisClient.GetVirtualMachineScaleSetNetworkInterface(ctx, a.client.resourceGroup, vmssName, vmssIndex, vmssName+"-netconfig", "")
	if err != nil {
		return nil, fmt.Errorf("getting info for VMSS network interface %q #%s: %w", vmssName, vmssIndex, err)
	}

	var addrs []string
	var challengeEndpoints []string
	for _, ipc := range *ni.IPConfigurations {
		if ipc.PrivateIPAddress != nil {
			addrs = append(addrs, *ipc.PrivateIPAddress)
			challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(*ipc.PrivateIPAddress, strconv.Itoa(wellknownports.NodeupChallenge)))
		}
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("determining challenge endpoint for VMSS %q virtual machine #%s", vmssName, vmssIndex)
	}
	if len(challengeEndpoints) == 0 {
		return nil, fmt.Errorf("determining challenge endpoint for VMSS %q virtual machine #%s", vmssName, vmssIndex)
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
	vmsClient     *compute.VirtualMachineScaleSetVMsClient
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

	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("creating authorizer: %w", err)
	}

	nisClient := network.NewInterfacesClient(m.Compute.SubscriptionID)
	nisClient.Authorizer = authorizer
	vmsClient := compute.NewVirtualMachineScaleSetVMsClient(m.Compute.SubscriptionID)
	vmsClient.Authorizer = authorizer

	return &client{
		resourceGroup: m.Compute.ResourceGroupName,
		nisClient:     &nisClient,
		vmsClient:     &vmsClient,
	}, nil
}

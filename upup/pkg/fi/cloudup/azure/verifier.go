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
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/wellknownports"
)

const (
	// InstanceGroupNameTag is the key of the tag used to identify an instance group that VM belongs to.
	InstanceGroupNameTag = "kops.k8s.io_instancegroup"
)

// AzureVerifierOptions configures the Azure bootstrap token verifier.
type AzureVerifierOptions struct {
	ClusterName string `json:"clusterName,omitempty"`
}

type azureVerifier struct {
	client      *client
	clusterName string
}

var _ bootstrap.Verifier = (*azureVerifier)(nil)

// NewAzureVerifier returns a verifier that validates Azure IMDS attestation
// tokens and resolves the claimed VM identity through the Azure API.
func NewAzureVerifier(ctx context.Context, opt *AzureVerifierOptions) (bootstrap.Verifier, error) {
	azureClient, err := newVerifierClient()
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

// vmLogIDFromResource returns a short human-readable identifier for a VM
// resource, used in log messages throughout the verifier.
func vmLogIDFromResource(res *arm.ResourceID) string {
	if res == nil {
		return "<nil>"
	}

	switch res.ResourceType.String() {
	case "Microsoft.Compute/virtualMachines":
		return res.Name
	case "Microsoft.Compute/virtualMachineScaleSets/virtualMachines":
		return res.Parent.Name + "/" + res.Name
	default:
		return res.ResourceType.String() + "/" + res.Name
	}
}

// VerifyToken validates the Azure attestation token, confirms the claimed VM
// through the Azure API, and returns the node bootstrap identity.
func (a azureVerifier) VerifyToken(ctx context.Context, rawRequest *http.Request, token string, body []byte) (*bootstrap.VerifyResult, error) {
	if !strings.HasPrefix(token, AzureAuthenticationTokenPrefix) {
		return nil, bootstrap.ErrNotThisVerifier
	}

	// Token format: "x-azure-id <resourceID> <base64-pkcs7-signature>"
	tokenPayload := strings.TrimPrefix(token, AzureAuthenticationTokenPrefix)
	resourceID, signature, ok := strings.Cut(tokenPayload, " ")
	if !ok || resourceID == "" || signature == "" {
		return nil, fmt.Errorf("incorrect token format")
	}

	// Parse the resource ID early to reject malformed tokens before expensive crypto.
	res, err := arm.ParseResourceID(resourceID)
	if err != nil {
		return nil, fmt.Errorf("parsing resource ID: %w", err)
	}
	vmLogID := vmLogIDFromResource(res)
	resourceType := res.ResourceType.String()
	klog.V(4).Infof("Azure verifier for VM %q parsed resource ID: subscription=%q resourceGroup=%q", vmLogID, res.SubscriptionID, res.ResourceGroupName)

	// Reject resource IDs outside the verifier's own subscription / resource
	// group. The Azure API lookup below is already scoped to kops-controller's
	// subscription and resource group, so any claim that names a different
	// location cannot describe a cluster VM. Failing here avoids a wasted
	// Azure API call and makes the scope explicit instead of implicit.
	if !strings.EqualFold(res.SubscriptionID, a.client.subscriptionID) {
		return nil, fmt.Errorf("resource ID subscription %q does not match verifier subscription %q", res.SubscriptionID, a.client.subscriptionID)
	}
	if !strings.EqualFold(res.ResourceGroupName, a.client.resourceGroup) {
		return nil, fmt.Errorf("resource ID resource group %q does not match verifier resource group %q", res.ResourceGroupName, a.client.resourceGroup)
	}
	switch resourceType {
	case "Microsoft.Compute/virtualMachines":
	case "Microsoft.Compute/virtualMachineScaleSets/virtualMachines":
		if !strings.HasSuffix(res.Parent.Name, "."+a.clusterName) {
			return nil, fmt.Errorf("resource ID VMSS name %q does not match cluster name %q", res.Parent.Name, a.clusterName)
		}
	default:
		return nil, fmt.Errorf("unsupported resource type %q", resourceType)
	}

	// Verify the PKCS7 attested document: signature, certificate chain, nonce, and expiration.
	data, err := verifyAttestedDocument(signature, body)
	if err != nil {
		return nil, err
	}
	klog.V(2).Infof("Azure verifier for VM %q verified attested document", vmLogID)
	if !strings.EqualFold(data.SubscriptionId, a.client.subscriptionID) {
		return nil, fmt.Errorf("attested subscriptionId %q does not match verifier subscription %q", data.SubscriptionId, a.client.subscriptionID)
	}

	// Look up the VM or VMSS VM via the Azure API using the resource ID,
	// cross-verify the attested vmId, and extract node identity.
	var nodeName, igName string
	var addrs, challengeEndpoints []string

	switch resourceType {
	case "Microsoft.Compute/virtualMachines":
		vmName := res.Name
		klog.V(2).Infof("Azure verifier for VM %q looking up Azure API object", vmLogID)

		// Fetch the VM from the Azure API.
		vm, err := a.client.vmsClient.Get(ctx, a.client.resourceGroup, vmName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VM %q: %w", vmName, err)
		}
		if vm.Properties == nil || vm.Properties.VMID == nil {
			return nil, fmt.Errorf("determining VMID for VM %q", vmName)
		}

		// Cross-verify: the vmId from the cryptographically signed attested document
		// must match the vmId from the Azure API for the claimed resource ID.
		klog.V(4).Infof("Azure verifier for VM %q cross-verifying vmId: attested=%q api=%q", vmLogID, data.VMId, *vm.Properties.VMID)
		if data.VMId != *vm.Properties.VMID {
			return nil, fmt.Errorf("attested vmId %q does not match VM %q (API vmId %q)", data.VMId, vmName, *vm.Properties.VMID)
		}
		if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil || *vm.Properties.OSProfile.ComputerName == "" {
			return nil, fmt.Errorf("determining ComputerName for VM %q", vmName)
		}

		// Extract node name and instance group from VM metadata.
		nodeName = strings.ToLower(*vm.Properties.OSProfile.ComputerName)
		if igNameTag, ok := vm.Tags[InstanceGroupNameTag]; ok && igNameTag != nil {
			igName = *igNameTag
		} else {
			return nil, fmt.Errorf("determining IG name for VM %q", vmName)
		}
		klog.V(4).Infof("Azure verifier for VM %q resolved identity: node=%q instanceGroup=%q", vmLogID, nodeName, igName)

		// Collect private IP addresses from the VM's network interface.
		ni, err := a.client.nisClient.Get(ctx, a.client.resourceGroup, nodeName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VM network interface %q: %w", nodeName, err)
		}
		if ni.Properties == nil {
			return nil, fmt.Errorf("determining IP configurations for VM network interface %q", nodeName)
		}

		for _, ipc := range ni.Properties.IPConfigurations {
			addrs, challengeEndpoints = appendPrivateIP(addrs, challengeEndpoints, ipc)
		}

	case "Microsoft.Compute/virtualMachineScaleSets/virtualMachines":
		vmssName := res.Parent.Name
		vmssIndex := res.Name
		klog.V(2).Infof("Azure verifier for VM %q looking up Azure API object", vmLogID)

		// Fetch the VMSS VM from the Azure API.
		vm, err := a.client.vmssVMsClient.Get(ctx, a.client.resourceGroup, vmssName, vmssIndex, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VMSS %q VM #%s: %w", vmssName, vmssIndex, err)
		}
		if vm.Properties == nil || vm.Properties.VMID == nil {
			return nil, fmt.Errorf("determining VMID for VMSS %q VM #%s", vmssName, vmssIndex)
		}

		// Cross-verify: the vmId from the cryptographically signed attested document
		// must match the vmId from the Azure API for the claimed resource ID.
		klog.V(4).Infof("Azure verifier for VM %q cross-verifying vmId: attested=%q api=%q", vmLogID, data.VMId, *vm.Properties.VMID)
		if data.VMId != *vm.Properties.VMID {
			return nil, fmt.Errorf("attested vmId %q does not match VMSS %q VM #%s (API vmId %q)", data.VMId, vmssName, vmssIndex, *vm.Properties.VMID)
		}
		if vm.Properties.OSProfile == nil || vm.Properties.OSProfile.ComputerName == nil || *vm.Properties.OSProfile.ComputerName == "" {
			return nil, fmt.Errorf("determining ComputerName for VMSS %q VM #%s", vmssName, vmssIndex)
		}

		// Extract node name and instance group from VMSS VM metadata.
		nodeName = strings.ToLower(*vm.Properties.OSProfile.ComputerName)
		if igNameTag, ok := vm.Tags[InstanceGroupNameTag]; ok && igNameTag != nil {
			igName = *igNameTag
		} else {
			return nil, fmt.Errorf("determining IG name for VMSS %q VM #%s", vmssName, vmssIndex)
		}
		klog.V(4).Infof("Azure verifier for VM %q resolved identity: node=%q instanceGroup=%q", vmLogID, nodeName, igName)

		// Collect private IP addresses from the VMSS VM's network interface.
		ni, err := a.client.nisClient.GetVirtualMachineScaleSetNetworkInterface(ctx, a.client.resourceGroup, vmssName, vmssIndex, vmssName, nil)
		if err != nil {
			return nil, fmt.Errorf("getting info for VMSS %q VM #%s network interface: %w", vmssName, vmssIndex, err)
		}
		if ni.Properties == nil {
			return nil, fmt.Errorf("determining IP configurations for VMSS %q VM #%s network interface", vmssName, vmssIndex)
		}

		for _, ipc := range ni.Properties.IPConfigurations {
			addrs, challengeEndpoints = appendPrivateIP(addrs, challengeEndpoints, ipc)
		}

	default:
		return nil, fmt.Errorf("unsupported resource type %q", resourceType)
	}

	// Validate that we found at least one address and challenge endpoint.
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

	klog.V(2).Infof("Azure verifier for VM %q verified as node %q in instance group %q", vmLogID, nodeName, igName)
	return result, nil
}

// appendPrivateIP appends the private IP address from an interface IP
// configuration to the address and challenge endpoint slices.
func appendPrivateIP(addrs, challengeEndpoints []string, ipc *network.InterfaceIPConfiguration) ([]string, []string) {
	if ipc.Properties != nil && ipc.Properties.PrivateIPAddress != nil {
		addrs = append(addrs, *ipc.Properties.PrivateIPAddress)
		challengeEndpoints = append(challengeEndpoints, net.JoinHostPort(*ipc.Properties.PrivateIPAddress, strconv.Itoa(wellknownports.NodeupChallenge)))
	}
	return addrs, challengeEndpoints
}

// client is an Azure client.
type client struct {
	subscriptionID string
	resourceGroup  string
	nisClient      *network.InterfacesClient
	vmsClient      *compute.VirtualMachinesClient
	vmssVMsClient  *compute.VirtualMachineScaleSetVMsClient
}

// newVerifierClient builds Azure API clients scoped to the local instance's
// subscription and resource group from IMDS metadata.
func newVerifierClient() (*client, error) {
	metadata, err := QueryComputeInstanceMetadata()
	if err != nil || metadata == nil {
		return nil, fmt.Errorf("getting instance metadata: %w", err)
	}
	if metadata.ResourceGroupName == "" {
		return nil, fmt.Errorf("empty resource group name")
	}
	if metadata.SubscriptionID == "" {
		return nil, fmt.Errorf("empty subscription ID")
	}
	klog.V(4).Infof("Azure verifier client using subscription %q resource group %q", metadata.SubscriptionID, metadata.ResourceGroupName)

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
		subscriptionID: metadata.SubscriptionID,
		resourceGroup:  metadata.ResourceGroupName,
		nisClient:      nisClient,
		vmsClient:      vmsClient,
		vmssVMsClient:  vmssVMsClient,
	}, nil
}

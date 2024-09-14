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

package protokube

import (
	"context"
	"fmt"

	compute "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	network "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipazure "k8s.io/kops/protokube/pkg/gossip/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
)

type client interface {
	ListVMScaleSets(ctx context.Context) ([]*compute.VirtualMachineScaleSet, error)
	ListVMSSNetworkInterfaces(ctx context.Context, vmScaleSetName string) ([]*network.Interface, error)
	GetName() string
	GetTags() (map[string]string, error)
}

var _ client = &gossipazure.Client{}

// AzureCloudProvider implements the CloudProvider interface for Azure.
type AzureCloudProvider struct {
	client client

	clusterTag string
	instanceID string
}

var _ CloudProvider = &AzureCloudProvider{}

// NewAzureCloudProvider returns a new AzureCloudProvider.
func NewAzureCloudProvider() (*AzureCloudProvider, error) {
	client, err := gossipazure.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error creating a new Azure client: %s", err)
	}

	tags, err := client.GetTags()
	if err != nil {
		return nil, fmt.Errorf("error querying tags: %s", err)
	}
	clusterTag := tags[azure.TagClusterName]
	if clusterTag == "" {
		return nil, fmt.Errorf("cluster tag %q not found", azure.TagClusterName)
	}
	instanceID := client.GetName()
	if instanceID == "" {
		return nil, fmt.Errorf("empty name")
	}
	return &AzureCloudProvider{
		client:     client,
		clusterTag: clusterTag,
		instanceID: instanceID,
	}, nil
}

// InstanceID implements CloudProvider InstanceID.
func (a *AzureCloudProvider) InstanceID() string {
	return a.instanceID
}

// GossipSeeds implements CloudProvider GossipSeeds.
func (a *AzureCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	tags := map[string]string{
		azure.TagClusterName: a.clusterTag,
	}
	return gossipazure.NewSeedProvider(a.client, tags)
}

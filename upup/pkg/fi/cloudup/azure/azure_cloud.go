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
	"errors"

	"github.com/Azure/go-autorest/autorest/azure/auth"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	TagClusterName = "KubernetesCluster"
	// Use dash (_) as a splitter. Other CSPs use slash (/), but slash is not
	// allowed as a tag key in Azure.
	TagNameRolePrefix        = "k8s.io_role_"
	TagRoleMaster            = "master"
	TagNameEtcdClusterPrefix = "k8s.io_etcd_"
)

// AzureCloud provides clients to make API calls to Azure.
type AzureCloud interface {
	fi.Cloud

	AddClusterTags(tags map[string]*string)
	FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error)
	GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error)

	SubscriptionID() string
	ResourceGroup() ResourceGroupsClient
	VirtualNetwork() VirtualNetworksClient
	Subnet() SubnetsClient
	RouteTable() RouteTablesClient
	VMScaleSet() VMScaleSetsClient
	VMScaleSetVM() VMScaleSetVMsClient
	Disk() DisksClient
	RoleAssignment() RoleAssignmentsClient
}

type azureCloudImplementation struct {
	subscriptionID        string
	location              string
	tags                  map[string]string
	resourceGroupsClient  ResourceGroupsClient
	vnetsClient           VirtualNetworksClient
	subnetsClient         SubnetsClient
	routeTablesClient     RouteTablesClient
	vmscaleSetsClient     VMScaleSetsClient
	vmscaleSetVMsClient   VMScaleSetVMsClient
	disksClient           DisksClient
	roleAssignmentsClient RoleAssignmentsClient
}

var _ fi.Cloud = &azureCloudImplementation{}

// NewAzureCloud creates a new AzureCloud.
func NewAzureCloud(subscriptionID, location string, tags map[string]string) (AzureCloud, error) {
	authorizer, err := auth.NewAuthorizerFromEnvironment()
	if err != nil {
		return nil, err
	}

	return &azureCloudImplementation{
		subscriptionID:        subscriptionID,
		location:              location,
		tags:                  tags,
		resourceGroupsClient:  newResourceGroupsClientImpl(subscriptionID, authorizer),
		vnetsClient:           newVirtualNetworksClientImpl(subscriptionID, authorizer),
		subnetsClient:         newSubnetsClientImpl(subscriptionID, authorizer),
		routeTablesClient:     newRouteTablesClientImpl(subscriptionID, authorizer),
		vmscaleSetsClient:     newVMScaleSetsClientImpl(subscriptionID, authorizer),
		vmscaleSetVMsClient:   newVMScaleSetVMsClientImpl(subscriptionID, authorizer),
		disksClient:           newDisksClientImpl(subscriptionID, authorizer),
		roleAssignmentsClient: newRoleAssignmentsClientImpl(subscriptionID, authorizer),
	}, nil
}

func (c *azureCloudImplementation) Region() string {
	return c.location
}

func (c *azureCloudImplementation) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderAzure
}

func (c *azureCloudImplementation) DNS() (dnsprovider.Interface, error) {
	return nil, errors.New("DNS not implemented on azureCloud")
}

func (c *azureCloudImplementation) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, errors.New("FindVPCInfo not implemented on azureCloud")
}

func (c *azureCloudImplementation) DeleteInstance(i *cloudinstances.CloudInstance) error {
	return errors.New("DeleteInstance not implemented on azureCloud")
}

func (c *azureCloudImplementation) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return errors.New("DeleteGroup not implemented on azureCloud")
}

func (c *azureCloudImplementation) DetachInstance(i *cloudinstances.CloudInstance) error {
	return errors.New("DetachInstance not implemented on azureCloud")
}

// AddClusterTags adds cluster tags to the resource.
func (c *azureCloudImplementation) AddClusterTags(tags map[string]*string) {
	for k, v := range c.tags {
		tags[k] = &v
	}
}

func (c *azureCloudImplementation) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	// TODO(kenji): Implement this. Currently we return nil as we
	// don't create any resources for ingress to the API server.
	return nil, nil
}

func (c *azureCloudImplementation) SubscriptionID() string {
	return c.subscriptionID
}

func (c *azureCloudImplementation) ResourceGroup() ResourceGroupsClient {
	return c.resourceGroupsClient
}

func (c *azureCloudImplementation) VirtualNetwork() VirtualNetworksClient {
	return c.vnetsClient
}

func (c *azureCloudImplementation) Subnet() SubnetsClient {
	return c.subnetsClient
}

func (c *azureCloudImplementation) RouteTable() RouteTablesClient {
	return c.routeTablesClient
}

func (c *azureCloudImplementation) VMScaleSet() VMScaleSetsClient {
	return c.vmscaleSetsClient
}

func (c *azureCloudImplementation) VMScaleSetVM() VMScaleSetVMsClient {
	return c.vmscaleSetVMsClient
}

func (c *azureCloudImplementation) Disk() DisksClient {
	return c.disksClient
}

func (c *azureCloudImplementation) RoleAssignment() RoleAssignmentsClient {
	return c.roleAssignmentsClient
}

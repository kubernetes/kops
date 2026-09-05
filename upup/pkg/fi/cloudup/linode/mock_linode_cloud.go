/*
Copyright 2026 The Kubernetes Authors.

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

package linode

import (
	"context"

	"github.com/linode/linodego/v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

type MockLinodeCloud struct {
	Region_ string
	Client_ LinodeClient
}

var _ LinodeCloud = &MockLinodeCloud{}

func (c *MockLinodeCloud) Client() LinodeClient {
	return c.Client_
}

func (c *MockLinodeCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderLinode
}

func (c *MockLinodeCloud) DNS() (dnsprovider.Interface, error) {
	return nil, nil
}

func (c *MockLinodeCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, nil
}

func (c *MockLinodeCloud) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (c *MockLinodeCloud) DeregisterInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (c *MockLinodeCloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	return nil
}

func (c *MockLinodeCloud) DetachInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (c *MockLinodeCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, nil
}

func (c *MockLinodeCloud) Region() string {
	return c.Region_
}

func (c *MockLinodeCloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return &kops.ClusterStatus{}, nil
}

func (c *MockLinodeCloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, nil
}

type MockLinodeClient struct {
	ListVPCsResponse []linodego.VPC
	ListVPCsError    error
	ListVPCsCalls    int
	LastListVPCsOpts *linodego.ListOptions

	CreateVPCResponse *linodego.VPC
	CreateVPCError    error
	CreateVPCCalls    int
	LastCreateVPCOpts linodego.VPCCreateOptions

	UpdateVPCResponse *linodego.VPC
	UpdateVPCError    error
	UpdateVPCCalls    int
	UpdatedVPCIDs     []int
	LastUpdateVPCOpts linodego.VPCUpdateOptions

	DeleteVPCError error
	DeleteVPCCalls int
	DeletedVPCIDs  []int

	ListSSHKeysResponse []linodego.SSHKey
	ListSSHKeysError    error
	ListSSHKeysCalls    int

	CreateSSHKeyResponse *linodego.SSHKey
	CreateSSHKeyError    error
	CreateSSHKeyCalls    int
	LastCreateSSHKeyOpts linodego.SSHKeyCreateOptions

	DeleteSSHKeyError error
	DeleteSSHKeyCalls int
	DeletedSSHKeyIDs  []int

	ListVPCSubnetsResponse  []linodego.VPCSubnet
	ListVPCSubnetsResponses map[int][]linodego.VPCSubnet
	ListVPCSubnetsError     error
	ListVPCSubnetsCalls     int
	LastListVPCSubnetsOpts  *linodego.ListOptions
	LastListVPCSubnetsVPCID int

	CreateVPCSubnetResponse  *linodego.VPCSubnet
	CreateVPCSubnetError     error
	CreateVPCSubnetCalls     int
	LastCreateVPCSubnetOpts  linodego.VPCSubnetCreateOptions
	LastCreateVPCSubnetVPCID int

	UpdateVPCSubnetResponse  *linodego.VPCSubnet
	UpdateVPCSubnetError     error
	UpdateVPCSubnetCalls     int
	LastUpdateVPCSubnetOpts  linodego.VPCSubnetUpdateOptions
	LastUpdateVPCSubnetVPCID int
	LastUpdateVPCSubnetID    int

	DeleteVPCSubnetError   error
	DeleteVPCSubnetCalls   int
	DeletedVPCSubnetVPCIDs []int
	DeletedVPCSubnetIDs    []int

	ListFirewallsResponse []linodego.Firewall
	ListFirewallsError    error
	ListFirewallsCalls    int
	LastListFirewallsOpts *linodego.ListOptions

	CreateFirewallResponse *linodego.Firewall
	CreateFirewallError    error
	CreateFirewallCalls    int
	LastCreateFirewallOpts linodego.FirewallCreateOptions

	UpdateFirewallResponse *linodego.Firewall
	UpdateFirewallError    error
	UpdateFirewallCalls    int
	UpdatedFirewallIDs     []int
	LastUpdateFirewallOpts linodego.FirewallUpdateOptions

	UpdateFirewallRulesResponse *linodego.FirewallRules
	UpdateFirewallRulesError    error
	UpdateFirewallRulesCalls    int
	UpdatedFirewallRulesIDs     []int
	LastUpdateFirewallRulesOpts linodego.FirewallRulesUpdateOptions

	CreateInstanceResponse *linodego.Instance
	CreateInstanceError    error
	CreateInstanceCalls    int
	LastCreateInstanceOpts linodego.InstanceCreateOptions

	UpdateInstanceResponse *linodego.Instance
	UpdateInstanceError    error
	UpdateInstanceCalls    int
	LastUpdateInstanceOpts linodego.InstanceUpdateOptions
	LastUpdateInstanceID   int

	DeleteInstanceError error
	DeleteInstanceCalls int
	DeletedInstanceIDs  []int

	ListInstancesResponse []linodego.Instance
	ListInstancesError    error
	ListInstancesCalls    int
	LastListInstancesOpts *linodego.ListOptions

	ListInterfacesResponse  []linodego.LinodeInterface
	ListInterfacesResponses map[int][]linodego.LinodeInterface
	ListInterfacesError     error
	ListInterfacesCalls     int
	LastListInterfacesOpts  *linodego.ListOptions
	LastListInterfacesID    int

	ListVolumesResponse []linodego.Volume
	ListVolumesError    error
	ListVolumesCalls    int
	LastListVolumesOpts *linodego.ListOptions

	CreateVolumeResponse *linodego.Volume
	CreateVolumeError    error
	CreateVolumeCalls    int
	LastCreateVolumeOpts linodego.VolumeCreateOptions

	DeleteVolumeError error
	DeleteVolumeCalls int
	DeletedVolumeIDs  []int

	ResizeVolumeError    error
	ResizeVolumeCalls    int
	ResizedVolumeIDs     []int
	LastResizeVolumeOpts linodego.VolumeResizeOptions
}

var _ LinodeClient = &MockLinodeClient{}

func (c *MockLinodeClient) ListVPCs(ctx context.Context, opts *linodego.ListOptions) ([]linodego.VPC, error) {
	c.ListVPCsCalls++
	c.LastListVPCsOpts = opts
	if c.ListVPCsError != nil {
		return nil, c.ListVPCsError
	}
	return c.ListVPCsResponse, nil
}

func (c *MockLinodeClient) CreateVPC(ctx context.Context, opts linodego.VPCCreateOptions) (*linodego.VPC, error) {
	c.CreateVPCCalls++
	c.LastCreateVPCOpts = opts
	if c.CreateVPCError != nil {
		return nil, c.CreateVPCError
	}
	if c.CreateVPCResponse == nil {
		return &linodego.VPC{}, nil
	}
	return c.CreateVPCResponse, nil
}

func (c *MockLinodeClient) UpdateVPC(ctx context.Context, vpcID int, opts linodego.VPCUpdateOptions) (*linodego.VPC, error) {
	c.UpdateVPCCalls++
	c.UpdatedVPCIDs = append(c.UpdatedVPCIDs, vpcID)
	c.LastUpdateVPCOpts = opts
	if c.UpdateVPCError != nil {
		return nil, c.UpdateVPCError
	}
	if c.UpdateVPCResponse == nil {
		return &linodego.VPC{}, nil
	}
	return c.UpdateVPCResponse, nil
}

func (c *MockLinodeClient) DeleteVPC(ctx context.Context, vpcID int) error {
	c.DeleteVPCCalls++
	c.DeletedVPCIDs = append(c.DeletedVPCIDs, vpcID)
	return c.DeleteVPCError
}

func (c *MockLinodeClient) ListSSHKeys(ctx context.Context, opts *linodego.ListOptions) ([]linodego.SSHKey, error) {
	c.ListSSHKeysCalls++
	if c.ListSSHKeysError != nil {
		return nil, c.ListSSHKeysError
	}
	return c.ListSSHKeysResponse, nil
}

func (c *MockLinodeClient) CreateSSHKey(ctx context.Context, opts linodego.SSHKeyCreateOptions) (*linodego.SSHKey, error) {
	c.CreateSSHKeyCalls++
	c.LastCreateSSHKeyOpts = opts
	if c.CreateSSHKeyError != nil {
		return nil, c.CreateSSHKeyError
	}
	if c.CreateSSHKeyResponse == nil {
		return &linodego.SSHKey{}, nil
	}
	return c.CreateSSHKeyResponse, nil
}

func (c *MockLinodeClient) DeleteSSHKey(ctx context.Context, sshKeyID int) error {
	c.DeleteSSHKeyCalls++
	c.DeletedSSHKeyIDs = append(c.DeletedSSHKeyIDs, sshKeyID)
	return c.DeleteSSHKeyError
}

func (c *MockLinodeClient) ListVPCSubnets(ctx context.Context, vpcID int, opts *linodego.ListOptions) ([]linodego.VPCSubnet, error) {
	c.ListVPCSubnetsCalls++
	c.LastListVPCSubnetsOpts = opts
	c.LastListVPCSubnetsVPCID = vpcID
	if c.ListVPCSubnetsError != nil {
		return nil, c.ListVPCSubnetsError
	}
	if c.ListVPCSubnetsResponses != nil {
		return c.ListVPCSubnetsResponses[vpcID], nil
	}
	return c.ListVPCSubnetsResponse, nil
}

func (c *MockLinodeClient) CreateVPCSubnet(ctx context.Context, opts linodego.VPCSubnetCreateOptions, vpcID int) (*linodego.VPCSubnet, error) {
	c.CreateVPCSubnetCalls++
	c.LastCreateVPCSubnetOpts = opts
	c.LastCreateVPCSubnetVPCID = vpcID
	if c.CreateVPCSubnetError != nil {
		return nil, c.CreateVPCSubnetError
	}
	if c.CreateVPCSubnetResponse == nil {
		return &linodego.VPCSubnet{}, nil
	}
	return c.CreateVPCSubnetResponse, nil
}

func (c *MockLinodeClient) UpdateVPCSubnet(ctx context.Context, vpcID int, subnetID int, opts linodego.VPCSubnetUpdateOptions) (*linodego.VPCSubnet, error) {
	c.UpdateVPCSubnetCalls++
	c.LastUpdateVPCSubnetOpts = opts
	c.LastUpdateVPCSubnetVPCID = vpcID
	c.LastUpdateVPCSubnetID = subnetID
	if c.UpdateVPCSubnetError != nil {
		return nil, c.UpdateVPCSubnetError
	}
	if c.UpdateVPCSubnetResponse == nil {
		return &linodego.VPCSubnet{}, nil
	}
	return c.UpdateVPCSubnetResponse, nil
}

func (c *MockLinodeClient) DeleteVPCSubnet(ctx context.Context, vpcID int, subnetID int) error {
	c.DeleteVPCSubnetCalls++
	c.DeletedVPCSubnetIDs = append(c.DeletedVPCSubnetIDs, subnetID)
	c.DeletedVPCSubnetVPCIDs = append(c.DeletedVPCSubnetVPCIDs, vpcID)
	return c.DeleteVPCSubnetError
}

func (c *MockLinodeClient) ListFirewalls(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Firewall, error) {
	c.ListFirewallsCalls++
	c.LastListFirewallsOpts = opts
	if c.ListFirewallsError != nil {
		return nil, c.ListFirewallsError
	}
	return c.ListFirewallsResponse, nil
}

func (c *MockLinodeClient) CreateFirewall(ctx context.Context, opts linodego.FirewallCreateOptions) (*linodego.Firewall, error) {
	c.CreateFirewallCalls++
	c.LastCreateFirewallOpts = opts
	if c.CreateFirewallError != nil {
		return nil, c.CreateFirewallError
	}
	if c.CreateFirewallResponse == nil {
		return &linodego.Firewall{}, nil
	}
	return c.CreateFirewallResponse, nil
}

func (c *MockLinodeClient) UpdateFirewall(ctx context.Context, firewallID int, opts linodego.FirewallUpdateOptions) (*linodego.Firewall, error) {
	c.UpdateFirewallCalls++
	c.UpdatedFirewallIDs = append(c.UpdatedFirewallIDs, firewallID)
	c.LastUpdateFirewallOpts = opts
	if c.UpdateFirewallError != nil {
		return nil, c.UpdateFirewallError
	}
	if c.UpdateFirewallResponse == nil {
		return &linodego.Firewall{}, nil
	}
	return c.UpdateFirewallResponse, nil
}

func (c *MockLinodeClient) UpdateFirewallRules(ctx context.Context, firewallID int, opts linodego.FirewallRulesUpdateOptions) (*linodego.FirewallRules, error) {
	c.UpdateFirewallRulesCalls++
	c.UpdatedFirewallRulesIDs = append(c.UpdatedFirewallRulesIDs, firewallID)
	c.LastUpdateFirewallRulesOpts = opts
	if c.UpdateFirewallRulesError != nil {
		return nil, c.UpdateFirewallRulesError
	}
	if c.UpdateFirewallRulesResponse == nil {
		return &linodego.FirewallRules{}, nil
	}
	return c.UpdateFirewallRulesResponse, nil
}

func (c *MockLinodeClient) ListInstances(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Instance, error) {
	c.ListInstancesCalls++
	c.LastListInstancesOpts = opts
	if c.ListInstancesError != nil {
		return nil, c.ListInstancesError
	}
	return c.ListInstancesResponse, nil
}

func (c *MockLinodeClient) CreateInstance(ctx context.Context, opts linodego.InstanceCreateOptions) (*linodego.Instance, error) {
	c.CreateInstanceCalls++
	c.LastCreateInstanceOpts = opts
	if c.CreateInstanceError != nil {
		return nil, c.CreateInstanceError
	}
	if c.CreateInstanceResponse == nil {
		return &linodego.Instance{}, nil
	}
	return c.CreateInstanceResponse, nil
}

func (c *MockLinodeClient) UpdateInstance(ctx context.Context, instanceID int, opts linodego.InstanceUpdateOptions) (*linodego.Instance, error) {
	c.UpdateInstanceCalls++
	c.LastUpdateInstanceOpts = opts
	c.LastUpdateInstanceID = instanceID
	if c.UpdateInstanceError != nil {
		return nil, c.UpdateInstanceError
	}
	if c.UpdateInstanceResponse == nil {
		return &linodego.Instance{}, nil
	}
	return c.UpdateInstanceResponse, nil
}

func (c *MockLinodeClient) DeleteInstance(ctx context.Context, instanceID int) error {
	c.DeleteInstanceCalls++
	c.DeletedInstanceIDs = append(c.DeletedInstanceIDs, instanceID)
	return c.DeleteInstanceError
}

func (c *MockLinodeClient) ListInterfaces(ctx context.Context, instanceID int, opts *linodego.ListOptions) ([]linodego.LinodeInterface, error) {
	c.ListInterfacesCalls++
	c.LastListInterfacesOpts = opts
	c.LastListInterfacesID = instanceID
	if c.ListInterfacesError != nil {
		return nil, c.ListInterfacesError
	}
	if c.ListInterfacesResponses != nil {
		return c.ListInterfacesResponses[instanceID], nil
	}
	return c.ListInterfacesResponse, nil
}

func (c *MockLinodeClient) CreateVolume(ctx context.Context, opts linodego.VolumeCreateOptions) (*linodego.Volume, error) {
	c.CreateVolumeCalls++
	c.LastCreateVolumeOpts = opts
	if c.CreateVolumeError != nil {
		return nil, c.CreateVolumeError
	}
	if c.CreateVolumeResponse == nil {
		return &linodego.Volume{}, nil
	}
	return c.CreateVolumeResponse, nil
}

func (c *MockLinodeClient) ListVolumes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Volume, error) {
	c.ListVolumesCalls++
	c.LastListVolumesOpts = opts
	if c.ListVolumesError != nil {
		return nil, c.ListVolumesError
	}
	return c.ListVolumesResponse, nil
}

func (c *MockLinodeClient) DeleteVolume(ctx context.Context, volumeID int) error {
	c.DeleteVolumeCalls++
	c.DeletedVolumeIDs = append(c.DeletedVolumeIDs, volumeID)
	return c.DeleteVolumeError
}

func (c *MockLinodeClient) ResizeVolume(ctx context.Context, volumeID int, opts linodego.VolumeResizeOptions) error {
	c.ResizeVolumeCalls++
	c.ResizedVolumeIDs = append(c.ResizedVolumeIDs, volumeID)
	c.LastResizeVolumeOpts = opts
	return c.ResizeVolumeError
}

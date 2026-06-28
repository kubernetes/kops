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

	"github.com/linode/linodego"
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

	CreateSSHKeyResponse *linodego.SSHKey
	CreateSSHKeyError    error
	CreateSSHKeyCalls    int
	LastCreateSSHKeyOpts linodego.SSHKeyCreateOptions

	DeleteSSHKeyError error
	DeleteSSHKeyCalls int
	DeletedSSHKeyIDs  []int
}

var _ LinodeClient = &MockLinodeClient{}

func (c *MockLinodeClient) ListVPCs(ctx context.Context, opts *linodego.ListOptions) ([]linodego.VPC, error) {
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

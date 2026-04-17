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

// MockLinodeClient implements LinodeClient for use in tests.
type MockLinodeClient struct {
	// SSH keys
	ListSSHKeysResponse  []linodego.SSHKey
	ListSSHKeysError     error
	CreateSSHKeyResponse *linodego.SSHKey
	CreateSSHKeyError    error
	CreateSSHKeyCalls    int
	LastCreateSSHKeyOpts linodego.SSHKeyCreateOptions

	// Instances
	ListInstancesResponse  []linodego.Instance
	ListInstancesError     error
	CreateInstanceResponse *linodego.Instance
	CreateInstanceError    error
	CreateInstanceCalls    int
	LastCreateInstanceOpts linodego.InstanceCreateOptions
	DeleteInstanceErr      error
	DeleteInstanceErrByID  map[int]error
	DeletedInstanceIDs     []int
	UpdateInstanceErr      error
	UpdateInstanceErrByID  map[int]error
	UpdatedInstanceIDs     []int
	UpdatedTagsByID        map[int][]string

	// Volumes
	ListVolumesResponse  []linodego.Volume
	ListVolumesError     error
	CreateVolumeResponse *linodego.Volume
	CreateVolumeError    error
	CreateVolumeCalls    int
	LastCreateVolumeOpts linodego.VolumeCreateOptions

	// NodeBalancers
	ListNodeBalancersResponse  []linodego.NodeBalancer
	ListNodeBalancersError     error
	CreateNodeBalancerResponse *linodego.NodeBalancer
	CreateNodeBalancerError    error
	CreateNodeBalancerCalls    int
	LastCreateNodeBalancerOpts linodego.NodeBalancerCreateOptions

	// NodeBalancer configs
	ListNodeBalancerConfigsResponse []linodego.NodeBalancerConfig
	ListNodeBalancerConfigsError    error
	CreateNodeBalancerConfigCalls   int
	CreateNodeBalancerConfigOpts    []linodego.NodeBalancerConfigCreateOptions
	RebuildNodeBalancerConfigCalls  int
	RebuildNodeBalancerConfigOpts   []linodego.NodeBalancerConfigRebuildOptions

	// NodeBalancer nodes
	ListNodeBalancerNodesResponse map[int][]linodego.NodeBalancerNode
	CreateNodeBalancerNodeCalls   int
	CreateNodeBalancerNodeOpts    []linodego.NodeBalancerNodeCreateOptions
	UpdateNodeBalancerNodeCalls   int
	UpdateNodeBalancerNodeOpts    []linodego.NodeBalancerNodeUpdateOptions
}

var _ LinodeClient = &MockLinodeClient{}

func (m *MockLinodeClient) ListSSHKeys(_ context.Context, _ *linodego.ListOptions) ([]linodego.SSHKey, error) {
	return m.ListSSHKeysResponse, m.ListSSHKeysError
}

func (m *MockLinodeClient) CreateSSHKey(_ context.Context, opts linodego.SSHKeyCreateOptions) (*linodego.SSHKey, error) {
	m.CreateSSHKeyCalls++
	m.LastCreateSSHKeyOpts = opts
	if m.CreateSSHKeyError != nil {
		return nil, m.CreateSSHKeyError
	}
	if m.CreateSSHKeyResponse != nil {
		return m.CreateSSHKeyResponse, nil
	}
	return &linodego.SSHKey{ID: 1, Label: opts.Label, SSHKey: opts.SSHKey}, nil
}

func (m *MockLinodeClient) DeleteSSHKey(_ context.Context, _ int) error { return nil }

func (m *MockLinodeClient) ListInstances(_ context.Context, _ *linodego.ListOptions) ([]linodego.Instance, error) {
	return m.ListInstancesResponse, m.ListInstancesError
}

func (m *MockLinodeClient) CreateInstance(_ context.Context, opts linodego.InstanceCreateOptions) (*linodego.Instance, error) {
	m.CreateInstanceCalls++
	m.LastCreateInstanceOpts = opts
	if m.CreateInstanceError != nil {
		return nil, m.CreateInstanceError
	}
	if m.CreateInstanceResponse != nil {
		return m.CreateInstanceResponse, nil
	}
	return &linodego.Instance{ID: 1, Label: opts.Label, Region: opts.Region, Type: opts.Type, Image: opts.Image, Tags: opts.Tags}, nil
}

func (m *MockLinodeClient) DeleteInstance(_ context.Context, linodeID int) error {
	m.DeletedInstanceIDs = append(m.DeletedInstanceIDs, linodeID)
	if m.DeleteInstanceErrByID != nil {
		if err := m.DeleteInstanceErrByID[linodeID]; err != nil {
			return err
		}
	}
	return m.DeleteInstanceErr
}

func (m *MockLinodeClient) UpdateInstance(_ context.Context, linodeID int, opts linodego.InstanceUpdateOptions) (*linodego.Instance, error) {
	m.UpdatedInstanceIDs = append(m.UpdatedInstanceIDs, linodeID)
	if opts.Tags != nil {
		if m.UpdatedTagsByID == nil {
			m.UpdatedTagsByID = make(map[int][]string)
		}
		m.UpdatedTagsByID[linodeID] = append([]string(nil), (*opts.Tags)...)
	}
	if m.UpdateInstanceErrByID != nil {
		if err := m.UpdateInstanceErrByID[linodeID]; err != nil {
			return nil, err
		}
	}
	if m.UpdateInstanceErr != nil {
		return nil, m.UpdateInstanceErr
	}
	updated := &linodego.Instance{ID: linodeID}
	if opts.Tags != nil {
		updated.Tags = append([]string(nil), (*opts.Tags)...)
	}
	return updated, nil
}

func (m *MockLinodeClient) ListVolumes(_ context.Context, _ *linodego.ListOptions) ([]linodego.Volume, error) {
	return m.ListVolumesResponse, m.ListVolumesError
}

func (m *MockLinodeClient) CreateVolume(_ context.Context, opts linodego.VolumeCreateOptions) (*linodego.Volume, error) {
	m.CreateVolumeCalls++
	m.LastCreateVolumeOpts = opts
	if m.CreateVolumeError != nil {
		return nil, m.CreateVolumeError
	}
	if m.CreateVolumeResponse != nil {
		return m.CreateVolumeResponse, nil
	}
	return &linodego.Volume{ID: 1, Label: opts.Label, Region: opts.Region, Size: opts.Size, Tags: opts.Tags}, nil
}

func (m *MockLinodeClient) DeleteVolume(_ context.Context, _ int) error { return nil }

func (m *MockLinodeClient) ListNodeBalancers(_ context.Context, _ *linodego.ListOptions) ([]linodego.NodeBalancer, error) {
	return m.ListNodeBalancersResponse, m.ListNodeBalancersError
}

func (m *MockLinodeClient) GetNodeBalancer(_ context.Context, nodebalancerID int) (*linodego.NodeBalancer, error) {
	for _, nb := range m.ListNodeBalancersResponse {
		if nb.ID == nodebalancerID {
			return &nb, nil
		}
	}
	return nil, nil
}

func (m *MockLinodeClient) CreateNodeBalancer(_ context.Context, opts linodego.NodeBalancerCreateOptions) (*linodego.NodeBalancer, error) {
	m.CreateNodeBalancerCalls++
	m.LastCreateNodeBalancerOpts = opts
	if m.CreateNodeBalancerError != nil {
		return nil, m.CreateNodeBalancerError
	}
	if m.CreateNodeBalancerResponse != nil {
		return m.CreateNodeBalancerResponse, nil
	}
	return &linodego.NodeBalancer{ID: 1, Label: opts.Label, Region: opts.Region, Tags: opts.Tags}, nil
}

func (m *MockLinodeClient) DeleteNodeBalancer(_ context.Context, _ int) error { return nil }

func (m *MockLinodeClient) ListNodeBalancerConfigs(_ context.Context, _ int, _ *linodego.ListOptions) ([]linodego.NodeBalancerConfig, error) {
	return m.ListNodeBalancerConfigsResponse, m.ListNodeBalancerConfigsError
}

func (m *MockLinodeClient) CreateNodeBalancerConfig(_ context.Context, _ int, opts linodego.NodeBalancerConfigCreateOptions) (*linodego.NodeBalancerConfig, error) {
	m.CreateNodeBalancerConfigCalls++
	m.CreateNodeBalancerConfigOpts = append(m.CreateNodeBalancerConfigOpts, opts)
	return &linodego.NodeBalancerConfig{ID: m.CreateNodeBalancerConfigCalls, Port: opts.Port}, nil
}

func (m *MockLinodeClient) RebuildNodeBalancerConfig(_ context.Context, _ int, configID int, opts linodego.NodeBalancerConfigRebuildOptions) (*linodego.NodeBalancerConfig, error) {
	m.RebuildNodeBalancerConfigCalls++
	m.RebuildNodeBalancerConfigOpts = append(m.RebuildNodeBalancerConfigOpts, opts)
	return &linodego.NodeBalancerConfig{ID: configID, Port: opts.Port}, nil
}

func (m *MockLinodeClient) ListNodeBalancerNodes(_ context.Context, _ int, configID int, _ *linodego.ListOptions) ([]linodego.NodeBalancerNode, error) {
	if m.ListNodeBalancerNodesResponse == nil {
		return nil, nil
	}
	return m.ListNodeBalancerNodesResponse[configID], nil
}

func (m *MockLinodeClient) CreateNodeBalancerNode(_ context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerNodeCreateOptions) (*linodego.NodeBalancerNode, error) {
	m.CreateNodeBalancerNodeCalls++
	m.CreateNodeBalancerNodeOpts = append(m.CreateNodeBalancerNodeOpts, opts)
	return &linodego.NodeBalancerNode{ID: m.CreateNodeBalancerNodeCalls, Address: opts.Address, Label: opts.Label, Mode: opts.Mode, ConfigID: configID, NodeBalancerID: nodebalancerID}, nil
}

func (m *MockLinodeClient) UpdateNodeBalancerNode(_ context.Context, nodebalancerID int, configID int, nodeID int, opts linodego.NodeBalancerNodeUpdateOptions) (*linodego.NodeBalancerNode, error) {
	m.UpdateNodeBalancerNodeCalls++
	m.UpdateNodeBalancerNodeOpts = append(m.UpdateNodeBalancerNodeOpts, opts)
	return &linodego.NodeBalancerNode{ID: nodeID, Address: opts.Address, Label: opts.Label, Mode: opts.Mode, Weight: opts.Weight, ConfigID: configID, NodeBalancerID: nodebalancerID}, nil
}

// MockLinodeCloud implements LinodeCloud for use in tests.
type MockLinodeCloud struct {
	Client_ *MockLinodeClient
}

var _ LinodeCloud = &MockLinodeCloud{}

func (m *MockLinodeCloud) ProviderID() kops.CloudProviderID                         { return kops.CloudProviderLinode }
func (m *MockLinodeCloud) DNS() (dnsprovider.Interface, error)                      { return nil, nil }
func (m *MockLinodeCloud) FindVPCInfo(_ string) (*fi.VPCInfo, error)                { return nil, nil }
func (m *MockLinodeCloud) DeleteInstance(_ *cloudinstances.CloudInstance) error     { return nil }
func (m *MockLinodeCloud) DeregisterInstance(_ *cloudinstances.CloudInstance) error { return nil }
func (m *MockLinodeCloud) DeleteGroup(_ *cloudinstances.CloudInstanceGroup) error   { return nil }
func (m *MockLinodeCloud) DetachInstance(_ *cloudinstances.CloudInstance) error     { return nil }
func (m *MockLinodeCloud) GetCloudGroups(_ *kops.Cluster, _ []*kops.InstanceGroup, _ bool, _ []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, nil
}
func (m *MockLinodeCloud) Region() string { return "us-east" }
func (m *MockLinodeCloud) FindClusterStatus(_ *kops.Cluster) (*kops.ClusterStatus, error) {
	return &kops.ClusterStatus{}, nil
}
func (m *MockLinodeCloud) GetApiIngressStatus(_ *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, nil
}
func (m *MockLinodeCloud) AccessToken() string  { return "test-token" }
func (m *MockLinodeCloud) Client() LinodeClient { return m.Client_ }

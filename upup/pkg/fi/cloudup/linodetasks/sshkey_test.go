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

package linodetasks

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/linode/linodego"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const testOpenSSHPublicKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCySdqIU+FhCWl3BNrAvPaOe5VfL2aCARUWwy91ZP+T7LBwFa9lhdttfjp/VX1D1/PVwntn2EhN079m8c2kfdmiZ/iCHqrLyIGSd+BOiCz0lT47znvANSfxYjLUuKrWWWeaXqerJkOsAD4PHchRLbZGPdbfoBKwtb/WT4GMRQmb9vmiaZYjsfdPPM9KkWI9ECoWFGjGehA8D+iYIPR711kRacb1xdYmnjHqxAZHFsb5L8wDWIeAyhy49cBD+lbzTiioq2xWLorXuFmXh6Do89PgzvHeyCLY6816f/kCX6wIFts8A2eaEHFL4rAOsuh6qHmSxGCR9peSyuRW8DxV725x justin@test"

type fakeLinodeClient struct {
	listResponse []linodego.SSHKey
	listError    error

	createResponse *linodego.SSHKey
	createError    error
	createCalls    int
	lastCreateOpts linodego.SSHKeyCreateOptions

	listInstancesResponse []linodego.Instance
	listInstancesError    error

	createInstanceResponse *linodego.Instance
	createInstanceError    error
	createInstanceCalls    int
	lastCreateInstanceOpts linodego.InstanceCreateOptions

	listVolumesResponse []linodego.Volume
	listVolumesError    error

	createVolumeResponse *linodego.Volume
	createVolumeError    error
	createVolumeCalls    int
	lastCreateVolumeOpts linodego.VolumeCreateOptions

	listNodeBalancersResponse []linodego.NodeBalancer
	listNodeBalancersError    error

	createNodeBalancerResponse *linodego.NodeBalancer
	createNodeBalancerError    error
	createNodeBalancerCalls    int
	lastCreateNodeBalancerOpts linodego.NodeBalancerCreateOptions

	listNodeBalancerConfigsResponse []linodego.NodeBalancerConfig
	listNodeBalancerConfigsError    error

	createNodeBalancerConfigCalls int
	createNodeBalancerConfigOpts  []linodego.NodeBalancerConfigCreateOptions

	rebuildNodeBalancerConfigCalls int
	rebuildNodeBalancerConfigOpts  []linodego.NodeBalancerConfigRebuildOptions

	listNodeBalancerNodesResponse map[int][]linodego.NodeBalancerNode

	createNodeBalancerNodeCalls int
	createNodeBalancerNodeOpts  []linodego.NodeBalancerNodeCreateOptions

	updateNodeBalancerNodeCalls int
	updateNodeBalancerNodeOpts  []linodego.NodeBalancerNodeUpdateOptions
}

func (f *fakeLinodeClient) ListSSHKeys(ctx context.Context, opts *linodego.ListOptions) ([]linodego.SSHKey, error) {
	if f.listError != nil {
		return nil, f.listError
	}
	return f.listResponse, nil
}

func (f *fakeLinodeClient) CreateSSHKey(ctx context.Context, opts linodego.SSHKeyCreateOptions) (*linodego.SSHKey, error) {
	f.createCalls++
	f.lastCreateOpts = opts

	if f.createError != nil {
		return nil, f.createError
	}
	if f.createResponse == nil {
		return &linodego.SSHKey{ID: 1, Label: opts.Label, SSHKey: opts.SSHKey}, nil
	}

	return f.createResponse, nil
}

func (f *fakeLinodeClient) DeleteSSHKey(ctx context.Context, keyID int) error {
	return nil
}

func (f *fakeLinodeClient) ListInstances(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Instance, error) {
	if f.listInstancesError != nil {
		return nil, f.listInstancesError
	}
	return f.listInstancesResponse, nil
}

func (f *fakeLinodeClient) CreateInstance(ctx context.Context, opts linodego.InstanceCreateOptions) (*linodego.Instance, error) {
	f.createInstanceCalls++
	f.lastCreateInstanceOpts = opts

	if f.createInstanceError != nil {
		return nil, f.createInstanceError
	}
	if f.createInstanceResponse == nil {
		return &linodego.Instance{ID: 1, Label: opts.Label, Region: opts.Region, Type: opts.Type, Image: opts.Image, Tags: opts.Tags}, nil
	}

	return f.createInstanceResponse, nil
}

func (f *fakeLinodeClient) DeleteInstance(ctx context.Context, linodeID int) error {
	return nil
}

func (f *fakeLinodeClient) UpdateInstance(ctx context.Context, linodeID int, opts linodego.InstanceUpdateOptions) (*linodego.Instance, error) {
	return &linodego.Instance{ID: linodeID}, nil
}

func (f *fakeLinodeClient) ListVolumes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Volume, error) {
	if f.listVolumesError != nil {
		return nil, f.listVolumesError
	}
	return f.listVolumesResponse, nil
}

func (f *fakeLinodeClient) CreateVolume(ctx context.Context, opts linodego.VolumeCreateOptions) (*linodego.Volume, error) {
	f.createVolumeCalls++
	f.lastCreateVolumeOpts = opts

	if f.createVolumeError != nil {
		return nil, f.createVolumeError
	}
	if f.createVolumeResponse == nil {
		return &linodego.Volume{ID: 1, Label: opts.Label, Region: opts.Region, Size: opts.Size, Tags: opts.Tags}, nil
	}

	return f.createVolumeResponse, nil
}

func (f *fakeLinodeClient) DeleteVolume(ctx context.Context, volumeID int) error {
	return nil
}

func (f *fakeLinodeClient) ListNodeBalancers(ctx context.Context, opts *linodego.ListOptions) ([]linodego.NodeBalancer, error) {
	if f.listNodeBalancersError != nil {
		return nil, f.listNodeBalancersError
	}
	return f.listNodeBalancersResponse, nil
}

func (f *fakeLinodeClient) GetNodeBalancer(ctx context.Context, nodebalancerID int) (*linodego.NodeBalancer, error) {
	return nil, nil
}

func (f *fakeLinodeClient) CreateNodeBalancer(ctx context.Context, opts linodego.NodeBalancerCreateOptions) (*linodego.NodeBalancer, error) {
	f.createNodeBalancerCalls++
	f.lastCreateNodeBalancerOpts = opts

	if f.createNodeBalancerError != nil {
		return nil, f.createNodeBalancerError
	}
	if f.createNodeBalancerResponse == nil {
		return &linodego.NodeBalancer{ID: 1, Label: opts.Label, Region: opts.Region, Tags: opts.Tags}, nil
	}

	return f.createNodeBalancerResponse, nil
}

func (f *fakeLinodeClient) DeleteNodeBalancer(ctx context.Context, nodebalancerID int) error {
	return nil
}

func (f *fakeLinodeClient) ListNodeBalancerConfigs(ctx context.Context, nodebalancerID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerConfig, error) {
	if f.listNodeBalancerConfigsError != nil {
		return nil, f.listNodeBalancerConfigsError
	}
	return f.listNodeBalancerConfigsResponse, nil
}

func (f *fakeLinodeClient) CreateNodeBalancerConfig(ctx context.Context, nodebalancerID int, opts linodego.NodeBalancerConfigCreateOptions) (*linodego.NodeBalancerConfig, error) {
	f.createNodeBalancerConfigCalls++
	f.createNodeBalancerConfigOpts = append(f.createNodeBalancerConfigOpts, opts)
	return &linodego.NodeBalancerConfig{ID: f.createNodeBalancerConfigCalls, Port: opts.Port}, nil
}

func (f *fakeLinodeClient) RebuildNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerConfigRebuildOptions) (*linodego.NodeBalancerConfig, error) {
	f.rebuildNodeBalancerConfigCalls++
	f.rebuildNodeBalancerConfigOpts = append(f.rebuildNodeBalancerConfigOpts, opts)
	return &linodego.NodeBalancerConfig{ID: configID, Port: opts.Port}, nil
}

func (f *fakeLinodeClient) ListNodeBalancerNodes(ctx context.Context, nodebalancerID int, configID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerNode, error) {
	if f.listNodeBalancerNodesResponse == nil {
		return nil, nil
	}
	return f.listNodeBalancerNodesResponse[configID], nil
}

func (f *fakeLinodeClient) CreateNodeBalancerNode(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerNodeCreateOptions) (*linodego.NodeBalancerNode, error) {
	f.createNodeBalancerNodeCalls++
	f.createNodeBalancerNodeOpts = append(f.createNodeBalancerNodeOpts, opts)
	return &linodego.NodeBalancerNode{ID: f.createNodeBalancerNodeCalls, Address: opts.Address, Label: opts.Label, Mode: opts.Mode, ConfigID: configID, NodeBalancerID: nodebalancerID}, nil
}

func (f *fakeLinodeClient) UpdateNodeBalancerNode(ctx context.Context, nodebalancerID int, configID int, nodeID int, opts linodego.NodeBalancerNodeUpdateOptions) (*linodego.NodeBalancerNode, error) {
	f.updateNodeBalancerNodeCalls++
	f.updateNodeBalancerNodeOpts = append(f.updateNodeBalancerNodeOpts, opts)
	return &linodego.NodeBalancerNode{ID: nodeID, Address: opts.Address, Label: opts.Label, Mode: opts.Mode, Weight: opts.Weight, ConfigID: configID, NodeBalancerID: nodebalancerID}, nil
}

type fakeLinodeCloud struct {
	client linode.LinodeClient
}

var _ linode.LinodeCloud = &fakeLinodeCloud{}

func (f *fakeLinodeCloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderLinode
}

func (f *fakeLinodeCloud) DNS() (dnsprovider.Interface, error) {
	return nil, nil
}

func (f *fakeLinodeCloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, nil
}

func (f *fakeLinodeCloud) DeleteInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (f *fakeLinodeCloud) DeregisterInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (f *fakeLinodeCloud) DeleteGroup(group *cloudinstances.CloudInstanceGroup) error {
	return nil
}

func (f *fakeLinodeCloud) DetachInstance(instance *cloudinstances.CloudInstance) error {
	return nil
}

func (f *fakeLinodeCloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	return nil, nil
}

func (f *fakeLinodeCloud) Region() string {
	return "us-east"
}

func (f *fakeLinodeCloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return &kops.ClusterStatus{}, nil
}

func (f *fakeLinodeCloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, nil
}

func (f *fakeLinodeCloud) AccessToken() string {
	return "test-token"
}

func (f *fakeLinodeCloud) Client() linode.LinodeClient {
	return f.client
}

func newTestPublicKeyResource() *fi.Resource {
	resource := fi.Resource(fi.NewStringResource(testOpenSSHPublicKey))
	return &resource
}

func newTestCloudupContext(t *testing.T, cloud linode.LinodeCloud) *fi.CloudupContext {
	t.Helper()

	ctx, err := fi.NewCloudupContext(
		context.Background(),
		fi.DeletionProcessingModeDeleteIncludingDeferred,
		linode.NewAPITarget(cloud),
		nil,
		cloud,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		t.Fatalf("unexpected error creating context: %v", err)
	}

	return ctx
}

func TestSSHKeyFindMatch(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &fakeLinodeClient{
		listResponse: []linodego.SSHKey{{
			ID:     123,
			Label:  "kubernetes.example.k8s.local-1234",
			SSHKey: testOpenSSHPublicKey,
		}},
	}
	cloud := &fakeLinodeCloud{client: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{
		Name:      fi.PtrTo("kubernetes.example.k8s.local-1234"),
		PublicKey: publicKey,
		Lifecycle: fi.LifecycleSync,
	}

	actual, err := task.Find(ctx)
	if err != nil {
		t.Fatalf("Find returned error: %v", err)
	}
	if actual == nil {
		t.Fatalf("expected to find SSH key")
	}
	if got, want := fi.ValueOf(actual.ID), 123; got != want {
		t.Fatalf("unexpected ID: got %d, want %d", got, want)
	}
}

func TestSSHKeyFindDuplicate(t *testing.T) {
	client := &fakeLinodeClient{
		listResponse: []linodego.SSHKey{
			{ID: 1, Label: "kubernetes.example.k8s.local-1234", SSHKey: "ssh-rsa AAAA test"},
			{ID: 2, Label: "kubernetes.example.k8s.local-1234", SSHKey: "ssh-rsa AAAA test"},
		},
	}
	cloud := &fakeLinodeCloud{client: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234")}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected duplicate name error")
	}
	if !strings.Contains(err.Error(), "found multiple SSH keys named") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyFindPublicKeyMismatch(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &fakeLinodeClient{
		listResponse: []linodego.SSHKey{{
			ID:     123,
			Label:  "kubernetes.example.k8s.local-1234",
			SSHKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDbadkey",
		}},
	}
	cloud := &fakeLinodeCloud{client: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), PublicKey: publicKey}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected mismatch error")
	}
	if !strings.Contains(err.Error(), "public key data did not match") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyFindListError(t *testing.T) {
	client := &fakeLinodeClient{listError: errors.New("api unavailable")}
	cloud := &fakeLinodeCloud{client: client}
	ctx := newTestCloudupContext(t, cloud)

	task := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234")}
	_, err := task.Find(ctx)
	if err == nil {
		t.Fatalf("expected list error")
	}
	if !strings.Contains(err.Error(), "error listing Linode (Akamai) SSH keys") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSSHKeyRenderLinodeCreate(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &fakeLinodeClient{createResponse: &linodego.SSHKey{ID: 42, Label: "kubernetes.example.k8s.local-1234"}}
	cloud := &fakeLinodeCloud{client: client}
	target := linode.NewAPITarget(cloud)

	expected := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), PublicKey: publicKey}
	err := (&SSHKey{}).RenderLinode(target, nil, expected, nil)
	if err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got, want := client.createCalls, 1; got != want {
		t.Fatalf("unexpected create calls: got %d, want %d", got, want)
	}
	if got, want := client.lastCreateOpts.Label, "kubernetes.example.k8s.local-1234"; got != want {
		t.Fatalf("unexpected create label: got %q, want %q", got, want)
	}
	if fi.ValueOf(expected.ID) != 42 {
		t.Fatalf("expected task ID to be populated from create response")
	}
}

func TestSSHKeyRenderLinodeNoopWhenActualExists(t *testing.T) {
	publicKey := newTestPublicKeyResource()
	client := &fakeLinodeClient{}
	cloud := &fakeLinodeCloud{client: client}
	target := linode.NewAPITarget(cloud)

	actual := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), ID: fi.PtrTo(11)}
	expected := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234"), PublicKey: publicKey}
	err := (&SSHKey{}).RenderLinode(target, actual, expected, nil)
	if err != nil {
		t.Fatalf("RenderLinode returned error: %v", err)
	}
	if got := client.createCalls; got != 0 {
		t.Fatalf("unexpected create calls: got %d, want 0", got)
	}
}

func TestSSHKeyRenderLinodeRequiresPublicKey(t *testing.T) {
	client := &fakeLinodeClient{}
	cloud := &fakeLinodeCloud{client: client}
	target := linode.NewAPITarget(cloud)

	expected := &SSHKey{Name: fi.PtrTo("kubernetes.example.k8s.local-1234")}
	err := (&SSHKey{}).RenderLinode(target, nil, expected, nil)
	if err == nil {
		t.Fatalf("expected missing PublicKey error")
	}
	if !strings.Contains(err.Error(), "PublicKey") {
		t.Fatalf("unexpected error: %v", err)
	}
}

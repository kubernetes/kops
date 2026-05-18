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
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/linode/linodego"
	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	cloudlinode "k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

type fakeLinodeClient struct {
	listInstancesResponse []linodego.Instance
	listInstancesErr      error

	listVolumesResponse []linodego.Volume
	listVolumesErr      error

	listSSHKeysResponse []linodego.SSHKey
	listSSHKeysErr      error

	listNodeBalancersResponse []linodego.NodeBalancer
	listNodeBalancersErr      error

	deleteSSHKeyErrByID map[int]error
	deletedSSHKeyIDs    []int

	deleteVolumeErrByID map[int]error
	deletedVolumeIDs    []int

	deleteNodeBalancerErrByID map[int]error
	deletedNodeBalancerIDs    []int
}

func (f *fakeLinodeClient) ListSSHKeys(ctx context.Context, opts *linodego.ListOptions) ([]linodego.SSHKey, error) {
	if f.listSSHKeysErr != nil {
		return nil, f.listSSHKeysErr
	}

	return f.listSSHKeysResponse, nil
}

func (f *fakeLinodeClient) CreateSSHKey(ctx context.Context, opts linodego.SSHKeyCreateOptions) (*linodego.SSHKey, error) {
	return nil, nil
}

func (f *fakeLinodeClient) DeleteSSHKey(ctx context.Context, keyID int) error {
	f.deletedSSHKeyIDs = append(f.deletedSSHKeyIDs, keyID)
	if f.deleteSSHKeyErrByID != nil {
		if err := f.deleteSSHKeyErrByID[keyID]; err != nil {
			return err
		}
	}

	return nil
}

func (f *fakeLinodeClient) ListInstances(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Instance, error) {
	if f.listInstancesErr != nil {
		return nil, f.listInstancesErr
	}

	return f.listInstancesResponse, nil
}

func (f *fakeLinodeClient) CreateInstance(ctx context.Context, opts linodego.InstanceCreateOptions) (*linodego.Instance, error) {
	return nil, nil
}

func (f *fakeLinodeClient) DeleteInstance(ctx context.Context, linodeID int) error {
	return nil
}

func (f *fakeLinodeClient) UpdateInstance(ctx context.Context, linodeID int, opts linodego.InstanceUpdateOptions) (*linodego.Instance, error) {
	return &linodego.Instance{ID: linodeID}, nil
}

func (f *fakeLinodeClient) ListVolumes(ctx context.Context, opts *linodego.ListOptions) ([]linodego.Volume, error) {
	if f.listVolumesErr != nil {
		return nil, f.listVolumesErr
	}

	return f.listVolumesResponse, nil
}

func (f *fakeLinodeClient) CreateVolume(ctx context.Context, opts linodego.VolumeCreateOptions) (*linodego.Volume, error) {
	return nil, nil
}

func (f *fakeLinodeClient) DeleteVolume(ctx context.Context, volumeID int) error {
	f.deletedVolumeIDs = append(f.deletedVolumeIDs, volumeID)
	if f.deleteVolumeErrByID != nil {
		if err := f.deleteVolumeErrByID[volumeID]; err != nil {
			return err
		}
	}

	return nil
}

func (f *fakeLinodeClient) ListNodeBalancers(ctx context.Context, opts *linodego.ListOptions) ([]linodego.NodeBalancer, error) {
	if f.listNodeBalancersErr != nil {
		return nil, f.listNodeBalancersErr
	}

	return f.listNodeBalancersResponse, nil
}

func (f *fakeLinodeClient) GetNodeBalancer(ctx context.Context, nodebalancerID int) (*linodego.NodeBalancer, error) {
	return nil, nil
}

func (f *fakeLinodeClient) CreateNodeBalancer(ctx context.Context, opts linodego.NodeBalancerCreateOptions) (*linodego.NodeBalancer, error) {
	return nil, nil
}

func (f *fakeLinodeClient) DeleteNodeBalancer(ctx context.Context, nodebalancerID int) error {
	f.deletedNodeBalancerIDs = append(f.deletedNodeBalancerIDs, nodebalancerID)
	if f.deleteNodeBalancerErrByID != nil {
		if err := f.deleteNodeBalancerErrByID[nodebalancerID]; err != nil {
			return err
		}
	}

	return nil
}

func (f *fakeLinodeClient) ListNodeBalancerConfigs(ctx context.Context, nodebalancerID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerConfig, error) {
	return nil, nil
}

func (f *fakeLinodeClient) CreateNodeBalancerConfig(ctx context.Context, nodebalancerID int, opts linodego.NodeBalancerConfigCreateOptions) (*linodego.NodeBalancerConfig, error) {
	return nil, nil
}

func (f *fakeLinodeClient) RebuildNodeBalancerConfig(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerConfigRebuildOptions) (*linodego.NodeBalancerConfig, error) {
	return nil, nil
}

func (f *fakeLinodeClient) ListNodeBalancerNodes(ctx context.Context, nodebalancerID int, configID int, opts *linodego.ListOptions) ([]linodego.NodeBalancerNode, error) {
	return nil, nil
}

func (f *fakeLinodeClient) CreateNodeBalancerNode(ctx context.Context, nodebalancerID int, configID int, opts linodego.NodeBalancerNodeCreateOptions) (*linodego.NodeBalancerNode, error) {
	return nil, nil
}

func (f *fakeLinodeClient) UpdateNodeBalancerNode(ctx context.Context, nodebalancerID int, configID int, nodeID int, opts linodego.NodeBalancerNodeUpdateOptions) (*linodego.NodeBalancerNode, error) {
	return nil, nil
}

type fakeLinodeCloud struct {
	client             cloudlinode.LinodeClient
	deletedInstanceIDs []string
	deleteInstanceErr  error
}

var _ cloudlinode.LinodeCloud = &fakeLinodeCloud{}

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
	f.deletedInstanceIDs = append(f.deletedInstanceIDs, instance.ID)
	return f.deleteInstanceErr
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

func (f *fakeLinodeCloud) Client() cloudlinode.LinodeClient {
	return f.client
}

func TestListResources(t *testing.T) {
	client := &fakeLinodeClient{
		listInstancesResponse: []linodego.Instance{
			{ID: 101, Label: "nodes-1", Tags: []string{"kops.k8s.io/cluster:example.k8s.local"}},
			{ID: 102, Label: "nodes-2", Tags: []string{"kops.k8s.io/cluster:other.k8s.local"}},
		},
		listVolumesResponse: []linodego.Volume{
			{ID: 301, Label: "cp-0.etcd-main.example.k8s.local", Tags: []string{"kops.k8s.io/cluster:example.k8s.local", "kops.k8s.io/etcd:main"}},
			{ID: 302, Label: "cp-0.etcd-main.other.k8s.local", Tags: []string{"kops.k8s.io/cluster:other.k8s.local", "kops.k8s.io/etcd:main"}},
		},
		listSSHKeysResponse: []linodego.SSHKey{
			{ID: 501, Label: "kubernetes-example-k8s-local-abc"},
			{ID: 502, Label: "unrelated-key"},
		},
		listNodeBalancersResponse: []linodego.NodeBalancer{
			{ID: 601, Label: linodego.Pointer("api-kops-test-linode-k8s-local"), Tags: []string{"kops.k8s.io/cluster:example.k8s.local"}},
			{ID: 602, Label: linodego.Pointer("api-other"), Tags: []string{"kops.k8s.io/cluster:other.k8s.local"}},
		},
	}
	cloud := &fakeLinodeCloud{client: client}

	resourceMap, err := ListResources(cloud, resources.ClusterInfo{Name: "example.k8s.local"})
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	wantKeys := []string{"instance:101", "nodebalancer:601", "ssh-key:501", "volume:301"}
	if gotKeys := sortedResourceKeys(resourceMap); !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("unexpected resources\nwant: %v\n got: %v", wantKeys, gotKeys)
	}

	if r := resourceMap["instance:101"]; r == nil {
		t.Fatalf("missing instance:101")
	} else if got, want := r.Name, "nodes-1"; got != want {
		t.Fatalf("unexpected instance Name: got %q, want %q", got, want)
	} else if got, want := r.Type, resourceTypeInstance; got != want {
		t.Fatalf("unexpected instance Type: got %q, want %q", got, want)
	}

	if r := resourceMap["ssh-key:501"]; r == nil {
		t.Fatalf("missing ssh-key:501")
	} else if got, want := r.Type, resourceTypeSSHKey; got != want {
		t.Fatalf("unexpected SSH key Type: got %q, want %q", got, want)
	}
}

func TestListResources_CustomSSHKeyName(t *testing.T) {
	client := &fakeLinodeClient{
		listSSHKeysResponse: []linodego.SSHKey{
			{ID: 501, Label: "kubernetes.example.k8s.local-abc"},
			{ID: 777, Label: "my-custom-key-name"},
		},
	}
	cloud := &fakeLinodeCloud{client: client}

	resourceMap, err := ListResources(cloud, resources.ClusterInfo{Name: "example.k8s.local", LinodeSSHKeyName: "my.custom:key:name"})
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	wantKeys := []string{"ssh-key:777"}
	if gotKeys := sortedResourceKeys(resourceMap); !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("unexpected resources\nwant: %v\n got: %v", wantKeys, gotKeys)
	}
}

func TestListResources_DefaultSSHKeyNameLegacyAndNormalized(t *testing.T) {
	client := &fakeLinodeClient{
		listSSHKeysResponse: []linodego.SSHKey{
			{ID: 501, Label: "kubernetes.example.k8s.local-legacy"},
			{ID: 502, Label: "kubernetes-example-k8s-local-normalized"},
			{ID: 503, Label: "unrelated-key"},
		},
	}
	cloud := &fakeLinodeCloud{client: client}

	resourceMap, err := ListResources(cloud, resources.ClusterInfo{Name: "example.k8s.local"})
	if err != nil {
		t.Fatalf("ListResources returned error: %v", err)
	}

	wantKeys := []string{"ssh-key:501", "ssh-key:502"}
	if gotKeys := sortedResourceKeys(resourceMap); !reflect.DeepEqual(gotKeys, wantKeys) {
		t.Fatalf("unexpected resources\nwant: %v\n got: %v", wantKeys, gotKeys)
	}
}

func TestDeleteInstance(t *testing.T) {
	cloud := &fakeLinodeCloud{client: &fakeLinodeClient{}}

	tracker := &resources.Resource{Name: "nodes-1", ID: "101", Type: resourceTypeInstance}
	if err := deleteInstance(cloud, tracker); err != nil {
		t.Fatalf("deleteInstance returned error: %v", err)
	}

	if !reflect.DeepEqual(cloud.deletedInstanceIDs, []string{"101"}) {
		t.Fatalf("unexpected deleted instance IDs: %v", cloud.deletedInstanceIDs)
	}
}

func TestDeleteSSHKey(t *testing.T) {
	client := &fakeLinodeClient{}
	cloud := &fakeLinodeCloud{client: client}

	tracker := &resources.Resource{Name: "kubernetes.example.k8s.local-abc", ID: "501", Type: resourceTypeSSHKey}
	if err := deleteSSHKey(cloud, tracker); err != nil {
		t.Fatalf("deleteSSHKey returned error: %v", err)
	}

	if !reflect.DeepEqual(client.deletedSSHKeyIDs, []int{501}) {
		t.Fatalf("unexpected deleted SSH key IDs: %v", client.deletedSSHKeyIDs)
	}
}

func TestDeleteSSHKey_NotFound(t *testing.T) {
	client := &fakeLinodeClient{
		deleteSSHKeyErrByID: map[int]error{
			501: &linodego.Error{Code: 404, Message: "not found"},
		},
	}
	cloud := &fakeLinodeCloud{client: client}

	tracker := &resources.Resource{Name: "kubernetes.example.k8s.local-abc", ID: "501", Type: resourceTypeSSHKey}
	if err := deleteSSHKey(cloud, tracker); err != nil {
		t.Fatalf("deleteSSHKey returned error for not found response: %v", err)
	}
}

func TestDeleteVolume(t *testing.T) {
	client := &fakeLinodeClient{}
	cloud := &fakeLinodeCloud{client: client}

	tracker := &resources.Resource{Name: "cp-0.etcd-main.example.k8s.local", ID: "301", Type: resourceTypeVolume}
	if err := deleteVolume(cloud, tracker); err != nil {
		t.Fatalf("deleteVolume returned error: %v", err)
	}

	if !reflect.DeepEqual(client.deletedVolumeIDs, []int{301}) {
		t.Fatalf("unexpected deleted volume IDs: %v", client.deletedVolumeIDs)
	}
}

func TestDeleteVolume_NotFound(t *testing.T) {
	client := &fakeLinodeClient{
		deleteVolumeErrByID: map[int]error{
			301: &linodego.Error{Code: 404, Message: "not found"},
		},
	}
	cloud := &fakeLinodeCloud{client: client}

	tracker := &resources.Resource{Name: "cp-0.etcd-main.example.k8s.local", ID: "301", Type: resourceTypeVolume}
	if err := deleteVolume(cloud, tracker); err != nil {
		t.Fatalf("deleteVolume returned error for not found response: %v", err)
	}
}

func TestDeleteNodeBalancer(t *testing.T) {
	client := &fakeLinodeClient{}
	cloud := &fakeLinodeCloud{client: client}

	tracker := &resources.Resource{Name: "api-kops-test-linode-k8s-local", ID: "601", Type: resourceTypeNodeBalancer}
	if err := deleteNodeBalancer(cloud, tracker); err != nil {
		t.Fatalf("deleteNodeBalancer returned error: %v", err)
	}

	if !reflect.DeepEqual(client.deletedNodeBalancerIDs, []int{601}) {
		t.Fatalf("unexpected deleted node balancer IDs: %v", client.deletedNodeBalancerIDs)
	}
}

func TestDeleteNodeBalancer_NotFound(t *testing.T) {
	client := &fakeLinodeClient{
		deleteNodeBalancerErrByID: map[int]error{
			601: &linodego.Error{Code: 404, Message: "not found"},
		},
	}
	cloud := &fakeLinodeCloud{client: client}

	tracker := &resources.Resource{Name: "api-kops-test-linode-k8s-local", ID: "601", Type: resourceTypeNodeBalancer}
	if err := deleteNodeBalancer(cloud, tracker); err != nil {
		t.Fatalf("deleteNodeBalancer returned error for not found response: %v", err)
	}
}

func TestListResources_PropagatesErrors(t *testing.T) {
	client := &fakeLinodeClient{listInstancesErr: errors.New("instances API down")}
	cloud := &fakeLinodeCloud{client: client}

	_, err := ListResources(cloud, resources.ClusterInfo{Name: "example.k8s.local"})
	if err == nil {
		t.Fatalf("expected error when listing instances")
	}
}

func sortedResourceKeys(m map[string]*resources.Resource) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

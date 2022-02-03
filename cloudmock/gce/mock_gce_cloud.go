/*
Copyright 2017 The Kubernetes Authors.

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

package gce

import (
	"fmt"

	"google.golang.org/api/cloudresourcemanager/v1"
	compute "google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/cloudmock/gce/mockcloudresourcemanager"
	mockcompute "k8s.io/kops/cloudmock/gce/mockcompute"
	"k8s.io/kops/cloudmock/gce/mockdns"
	"k8s.io/kops/cloudmock/gce/mockiam"
	"k8s.io/kops/cloudmock/gce/mockstorage"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dnsproviderclouddns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/google/clouddns"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

// MockGCECloud is a mock implementation of GCECloud for testing
type MockGCECloud struct {
	project string
	region  string
	labels  map[string]string

	computeClient              *mockcompute.MockClient
	dnsClient                  *mockdns.MockClient
	iamClient                  *mockiam.MockClient
	storageClient              *storage.Service
	cloudResourceManagerClient *cloudresourcemanager.Service
}

var _ gce.GCECloud = &MockGCECloud{}

// InstallMockGCECloud registers a MockGCECloud implementation for the specified region & project
func InstallMockGCECloud(region string, project string) *MockGCECloud {
	c := &MockGCECloud{
		project:                    project,
		region:                     region,
		computeClient:              mockcompute.NewMockClient(project),
		dnsClient:                  mockdns.NewMockClient(),
		iamClient:                  mockiam.NewMockClient(project),
		storageClient:              mockstorage.New(),
		cloudResourceManagerClient: mockcloudresourcemanager.New(),
	}
	gce.CacheGCECloudInstance(region, project, c)
	return c
}

func (c *MockGCECloud) AllResources() map[string]interface{} {
	return c.computeClient.AllResources()
}

// GetCloudGroups is not implemented yet
func (c *MockGCECloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Infof("MockGCECloud cloud provider GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("MockGCECloud cloud provider does not support getting cloud groups at this time")
}

// Zones is not implemented yet
func (c *MockGCECloud) Zones() ([]string, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// WithLabels returns a copy of the MockGCECloud bound to the specified labels
func (c *MockGCECloud) WithLabels(labels map[string]string) gce.GCECloud {
	i := &MockGCECloud{}
	*i = *c
	i.labels = labels
	return i
}

// ProviderID implements fi.Cloud::ProviderID
func (c *MockGCECloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderGCE
}

// FindVPCInfo implements fi.Cloud::FindVPCInfo
func (c *MockGCECloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("MockGCECloud::MockGCECloud not implemented")
}

// DNS implements fi.Cloud::DNS
func (c *MockGCECloud) DNS() (dnsprovider.Interface, error) {
	return dnsproviderclouddns.NewFakeInterface()
}

// Compute implements GCECloud::Compute
func (c *MockGCECloud) Compute() gce.ComputeClient {
	return c.computeClient
}

// Storage implements GCECloud::Storage
func (c *MockGCECloud) Storage() *storage.Service {
	return c.storageClient
}

// IAM returns the IAM client
func (c *MockGCECloud) IAM() gce.IamClient {
	return c.iamClient
}

// CloudResourceManager returns the client for the cloudresourcemanager API
func (c *MockGCECloud) CloudResourceManager() *cloudresourcemanager.Service {
	return c.cloudResourceManagerClient
}

// CloudDNS returns the DNS client
func (c *MockGCECloud) CloudDNS() gce.DNSClient {
	return c.dnsClient
}

// WaitForOp implements GCECloud::WaitForOp
func (c *MockGCECloud) WaitForOp(op *compute.Operation) error {
	if op.Status != "DONE" {
		return fmt.Errorf("unexpected operation: %+v", op)
	}
	return nil
}

// FindClusterStatus implements GCECloud::FindClusterStatus
func (c *MockGCECloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return nil, fmt.Errorf("MockGCECloud::FindClusterStatus not implemented")
}

// GetApiIngressStatus implements GCECloud::GetApiIngressStatus
func (c *MockGCECloud) GetApiIngressStatus(cluster *kops.Cluster) ([]fi.ApiIngressStatus, error) {
	return nil, fmt.Errorf("MockGCECloud::GetApiIngressStatus not implemented")
}

// Region implements GCECloud::Region
func (c *MockGCECloud) Region() string {
	return c.region
}

// Project implements GCECloud::Project
func (c *MockGCECloud) Project() string {
	return c.project
}

// ServiceAccount implements GCECloud::ServiceAccount
func (c *MockGCECloud) ServiceAccount() (string, error) {
	return "12345678-compute@developer.gserviceaccount.com", nil
}

// Labels implements GCECloud::Labels
func (c *MockGCECloud) Labels() map[string]string {
	return c.labels
}

// DeleteGroup implements fi.Cloud::DeleteGroup
func (c *MockGCECloud) DeleteGroup(g *cloudinstances.CloudInstanceGroup) error {
	return nil
	//	return deleteCloudInstanceGroup(c, g)
}

// DeleteInstance deletes a GCE instance
func (c *MockGCECloud) DeleteInstance(i *cloudinstances.CloudInstance) error {
	return nil
	//	return recreateCloudInstance(c, i)
}

func (c *MockGCECloud) DeregisterInstance(i *cloudinstances.CloudInstance) error {
	return nil
}

// DetachInstance is not implemented yet. It needs to cause a cloud instance to no longer be counted against the group's size limits.
func (c *MockGCECloud) DetachInstance(i *cloudinstances.CloudInstance) error {
	return nil
}

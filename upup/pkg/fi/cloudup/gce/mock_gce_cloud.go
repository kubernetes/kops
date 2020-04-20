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

	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/dns/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/storage/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	dnsproviderclouddns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/google/clouddns"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi"
)

// mockGCECloud is a mock implementation of GCECloud for testing
type mockGCECloud struct {
	region  string
	project string
	labels  map[string]string
}

var _ GCECloud = &mockGCECloud{}

// InstallMockGCECloud registers a mockGCECloud implementation for the specified region & project
func InstallMockGCECloud(region string, project string) *mockGCECloud {
	i := buildMockGCECloud(region, project)
	gceCloudInstances[region+"::"+project] = i
	return i
}

// buildMockGCECloud creates a mockGCECloud implementation for the specified region & project
func buildMockGCECloud(region string, project string) *mockGCECloud {
	i := &mockGCECloud{region: region, project: project}
	return i
}

// GetCloudGroups is not implemented yet
func (c *mockGCECloud) GetCloudGroups(cluster *kops.Cluster, instancegroups []*kops.InstanceGroup, warnUnmatched bool, nodes []v1.Node) (map[string]*cloudinstances.CloudInstanceGroup, error) {
	klog.V(8).Infof("mockGCECloud cloud provider GetCloudGroups not implemented yet")
	return nil, fmt.Errorf("mockGCECloud cloud provider does not support getting cloud groups at this time")
}

// Zones is not implemented yet
func (c *mockGCECloud) Zones() ([]string, error) {
	return nil, fmt.Errorf("not yet implemented")
}

// WithLabels returns a copy of the mockGCECloud bound to the specified labels
func (c *mockGCECloud) WithLabels(labels map[string]string) GCECloud {
	i := &mockGCECloud{}
	*i = *c
	i.labels = labels
	return i
}

// ProviderID implements fi.Cloud::ProviderID
func (c *mockGCECloud) ProviderID() kops.CloudProviderID {
	return kops.CloudProviderGCE
}

// FindVPCInfo implements fi.Cloud::FindVPCInfo
func (c *mockGCECloud) FindVPCInfo(id string) (*fi.VPCInfo, error) {
	return nil, fmt.Errorf("mockGCECloud::mockGCECloud not implemented")
}

// DNS implements fi.Cloud::DNS
func (c *mockGCECloud) DNS() (dnsprovider.Interface, error) {
	return dnsproviderclouddns.NewFakeInterface()
}

// Compute implements GCECloud::Compute
func (c *mockGCECloud) Compute() *compute.Service {
	klog.Fatalf("mockGCECloud::Compute not implemented")
	return nil
}

// Storage implements GCECloud::Storage
func (c *mockGCECloud) Storage() *storage.Service {
	klog.Fatalf("mockGCECloud::Storage not implemented")
	return nil
}

// IAM returns the IAM client
func (c *mockGCECloud) IAM() *iam.Service {
	klog.Fatalf("mockGCECloud::IAM not implemented")
	return nil
}

// NameService returns the DNS client
func (c *mockGCECloud) CloudDNS() *dns.Service {
	klog.Fatalf("mockGCECloud::CloudDNS not implemented")
	return nil
}

// WaitForOp implements GCECloud::WaitForOp
func (c *mockGCECloud) WaitForOp(op *compute.Operation) error {
	return fmt.Errorf("mockGCECloud::WaitForOp not implemented")
}

// FindClusterStatus implements GCECloud::FindClusterStatus
func (c *mockGCECloud) FindClusterStatus(cluster *kops.Cluster) (*kops.ClusterStatus, error) {
	return nil, fmt.Errorf("mockGCECloud::FindClusterStatus not implemented")
}

// GetApiIngressStatus implements GCECloud::GetApiIngressStatus
func (c *mockGCECloud) GetApiIngressStatus(cluster *kops.Cluster) ([]kops.ApiIngressStatus, error) {
	return nil, fmt.Errorf("mockGCECloud::GetApiIngressStatus not implemented")
}

// Region implements GCECloud::Region
func (c *mockGCECloud) Region() string {
	return c.region
}

// Project implements GCECloud::Project
func (c *mockGCECloud) Project() string {
	return c.region
}

// ServiceAccount implements GCECloud::ServiceAccount
func (c *mockGCECloud) ServiceAccount() (string, error) {
	return "12345678-compute@developer.gserviceaccount.com", nil
}

// Labels implements GCECloud::Labels
func (c *mockGCECloud) Labels() map[string]string {
	return c.labels
}

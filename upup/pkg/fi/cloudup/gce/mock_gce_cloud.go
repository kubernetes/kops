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
	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/storage/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	dnsproviderclouddns "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
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

// mockGCECloud returns a copy of the mockGCECloud bound to the specified labels
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
	glog.Fatalf("mockGCECloud::Compute not implemented")
	return nil
}

// Storage implements GCECloud::Storage
func (c *mockGCECloud) Storage() *storage.Service {
	glog.Fatalf("mockGCECloud::Storage not implemented")
	return nil
}

// WaitForOp implements GCECloud::WaitForOp
func (c *mockGCECloud) WaitForOp(op *compute.Operation) error {
	return fmt.Errorf("mockGCECloud::WaitForOp not implemented")
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

// Labels implements GCECloud::Labels
func (c *mockGCECloud) Labels() map[string]string {
	return c.labels
}

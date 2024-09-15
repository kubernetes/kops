/*
Copyright 2019 The Kubernetes Authors.

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
	"fmt"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipos "k8s.io/kops/protokube/pkg/gossip/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// OpenStackCloudProvider is the CloudProvider implementation for OpenStack
type OpenStackCloudProvider struct {
	cloud openstack.OpenstackCloud

	meta *openstack.InstanceMetadata

	clusterName  string
	project      string
	instanceName string
	storageZone  string
}

var _ CloudProvider = &OpenStackCloudProvider{}

// NewOpenStackCloudProvider builds a OpenStackCloudProvider
func NewOpenStackCloudProvider() (*OpenStackCloudProvider, error) {
	metadata, err := openstack.GetLocalMetadata()
	if err != nil {
		return nil, fmt.Errorf("Failed to get server metadata: %v", err)
	}

	oscloud, err := openstack.NewOpenstackCloud(nil, "protokube")
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize OpenStackCloudProvider: %v", err)
	}

	a := &OpenStackCloudProvider{
		cloud: oscloud,
		meta:  metadata,
	}

	err = a.discoverTags()
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Project returns the current OpenStack project
func (a *OpenStackCloudProvider) Project() string {
	return a.meta.ProjectID
}

func (a *OpenStackCloudProvider) discoverTags() error {
	// Cluster Name
	{
		a.clusterName = strings.TrimSpace(string(a.meta.UserMeta.ClusterName))
		if a.clusterName == "" {
			return fmt.Errorf("cluster name metadata was empty")
		}
		klog.Infof("Found cluster name=%q", a.clusterName)
	}

	// Project ID
	{
		a.project = strings.TrimSpace(a.meta.ProjectID)
		if a.project == "" {
			return fmt.Errorf("project metadata was empty")
		}
		klog.Infof("Found project=%q", a.project)
	}

	// Storage Availability Zone
	az, err := a.cloud.GetStorageAZFromCompute(a.meta.AvailabilityZone)
	if err != nil {
		return fmt.Errorf("Could not establish storage availability zone: %v", err)
	}
	a.storageZone = az.ZoneName
	klog.Infof("Found zone=%q", a.storageZone)

	// Instance Name
	{
		a.instanceName = strings.TrimSpace(a.meta.Name)
		if a.instanceName == "" {
			return fmt.Errorf("instance name metadata was empty")
		}
		klog.Infof("Found instanceName=%q", a.instanceName)
	}

	return nil
}

func (g *OpenStackCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	return gossipos.NewSeedProvider(g.cloud.ComputeClient(), g.clusterName, g.project)
}

func (g *OpenStackCloudProvider) InstanceID() string {
	return g.instanceName
}

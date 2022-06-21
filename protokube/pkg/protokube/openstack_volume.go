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
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipos "k8s.io/kops/protokube/pkg/gossip/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const MetadataLatest string = "http://169.254.169.254/openstack/latest/meta_data.json"

type Metadata struct {
	// Matches openstack.TagClusterName
	ClusterName string `json:"KubernetesCluster"`
}

type InstanceMetadata struct {
	Name             string    `json:"name"`
	UserMeta         *Metadata `json:"meta"`
	ProjectID        string    `json:"project_id"`
	AvailabilityZone string    `json:"availability_zone"`
	Hostname         string    `json:"hostname"`
	ServerID         string    `json:"uuid"`
}

// OpenStackCloudProvider is the CloudProvider implementation for OpenStack
type OpenStackCloudProvider struct {
	cloud openstack.OpenstackCloud

	meta *InstanceMetadata

	clusterName  string
	project      string
	instanceName string
	internalIP   net.IP
	storageZone  string
}

var _ CloudProvider = &OpenStackCloudProvider{}

func getLocalMetadata() (*InstanceMetadata, error) {
	var meta InstanceMetadata
	var client http.Client
	resp, err := client.Get(MetadataLatest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(bodyBytes, &meta)
		if err != nil {
			return nil, err
		}
		return &meta, nil
	}
	return nil, err
}

// NewOpenStackCloudProvider builds a OpenStackCloudProvider
func NewOpenStackCloudProvider() (*OpenStackCloudProvider, error) {
	metadata, err := getLocalMetadata()
	if err != nil {
		return nil, fmt.Errorf("Failed to get server metadata: %v", err)
	}

	tags := make(map[string]string)
	// Cluster name needed to bypass missing designate options
	tags[openstack.TagClusterName] = metadata.UserMeta.ClusterName

	oscloud, err := openstack.NewOpenstackCloud(tags, nil, "protokube")
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

// InstanceInternalIP implements CloudProvider InstanceInternalIP
func (a *OpenStackCloudProvider) InstanceInternalIP() net.IP {
	return a.internalIP
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

	// Internal IP
	{
		server, err := a.cloud.GetInstance(strings.TrimSpace(a.meta.ServerID))
		if err != nil {
			return fmt.Errorf("error getting instance from ID: %v", err)
		}
		// find kopsNetwork from metadata, fallback to clustername
		ifName := a.clusterName
		if val, ok := server.Metadata[openstack.TagKopsNetwork]; ok {
			ifName = val
		}
		ip, err := openstack.GetServerFixedIP(server, ifName)
		if err != nil {
			return fmt.Errorf("error querying InternalIP from name: %v", err)
		}
		a.internalIP = net.ParseIP(ip)
		klog.Infof("Found internalIP=%q", a.internalIP)
	}

	return nil
}

func (g *OpenStackCloudProvider) GossipSeeds() (gossip.SeedProvider, error) {
	return gossipos.NewSeedProvider(g.cloud.ComputeClient(), g.clusterName, g.project)
}

func (g *OpenStackCloudProvider) InstanceID() string {
	return g.instanceName
}

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
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	cinderv3 "github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/volumeattach"
	"k8s.io/klog"
	"k8s.io/kops/protokube/pkg/etcd"
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

// GCEVolumes is the Volumes implementation for GCE
type OpenstackVolumes struct {
	cloud openstack.OpenstackCloud

	meta *InstanceMetadata

	clusterName  string
	project      string
	instanceName string
	internalIP   net.IP
	storageZone  string
}

var _ Volumes = &OpenstackVolumes{}

func getLocalMetadata() (*InstanceMetadata, error) {
	var meta InstanceMetadata
	var client http.Client
	resp, err := client.Get(MetadataLatest)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
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

// NewOpenstackVolumes builds a OpenstackVolume
func NewOpenstackVolumes() (*OpenstackVolumes, error) {

	metadata, err := getLocalMetadata()
	if err != nil {
		return nil, fmt.Errorf("Failed to get server metadata: %v", err)
	}

	tags := make(map[string]string)
	// Cluster name needed to bypass missing designate options
	tags[openstack.TagClusterName] = metadata.UserMeta.ClusterName

	oscloud, err := openstack.NewOpenstackCloud(tags, nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize OpenstackVolumes: %v", err)
	}

	a := &OpenstackVolumes{
		cloud: oscloud,
		meta:  metadata,
	}

	err = a.discoverTags()
	if err != nil {
		return nil, err
	}

	return a, nil
}

// ClusterID implements Volumes ClusterID
func (a *OpenstackVolumes) ClusterID() string {
	return a.meta.UserMeta.ClusterName
}

// Project returns the current GCE project
func (a *OpenstackVolumes) Project() string {
	return a.meta.ProjectID
}

// InternalIP implements Volumes InternalIP
func (a *OpenstackVolumes) InternalIP() net.IP {
	return a.internalIP
}

func (a *OpenstackVolumes) discoverTags() error {

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

func (v *OpenstackVolumes) buildOpenstackVolume(d *cinderv3.Volume) (*Volume, error) {
	volumeName := d.Name
	vol := &Volume{
		ID: d.ID,
		Info: VolumeInfo{
			Description: volumeName,
		},
	}

	vol.Status = d.Status

	for _, attachedTo := range d.Attachments {
		vol.AttachedTo = attachedTo.HostName
		if attachedTo.ServerID == v.meta.ServerID {
			vol.LocalDevice = attachedTo.Device
		}
	}

	// FIXME: Zone matters, broken in my env

	for k, v := range d.Metadata {
		if strings.HasPrefix(k, openstack.TagNameEtcdClusterPrefix) {
			etcdClusterName := k[len(openstack.TagNameEtcdClusterPrefix):]
			spec, err := etcd.ParseEtcdClusterSpec(etcdClusterName, v)
			if err != nil {
				return nil, fmt.Errorf("error parsing etcd cluster meta %q on volume %q: %v", v, d.Name, err)
			}
			vol.Info.EtcdClusters = append(vol.Info.EtcdClusters, spec)
		}
	}

	return vol, nil
}

func (v *OpenstackVolumes) FindVolumes() ([]*Volume, error) {
	var volumes []*Volume

	klog.V(2).Infof("Listing Openstack disks in %s/%s", v.project, v.meta.AvailabilityZone)

	vols, err := v.cloud.ListVolumes(cinderv3.ListOpts{
		TenantID: v.project,
	})
	if err != nil {
		return volumes, fmt.Errorf("FindVolumes: Failed to list volume.")
	}

	for _, volume := range vols {
		if clusterName, ok := volume.Metadata[openstack.TagClusterName]; ok && clusterName == v.clusterName {
			if _, isMasterRole := volume.Metadata[openstack.TagNameRolePrefix+"master"]; isMasterRole {
				vol, err := v.buildOpenstackVolume(&volume)
				if err != nil {
					klog.Errorf("FindVolumes: Failed to build openstack volume %s: %v", volume.Name, err)
					continue
				}
				volumes = append(volumes, vol)
			}
		}
	}

	return volumes, nil
}

// FindMountedVolume implements Volumes::FindMountedVolume
func (v *OpenstackVolumes) FindMountedVolume(volume *Volume) (string, error) {
	device := volume.LocalDevice

	_, err := os.Stat(pathFor(device))
	if err == nil {
		return device, nil
	}
	if os.IsNotExist(err) {
		return "", nil
	}
	return "", fmt.Errorf("error checking for device %q: %v", device, err)
}

// AttachVolume attaches the specified volume to this instance, returning the mountpoint & nil if successful
func (v *OpenstackVolumes) AttachVolume(volume *Volume) error {
	opts := volumeattach.CreateOpts{
		VolumeID: volume.ID,
	}
	attachment, err := v.cloud.AttachVolume(v.meta.ServerID, opts)
	if err != nil {
		return fmt.Errorf("AttachVolume: failed to attach volume: %s", err)
	}
	volume.LocalDevice = attachment.Device
	return nil
}

func (g *OpenstackVolumes) GossipSeeds() (gossip.SeedProvider, error) {
	return gossipos.NewSeedProvider(g.cloud.ComputeClient(), g.clusterName, g.project)
}

func (g *OpenstackVolumes) InstanceName() string {
	return g.instanceName
}

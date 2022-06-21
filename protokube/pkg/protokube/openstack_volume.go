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
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipos "k8s.io/kops/protokube/pkg/gossip/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/mount-utils"
	utilexec "k8s.io/utils/exec"
)

const (
	// MetadataLatestPath is the path to the metadata on the config drive
	MetadataLatestPath string = "openstack/latest/meta_data.json"

	// MetadataID is the identifier for the metadata service
	MetadataID string = "metadataService"

	// MetadataLastestServiceURL points to the latest metadata of the metadata service
	MetadataLatestServiceURL string = "http://169.254.169.254/" + MetadataLatestPath

	// ConfigDriveID is the identifier for the config drive containing metadata
	ConfigDriveID string = "configDrive"

	// ConfigDriveLabel identifies the config drive by label on the OS
	ConfigDriveLabel string = "config-2"

	// DefaultMetadataSearchOrder defines the default order in which the metadata services are queried
	DefaultMetadataSearchOrder string = ConfigDriveID + ", " + MetadataID

	DiskByLabelPath string = "/dev/disk/by-label/"
)

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

type MetadataService struct {
	serviceURL      string
	configDrivePath string
	mounter         *mount.SafeFormatAndMount
	mountTarget     string
	searchOrder     string
}

var _ CloudProvider = &OpenStackCloudProvider{}

// getFromConfigDrive tries to get metadata by mounting a config drive and returns it as InstanceMetadata
// It will return an error if there is no disk labelled as ConfigDriveLabel or other errors while mounting the disk, or reading the file occur.
func (mds MetadataService) getFromConfigDrive() (*InstanceMetadata, error) {
	dev := path.Join(DiskByLabelPath, ConfigDriveLabel)
	if _, err := os.Stat(dev); os.IsNotExist(err) {
		out, err := mds.mounter.Exec.Command(
			"blkid", "-l",
			"-t", fmt.Sprintf("LABEL=%s", ConfigDriveLabel),
			"-o", "device",
		).CombinedOutput()
		if err != nil {
			return nil, fmt.Errorf("unable to run blkid: %v", err)
		}
		dev = strings.TrimSpace(string(out))
	}

	err := mds.mounter.Mount(dev, mds.mountTarget, "iso9660", []string{"ro"})
	if err != nil {
		err = mds.mounter.Mount(dev, mds.mountTarget, "vfat", []string{"ro"})
	}
	if err != nil {
		return nil, fmt.Errorf("error mounting configdrive '%s': %v", dev, err)
	}
	defer mds.mounter.Unmount(mds.mountTarget)

	f, err := os.Open(
		path.Join(mds.mountTarget, mds.configDrivePath))
	if err != nil {
		return nil, fmt.Errorf("error reading '%s' on config drive: %v", mds.configDrivePath, err)
	}
	defer f.Close()

	return mds.parseMetadata(f)
}

// getFromMetadataService tries to get metadata from a metadata service endpoint and returns it as InstanceMetadata.
// If the service endpoint cannot be contacted or reports a different status than StatusOK it will return an error.
func (mds MetadataService) getFromMetadataService() (*InstanceMetadata, error) {
	var client http.Client

	resp, err := client.Get(mds.serviceURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return mds.parseMetadata(resp.Body)
	}

	err = fmt.Errorf("fetching metadata from '%s' returned status code '%d'", mds.serviceURL, resp.StatusCode)
	return nil, err
}

// parseMetadata reads JSON data from a Reader and returns it as InstanceMetadata.
func (mds MetadataService) parseMetadata(r io.Reader) (*InstanceMetadata, error) {
	var meta InstanceMetadata

	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

// getMetadata tries to get metadata for the instance by mounting the config drive and/or querying the metadata service endpoint.
// Depending on the searchOrder it will return data from the first source which successfully returns.
// If all the sources in searchOrder are erroneous it will propagate the last error to its caller.
func (mds MetadataService) getMetadata() (*InstanceMetadata, error) {
	// Note(ederst): I used and modified code for getting the config drive metadata to work from here:
	//   * https://github.com/kubernetes/cloud-provider-openstack/blob/27b6fc483451b6df2112a6a4a40a34ffc9093635/pkg/util/metadata/metadata.go

	var meta *InstanceMetadata
	var err error

	ids := strings.Split(mds.searchOrder, ",")
	for _, id := range ids {
		id = strings.TrimSpace(id)
		switch id {
		case ConfigDriveID:
			meta, err = mds.getFromConfigDrive()
		case MetadataID:
			meta, err = mds.getFromMetadataService()
		default:
			err = fmt.Errorf("%s is not a valid metadata search order option. Supported options are %s and %s", id, ConfigDriveID, MetadataID)
		}

		if err == nil {
			break
		}
	}

	return meta, err
}

func newMetadataService(serviceURL string, configDrivePath string, mounter *mount.SafeFormatAndMount, mountTarget string, searchOrder string) *MetadataService {
	return &MetadataService{
		serviceURL:      serviceURL,
		configDrivePath: configDrivePath,
		mounter:         mounter,
		mountTarget:     mountTarget,
		searchOrder:     searchOrder,
	}
}

// getDefaultMounter returns a mount and executor interface to use for getting metadata from a config drive
func getDefaultMounter() *mount.SafeFormatAndMount {
	mounter := mount.New("")
	exec := utilexec.New()
	return &mount.SafeFormatAndMount{
		Interface: mounter,
		Exec:      exec,
	}
}

func getLocalMetadata() (*InstanceMetadata, error) {
	mountTarget, err := ioutil.TempDir("", "configdrive")
	if err != nil {
		return nil, err
	}
	defer os.Remove(mountTarget)

	return newMetadataService(MetadataLatestServiceURL, MetadataLatestPath, getDefaultMounter(), mountTarget, DefaultMetadataSearchOrder).getMetadata()
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

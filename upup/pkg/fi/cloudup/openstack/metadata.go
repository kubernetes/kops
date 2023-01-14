/*
Copyright 2023 The Kubernetes Authors.

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

package openstack

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

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

type MetadataService struct {
	serviceURL      string
	configDrivePath string
	mounter         *mount.SafeFormatAndMount
	mountTarget     string
	searchOrder     string
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

// GetLocalMetadata returns a local metadata for the server
func GetLocalMetadata() (*InstanceMetadata, error) {
	mountTarget, err := ioutil.TempDir("", "configdrive")
	if err != nil {
		return nil, err
	}
	defer os.Remove(mountTarget)

	return newMetadataService(MetadataLatestServiceURL, MetadataLatestPath, getDefaultMounter(), mountTarget, DefaultMetadataSearchOrder).getMetadata()
}

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

// getDefaultMounter returns a mount and executor interface to use for getting metadata from a config drive
func getDefaultMounter() *mount.SafeFormatAndMount {
	mounter := mount.New("")
	exec := utilexec.New()
	return &mount.SafeFormatAndMount{
		Interface: mounter,
		Exec:      exec,
	}
}

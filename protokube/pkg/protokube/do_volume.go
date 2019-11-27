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
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	"k8s.io/kops/pkg/resources/digitalocean"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipdo "k8s.io/kops/protokube/pkg/gossip/do"
)

const (
	dropletRegionMetadataURL     = "http://169.254.169.254/metadata/v1/region"
	dropletNameMetadataURL       = "http://169.254.169.254/metadata/v1/hostname"
	dropletIDMetadataURL         = "http://169.254.169.254/metadata/v1/id"
	dropletIDMetadataTags        = "http://169.254.169.254/metadata/v1/tags"
	dropletInternalIPMetadataURL = "http://169.254.169.254/metadata/v1/interfaces/private/0/ipv4/address"
	localDevicePrefix            = "/dev/disk/by-id/scsi-0DO_Volume_"
)

type DOVolumes struct {
	ClusterID string
	Cloud     *digitalocean.Cloud

	region      string
	dropletName string
	dropletID   int
	dropletTags []string
}

var _ Volumes = &DOVolumes{}

func GetClusterID() (string, error) {
	var clusterID = ""

	dropletTags, err := getMetadataDropletTags()
	if err != nil {
		return clusterID, fmt.Errorf("GetClusterID failed - unable to retrieve droplet tags: %s", err)
	}

	for _, dropletTag := range dropletTags {
		if strings.Contains(dropletTag, "KubernetesCluster:") {
			clusterID = strings.Replace(dropletTag, ".", "-", -1)

			tokens := strings.Split(clusterID, ":")
			if len(tokens) != 2 {
				return clusterID, fmt.Errorf("invalid clusterID (expected two tokens): %q", clusterID)
			}

			clusterID := tokens[1]

			return clusterID, nil
		}
	}

	return clusterID, fmt.Errorf("failed to get droplet clusterID")
}

func NewDOVolumes() (*DOVolumes, error) {
	region, err := getMetadataRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet region: %s", err)
	}

	dropletID, err := getMetadataDropletID()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet id: %s", err)
	}

	dropletIDInt, err := strconv.Atoi(dropletID)
	if err != nil {
		return nil, fmt.Errorf("failed to convert droplet ID to int: %s", err)
	}

	dropletName, err := getMetadataDropletName()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet name: %s", err)
	}

	cloud, err := digitalocean.NewCloud(region)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize digitalocean cloud: %s", err)
	}

	dropletTags, err := getMetadataDropletTags()
	if err != nil {
		return nil, fmt.Errorf("failed to get droplet tags: %s", err)
	}

	clusterID, err := GetClusterID()
	if err != nil {
		return nil, fmt.Errorf("failed to get clusterID: %s", err)
	}

	return &DOVolumes{
		Cloud:       cloud,
		ClusterID:   clusterID,
		dropletID:   dropletIDInt,
		dropletName: dropletName,
		region:      region,
		dropletTags: dropletTags,
	}, nil
}

func (d *DOVolumes) AttachVolume(volume *Volume) error {
	for {
		action, _, err := d.Cloud.VolumeActions().Attach(context.TODO(), volume.ID, d.dropletID)
		if err != nil {
			return fmt.Errorf("error attaching volume: %s", err)
		}

		if action.Status != godo.ActionInProgress && action.Status != godo.ActionCompleted {
			return fmt.Errorf("invalid status for digitalocean volume: %s", volume.ID)
		}

		doVolume, err := d.getVolumeByID(volume.ID)
		if err != nil {
			return fmt.Errorf("error getting volume status: %s", err)
		}

		if len(doVolume.DropletIDs) == 1 {
			if doVolume.DropletIDs[0] != d.dropletID {
				return fmt.Errorf("digitalocean volume %s is attached to another droplet", doVolume.ID)
			}

			volume.LocalDevice = getLocalDeviceName(doVolume)
			return nil
		}

		time.Sleep(10 * time.Second)
	}
}

func (d *DOVolumes) FindVolumes() ([]*Volume, error) {
	doVolumes, err := getAllVolumesByRegion(d.Cloud, d.region)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %s", err)
	}

	var volumes []*Volume
	for _, doVolume := range doVolumes {
		// determine if this volume belongs to this cluster
		// check for string d.ClusterID but with strings "." replaced with "-"
		if !strings.Contains(doVolume.Name, strings.Replace(d.ClusterID, ".", "-", -1)) {
			continue
		}

		vol := &Volume{
			ID: doVolume.ID,
			Info: VolumeInfo{
				Description: doVolume.Description,
			},
		}

		if len(doVolume.DropletIDs) == 1 {
			vol.AttachedTo = strconv.Itoa(doVolume.DropletIDs[0])
			vol.LocalDevice = getLocalDeviceName(&doVolume)
		}

		etcdClusterSpec, err := d.getEtcdClusterSpec(doVolume)
		if err != nil {
			return nil, fmt.Errorf("failed to get etcd cluster spec: %s", err)
		}

		vol.Info.EtcdClusters = append(vol.Info.EtcdClusters, etcdClusterSpec)
		volumes = append(volumes, vol)
	}

	return volumes, nil
}

func getAllVolumesByRegion(cloud *digitalocean.Cloud, region string) ([]godo.Volume, error) {
	allVolumes := []godo.Volume{}

	opt := &godo.ListOptions{}
	for {
		volumes, resp, err := cloud.Volumes().ListVolumes(context.TODO(), &godo.ListVolumeParams{
			Region:      region,
			ListOptions: opt,
		})

		if err != nil {
			return nil, err
		}

		allVolumes = append(allVolumes, volumes...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allVolumes, nil

}

func (d *DOVolumes) FindMountedVolume(volume *Volume) (string, error) {
	device := volume.LocalDevice

	_, err := os.Stat(pathFor(device))
	if err == nil {
		return device, nil
	}

	if !os.IsNotExist(err) {
		return "", fmt.Errorf("error checking for device %q: %v", device, err)
	}

	return "", nil
}

func (d *DOVolumes) getVolumeByID(id string) (*godo.Volume, error) {
	vol, _, err := d.Cloud.Volumes().GetVolume(context.TODO(), id)
	return vol, err

}

// getEtcdClusterSpec returns etcd.EtcdClusterSpec which holds
// necessary information required for starting an etcd server.
// DigitalOcean support on kops only supports single master setup for now
// but in the future when it supports multiple masters this method be
// updated to handle that case.
// TODO: use tags once it's supported for volumes
func (d *DOVolumes) getEtcdClusterSpec(vol godo.Volume) (*etcd.EtcdClusterSpec, error) {
	nodeName := d.dropletName

	var clusterKey string
	if strings.Contains(vol.Name, "etcd-main") {
		clusterKey = "main"
	} else if strings.Contains(vol.Name, "etcd-events") {
		clusterKey = "events"
	} else {
		return nil, fmt.Errorf("could not determine etcd cluster type for volume: %s", vol.Name)
	}

	return &etcd.EtcdClusterSpec{
		ClusterKey: clusterKey,
		NodeName:   nodeName,
		NodeNames:  []string{nodeName},
	}, nil
}

func getLocalDeviceName(vol *godo.Volume) string {
	return localDevicePrefix + vol.Name
}

func (d *DOVolumes) GossipSeeds() (gossip.SeedProvider, error) {
	for _, dropletTag := range d.dropletTags {
		if strings.Contains(dropletTag, strings.Replace(d.ClusterID, ".", "-", -1)) {
			return gossipdo.NewSeedProvider(d.Cloud, dropletTag)
		}
	}

	return nil, fmt.Errorf("could not determine a matching droplet tag for gossip seeding")
}

func (d *DOVolumes) InstanceName() string {
	return d.dropletName
}

// GetDropletInternalIP gets the private IP of the droplet running this program
// This function is exported so it can be called from protokube
func GetDropletInternalIP() (net.IP, error) {
	addr, err := getMetadata(dropletInternalIPMetadataURL)
	if err != nil {
		return nil, err
	}

	return net.ParseIP(addr), nil
}

func getMetadataRegion() (string, error) {
	return getMetadata(dropletRegionMetadataURL)
}

func getMetadataDropletName() (string, error) {
	return getMetadata(dropletNameMetadataURL)
}

func getMetadataDropletID() (string, error) {
	return getMetadata(dropletIDMetadataURL)
}

func getMetadataDropletTags() ([]string, error) {

	tagString, err := getMetadata(dropletIDMetadataTags)
	return strings.Split(tagString, "\n"), err
}

func getMetadata(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("droplet metadata returned non-200 status code: %d", resp.StatusCode)
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(bodyBytes), nil
}

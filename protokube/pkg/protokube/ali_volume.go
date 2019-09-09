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
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/denverdino/aliyungo/common"
	"github.com/denverdino/aliyungo/ecs"
	"github.com/denverdino/aliyungo/metadata"
	"k8s.io/klog"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipali "k8s.io/kops/protokube/pkg/gossip/ali"
	"k8s.io/kops/upup/pkg/fi/cloudup/aliup"
)

// ALIVolumes is the Volumes implementation for Aliyun ECS
type ALIVolumes struct {
	client *ecs.Client

	clusterTag string
	region     string
	zone       string
	instanceId string
	internalIP net.IP
}

var _ Volumes = &ALIVolumes{}

func NewALIVolumes() (*ALIVolumes, error) {
	accessKeyId := os.Getenv("ALIYUN_ACCESS_KEY_ID")
	if accessKeyId == "" {
		return nil, fmt.Errorf("error initialing ALIVolumes: ALIYUN_ACCESS_KEY_ID cannot be empty")
	}
	accessKeySecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
	if accessKeySecret == "" {
		return nil, fmt.Errorf("error initialing ALIVolumes: ALIYUN_ACCESS_KEY_SECRET cannot be empty")
	}
	ecsEndpoint := os.Getenv("ALIYUN_ECS_ENDPOINT")
	if ecsEndpoint == "" {
		// TODO: shall we raise error here?
		ecsEndpoint = ecs.ECSDefaultEndpoint
	}

	client := ecs.NewClientWithEndpoint(ecsEndpoint, accessKeyId, accessKeySecret)
	a := &ALIVolumes{
		client: client,
	}

	err := a.discoverTags()
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ClusterID implements Volumes ClusterID
func (a *ALIVolumes) ClusterID() string {
	return a.clusterTag
}

// InstanceID implements Volumes InstanceID
func (a *ALIVolumes) InstanceID() string {
	return a.instanceId
}

// InternalIP implements Volumes InternalIP
func (a *ALIVolumes) InternalIP() net.IP {
	return a.internalIP
}

func (a *ALIVolumes) discoverTags() error {
	metadataClient := metadata.NewMetaData(&http.Client{})
	// Region
	{
		region, err := metadataClient.Region()
		if err != nil {
			return fmt.Errorf("error reading region from Aliyun: %v", err)
		}
		a.region = region
		if a.region == "" {
			return fmt.Errorf("region metadata was empty")
		}
		klog.Infof("Found region=%q", a.region)
	}

	// Zone
	{
		zone, err := metadataClient.Zone()
		if err != nil {
			return fmt.Errorf("error reading zone from Aliyun: %v", err)
		}
		a.zone = zone
		if a.zone == "" {
			return fmt.Errorf("zone metadata was empty")
		}
		klog.Infof("Found zone=%q", a.zone)
	}

	// Instance Name
	{
		instanceId, err := metadataClient.InstanceID()
		if err != nil {
			return fmt.Errorf("error reading instance ID from Aliyun: %v", err)
		}
		a.instanceId = instanceId
		if a.instanceId == "" {
			return fmt.Errorf("instance ID metadata was empty")
		}
		klog.Infof("Found instanceId=%q", a.instanceId)
	}

	// Internal IP
	{
		internalIP, err := metadataClient.PrivateIPv4()
		if err != nil {
			return fmt.Errorf("error querying InternalIP from Aliyun: %v", err)
		}
		if internalIP == "" {
			return fmt.Errorf("InternalIP from metadata was empty")
		}
		a.internalIP = net.ParseIP(internalIP)
		if a.internalIP == nil {
			return fmt.Errorf("InternalIP from metadata was not parseable(%q)", internalIP)
		}
		klog.Infof("Found internalIP=%q", a.internalIP)
	}

	// Cluster Tag
	{
		describeTagsArgs := &ecs.DescribeTagsArgs{
			RegionId:     common.Region(a.region),
			ResourceType: ecs.TagResourceInstance,
			ResourceId:   a.instanceId,
		}
		result, _, err := a.client.DescribeTags(describeTagsArgs)
		if err != nil {
			return fmt.Errorf("error querying Aliyun instance tags: %v", err)
		}
		for _, tag := range result {
			if tag.TagKey == aliup.TagClusterName {
				a.clusterTag = tag.TagValue
			}
		}
		if a.clusterTag == "" {
			return fmt.Errorf("cluster tag metadata was empty")
		}
	}

	return nil
}

// AttachVolume attaches the specified volume to this instance, returning the mountpoint & nil if successful
func (a *ALIVolumes) AttachVolume(volume *Volume) error {
	// TODO: what if this volume has already been attached to another instance?
	// Aliyun Disk can only be attached to one instance
	if volume.LocalDevice == "" && volume.AttachedTo == "" {
		attachDiskArgs := &ecs.AttachDiskArgs{
			InstanceId: a.instanceId,
			DiskId:     volume.ID,
			// TODO: DeleteWithInstance?
		}
		err := a.client.AttachDisk(attachDiskArgs)
		if err != nil {
			return fmt.Errorf("error attach disk %q: %v", volume.ID, err)
		}

		// TODO: Do we have to wait for attach to complete?
		// retrieve device info
		args := &ecs.DescribeDisksArgs{
			RegionId: common.Region(a.region),
			ZoneId:   a.zone,
			DiskIds:  []string{volume.ID},
		}
		disks, _, err := a.client.DescribeDisks(args)
		if err != nil || len(disks) == 0 {
			return fmt.Errorf("error querying Aliyun disk %q: %v", volume.ID, err)
		}

		volume.LocalDevice = disks[0].Device
		volume.AttachedTo = a.instanceId
	} else if volume.AttachedTo != a.instanceId {
		return fmt.Errorf("cannot reattach an attached disk without detaching it first")
	}
	return nil
}

func (a *ALIVolumes) FindVolumes() ([]*Volume, error) {
	klog.V(2).Infof("Listing Aliyun disks in %s", a.zone)

	var volumes []*Volume

	var disks []ecs.DiskItemType
	// We could query at most 50 disks at a time on Aliyun ECS
	maxPageSize := 50
	tags := make(map[string]string)
	tags[aliup.TagClusterName] = a.clusterTag
	tags[aliup.TagNameRolePrefix+"master"] = "1"
	args := &ecs.DescribeDisksArgs{
		RegionId: common.Region(a.region),
		ZoneId:   a.zone,
		Tag:      tags,
		Pagination: common.Pagination{
			PageNumber: 1,
			PageSize:   maxPageSize,
		},
	}
	for {
		resp, page, err := a.client.DescribeDisks(args)
		if err != nil {
			return nil, fmt.Errorf("error querying Aliyun disks: %v", err)
		}
		disks = append(disks, resp...)

		if page.NextPage() == nil {
			break
		}
		args.Pagination = *(page.NextPage())
	}

	for _, disk := range disks {
		volume := &Volume{
			ID: disk.DiskId,
			Info: VolumeInfo{
				Description: disk.Description,
			},
			Status:     string(disk.Status),
			AttachedTo: disk.InstanceId,
		}
		if volume.AttachedTo == a.instanceId {
			volume.LocalDevice = disk.Device
		}

		describeTagsArgs := &ecs.DescribeTagsArgs{
			RegionId:     common.Region(a.region),
			ResourceType: ecs.TagResourceDisk,
			ResourceId:   disk.DiskId,
		}
		result, _, err := a.client.DescribeTags(describeTagsArgs)
		if err != nil {
			return nil, fmt.Errorf("error querying Aliyun disk tags: %v", err)
		}

		skipVolume := false
		for _, tag := range result {
			switch tag.TagKey {
			case aliup.TagClusterName:
				{
					// Ignore
				}
			default:
				if strings.HasPrefix(tag.TagKey, aliup.TagNameEtcdClusterPrefix) {
					etcdClusterName := strings.TrimPrefix(tag.TagKey, aliup.TagNameEtcdClusterPrefix)
					spec, err := etcd.ParseEtcdClusterSpec(etcdClusterName, tag.TagValue)
					if err != nil {
						// Fail safe
						klog.Warningf("error parsing etcd cluster tag %q on volume %q; skipping volume: %v", tag.TagValue, volume.ID, err)
						skipVolume = true
					}
					volume.Info.EtcdClusters = append(volume.Info.EtcdClusters, spec)
				} else if strings.HasPrefix(tag.TagKey, aliup.TagNameRolePrefix) {
					// Ignore
				} else {
					klog.Warningf("unknown tag on volume %q: %s=%s", volume.ID, tag.TagKey, tag.TagValue)
				}
			}
		}
		if !skipVolume {
			volumes = append(volumes, volume)
		}
	}
	return volumes, nil
}

// FindMountedVolume implements Volumes::FindMountedVolume
func (a *ALIVolumes) FindMountedVolume(volume *Volume) (string, error) {
	device := volume.LocalDevice

	_, err := os.Stat(pathFor(device))
	if err == nil {
		return device, nil
	}
	if os.IsNotExist(err) {
		if strings.HasPrefix(device, "/dev/xvd") {
			device = "/dev/vd" + strings.TrimPrefix(device, "/dev/xvd")
			_, err = os.Stat(pathFor(device))
			return device, err
		} else if strings.HasPrefix(device, "/dev/vd") {
			device = "/dev/xvd" + strings.TrimPrefix(device, "/dev/vd")
			_, err = os.Stat(pathFor(device))
			return device, err
		}
		return "", nil
	}
	return "", fmt.Errorf("error checking for device %q: %v", device, err)
}

func (a *ALIVolumes) GossipSeeds() (gossip.SeedProvider, error) {
	tags := make(map[string]string)
	tags[aliup.TagClusterName] = a.clusterTag

	return gossipali.NewSeedProvider(a.client, a.region, tags)
}

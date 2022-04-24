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
	"net"
	"os"
	"strings"

	"cloud.google.com/go/compute/metadata"
	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/protokube/pkg/gossip"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcediscovery"
)

// GCEVolumes is the Volumes implementation for GCE
type GCEVolumes struct {
	compute   *compute.Service
	discovery *gcediscovery.Discovery

	project      string
	zone         string
	region       string
	clusterName  string
	instanceName string
	internalIP   net.IP
}

var _ Volumes = &GCEVolumes{}

// NewGCEVolumes builds a GCEVolumes
func NewGCEVolumes() (*GCEVolumes, error) {
	discovery, err := gcediscovery.New()
	if err != nil {
		return nil, err
	}

	a := &GCEVolumes{
		discovery: discovery,
		compute:   discovery.Compute(),
	}

	err = a.discoverTags()
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Project returns the current GCE project
func (a *GCEVolumes) Project() string {
	return a.project
}

// InternalIP implements Volumes InternalIP
func (a *GCEVolumes) InternalIP() net.IP {
	return a.internalIP
}

func (a *GCEVolumes) discoverTags() error {
	// Cluster Name
	{
		a.clusterName = a.discovery.ClusterName()
		if a.clusterName == "" {
			return fmt.Errorf("cluster-name metadata was empty")
		}
	}

	// Project ID
	{
		a.project = a.discovery.ProjectID()
		if a.project == "" {
			return fmt.Errorf("project metadata was empty")
		}
		klog.Infof("Found project=%q", a.project)
	}

	// Zone
	{
		a.zone = a.discovery.Zone()
		if a.zone == "" {
			return fmt.Errorf("zone metadata was empty")
		}
		klog.Infof("Found zone=%q", a.zone)

		a.region = a.discovery.Region()
		klog.Infof("Found region=%q", a.region)
	}

	// Instance Name
	{
		instanceName, err := metadata.InstanceName()
		if err != nil {
			return fmt.Errorf("error reading instance name from GCE: %v", err)
		}
		a.instanceName = strings.TrimSpace(instanceName)
		if a.instanceName == "" {
			return fmt.Errorf("instance name metadata was empty")
		}
		klog.Infof("Found instanceName=%q", a.instanceName)
	}

	// Internal IP
	{
		internalIP, err := metadata.InternalIP()
		if err != nil {
			return fmt.Errorf("error querying InternalIP from GCE: %v", err)
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

	return nil
}

func (v *GCEVolumes) buildGCEVolume(d *compute.Disk) (*Volume, error) {
	volumeName := d.Name
	vol := &Volume{
		ID: volumeName,
		Info: VolumeInfo{
			Description: volumeName,
		},
	}

	vol.Status = d.Status

	for _, attachedTo := range d.Users {
		u, err := gce.ParseGoogleCloudURL(attachedTo)
		if err != nil {
			return nil, fmt.Errorf("error parsing disk attachment url %q: %v", attachedTo, err)
		}

		vol.AttachedTo = u.Name

		if u.Project == v.project && u.Zone == v.zone && u.Name == v.instanceName {
			devicePath := "/dev/disk/by-id/google-" + volumeName
			vol.LocalDevice = devicePath
			klog.V(2).Infof("volume %q is attached to this instance at %s", d.Name, devicePath)
		} else {
			klog.V(2).Infof("volume %q is attached to another instance %q", d.Name, attachedTo)
		}
	}

	for k, v := range d.Labels {
		switch k {
		case gce.GceLabelNameKubernetesCluster:
			{
				// Ignore
			}

		default:
			if strings.HasPrefix(k, gce.GceLabelNameEtcdClusterPrefix) {
				etcdClusterName := k[len(gce.GceLabelNameEtcdClusterPrefix):]

				value, err := gce.DecodeGCELabel(v)
				if err != nil {
					return nil, fmt.Errorf("Error decoding GCE label: %s=%q", k, v)
				}
				spec, err := etcd.ParseEtcdClusterSpec(etcdClusterName, value)
				if err != nil {
					return nil, fmt.Errorf("error parsing etcd cluster label %q on volume %q: %v", value, volumeName, err)
				}
				vol.Info.EtcdClusters = append(vol.Info.EtcdClusters, spec)
			} else if strings.HasPrefix(k, gce.GceLabelNameRolePrefix) {
				// Ignore
			} else {
				klog.Warningf("unknown label on volume %q: %s=%s", volumeName, k, v)
			}
		}
	}

	return vol, nil
}

func (v *GCEVolumes) FindVolumes() ([]*Volume, error) {
	var volumes []*Volume

	klog.V(2).Infof("Listing GCE disks in %s/%s", v.project, v.zone)

	// TODO: Apply filters
	ctx := context.Background()
	err := v.compute.Disks.List(v.project, v.zone).Pages(ctx, func(page *compute.DiskList) error {
		for _, d := range page.Items {
			klog.V(4).Infof("Found disk %q with labels %v", d.Name, d.Labels)

			diskClusterName := d.Labels[gce.GceLabelNameKubernetesCluster]
			if diskClusterName == "" {
				klog.V(4).Infof("Skipping disk %q with no cluster name", d.Name)
				continue
			}
			// Note that the cluster name is _not_ encoded with EncodeGCELabel
			// this is because it is also used by k8s itself, e.g. in the route controller,
			// and that is not encoded (issue #28436)
			// Instead we use the much simpler SafeClusterName sanitizer
			findClusterName := gce.SafeClusterName(v.clusterName)
			if diskClusterName != findClusterName {
				klog.V(2).Infof("Skipping disk %q with cluster name that does not match: %s=%s (looking for %s)", d.Name, gce.GceLabelNameKubernetesCluster, diskClusterName, findClusterName)
				continue
			}

			roles := make(map[string]string)
			for k, v := range d.Labels {
				if strings.HasPrefix(k, gce.GceLabelNameRolePrefix) {
					roleName := strings.TrimPrefix(k, gce.GceLabelNameRolePrefix)

					value, err := gce.DecodeGCELabel(v)
					if err != nil {
						klog.Warningf("error decoding GCE role label: %s=%s", k, v)
						continue
					}
					roles[roleName] = value
				}
			}

			_, isMaster := roles["master"]
			if !isMaster {
				klog.V(2).Infof("Skipping disk %q - no master role", d.Name)
				continue
			}

			vol, err := v.buildGCEVolume(d)
			if err != nil {
				// Fail safe
				klog.Warningf("skipping malformed volume %q: %v", d.Name, err)
				continue
			}
			volumes = append(volumes, vol)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error querying GCE disks: %v", err)
	}

	//instance, err := v.compute.Instances.Get(v.project, v.zone, v.instanceName).Do()
	//for _, d := range instance.Disks {
	//	var found *Volume
	//	source := d.Source
	//	for _, v := range volumes {
	//		if v.ID == source {
	//			if found != nil {
	//				return nil, fmt.Errorf("Found multiple volumes with name %q", v.ID)
	//			}
	//			found = v
	//		}
	//	}
	//
	//	if found != nil {
	//		if d.DeviceName == "" {
	//			return fmt.Errorf("DeviceName for mounted disk %q was unexpected empty", d.Source)
	//		}
	//		found.LocalDevice = d.DeviceName
	//	}
	//}

	return volumes, nil
}

// FindMountedVolume implements Volumes::FindMountedVolume
func (v *GCEVolumes) FindMountedVolume(volume *Volume) (string, error) {
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
func (v *GCEVolumes) AttachVolume(volume *Volume) error {
	volumeName := volume.ID

	volumeURL := gce.GoogleCloudURL{
		Project: v.project,
		Zone:    v.zone,
		Name:    volumeName,
		Type:    "disks",
	}

	attachedDisk := &compute.AttachedDisk{
		DeviceName: volumeName,
		// TODO: The k8s GCE provider sets Kind, but this seems wrong.  Open an issue?
		// Kind:       disk.Kind,
		Mode:   "READ_WRITE",
		Source: volumeURL.BuildURL(),
		Type:   "PERSISTENT",
	}

	attachOp, err := v.compute.Instances.AttachDisk(v.project, v.zone, v.instanceName, attachedDisk).Do()
	if err != nil {
		return fmt.Errorf("error attach disk %q: %v", volumeName, err)
	}

	err = gce.WaitForOp(v.compute, attachOp)
	if err != nil {
		return fmt.Errorf("error waiting for disk attach to complete %q: %v", volumeName, err)
	}

	devicePath := "/dev/disk/by-id/google-" + volumeName

	// TODO: Wait for device to appear?

	volume.LocalDevice = devicePath

	return nil
}

func (g *GCEVolumes) GossipSeeds() (gossip.SeedProvider, error) {
	return g.discovery, nil
}

func (g *GCEVolumes) InstanceName() string {
	return g.instanceName
}

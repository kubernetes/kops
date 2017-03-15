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

package protokube

import (
	"errors"
	"github.com/golang/glog"
	"net"
)

const EtcdDataKey = "01"
const EtcdDataVolPath = "/mnt/master-" + EtcdDataKey
const EtcdEventKey = "02"
const EtcdEventVolPath = "/mnt/master-" + EtcdEventKey

// TODO Use lsblk or counterpart command to find the actual device details.
const LocalDeviceForDataVol = "/dev/sdb1"
const LocalDeviceForEventsVol = "/dev/sdc1"
const VolStatusValue = "attached"
const EtcdNodeName = "a"
const EtcdClusterName = "main"
const EtcdEventsClusterName = "events"

type VSphereVolumes struct {
	// Dummy property. Not getting used any where for now.
	paths map[string]string
}

var _ Volumes = &VSphereVolumes{}
var machineIp net.IP

func NewVSphereVolumes() (*VSphereVolumes, error) {
	vsphereVolumes := &VSphereVolumes{
		paths: make(map[string]string),
	}
	vsphereVolumes.paths[EtcdDataKey] = EtcdDataVolPath
	vsphereVolumes.paths[EtcdEventKey] = EtcdEventVolPath
	return vsphereVolumes, nil
}

func (v *VSphereVolumes) FindVolumes() ([]*Volume, error) {
	var volumes []*Volume
	ip := v.InternalIp()
	attachedTo := ""
	if ip != nil {
		attachedTo = ip.String()
	}

	// etcd data volume and etcd cluster spec.
	{
		vol := &Volume{
			ID:          EtcdDataKey,
			LocalDevice: LocalDeviceForDataVol,
			AttachedTo:  attachedTo,
			Mountpoint:  EtcdDataVolPath,
			Status:      VolStatusValue,
			Info: VolumeInfo{
				Description: EtcdClusterName,
			},
		}
		etcdSpec := &EtcdClusterSpec{
			ClusterKey: EtcdClusterName,
			NodeName:   EtcdNodeName,
			NodeNames:  []string{EtcdNodeName},
		}
		vol.Info.EtcdClusters = []*EtcdClusterSpec{etcdSpec}
		volumes = append(volumes, vol)
	}

	// etcd events volume and etcd events cluster spec.
	{
		vol := &Volume{
			ID:          EtcdEventKey,
			LocalDevice: LocalDeviceForEventsVol,
			AttachedTo:  attachedTo,
			Mountpoint:  EtcdEventVolPath,
			Status:      VolStatusValue,
			Info: VolumeInfo{
				Description: EtcdEventsClusterName,
			},
		}
		etcdSpec := &EtcdClusterSpec{
			ClusterKey: EtcdEventsClusterName,
			NodeName:   EtcdNodeName,
			NodeNames:  []string{EtcdNodeName},
		}
		vol.Info.EtcdClusters = []*EtcdClusterSpec{etcdSpec}
		volumes = append(volumes, vol)
	}
	glog.Infof("Found volumes: %v", volumes)
	return volumes, nil
}

func (v *VSphereVolumes) AttachVolume(volume *Volume) error {
	// Currently this is a no-op for vSphere. The virtual disks should already be mounted on this VM.
	glog.Infof("All volumes should already be attached. No operation done.")
	return nil
}

func (v *VSphereVolumes) InternalIp() net.IP {
	if machineIp == nil {
		ip, err := getMachineIp()
		if err != nil {
			return ip
		}
		machineIp = ip
	}
	return machineIp
}

func getMachineIp() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip, nil
		}
	}
	return nil, errors.New("No IP found.")
}

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

// vsphere_volume houses vSphere volume and implements relevant interfaces.

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"k8s.io/klog"
	etcdmanager "k8s.io/kops/protokube/pkg/etcd"
	"k8s.io/kops/upup/pkg/fi/cloudup/vsphere"
)

const (
	VolumeMetaDataFile = "/vol-metadata/metadata.json"
	VolStatusValue     = "attached"
)

// VSphereVolumes represents vSphere volume and implements Volumes interface.
type VSphereVolumes struct{}

var _ Volumes = &VSphereVolumes{}
var machineIp net.IP

// NewVSphereVolumes returns instance of VSphereVolumes type.
func NewVSphereVolumes() (*VSphereVolumes, error) {
	vsphereVolumes := &VSphereVolumes{}
	return vsphereVolumes, nil
}

// FindVolumes returns Volume instances associated with this VSphereVolumes.
// EtcdClusterSpec is populated using vSphere volume metadata.
func (v *VSphereVolumes) FindVolumes() ([]*Volume, error) {
	var volumes []*Volume
	ip := v.InternalIp()
	attachedTo := ""
	if ip != nil {
		attachedTo = ip.String()
	}

	etcdClusters, err := getVolMetadata()

	if err != nil {
		return nil, err
	}

	for _, etcd := range etcdClusters {
		mountPoint := vsphere.GetMountPoint(etcd.VolumeId)
		localDevice, err := getDevice(mountPoint)
		if err != nil {
			return nil, err
		}
		vol := &Volume{
			ID:          etcd.VolumeId,
			LocalDevice: localDevice,
			AttachedTo:  attachedTo,
			Mountpoint:  mountPoint,
			Status:      VolStatusValue,
			Info: VolumeInfo{
				Description: etcd.EtcdClusterName,
			},
		}

		etcdSpec := &etcdmanager.EtcdClusterSpec{
			ClusterKey: etcd.EtcdClusterName,
			NodeName:   etcd.EtcdNodeName,
		}

		var nodeNames []string
		for _, member := range etcd.Members {
			nodeNames = append(nodeNames, member.Name)
		}
		etcdSpec.NodeNames = nodeNames
		vol.Info.EtcdClusters = []*etcdmanager.EtcdClusterSpec{etcdSpec}
		volumes = append(volumes, vol)
	}
	klog.V(4).Infof("Found volumes: %v", volumes)
	return volumes, nil
}

// FindMountedVolume implements Volumes::FindMountedVolume
func (v *VSphereVolumes) FindMountedVolume(volume *Volume) (string, error) {
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

func getDevice(mountPoint string) (string, error) {
	if runtime.GOOS == "linux" {
		cmd := "lsblk"
		arg := "-l"
		out, err := exec.Command(cmd, arg).Output()
		if err != nil {
			return "", err
		}

		if Containerized {
			mountPoint = pathFor(mountPoint)
		}
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, mountPoint) {
				lsblkOutput := strings.Split(line, " ")
				klog.V(4).Infof("Found device: %v ", lsblkOutput[0])
				return "/dev/" + lsblkOutput[0], nil
			}
		}
	} else {
		return "", fmt.Errorf("Failed to find device. OS %v is not supported for vSphere.", runtime.GOOS)
	}
	return "", fmt.Errorf("No device has been mounted on mountPoint %v.", mountPoint)
}

func getVolMetadata() ([]vsphere.VolumeMetadata, error) {
	rawData, err := ioutil.ReadFile(pathFor(VolumeMetaDataFile))

	if err != nil {
		return nil, err
	}

	return vsphere.UnmarshalVolumeMetadata(string(rawData))
}

// AttachVolume attaches given volume. In case of vSphere, volumes are statically mounted, so no operation is performed.
func (v *VSphereVolumes) AttachVolume(volume *Volume) error {
	// Currently this is a no-op for vSphere. The virtual disks should already be mounted on this VM.
	klog.Infof("All volumes should already be attached. No operation done.")
	return nil
}

// InternalIp returns IP of machine associated with this volume.
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

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

package dns

import (
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"os"
	"path"
)

// DockerHostsFile updates /etc/hosts and any docker containers than mount it
type DockerHostsFile struct {
	HostsFile

	ContainersPath string
}

var _ DNSTarget = &DockerHostsFile{}

type DockerHostConfig struct {
	NetworkMode string `json:"NetworkMode"`

	// {"Binds":null,"ContainerIDFile":"","LogConfig":{"Type":"","Config":null},"NetworkMode":"host","PortBindings":{},
	// "RestartPolicy":{"Name":"","MaximumRetryCount":0},"AutoRemove":false,"VolumeDriver":"","VolumesFrom":null,
	// "CapAdd":null,"CapDrop":null,"Dns":null,"DnsOptions":null,"DnsSearch":null,"ExtraHosts":null,"GroupAdd":null,
	// "IpcMode":"","Cgroup":"","Links":[],"OomScoreAdj":-998,"PidMode":"","Privileged":false,"PublishAllPorts":false,
	// "ReadonlyRootfs":false,"SecurityOpt":["seccomp=unconfined"],"StorageOpt":null,"UTSMode":"","UsernsMode":"",
	// "ShmSize":67108864,"ConsoleSize":[0,0],"Isolation":"","CpuShares":2,"Memory":0,
	// "CgroupParent":"/kubepods/burstable/podf09c109b00821fac8d9eaa8c4591c367","BlkioWeight":0,
	// "BlkioWeightDevice":null,"BlkioDeviceReadBps":null,"BlkioDeviceWriteBps":null,"BlkioDeviceReadIOps":null,
	// "BlkioDeviceWriteIOps":null,"CpuPeriod":0,"CpuQuota":0,"CpusetCpus":"","CpusetMems":"","Devices":null,
	// "DiskQuota":0,"KernelMemory":0,"MemoryReservation":0,"MemorySwap":0,"MemorySwappiness":-1,
	// "OomKillDisable":false,"PidsLimit":0,"Ulimits":null,"CpuCount":0,"CpuPercent":0,"BlkioIOps":0,
	// "BlkioBps":0,"SandboxSize":0}
}

func (h *DockerHostsFile) Update(snapshot *DNSViewSnapshot) error {
	err := h.HostsFile.Update(snapshot)
	if err != nil {
		return err
	}

	hostsFileBytes, err := ioutil.ReadFile(h.HostsFile.Path)
	if err != nil {
		return fmt.Errorf("error reading hosts file %q for copying to docker containers: %v", h.HostsFile.Path, err)
	}

	containerDirs, err := ioutil.ReadDir(h.ContainersPath)
	if err != nil {
		return fmt.Errorf("error listing container dir %q: %v", h.ContainersPath, err)
	}

	var errors []error

	for _, containerDir := range containerDirs {
		containerDirPath := path.Join(h.ContainersPath, containerDir.Name())

		// Make sure this is a container directory
		if !containerDir.IsDir() {
			glog.V(4).Infof("ignoring non-directory %q in %q", containerDirPath)
			continue
		}

		// Sanity check for a hosts file
		hostsPath := path.Join(containerDirPath, "hosts")
		hostsStat, err := os.Stat(hostsPath)
		if err != nil {
			if os.IsNotExist(err) {
				glog.V(6).Infof("skipping container without hosts file %q", hostsPath)
				continue
			} else {
				errors = append(errors, err)
				glog.Infof("skipping container where could not read hosts file %q", hostsPath)
				continue
			}
		}
		if !hostsStat.Mode().IsRegular() {
			glog.Infof("skipping container where hosts file was not regular file %q", hostsPath)
			continue
		}

		// Read hostconfig
		hostConfigPath := path.Join(containerDirPath, "hostconfig.json")
		hostConfigBytes, err := ioutil.ReadFile(hostConfigPath)
		if err != nil {
			if os.IsNotExist(err) {
				glog.V(6).Infof("skipping container without hostconfig.json file %q", hostConfigPath)
				continue
			} else {
				errors = append(errors, err)
				glog.Infof("skipping container where could not read hostconfig.json file %q: %v", hostConfigPath, err)
				continue
			}
		}

		hostConfig := &DockerHostConfig{}
		if err := json.Unmarshal(hostConfigBytes, hostConfig); err != nil {
			errors = append(errors, err)
			glog.Infof("skipping container where could not parse hostconfig.json file %q: %v", hostConfigPath, err)
			continue
		}

		if hostConfig.NetworkMode != "host" {
			glog.V(6).Infof("skipping container not running with NetworkMode=host %q", hostConfigPath)
			continue
		}

		if err := atomicWriteFile(hostsPath, hostsFileBytes, hostsStat.Mode()); err != nil {
			errors = append(errors, err)
			glog.Warningf("error writing hosts file %q: %v", hostsPath, err)
			continue
		}

		glog.V(4).Infof("updated hosts file %q", hostsPath)
	}

	if len(errors) == 0 {
		return nil
	}
	return errors[0]
}

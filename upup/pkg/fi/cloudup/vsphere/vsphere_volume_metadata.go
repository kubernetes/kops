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

package vsphere

import (
	"encoding/json"
	"strconv"
)

type VolumeMetadata struct {
	// EtcdClusterName is the name of the etcd cluster (main, events etc)
	EtcdClusterName string `json:"etcdClusterName,omitempty"`
	// EtcdNodeName is the name of a node in etcd cluster for which this volume will be used
	EtcdNodeName string `json:"etcdNodeName,omitempty"`
	// EtcdMember stores the configurations for each member of the cluster
	Members []EtcdMemberSpec `json:"etcdMembers,omitempty"`
	// Volume id
	VolumeId string `json:"volumeId,omitempty"`
}

type EtcdMemberSpec struct {
	// Name is the name of the member within the etcd cluster
	Name          string `json:"name,omitempty"`
	InstanceGroup string `json:"instanceGroup,omitempty"`
}

func MarshalVolumeMetadata(v []VolumeMetadata) (string, error) {
	metadata, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(metadata), nil
}

func UnmarshalVolumeMetadata(text string) ([]VolumeMetadata, error) {
	var v []VolumeMetadata
	err := json.Unmarshal([]byte(text), &v)
	return v, err
}

func GetVolumeId(i int) string {
	return "0" + strconv.Itoa(i)
}

/*
 * GetMountPoint will return the mount point where the volume is expected to be mounted.
 * This path would be /mnt/master-<volumeId>, eg: /mnt/master-01.
 */
func GetMountPoint(volumeId string) string {
	return "/mnt/master-" + volumeId
}

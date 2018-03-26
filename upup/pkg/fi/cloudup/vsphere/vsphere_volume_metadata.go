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

// vsphere_volume_metadata houses the volume metadata and related methods for vSphere cloud.

import (
	"encoding/json"
	"strconv"
)

// VolumeMetadata represents metadata for vSphere volumes. Unlike aws and gce clouds, vSphere doesn't support tags for volumes/vmdks yet. This metadata is used to pass the information that aws and gce clouds associate with volumes using tags.
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

// EtcdMemberSpec is the specification of members of etcd cluster, to be associated with this volume.
type EtcdMemberSpec struct {
	// Name is the name of the member within the etcd cluster
	Name          string `json:"name,omitempty"`
	InstanceGroup string `json:"instanceGroup,omitempty"`
}

// MarshalVolumeMetadata marshals given VolumeMetadata to json string.
func MarshalVolumeMetadata(v []VolumeMetadata) (string, error) {
	metadata, err := json.Marshal(v)
	if err != nil {
		return "", err
	}

	return string(metadata), nil
}

// UnmarshalVolumeMetadata unmarshals given json string into VolumeMetadata.
func UnmarshalVolumeMetadata(text string) ([]VolumeMetadata, error) {
	var v []VolumeMetadata
	err := json.Unmarshal([]byte(text), &v)
	return v, err
}

// GetVolumeId returns given integer value to VolumeId format, eg: for i=2, volume id="02".
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

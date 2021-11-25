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

package kops

type ClusterStatus struct {
	// EtcdClusters stores the status for each cluster
	EtcdClusters []EtcdClusterStatus `json:"etcdClusters,omitempty"`
}

// EtcdClusterStatus represents the status of etcd: because etcd only allows limited reconfiguration, we have to block changes once etcd has been initialized.
type EtcdClusterStatus struct {
	// Name is the name of the etcd cluster (main, events etc)
	Name string `json:"name,omitempty"`
	// EtcdMember stores the configurations for each member of the cluster (including the data volume)
	Members []*EtcdMemberStatus `json:"etcdMembers,omitempty"`
}

type EtcdMemberStatus struct {
	// Name is the name of the member within the etcd cluster
	Name string `json:"name,omitempty"`

	// VolumeID is the id of the cloud volume (e.g. the AWS volume id)
	VolumeID string `json:"volumeID,omitempty"`
}

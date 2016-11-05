/*
Copyright 2016 The Kubernetes Authors.

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
	"strings"
)

type Volumes interface {
	AttachVolume(volume *Volume) error
	FindVolumes() ([]*Volume, error)

	ClusterID() string
}

type Volume struct {
	// ID is the cloud-provider identifier for the volume
	ID string

	// Device is set if the volume is attached to the local machine
	LocalDevice string

	// AttachedTo is set to the ID of the machine the volume is attached to, or "" if not attached
	AttachedTo string

	// Mountpoint is the path on which the volume is mounted, if mounted
	// It will likely be "/mnt/master-" + ID
	Mountpoint string

	// Status is a volume provider specific Status string; it makes it easier for the volume provider
	Status string

	Info VolumeInfo
}

func (v *Volume) String() string {
	return DebugString(v)
}

type VolumeInfo struct {
	Description string `json:"description,omitempty"`
	MasterID    int    `json:"masterId,omitempty"`
	// TODO: Maybe the events cluster can just be a PetSet - do we need it for boot?
	EtcdClusters []*EtcdClusterSpec `json:"etcdClusters,omitempty"`
}

func (v *VolumeInfo) String() string {
	return DebugString(v)
}

// Parses a tag on a volume that encodes an etcd cluster role
// The format is "<myname>/<allnames>", e.g. "node1/node1,node2,node3"
func ParseEtcdClusterSpec(clusterKey, v string) (*EtcdClusterSpec, error) {
	v = strings.TrimSpace(v)

	tokens := strings.Split(v, "/")
	if len(tokens) != 2 {
		return nil, fmt.Errorf("invalid EtcdClusterSpec (expected two tokens): %q", v)
	}

	nodeName := tokens[0]
	nodeNames := strings.Split(tokens[1], ",")

	found := false
	for _, s := range nodeNames {
		if s == nodeName {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid EtcdClusterSpec (member not found in all nodes): %q", v)
	}

	c := &EtcdClusterSpec{
		ClusterKey: clusterKey,
		NodeName:   nodeName,
		NodeNames:  nodeNames,
	}
	return c, nil
}

package protokube

import (
	"fmt"
	"strings"
)

type Volumes interface {
	AttachVolume(volume *Volume) error
	FindVolumes() ([]*Volume, error)
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
	Description string
	MasterID    int
	// TODO: Maybe the events cluster can just be a PetSet - do we need it for boot?
	EtcdClusters []*EtcdClusterSpec
}

func (v *VolumeInfo) String() string {
	return DebugString(v)
}

type EtcdClusterSpec struct {
	ClusterKey string

	NodeName  string
	NodeNames []string
}

func (e *EtcdClusterSpec) String() string {
	return DebugString(e)
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

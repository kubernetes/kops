package protokube

import (
	"fmt"
	"strings"
)

type Volumes interface {
	AttachVolume(volume *Volume) (string, error)
	FindMountedVolumes() ([]*Volume, error)
	FindMountableVolumes() ([]*Volume, error)
}

type Volume struct {
	Name      string
	Device    string
	Available bool

	Info VolumeInfo
}

func (v *Volume) String() string {
	return DebugString(v)
}

type VolumeInfo struct {
	Name     string
	MasterID int
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

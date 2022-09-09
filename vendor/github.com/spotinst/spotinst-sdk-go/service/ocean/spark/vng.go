package spark

type DedicatedVirtualNodeGroup struct {
	OceanClusterID      *string `json:"oceanClusterId,omitempty"`
	OceanSparkClusterID *string `json:"oceanSparkClusterId,omitempty"`
	VngID               *string `json:"vngId,omitempty"`
}

type AttachVirtualNodeGroupRequest struct {
	VngID *string `json:"id,omitempty"`
}

type AttachVngInput struct {
	ClusterID        *string                        `json:"-"`
	VirtualNodeGroup *AttachVirtualNodeGroupRequest `json:"virtualNodeGroup,omitempty"`
}

type AttachVngOutput struct {
	VirtualNodeGroup *DedicatedVirtualNodeGroup `json:"virtualNodeGroup,omitempty"`
}

type DetachVngInput struct {
	ClusterID *string `json:"clusterId,omitempty"`
	VngID     *string `json:"vngId,omitempty"`
}

type DetachVngOutput struct{}

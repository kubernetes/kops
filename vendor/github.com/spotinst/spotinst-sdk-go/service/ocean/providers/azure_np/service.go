package azure_np

import (
	"context"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

// Service provides the API operation methods for making requests to endpoints
// of the Spotinst API. See this package's package overview docs for details on
// the service.
type Service interface {
	ListClusters(context.Context) (*ListClustersOutput, error)
	CreateCluster(context.Context, *CreateClusterInput) (*CreateClusterOutput, error)
	ReadCluster(context.Context, *ReadClusterInput) (*ReadClusterOutput, error)
	UpdateCluster(context.Context, *UpdateClusterInput) (*UpdateClusterOutput, error)
	DeleteCluster(context.Context, *DeleteClusterInput) (*DeleteClusterOutput, error)
	ImportCluster(context.Context, *ImportClusterInput) (*ImportClusterOutput, error)
	CreateRoll(context.Context, *CreateRollInput) (*CreateRollOutput, error)
	ReadRoll(ctx context.Context, input *ReadRollInput) (*ReadRollOutput, error)
	ListRolls(ctx context.Context, input *ListRollsInput) (*ListRollsOutput, error)
	StopRoll(ctx context.Context, input *StopRollInput) (*StopRollOutput, error)

	ListVirtualNodeGroups(context.Context, *ListVirtualNodeGroupsInput) (*ListVirtualNodeGroupsOutput, error)
	CreateVirtualNodeGroup(context.Context, *CreateVirtualNodeGroupInput) (*CreateVirtualNodeGroupOutput, error)
	ReadVirtualNodeGroup(context.Context, *ReadVirtualNodeGroupInput) (*ReadVirtualNodeGroupOutput, error)
	UpdateVirtualNodeGroup(context.Context, *UpdateVirtualNodeGroupInput) (*UpdateVirtualNodeGroupOutput, error)
	DeleteVirtualNodeGroup(context.Context, *DeleteVirtualNodeGroupInput) (*DeleteVirtualNodeGroupOutput, error)
	LaunchNewNodes(context.Context, *LaunchNewNodesInput) (*LaunchNewNodesOutput, error)
}

type ServiceOp struct {
	Client *client.Client
}

var _ Service = &ServiceOp{}

func New(sess *session.Session, cfgs ...*spotinst.Config) *ServiceOp {
	cfg := &spotinst.Config{}
	cfg.Merge(sess.Config)
	cfg.Merge(cfgs...)

	return &ServiceOp{
		Client: client.New(sess.Config),
	}
}

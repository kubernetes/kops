package spark

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
	ReadCluster(context.Context, *ReadClusterInput) (*ReadClusterOutput, error)
	ListClusters(context.Context, *ListClustersInput) (*ListClustersOutput, error)
	DeleteCluster(context.Context, *DeleteClusterInput) (*DeleteClusterOutput, error)
	CreateCluster(context.Context, *CreateClusterInput) (*CreateClusterOutput, error)
	UpdateCluster(context.Context, *UpdateClusterInput) (*UpdateClusterOutput, error)
	ListVirtualNodeGroups(context.Context, *ListVngsInput) (*ListVngsOutput, error)
	DetachVirtualNodeGroup(context.Context, *DetachVngInput) (*DetachVngOutput, error)
	AttachVirtualNodeGroup(context.Context, *AttachVngInput) (*AttachVngOutput, error)
}

type ClusterManager interface {
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
		Client: client.New(cfg),
	}
}

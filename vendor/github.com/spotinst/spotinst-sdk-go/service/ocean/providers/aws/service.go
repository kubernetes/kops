package aws

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
	ListClusters(context.Context, *ListClustersInput) (*ListClustersOutput, error)
	CreateCluster(context.Context, *CreateClusterInput) (*CreateClusterOutput, error)
	ReadCluster(context.Context, *ReadClusterInput) (*ReadClusterOutput, error)
	UpdateCluster(context.Context, *UpdateClusterInput) (*UpdateClusterOutput, error)
	DeleteCluster(context.Context, *DeleteClusterInput) (*DeleteClusterOutput, error)

	ListLaunchSpecs(context.Context, *ListLaunchSpecsInput) (*ListLaunchSpecsOutput, error)
	CreateLaunchSpec(context.Context, *CreateLaunchSpecInput) (*CreateLaunchSpecOutput, error)
	ReadLaunchSpec(context.Context, *ReadLaunchSpecInput) (*ReadLaunchSpecOutput, error)
	UpdateLaunchSpec(context.Context, *UpdateLaunchSpecInput) (*UpdateLaunchSpecOutput, error)
	DeleteLaunchSpec(context.Context, *DeleteLaunchSpecInput) (*DeleteLaunchSpecOutput, error)

	ListClusterInstances(context.Context, *ListClusterInstancesInput) (*ListClusterInstancesOutput, error)
	DetachClusterInstances(context.Context, *DetachClusterInstancesInput) (*DetachClusterInstancesOutput, error)

	Roll(context.Context, *RollClusterInput) (*RollClusterOutput, error)
	ListECSClusters(context.Context, *ListECSClustersInput) (*ListECSClustersOutput, error)
	CreateECSCluster(context.Context, *CreateECSClusterInput) (*CreateECSClusterOutput, error)
	ReadECSCluster(context.Context, *ReadECSClusterInput) (*ReadECSClusterOutput, error)
	UpdateECSCluster(context.Context, *UpdateECSClusterInput) (*UpdateECSClusterOutput, error)
	DeleteECSCluster(context.Context, *DeleteECSClusterInput) (*DeleteECSClusterOutput, error)

	ListECSLaunchSpecs(context.Context, *ListECSLaunchSpecsInput) (*ListECSLaunchSpecsOutput, error)
	CreateECSLaunchSpec(context.Context, *CreateECSLaunchSpecInput) (*CreateECSLaunchSpecOutput, error)
	ReadECSLaunchSpec(context.Context, *ReadECSLaunchSpecInput) (*ReadECSLaunchSpecOutput, error)
	UpdateECSLaunchSpec(context.Context, *UpdateECSLaunchSpecInput) (*UpdateECSLaunchSpecOutput, error)
	DeleteECSLaunchSpec(context.Context, *DeleteECSLaunchSpecInput) (*DeleteECSLaunchSpecOutput, error)

	RollECS(context.Context, *ECSRollClusterInput) (*ECSRollClusterOutput, error)
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

package gcp

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
	Create(context.Context, *CreateGroupInput) (*CreateGroupOutput, error)
	Read(context.Context, *ReadGroupInput) (*ReadGroupOutput, error)
	Update(context.Context, *UpdateGroupInput) (*UpdateGroupOutput, error)
	Delete(context.Context, *DeleteGroupInput) (*DeleteGroupOutput, error)
	List(context.Context, *ListGroupsInput) (*ListGroupsOutput, error)
	ImportGKECluster(context.Context, *ImportGKEClusterInput) (*ImportGKEClusterOutput, error)
	Status(context.Context, *StatusGroupInput) (*StatusGroupOutput, error)
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

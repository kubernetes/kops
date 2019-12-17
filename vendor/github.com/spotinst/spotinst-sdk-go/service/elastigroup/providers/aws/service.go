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
	List(context.Context, *ListGroupsInput) (*ListGroupsOutput, error)
	Create(context.Context, *CreateGroupInput) (*CreateGroupOutput, error)
	Read(context.Context, *ReadGroupInput) (*ReadGroupOutput, error)
	Update(context.Context, *UpdateGroupInput) (*UpdateGroupOutput, error)
	Delete(context.Context, *DeleteGroupInput) (*DeleteGroupOutput, error)
	Status(context.Context, *StatusGroupInput) (*StatusGroupOutput, error)
	DeploymentStatus(context.Context, *DeploymentStatusInput) (*RollGroupOutput, error)
	StopDeployment(context.Context, *StopDeploymentInput) (*StopDeploymentOutput, error)
	Detach(context.Context, *DetachGroupInput) (*DetachGroupOutput, error)
	Roll(context.Context, *RollGroupInput) (*RollGroupOutput, error)
	RollECS(context.Context, *RollECSGroupInput) (*RollGroupOutput, error)
	Scale(context.Context, *ScaleGroupInput) (*ScaleGroupOutput, error)
	GetInstanceHealthiness(context.Context, *GetInstanceHealthinessInput) (*GetInstanceHealthinessOutput, error)
	GetGroupEvents(context.Context, *GetGroupEventsInput) (*GetGroupEventsOutput, error)
	ImportBeanstalkEnv(context.Context, *ImportBeanstalkInput) (*ImportBeanstalkOutput, error)
	StartBeanstalkMaintenance(context.Context, *BeanstalkMaintenanceInput) (*BeanstalkMaintenanceOutput, error)
	FinishBeanstalkMaintenance(context.Context, *BeanstalkMaintenanceInput) (*BeanstalkMaintenanceOutput, error)
	GetBeanstalkMaintenanceStatus(context.Context, *BeanstalkMaintenanceInput) (*string, error)
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

package ocean

import (
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/aws"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/azure_np"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/providers/gcp"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/right_sizing"
	"github.com/spotinst/spotinst-sdk-go/service/ocean/spark"
	"github.com/spotinst/spotinst-sdk-go/spotinst"
	"github.com/spotinst/spotinst-sdk-go/spotinst/client"
	"github.com/spotinst/spotinst-sdk-go/spotinst/session"
)

// Service provides the API operation methods for making requests to endpoints
// of the Spotinst API. See this package's package overview docs for details on
// the service.
type Service interface {
	CloudProviderAWS() aws.Service
	CloudProviderGCP() gcp.Service
	Spark() spark.Service
	CloudProviderAzureNP() azure_np.Service
	RightSizing() right_sizing.Service
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

func (s *ServiceOp) CloudProviderAWS() aws.Service {
	return &aws.ServiceOp{
		Client: s.Client,
	}
}

func (s *ServiceOp) CloudProviderGCP() gcp.Service {
	return &gcp.ServiceOp{
		Client: s.Client,
	}
}

func (s *ServiceOp) Spark() spark.Service {
	return &spark.ServiceOp{
		Client: s.Client,
	}
}

func (s *ServiceOp) CloudProviderAzureNP() azure_np.Service {
	return &azure_np.ServiceOp{
		Client: s.Client,
	}
}

func (s *ServiceOp) RightSizing() right_sizing.Service {
	return &right_sizing.ServiceOp{
		Client: s.Client,
	}
}

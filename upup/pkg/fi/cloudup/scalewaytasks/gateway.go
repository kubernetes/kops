package scalewaytasks

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

const GatewayDefaultType = "VPC-GW-S"

// +kops:fitask
type Gateway struct {
	ID   *string
	Name *string
	Zone *string
	Tags []string

	Lifecycle fi.Lifecycle
}

var _ fi.CloudupTask = &Gateway{}
var _ fi.CompareWithID = &Gateway{}

func (g *Gateway) CompareWithID() *string {
	return g.ID
}

func (g *Gateway) Find(context *fi.CloudupContext) (*Gateway, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	gateways, err := cloud.GatewayService().ListGateways(&vpcgw.ListGatewaysRequest{
		Zone: scw.Zone(cloud.Zone()),
		Name: g.Name,
		Tags: []string{fmt.Sprintf("%s=%s", scaleway.TagClusterName, scaleway.ClusterNameFromTags(g.Tags))},
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing gateways: %w", err)
	}

	if gateways.TotalCount == 0 {
		return nil, nil
	}
	if gateways.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 gateway, got %d", gateways.TotalCount)
	}
	gatewayFound := gateways.Gateways[0]

	return &Gateway{
		ID:        fi.PtrTo(gatewayFound.ID),
		Name:      fi.PtrTo(gatewayFound.Name),
		Zone:      fi.PtrTo(gatewayFound.Zone.String()),
		Tags:      gatewayFound.Tags,
		Lifecycle: g.Lifecycle,
	}, nil
}

func (g *Gateway) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(g, context)
}

func (_ *Gateway) CheckChanges(actual, expected, changes *Gateway) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *Gateway) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *Gateway) error {
	if actual != nil {
		//TODO(Mia-Cross): update tags
		return nil
	}

	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))

	gatewayCreated, err := cloud.GatewayService().CreateGateway(&vpcgw.CreateGatewayRequest{
		Zone: zone,
		Name: fi.ValueOf(expected.Name),
		Tags: expected.Tags,
		Type: GatewayDefaultType,
		//UpstreamDNSServers: nil,
		//TODO, if not work
		//IPID:               nil,
		//EnableSMTP:         false,
		EnableBastion: true,
		BastionPort:   scw.Uint32Ptr(1042),
	})
	if err != nil {
		return fmt.Errorf("creating gateway: %w", err)
	}

	_, err = cloud.GatewayService().WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID: gatewayCreated.ID,
		Zone:      zone,
	})
	if err != nil {
		return fmt.Errorf("waiting for gateway: %w", err)
	}

	expected.ID = &gatewayCreated.ID

	return nil
}

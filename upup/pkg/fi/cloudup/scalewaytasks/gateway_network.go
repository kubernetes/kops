package scalewaytasks

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type GatewayNetwork struct {
	ID   *string
	Name *string
	Zone *string

	Lifecycle      fi.Lifecycle
	DHCPConfig     *DHCPConfig
	Gateway        *Gateway
	PrivateNetwork *PrivateNetwork
}

//	func (g *GatewayNetwork) GetName() *string {
//		return g.Name
//	}
//
// var _ fi.HasName = &GatewayNetwork{}
var _ fi.CloudupTask = &GatewayNetwork{}
var _ fi.CompareWithID = &GatewayNetwork{}
var _ fi.CloudupHasDependencies = &GatewayNetwork{}

func (g *GatewayNetwork) CompareWithID() *string {
	return g.ID
}

func (g *GatewayNetwork) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*DHCPConfig); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*PrivateNetwork); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (g *GatewayNetwork) Find(context *fi.CloudupContext) (*GatewayNetwork, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	gwns, err := cloud.GatewayService().ListGatewayNetworks(&vpcgw.ListGatewayNetworksRequest{
		Zone:             scw.Zone(cloud.Zone()),
		GatewayID:        g.Gateway.ID,
		PrivateNetworkID: g.PrivateNetwork.ID,
		DHCPID:           g.DHCPConfig.ID,
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing gateway networks: %w", err)
	}

	if gwns.TotalCount == 0 {
		return nil, nil
	}
	if gwns.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 gateway network, got %d", gwns.TotalCount)
	}
	gwnFound := gwns.GatewayNetworks[0]

	return &GatewayNetwork{
		ID:             fi.PtrTo(gwnFound.ID),
		Zone:           fi.PtrTo(gwnFound.Zone.String()),
		Lifecycle:      g.Lifecycle,
		DHCPConfig:     g.DHCPConfig,
		Gateway:        g.Gateway,
		PrivateNetwork: g.PrivateNetwork,
	}, nil
}

func (g *GatewayNetwork) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(g, context)
}

func (_ *GatewayNetwork) CheckChanges(actual, expected, changes *GatewayNetwork) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
	} else {
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *GatewayNetwork) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *GatewayNetwork) error {
	if actual != nil {
		//TODO(Mia-Cross): update tags
		return nil
	}

	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))

	gwnCreated, err := cloud.GatewayService().CreateGatewayNetwork(&vpcgw.CreateGatewayNetworkRequest{
		Zone:             zone,
		GatewayID:        fi.ValueOf(expected.Gateway.ID),
		PrivateNetworkID: fi.ValueOf(expected.PrivateNetwork.ID),
		EnableMasquerade: true,
		//EnableMasquerade: false,
		EnableDHCP: scw.BoolPtr(true),
		DHCPID:     expected.DHCPConfig.ID,
		DHCP:       nil,
		Address:    nil,
		IpamConfig: nil,
	})
	if err != nil {
		return fmt.Errorf("creating gateway network: %w", err)
	}

	_, err = cloud.GatewayService().WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gwnCreated.ID,
		Zone:             zone,
	})
	if err != nil {
		return fmt.Errorf("waiting for gateway: %v", err)
	}

	expected.ID = &gwnCreated.ID

	return nil
}

package scalewaytasks

import (
	"fmt"
	"net"
	"os"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type Network struct {
	Name      *string
	ID        *string
	Lifecycle fi.Lifecycle
	IPRange   *string
	Zone      *string
	Tags      []string
	//DHCP      vpcgw.DHCP
	//Gateway   vpcgw.Gateway
	//Connexion
}

var _ fi.CompareWithID = &Network{}

func (v *Network) CompareWithID() *string {
	return v.ID
}

func (v *Network) Find(c *fi.Context) (*Network, error) {
	cloud := c.Cloud.(scaleway.ScwCloud)
	vpcService := cloud.VPCService()

	vpcs, err := vpcService.ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Zone: scw.Zone(cloud.Zone()),
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("error listing private networks: %s", err)
	}

	for _, vpc := range vpcs.PrivateNetworks {
		if vpc.Name == fi.StringValue(v.Name) {
			subnet := ""
			if len(vpc.Subnets) > 0 {
				subnet = vpc.Subnets[0].String()
			}
			return &Network{
				Name:      fi.String(vpc.Name),
				ID:        fi.String(vpc.ID),
				Lifecycle: v.Lifecycle,
				IPRange:   &subnet,
				Zone:      fi.String(string(vpc.Zone)),
				Tags:      vpc.Tags,
			}, nil
		}
	}
	return nil, nil
}

func (v *Network) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Network) CheckChanges(a, e, changes *Network) error {
	if a != nil {
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
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Zone == nil {
			return fi.RequiredField("Zone")
		}
	}
	return nil
}

func (_ *Network) RenderScw(t *scaleway.ScwAPITarget, a, e, changes *Network) error {
	if a != nil {
		return nil
	}

	vpcService := t.Cloud.VPCService()
	gwService := t.Cloud.GatewayService()

	// We create a private network
	pn, err := vpcService.CreatePrivateNetwork(&vpc.CreatePrivateNetworkRequest{
		Zone:      scw.Zone(fi.StringValue(e.Zone)),
		Name:      fi.StringValue(e.Name),
		ProjectID: os.Getenv("SCW_DEFAULT_PROJECT_ID"),
		Tags:      e.Tags,
	})
	if err != nil {
		return fmt.Errorf("error rendering network: %s", err)
	}

	// We create a public gateway
	gw, err := gwService.CreateGateway(&vpcgw.CreateGatewayRequest{
		Zone:               scw.Zone(fi.StringValue(e.Zone)),
		ProjectID:          os.Getenv("SCW_DEFAULT_PROJECT_ID"),
		Name:               fi.StringValue(e.Name),
		Tags:               e.Tags,
		Type:               "VPC-GW-S",
		UpstreamDNSServers: nil,
		IPID:               nil,
		EnableSMTP:         false,
		EnableBastion:      true,
		BastionPort:        scw.Uint32Ptr(1042), // TODO(Mia-Cross): drop the bastion if it doesn't work ??
	})
	if err != nil {
		return fmt.Errorf("error rendering gateway: %s", err)
	}

	_, subnet, err := net.ParseCIDR(fi.StringValue(e.IPRange))
	if err != nil {
		return fmt.Errorf("error parsing CIDR: %s", err)
	}

	// We create a DHCP server
	dhcp, err := gwService.CreateDHCP(&vpcgw.CreateDHCPRequest{
		Zone:               scw.Zone(fi.StringValue(e.Zone)),
		ProjectID:          os.Getenv("SCW_DEFAULT_PROJECT_ID"),
		Subnet:             scw.IPNet{IPNet: *subnet},
		Address:            nil,
		PoolLow:            nil,
		PoolHigh:           nil,
		EnableDynamic:      nil,
		ValidLifetime:      nil,
		RenewTimer:         nil,
		RebindTimer:        nil,
		PushDefaultRoute:   nil,
		PushDNSServer:      nil,
		DNSServersOverride: nil,
		DNSSearch:          nil,
		DNSLocalName:       nil,
	})
	if err != nil {
		return fmt.Errorf("error rendering DHCP: %v", err)
	}

	// We link the gateway (with DHCP) to the private network once it's in a stable state
	_, err = gwService.WaitForGateway(&vpcgw.WaitForGatewayRequest{
		GatewayID: gw.ID,
		Zone:      scw.Zone(fi.StringValue(e.Zone)),
	})
	if err != nil {
		return fmt.Errorf("error waiting for gateway: %v", err)
	}
	gwn, err := gwService.CreateGatewayNetwork(&vpcgw.CreateGatewayNetworkRequest{
		Zone:             scw.Zone(fi.StringValue(e.Zone)),
		GatewayID:        gw.ID,
		PrivateNetworkID: pn.ID,
		EnableMasquerade: true,
		DHCPID:           scw.StringPtr(dhcp.ID),
		Address:          nil,
		EnableDHCP:       scw.BoolPtr(true),
	})
	if err != nil {
		return fmt.Errorf("error rendering gateway network with DHCP: %v", err)
	}
	_, err = gwService.WaitForGatewayNetwork(&vpcgw.WaitForGatewayNetworkRequest{
		GatewayNetworkID: gwn.ID,
		Zone:             scw.Zone(fi.StringValue(e.Zone)),
	})
	if err != nil {
		return fmt.Errorf("error waiting for gateway: %v", err)
	}

	return nil
}

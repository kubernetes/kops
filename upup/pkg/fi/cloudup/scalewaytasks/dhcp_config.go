package scalewaytasks

import (
	"fmt"
	"net"

	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type DHCPConfig struct {
	ID   *string
	Name *string
	Zone *string

	Subnet *string

	Lifecycle      fi.Lifecycle
	Gateway        *Gateway
	PrivateNetwork *PrivateNetwork
}

var _ fi.CloudupTask = &DHCPConfig{}
var _ fi.CompareWithID = &DHCPConfig{}
var _ fi.CloudupHasDependencies = &DHCPConfig{}

func (d *DHCPConfig) CompareWithID() *string {
	return d.ID
}

func (d *DHCPConfig) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Gateway); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*PrivateNetwork); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (d *DHCPConfig) Find(context *fi.CloudupContext) (*DHCPConfig, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)

	//_, addr, err := net.ParseCIDR(fi.ValueOf(d.Subnet))
	//if err != nil {
	//	return nil, fmt.Errorf("parsing CIDR: %w", err)
	//}
	//address := []byte(addr.String())

	dhcps, err := cloud.GatewayService().ListDHCPs(&vpcgw.ListDHCPsRequest{
		Zone: scw.Zone(cloud.Zone()),
		//Address: &net.IP(address),
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing DHCP configs: %w", err)
	}

	if dhcps.TotalCount == 0 {
		return nil, nil
	}
	//TODO(Mia-Cross): what if the same project ID already has an unrelated DHCP config ?? Checkout the Address and HasAddress filters
	if dhcps.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 DHCP , got %d", dhcps.TotalCount)
	}
	dhcpFound := dhcps.Dhcps[0]

	return &DHCPConfig{
		ID:        fi.PtrTo(dhcpFound.ID),
		Zone:      fi.PtrTo(dhcpFound.Zone.String()),
		Subnet:    fi.PtrTo(dhcpFound.Subnet.String()),
		Lifecycle: d.Lifecycle,
		//TODO(Mia-Cross): how do i fill the name of the GW and PN ? Do i still give the DHCP object a name even if it's not useful to find or create it ?
		Gateway:        nil,
		PrivateNetwork: nil,
	}, nil
}

func (d *DHCPConfig) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(d, context)
}

func (_ *DHCPConfig) CheckChanges(actual, expected, changes *DHCPConfig) error {
	if actual != nil {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Zone != nil {
			return fi.CannotChangeField("Zone")
		}
		//TODO(Mia-Cross): subnet ???
	} else {
		if expected.Zone == nil {
			return fi.RequiredField("Zone")
		}
		if expected.Subnet == nil {
			return fi.RequiredField("Subnet")
		}
	}
	return nil
}

func (_ *DHCPConfig) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *DHCPConfig) error {
	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))
	_, subnet, err := net.ParseCIDR(fi.ValueOf(expected.Subnet))
	if err != nil {
		return fmt.Errorf("parsing CIDR: %w", err)
	}

	if actual != nil {
		//TODO(Mia-Cross): update tags
		//TODO(Mia-Cross): update subnet ??
		return nil
	}

	dhcpCreated, err := cloud.GatewayService().CreateDHCP(&vpcgw.CreateDHCPRequest{
		Zone:               zone,
		Subnet:             scw.IPNet{IPNet: fi.ValueOf(subnet)},
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
		return fmt.Errorf("creating DHCP config: %w", err)
	}

	expected.ID = &dhcpCreated.ID

	return nil
}

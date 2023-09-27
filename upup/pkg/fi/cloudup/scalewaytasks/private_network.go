package scalewaytasks

import (
	"fmt"
	"net"

	"github.com/scaleway/scaleway-sdk-go/api/vpc/v2"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type PrivateNetwork struct {
	ID     *string
	Name   *string
	Region *string
	Tags   []string

	IPRange *string

	Lifecycle fi.Lifecycle
	VPC       *VPC
}

var _ fi.CloudupTask = &PrivateNetwork{}
var _ fi.CompareWithID = &PrivateNetwork{}
var _ fi.CloudupHasDependencies = &PrivateNetwork{}

func (p *PrivateNetwork) CompareWithID() *string {
	return p.ID
}

func (p *PrivateNetwork) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*VPC); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (p *PrivateNetwork) Find(context *fi.CloudupContext) (*PrivateNetwork, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	pns, err := cloud.VPCService().ListPrivateNetworks(&vpc.ListPrivateNetworksRequest{
		Region: scw.Region(cloud.Region()),
		Name:   p.Name,
		Tags:   []string{fmt.Sprintf("%s=%s", scaleway.TagClusterName, scaleway.ClusterNameFromTags(p.Tags))},
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing private networks: %w", err)
	}

	if pns.TotalCount == 0 {
		return nil, nil
	}
	if pns.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 private network, got %d", pns.TotalCount)
	}
	pnFound := pns.PrivateNetworks[0]

	var ipRange *string
	if len(pnFound.Subnets) > 0 {
		ipRange = fi.PtrTo(pnFound.Subnets[0].Subnet.String())
	}
	return &PrivateNetwork{
		ID:        fi.PtrTo(pnFound.ID),
		Name:      fi.PtrTo(pnFound.Name),
		Region:    fi.PtrTo(cloud.Region()),
		Tags:      pnFound.Tags,
		IPRange:   ipRange,
		Lifecycle: p.Lifecycle,
		VPC: &VPC{
			Name: fi.PtrTo(pnFound.Name),
		},
	}, nil
}

func (p *PrivateNetwork) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(p, c)
}

func (_ *PrivateNetwork) CheckChanges(actual, expected, changes *PrivateNetwork) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Region != nil {
			return fi.CannotChangeField("Region")
		}
		//TODO(Mia-Cross): IP Range ???
	} else {
		if expected.Name == nil {
			return fi.RequiredField("Name")
		}
		if expected.Region == nil {
			return fi.RequiredField("Region")
		}
		if expected.IPRange == nil {
			return fi.RequiredField("IPRange")
		}
	}
	return nil
}

func (_ *PrivateNetwork) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *PrivateNetwork) error {
	if actual != nil {
		//TODO(Mia-Cross): update tags
		//TODO(Mia-Cross): update IPRange ??
		return nil
	}

	cloud := t.Cloud.(scaleway.ScwCloud)
	region := scw.Region(fi.ValueOf(expected.Region))
	_, ipRange, err := net.ParseCIDR(fi.ValueOf(expected.IPRange))
	if err != nil {
		return fmt.Errorf("parsing CIDR: %w", err)
	}

	pnCreated, err := cloud.VPCService().CreatePrivateNetwork(&vpc.CreatePrivateNetworkRequest{
		Region: region,
		Name:   fi.ValueOf(expected.Name),
		Tags:   expected.Tags,
		Subnets: []scw.IPNet{
			{IPNet: fi.ValueOf(ipRange)},
		},
		VpcID: expected.VPC.ID,
	})
	if err != nil {
		return fmt.Errorf("creating private network: %w", err)
	}

	expected.ID = &pnCreated.ID

	// We create a public gateway
	// We create a DHCP server
	// We link the gateway (with DHCP) to the private network once it's in a stable state

	return nil
}

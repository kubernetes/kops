package scalewaytasks

import (
	"fmt"

	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	ipam "github.com/scaleway/scaleway-sdk-go/api/ipam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

// +kops:fitask
type PrivateNIC struct {
	ID   *string
	Name *string
	Zone *string
	Tags []string

	InstanceID   *string
	ForAPIServer bool

	Lifecycle      fi.Lifecycle
	InstanceGroup  *Instance
	PrivateNetwork *PrivateNetwork
}

func (p *PrivateNIC) IsForAPIServer() bool {
	return p.ForAPIServer
}

func (p *PrivateNIC) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	pNICFound, err := p.Find(context)
	if err != nil {
		return nil, err
	}

	cloud := context.T.Cloud.(scaleway.ScwCloud)
	region, err := scw.Zone(fi.ValueOf(p.Zone)).Region()
	if err != nil {
		return nil, fmt.Errorf("finding private NIC's region: %w", err)
	}
	ips, err := cloud.IPAMService().ListIPs(&ipam.ListIPsRequest{
		Region:           region,
		PrivateNetworkID: pNICFound.PrivateNetwork.ID,
		ResourceID:       pNICFound.ID,
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing load-balancer's IPs: %w", err)
	}

	pnicIPs := []string(nil)
	for _, ip := range ips.IPs {
		pnicIPs = append(pnicIPs, ip.Address.IP.String())
	}

	return pnicIPs, nil
}

var _ fi.CloudupTask = &PrivateNIC{}
var _ fi.CompareWithID = &PrivateNIC{}
var _ fi.CloudupHasDependencies = &PrivateNIC{}
var _ fi.HasAddress = &PrivateNIC{}

func (p *PrivateNIC) CompareWithID() *string {
	return p.Name
}

func (p *PrivateNIC) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask
	for _, task := range tasks {
		if _, ok := task.(*Instance); ok {
			deps = append(deps, task)
		}
		if _, ok := task.(*PrivateNetwork); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

func (p *PrivateNIC) Find(context *fi.CloudupContext) (*PrivateNIC, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	pNICs, err := cloud.InstanceService().ListPrivateNICs(&instance.ListPrivateNICsRequest{
		Zone:     scw.Zone(cloud.Zone()),
		ServerID: fi.ValueOf(p.InstanceID),
		Tags:     p.Tags,
	}, scw.WithContext(context.Context()), scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing private NICs for instance %s: %w", fi.ValueOf(p.InstanceID), err)
	}

	if pNICs.TotalCount == 0 {
		return nil, nil
	}
	if pNICs.TotalCount > 1 {
		return nil, fmt.Errorf("expected exactly 1 private NIC for instance %s, got %d", fi.ValueOf(p.InstanceID), pNICs.TotalCount)
	}
	pNICFound := pNICs.PrivateNics[0]

	forAPIServer := false
	instanceRole := scaleway.InstanceRoleFromTags(pNICFound.Tags)
	if instanceRole == scaleway.TagRoleControlPlane {
		forAPIServer = true
	}

	return &PrivateNIC{
		ID:             fi.PtrTo(pNICFound.ID),
		Name:           p.Name,
		Zone:           p.Zone,
		Tags:           pNICFound.Tags,
		InstanceID:     fi.PtrTo(pNICFound.ServerID),
		ForAPIServer:   forAPIServer,
		Lifecycle:      p.Lifecycle,
		InstanceGroup:  p.InstanceGroup,
		PrivateNetwork: p.PrivateNetwork,
	}, nil
}

func (p *PrivateNIC) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(p, context)
}

func (p *PrivateNIC) CheckChanges(actual, expected, changes *PrivateNIC) error {
	if actual != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
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
		if expected.InstanceID == nil {
			return fi.RequiredField("InstanceID")
		}
	}
	return nil
}

func (_ *PrivateNIC) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *PrivateNIC) error {
	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))

	if actual != nil {
		return nil
	}
	pNICCreated, err := cloud.InstanceService().CreatePrivateNIC(&instance.CreatePrivateNICRequest{
		Zone:             zone,
		ServerID:         fi.ValueOf(expected.InstanceID),
		PrivateNetworkID: fi.ValueOf(expected.PrivateNetwork.ID),
		//IPIDs:
	})
	if err != nil {
		return fmt.Errorf("creating private NIC between instance %s and private network %s: %w", fi.ValueOf(expected.InstanceID), fi.ValueOf(expected.PrivateNetwork.ID), err)
	}

	// We wait for the private nic to be ready before proceeding
	_, err = cloud.InstanceService().WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
		ServerID:     fi.ValueOf(expected.InstanceID),
		PrivateNicID: pNICCreated.PrivateNic.ID,
		Zone:         zone,
	})
	if err != nil {
		return fmt.Errorf("waiting for private NIC %s: %w", pNICCreated.PrivateNic.ID, err)
	}

	expected.ID = &pNICCreated.PrivateNic.ID
	expected.InstanceID = &pNICCreated.PrivateNic.ServerID
	instanceRole := scaleway.InstanceRoleFromTags(expected.Tags)
	if instanceRole == scaleway.TagRoleControlPlane {
		expected.ForAPIServer = true
	} else {
		expected.ForAPIServer = false

	}

	return nil
}

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
	//ID   *string
	Name *string
	Zone *string
	Tags []string

	ForAPIServer bool

	Lifecycle      fi.Lifecycle
	Instance       *Instance
	PrivateNetwork *PrivateNetwork
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

func (p *PrivateNIC) IsForAPIServer() bool {
	return p.ForAPIServer
}

func (p *PrivateNIC) FindAddresses(context *fi.CloudupContext) ([]string, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	region, err := scw.Zone(fi.ValueOf(p.Zone)).Region()
	if err != nil {
		return nil, fmt.Errorf("finding private NIC's region: %w", err)
	}

	servers, err := cloud.GetClusterServers(scaleway.ClusterNameFromTags(p.Tags), p.Name)
	if err != nil {
		return nil, err
	}

	var pnicIPs []string

	for _, server := range servers {

		pNICs, err := cloud.InstanceService().ListPrivateNICs(&instance.ListPrivateNICsRequest{
			Zone:     scw.Zone(cloud.Zone()),
			Tags:     p.Tags,
			ServerID: server.ID,
		}, scw.WithContext(context.Context()), scw.WithAllPages())
		if err != nil {
			return nil, fmt.Errorf("listing private NICs for instance %q: %w", fi.ValueOf(p.Name), err)
		}

		for _, pNIC := range pNICs.PrivateNics {

			ips, err := cloud.IPAMService().ListIPs(&ipam.ListIPsRequest{
				Region:           region,
				PrivateNetworkID: p.PrivateNetwork.ID,
				ResourceID:       &pNIC.ID,
			}, scw.WithContext(context.Context()), scw.WithAllPages())
			if err != nil {
				return nil, fmt.Errorf("listing private NIC's IPs: %w", err)
			}

			for _, ip := range ips.IPs {
				pnicIPs = append(pnicIPs, ip.Address.IP.String())
			}
		}
	}
	return pnicIPs, nil
}

func (p *PrivateNIC) Find(context *fi.CloudupContext) (*PrivateNIC, error) {
	cloud := context.T.Cloud.(scaleway.ScwCloud)
	servers, err := cloud.GetClusterServers(scaleway.ClusterNameFromTags(p.Tags), p.Name)
	if err != nil {
		return nil, err
	}

	var privateNICsFound []*instance.PrivateNIC
	for _, server := range servers {
		pNICs, err := cloud.InstanceService().ListPrivateNICs(&instance.ListPrivateNICsRequest{
			Zone:     scw.Zone(cloud.Zone()),
			Tags:     p.Tags,
			ServerID: server.ID,
		}, scw.WithContext(context.Context()), scw.WithAllPages())
		if err != nil {
			return nil, fmt.Errorf("listing private NICs for instance group %s: %w", fi.ValueOf(p.Name), err)
		}
		for _, pNIC := range pNICs.PrivateNics {
			privateNICsFound = append(privateNICsFound, pNIC)
		}
	}

	if len(privateNICsFound) == 0 {
		return nil, nil
	}
	pNICFound := privateNICsFound[0]

	forAPIServer := false
	instanceRole := scaleway.InstanceRoleFromTags(pNICFound.Tags)
	if instanceRole == scaleway.TagRoleControlPlane {
		forAPIServer = true
	}

	return &PrivateNIC{
		//ID:             fi.PtrTo(pNICFound.ID),
		Name: p.Name,
		Zone: p.Zone,
		Tags: pNICFound.Tags,
		//InstanceID:     fi.PtrTo(pNICFound.ServerID),
		ForAPIServer:   forAPIServer,
		Lifecycle:      p.Lifecycle,
		Instance:       p.Instance,
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
		//if expected.InstanceID == nil {
		//	return fi.RequiredField("InstanceID")
		//}
	}
	return nil
}

func (_ *PrivateNIC) RenderScw(t *scaleway.ScwAPITarget, actual, expected, changes *PrivateNIC) error {
	cloud := t.Cloud.(scaleway.ScwCloud)
	zone := scw.Zone(fi.ValueOf(expected.Zone))
	clusterName := scaleway.ClusterNameFromTags(expected.Instance.Tags)
	igName := fi.ValueOf(expected.Name)

	if actual != nil {
		//TODO(Mia-Cross): handle changes to tags
		return nil
	}

	servers, err := cloud.GetClusterServers(clusterName, &igName)
	if err != nil {
		return fmt.Errorf("rendering private NIC for instance group %q: getting servers: %w", igName, err)
	}

	for _, server := range servers {

		pNICCreated, err := cloud.InstanceService().CreatePrivateNIC(&instance.CreatePrivateNICRequest{
			Zone:             zone,
			ServerID:         server.ID,
			PrivateNetworkID: fi.ValueOf(expected.PrivateNetwork.ID),
			Tags:             expected.Tags,
			//IPIDs:
		})
		if err != nil {
			return fmt.Errorf("creating private NIC between instance %s and private network %s: %w", server.ID, fi.ValueOf(expected.PrivateNetwork.ID), err)
		}

		// We wait for the private nic to be ready before proceeding
		_, err = cloud.InstanceService().WaitForPrivateNIC(&instance.WaitForPrivateNICRequest{
			ServerID:     server.ID,
			PrivateNicID: pNICCreated.PrivateNic.ID,
			Zone:         zone,
		})
		if err != nil {
			return fmt.Errorf("waiting for private NIC %s: %w", pNICCreated.PrivateNic.ID, err)
		}

		//expected.ID = &pNICCreated.PrivateNic.ID
		//expected.InstanceID = &pNICCreated.PrivateNic.ServerID

		//instanceRole := scaleway.InstanceRoleFromTags(expected.Tags)
		//if instanceRole == scaleway.TagRoleControlPlane {
		//	expected.ForAPIServer = true
		//} else {
		//	expected.ForAPIServer = false
		//}
	}

	return nil
}

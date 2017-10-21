package portsbinding

import (
	"testing"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/acceptance/tools"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/portsbinding"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/ports"
)

// CreatePortsbinding will create a port on the specified subnet. An error will be
// returned if the port could not be created.
func CreatePortsbinding(t *testing.T, client *gophercloud.ServiceClient, networkID, subnetID, hostID string) (*portsbinding.Port, error) {
	portName := tools.RandomString("TESTACC-", 8)
	iFalse := false

	t.Logf("Attempting to create port: %s", portName)

	createOpts := portsbinding.CreateOpts{
		CreateOptsBuilder: ports.CreateOpts{
			NetworkID:    networkID,
			Name:         portName,
			AdminStateUp: &iFalse,
			FixedIPs:     []ports.IP{ports.IP{SubnetID: subnetID}},
		},
		HostID: hostID,
	}

	port, err := portsbinding.Create(client, createOpts).Extract()
	if err != nil {
		return port, err
	}

	t.Logf("Successfully created port: %s", portName)

	return port, nil
}

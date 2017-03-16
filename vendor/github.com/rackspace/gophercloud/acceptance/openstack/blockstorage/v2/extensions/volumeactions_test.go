// +build acceptance blockstorage

package extensions

import (
	"os"
	"testing"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v2/extensions/volumeactions"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v2/volumes"
	th "github.com/rackspace/gophercloud/testhelper"
)

func newClient(t *testing.T) (*gophercloud.ServiceClient, error) {
	ao, err := openstack.AuthOptionsFromEnv()
	th.AssertNoErr(t, err)

	client, err := openstack.AuthenticatedClient(ao)
	th.AssertNoErr(t, err)

	return openstack.NewBlockStorageV2(client, gophercloud.EndpointOpts{
		Region: os.Getenv("OS_REGION_NAME"),
	})
}

func TestVolumeAttach(t *testing.T) {
	client, err := newClient(t)
	th.AssertNoErr(t, err)

	t.Logf("Creating volume")
	cv, err := volumes.Create(client, &volumes.CreateOpts{
		Size: 1,
		Name: "blockv2-volume",
	}).Extract()
	th.AssertNoErr(t, err)

	defer func() {
		err = volumes.WaitForStatus(client, cv.ID, "available", 60)
		th.AssertNoErr(t, err)

		t.Logf("Deleting volume")
		err = volumes.Delete(client, cv.ID).ExtractErr()
		th.AssertNoErr(t, err)
	}()

	err = volumes.WaitForStatus(client, cv.ID, "available", 60)
	th.AssertNoErr(t, err)

	instanceID := os.Getenv("OS_INSTANCE_ID")
	if instanceID == "" {
		t.Fatal("Environment variable OS_INSTANCE_ID is required")
	}

	t.Logf("Attaching volume")
	err = volumeactions.Attach(client, cv.ID, &volumeactions.AttachOpts{
		MountPoint:   "/mnt",
		Mode:         "rw",
		InstanceUUID: instanceID,
	}).ExtractErr()
	th.AssertNoErr(t, err)

	err = volumes.WaitForStatus(client, cv.ID, "in-use", 60)
	th.AssertNoErr(t, err)

	t.Logf("Detaching volume")
	err = volumeactions.Detach(client, cv.ID).ExtractErr()
	th.AssertNoErr(t, err)
}

func TestVolumeReserve(t *testing.T) {
	client, err := newClient(t)
	th.AssertNoErr(t, err)

	t.Logf("Creating volume")
	cv, err := volumes.Create(client, &volumes.CreateOpts{
		Size: 1,
		Name: "blockv2-volume",
	}).Extract()
	th.AssertNoErr(t, err)

	defer func() {
		err = volumes.WaitForStatus(client, cv.ID, "available", 60)
		th.AssertNoErr(t, err)

		t.Logf("Deleting volume")
		err = volumes.Delete(client, cv.ID).ExtractErr()
		th.AssertNoErr(t, err)
	}()

	err = volumes.WaitForStatus(client, cv.ID, "available", 60)
	th.AssertNoErr(t, err)

	t.Logf("Reserving volume")
	err = volumeactions.Reserve(client, cv.ID).ExtractErr()
	th.AssertNoErr(t, err)

	err = volumes.WaitForStatus(client, cv.ID, "attaching", 60)
	th.AssertNoErr(t, err)

	t.Logf("Unreserving volume")
	err = volumeactions.Unreserve(client, cv.ID).ExtractErr()
	th.AssertNoErr(t, err)

	err = volumes.WaitForStatus(client, cv.ID, "available", 60)
	th.AssertNoErr(t, err)
}

func TestVolumeConns(t *testing.T) {
	client, err := newClient(t)
	th.AssertNoErr(t, err)

	t.Logf("Creating volume")
	cv, err := volumes.Create(client, &volumes.CreateOpts{
		Size: 1,
		Name: "blockv2-volume",
	}).Extract()
	th.AssertNoErr(t, err)

	defer func() {
		err = volumes.WaitForStatus(client, cv.ID, "available", 60)
		th.AssertNoErr(t, err)

		t.Logf("Deleting volume")
		err = volumes.Delete(client, cv.ID).ExtractErr()
		th.AssertNoErr(t, err)
	}()

	err = volumes.WaitForStatus(client, cv.ID, "available", 60)
	th.AssertNoErr(t, err)

	connOpts := &volumeactions.ConnectorOpts{
		IP:        "127.0.0.1",
		Host:      "stack",
		Initiator: "iqn.1994-05.com.redhat:17cf566367d2",
		Multipath: false,
		Platform:  "x86_64",
		OSType:    "linux2",
	}

	t.Logf("Initializing connection")
	_, err = volumeactions.InitializeConnection(client, cv.ID, connOpts).Extract()
	th.AssertNoErr(t, err)

	t.Logf("Terminating connection")
	err = volumeactions.TerminateConnection(client, cv.ID, connOpts).ExtractErr()
	th.AssertNoErr(t, err)
}

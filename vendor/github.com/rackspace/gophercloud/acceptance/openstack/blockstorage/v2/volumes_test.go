// +build acceptance blockstorage

package v2

import (
	"os"
	"testing"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack"
	"github.com/rackspace/gophercloud/openstack/blockstorage/v2/volumes"
	"github.com/rackspace/gophercloud/pagination"
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

func TestVolumes(t *testing.T) {
	client, err := newClient(t)
	th.AssertNoErr(t, err)

	cv, err := volumes.Create(client, &volumes.CreateOpts{
		Size: 1,
		Name: "blockv2-volume",
	}).Extract()
	th.AssertNoErr(t, err)
	defer func() {
		err = volumes.WaitForStatus(client, cv.ID, "available", 60)
		th.AssertNoErr(t, err)
		err = volumes.Delete(client, cv.ID).ExtractErr()
		th.AssertNoErr(t, err)
	}()

	_, err = volumes.Update(client, cv.ID, &volumes.UpdateOpts{
		Name: "blockv2-updated-volume",
	}).Extract()
	th.AssertNoErr(t, err)

	v, err := volumes.Get(client, cv.ID).Extract()
	th.AssertNoErr(t, err)
	t.Logf("Got volume: %+v\n", v)

	if v.Name != "blockv2-updated-volume" {
		t.Errorf("Unable to update volume: Expected name: blockv2-updated-volume\nActual name: %s", v.Name)
	}

	err = volumes.List(client, &volumes.ListOpts{Name: "blockv2-updated-volume"}).EachPage(func(page pagination.Page) (bool, error) {
		vols, err := volumes.ExtractVolumes(page)
		th.CheckEquals(t, 1, len(vols))
		return true, err
	})
	th.AssertNoErr(t, err)
}

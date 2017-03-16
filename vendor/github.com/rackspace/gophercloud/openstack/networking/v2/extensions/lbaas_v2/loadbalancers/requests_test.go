package loadbalancers

import (
	"testing"

	fake "github.com/rackspace/gophercloud/openstack/networking/v2/common"
	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestURLs(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.AssertEquals(t, th.Endpoint()+"v2.0/lbaas/loadbalancers", rootURL(fake.ServiceClient()))
	th.AssertEquals(t, th.Endpoint()+"v2.0/lbaas/loadbalancers/foo", resourceURL(fake.ServiceClient(), "foo"))
}

func TestListLoadbalancers(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerListSuccessfully(t)

	pages := 0
	err := List(fake.ServiceClient(), ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := ExtractLoadbalancers(page)
		if err != nil {
			return false, err
		}

		if len(actual) != 2 {
			t.Fatalf("Expected 2 loadbalancers, got %d", len(actual))
		}
		th.CheckDeepEquals(t, LoadbalancerWeb, actual[0])
		th.CheckDeepEquals(t, LoadbalancerDb, actual[1])

		return true, nil
	})

	th.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}

func TestListAllLoadbalancers(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerListSuccessfully(t)

	allPages, err := List(fake.ServiceClient(), ListOpts{}).AllPages()
	th.AssertNoErr(t, err)
	actual, err := ExtractLoadbalancers(allPages)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, LoadbalancerWeb, actual[0])
	th.CheckDeepEquals(t, LoadbalancerDb, actual[1])
}

func TestCreateLoadbalancer(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerCreationSuccessfully(t, SingleLoadbalancerBody)

	actual, err := Create(fake.ServiceClient(), CreateOpts{
		Name:         "db_lb",
		AdminStateUp: Up,
		VipSubnetID:  "9cedb85d-0759-4898-8a4b-fa5a5ea10086",
		VipAddress:   "10.30.176.48",
		Flavor:       "medium",
		Provider:     "haproxy",
	}).Extract()
	th.AssertNoErr(t, err)

	th.CheckDeepEquals(t, LoadbalancerDb, *actual)
}

func TestRequiredCreateOpts(t *testing.T) {
	res := Create(fake.ServiceClient(), CreateOpts{})
	if res.Err == nil {
		t.Fatalf("Expected error, got none")
	}
	res = Create(fake.ServiceClient(), CreateOpts{Name: "foo"})
	if res.Err == nil {
		t.Fatalf("Expected error, got none")
	}
	res = Create(fake.ServiceClient(), CreateOpts{Name: "foo", Description: "bar"})
	if res.Err == nil {
		t.Fatalf("Expected error, got none")
	}
	res = Create(fake.ServiceClient(), CreateOpts{Name: "foo", Description: "bar", VipAddress: "bar"})
	if res.Err == nil {
		t.Fatalf("Expected error, got none")
	}
}

func TestGetLoadbalancer(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerGetSuccessfully(t)

	client := fake.ServiceClient()
	actual, err := Get(client, "36e08a3e-a78f-4b40-a229-1e7e23eee1ab").Extract()
	if err != nil {
		t.Fatalf("Unexpected Get error: %v", err)
	}

	th.CheckDeepEquals(t, LoadbalancerDb, *actual)
}

func TestGetLoadbalancerStatusesTree(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerGetStatusesTree(t)

	client := fake.ServiceClient()
	actual, err := GetStatuses(client, "36e08a3e-a78f-4b40-a229-1e7e23eee1ab").ExtractStatuses()
	if err != nil {
		t.Fatalf("Unexpected Get error: %v", err)
	}

	th.CheckDeepEquals(t, LoadbalancerStatusesTree, *(actual.Loadbalancer))
}

func TestDeleteLoadbalancer(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerDeletionSuccessfully(t)

	res := Delete(fake.ServiceClient(), "36e08a3e-a78f-4b40-a229-1e7e23eee1ab")
	th.AssertNoErr(t, res.Err)
}

func TestUpdateLoadbalancer(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleLoadbalancerUpdateSuccessfully(t)

	client := fake.ServiceClient()
	actual, err := Update(client, "36e08a3e-a78f-4b40-a229-1e7e23eee1ab", UpdateOpts{
		Name: "NewLoadbalancerName",
	}).Extract()
	if err != nil {
		t.Fatalf("Unexpected Update error: %v", err)
	}

	th.CheckDeepEquals(t, LoadbalancerUpdated, *actual)
}

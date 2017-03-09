package pools

import (
	"testing"

	fake "github.com/rackspace/gophercloud/openstack/networking/v2/common"
	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestURLs(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.AssertEquals(t, th.Endpoint()+"v2.0/lbaas/pools", rootURL(fake.ServiceClient()))
}

func TestListPools(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePoolListSuccessfully(t)

	pages := 0
	err := List(fake.ServiceClient(), ListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := ExtractPools(page)
		if err != nil {
			return false, err
		}

		if len(actual) != 2 {
			t.Fatalf("Expected 2 pools, got %d", len(actual))
		}
		th.CheckDeepEquals(t, PoolWeb, actual[0])
		th.CheckDeepEquals(t, PoolDb, actual[1])

		return true, nil
	})

	th.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}

func TestListAllPools(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePoolListSuccessfully(t)

	allPages, err := List(fake.ServiceClient(), ListOpts{}).AllPages()
	th.AssertNoErr(t, err)
	actual, err := ExtractPools(allPages)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, PoolWeb, actual[0])
	th.CheckDeepEquals(t, PoolDb, actual[1])
}

func TestCreatePool(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePoolCreationSuccessfully(t, SinglePoolBody)

	actual, err := Create(fake.ServiceClient(), CreateOpts{
		LBMethod:       LBMethodRoundRobin,
		Protocol:       "HTTP",
		Name:           "Example pool",
		TenantID:       "2ffc6e22aae24e4795f87155d24c896f",
		LoadbalancerID: "79e05663-7f03-45d2-a092-8b94062f22ab",
	}).Extract()
	th.AssertNoErr(t, err)

	th.CheckDeepEquals(t, PoolDb, *actual)
}

func TestGetPool(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePoolGetSuccessfully(t)

	client := fake.ServiceClient()
	actual, err := Get(client, "c3741b06-df4d-4715-b142-276b6bce75ab").Extract()
	if err != nil {
		t.Fatalf("Unexpected Get error: %v", err)
	}

	th.CheckDeepEquals(t, PoolDb, *actual)
}

func TestDeletePool(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePoolDeletionSuccessfully(t)

	res := Delete(fake.ServiceClient(), "c3741b06-df4d-4715-b142-276b6bce75ab")
	th.AssertNoErr(t, res.Err)
}

func TestUpdatePool(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePoolUpdateSuccessfully(t)

	client := fake.ServiceClient()
	actual, err := Update(client, "c3741b06-df4d-4715-b142-276b6bce75ab", UpdateOpts{
		Name:     "NewPoolName",
		LBMethod: LBMethodLeastConnections,
	}).Extract()
	if err != nil {
		t.Fatalf("Unexpected Update error: %v", err)
	}

	th.CheckDeepEquals(t, PoolUpdated, *actual)
}

func TestRequiredPoolCreateOpts(t *testing.T) {
	res := Create(fake.ServiceClient(), CreateOpts{})
	if res.Err == nil {
		t.Fatalf("Expected error, got none")
	}
	res = Create(fake.ServiceClient(), CreateOpts{LBMethod: LBMethod("invalid"), Protocol: ProtocolHTTPS, LoadbalancerID: "69055154-f603-4a28-8951-7cc2d9e54a9a"})
	if res.Err == nil || res.Err != errValidLBMethodRequired {
		t.Fatalf("Expected '%s' error, but got '%s'", errValidLBMethodRequired, res.Err)
	}
	res = Create(fake.ServiceClient(), CreateOpts{LBMethod: LBMethodRoundRobin, Protocol: Protocol("invalid"), LoadbalancerID: "69055154-f603-4a28-8951-7cc2d9e54a9a"})
	if res.Err == nil || res.Err != errValidProtocolRequired {
		t.Fatalf("Expected '%s' error, but got '%s'", errValidProtocolRequired, res.Err)
	}
	res = Create(fake.ServiceClient(), CreateOpts{LBMethod: LBMethodRoundRobin, Protocol: ProtocolHTTPS})
	if res.Err == nil || res.Err != errLoadbalancerOrListenerRequired {
		t.Fatalf("Expected '%s' error, but got '%s'", errLoadbalancerOrListenerRequired, res.Err)
	}
}

func TestListMembers(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMemberListSuccessfully(t)

	pages := 0
	err := ListAssociateMembers(fake.ServiceClient(), "332abe93-f488-41ba-870b-2ac66be7f853", MemberListOpts{}).EachPage(func(page pagination.Page) (bool, error) {
		pages++

		actual, err := ExtractMembers(page)
		if err != nil {
			return false, err
		}

		if len(actual) != 2 {
			t.Fatalf("Expected 2 members, got %d", len(actual))
		}
		th.CheckDeepEquals(t, MemberWeb, actual[0])
		th.CheckDeepEquals(t, MemberDb, actual[1])

		return true, nil
	})

	th.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}

func TestListAllMembers(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMemberListSuccessfully(t)

	allPages, err := ListAssociateMembers(fake.ServiceClient(), "332abe93-f488-41ba-870b-2ac66be7f853", MemberListOpts{}).AllPages()
	th.AssertNoErr(t, err)
	actual, err := ExtractMembers(allPages)
	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, MemberWeb, actual[0])
	th.CheckDeepEquals(t, MemberDb, actual[1])
}

func TestCreateMember(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMemberCreationSuccessfully(t, SingleMemberBody)

	actual, err := CreateAssociateMember(fake.ServiceClient(), "332abe93-f488-41ba-870b-2ac66be7f853", MemberCreateOpts{
		Name:         "db",
		SubnetID:     "1981f108-3c48-48d2-b908-30f7d28532c9",
		TenantID:     "2ffc6e22aae24e4795f87155d24c896f",
		Address:      "10.0.2.11",
		ProtocolPort: 80,
		Weight:       10,
	}).ExtractMember()
	th.AssertNoErr(t, err)

	th.CheckDeepEquals(t, MemberDb, *actual)
}

func TestRequiredMemberCreateOpts(t *testing.T) {
	res := CreateAssociateMember(fake.ServiceClient(), "", MemberCreateOpts{})
	if res.Err == nil {
		t.Fatalf("Expected error, got none")
	}
	res = CreateAssociateMember(fake.ServiceClient(), "", MemberCreateOpts{Address: "1.2.3.4", ProtocolPort: 80})
	if res.Err == nil || res.Err != errPoolIdRequired {
		t.Fatalf("Expected '%s' error, but got '%s'", errPoolIdRequired, res.Err)
	}
	res = CreateAssociateMember(fake.ServiceClient(), "332abe93-f488-41ba-870b-2ac66be7f853", MemberCreateOpts{ProtocolPort: 80})
	if res.Err == nil || res.Err != errAddressRequired {
		t.Fatalf("Expected '%s' error, but got '%s'", errAddressRequired, res.Err)
	}
	res = CreateAssociateMember(fake.ServiceClient(), "332abe93-f488-41ba-870b-2ac66be7f853", MemberCreateOpts{Address: "1.2.3.4"})
	if res.Err == nil || res.Err != errProtocolPortRequired {
		t.Fatalf("Expected '%s' error, but got '%s'", errProtocolPortRequired, res.Err)
	}
}

func TestGetMember(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMemberGetSuccessfully(t)

	client := fake.ServiceClient()
	actual, err := GetAssociateMember(client, "332abe93-f488-41ba-870b-2ac66be7f853", "2a280670-c202-4b0b-a562-34077415aabf").ExtractMember()
	if err != nil {
		t.Fatalf("Unexpected Get error: %v", err)
	}

	th.CheckDeepEquals(t, MemberDb, *actual)
}

func TestDeleteMember(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMemberDeletionSuccessfully(t)

	res := DeleteMember(fake.ServiceClient(), "332abe93-f488-41ba-870b-2ac66be7f853", "2a280670-c202-4b0b-a562-34077415aabf")
	th.AssertNoErr(t, res.Err)
}

func TestUpdateMember(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleMemberUpdateSuccessfully(t)

	client := fake.ServiceClient()
	actual, err := UpdateAssociateMember(client, "332abe93-f488-41ba-870b-2ac66be7f853", "2a280670-c202-4b0b-a562-34077415aabf", MemberUpdateOpts{
		Name:   "newMemberName",
		Weight: 4,
	}).ExtractMember()
	if err != nil {
		t.Fatalf("Unexpected Update error: %v", err)
	}

	th.CheckDeepEquals(t, MemberUpdated, *actual)
}

// +build acceptance imageservice

package v2

import (
	"os"
	"testing"

	members "github.com/rackspace/gophercloud/openstack/imageservice/v2/members"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestImageMemberCreateListDelete(t *testing.T) {
	client := newClient(t)

	//creating image
	image := createTestImage(t, client)
	defer deleteImage(t, client, image)

	//creating member
	member, err := members.Create(client, image.ID, "tenant").Extract()
	th.AssertNoErr(t, err)
	th.AssertNotNil(t, member)

	t.Logf("Member has been created for image %s", image.ID)

	//listing member
	var mems *[]members.ImageMember
	mems, err = members.List(client, image.ID).Extract()
	th.AssertNoErr(t, err)
	th.AssertNotNil(t, mems)
	th.AssertEquals(t, 1, len(*mems))

	t.Logf("Members after adding one %v", mems)

	//checking just created member
	m := (*mems)[0]
	th.AssertEquals(t, "pending", m.Status)
	th.AssertEquals(t, "tenant", m.MemberID)

	//deleting member
	deleteResult := members.Delete(client, image.ID, "tenant")
	th.AssertNoErr(t, deleteResult.Err)

	//listing member
	mems, err = members.List(client, image.ID).Extract()
	th.AssertNoErr(t, err)
	th.AssertNotNil(t, mems)
	th.AssertEquals(t, 0, len(*mems))

	t.Logf("Members after deleting one %v", mems)
}

func TestImageMemberDetailsAndUpdate(t *testing.T) {
	// getting current tenant id
	memberTenantID := os.Getenv("OS_TENANT_ID")
	if memberTenantID == "" {
		t.Fatalf("Please define OS_TENANT_ID for image member updating test was '%s'", memberTenantID)
	}

	client := newClient(t)

	//creating image
	image := createTestImage(t, client)
	defer deleteImage(t, client, image)

	//creating member
	member, err := members.Create(client, image.ID, memberTenantID).Extract()
	th.AssertNoErr(t, err)
	th.AssertNotNil(t, member)

	//checking image member details
	member, err = members.Get(client, image.ID, memberTenantID).Extract()
	th.AssertNoErr(t, err)
	th.AssertNotNil(t, member)

	th.AssertEquals(t, memberTenantID, member.MemberID)
	th.AssertEquals(t, "pending", member.Status)

	t.Logf("Updating image's %s member status for tenant %s to 'accepted' ", image.ID, memberTenantID)

	//updating image
	member, err = members.Update(client, image.ID, memberTenantID, "accepted").Extract()
	th.AssertNoErr(t, err)
	th.AssertNotNil(t, member)
	th.AssertEquals(t, "accepted", member.Status)

}

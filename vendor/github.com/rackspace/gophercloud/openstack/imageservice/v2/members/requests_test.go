package members

import (
	"strings"
	"testing"
	"time"

	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
	fakeclient "github.com/rackspace/gophercloud/testhelper/client"
)

const createdAtString = "2013-09-20T19:22:19Z"
const updatedAtString = "2013-09-20T19:25:31Z"

func TestCreateMemberSuccessfully(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleCreateImageMemberSuccessfully(t)
	im, err := Create(fakeclient.ServiceClient(), "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		"8989447062e04a818baf9e073fd04fa7").Extract()
	th.AssertNoErr(t, err)

	createdAt, err := time.Parse(time.RFC3339, createdAtString)
	th.AssertNoErr(t, err)

	updatedAt, err := time.Parse(time.RFC3339, updatedAtString)
	th.AssertNoErr(t, err)

	th.AssertDeepEquals(t, ImageMember{
		CreatedAt: createdAt,
		ImageID:   "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		MemberID:  "8989447062e04a818baf9e073fd04fa7",
		Schema:    "/v2/schemas/member",
		Status:    "pending",
		UpdatedAt: updatedAt,
	}, *im)

}

func TestCreateMemberMemberConflict(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleCreateImageMemberConflict(t)

	result := Create(fakeclient.ServiceClient(), "da3b75d9-memberConflict",
		"8989447062e04a818baf9e073fd04fa7")

	if result.Err == nil {
		t.Fatalf("Expected error in result defined (Err: %v)", result.Err)
	}

	message := result.Err.Error()
	if !strings.Contains(message, "is already member for image") {
		t.Fatalf("Wrong error message: %s", message)
	}

}
func TestCreateMemberInvalidVisibility(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleCreateImageMemberInvalidVisibility(t)

	result := Create(fakeclient.ServiceClient(), "da3b75d9-invalid-visibility",
		"8989447062e04a818baf9e073fd04fa7")

	if result.Err == nil {
		t.Fatalf("Expected error in result defined (Err: %v)", result.Err)
	}

	message := result.Err.Error()
	if !strings.Contains(message, "which 'visibility' attribute is private") {
		t.Fatalf("Wrong error message: %s", message)
	}
}

func TestMemberListSuccessfully(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleImageMemberList(t)

	pager := List(fakeclient.ServiceClient(), "da3b75d9-3f4a-40e7-8a2c-bfab23927dea")
	t.Logf("Pager state %v", pager)
	count, pages := 0, 0
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		pages++
		t.Logf("Page %v", page)
		members, err := ExtractMembers(page)
		if err != nil {
			return false, err
		}

		for _, i := range members {
			t.Logf("%s\t%s\t%s\t%s\t\n", i.ImageID, i.MemberID, i.Status, i.Schema)
			count++
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
	th.AssertEquals(t, 1, pages)
	th.AssertEquals(t, 2, count)
}

func TestMemberListEmpty(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleImageMemberEmptyList(t)

	pager := List(fakeclient.ServiceClient(), "da3b75d9-3f4a-40e7-8a2c-bfab23927dea")
	t.Logf("Pager state %v", pager)
	count, pages := 0, 0
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		pages++
		t.Logf("Page %v", page)
		members, err := ExtractMembers(page)
		if err != nil {
			return false, err
		}

		for _, i := range members {
			t.Logf("%s\t%s\t%s\t%s\t\n", i.ImageID, i.MemberID, i.Status, i.Schema)
			count++
		}

		return true, nil
	})

	th.AssertNoErr(t, err)
	th.AssertEquals(t, 0, pages)
	th.AssertEquals(t, 0, count)
}

func TestShowMemberDetails(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	HandleImageMemberDetails(t)
	md, err := Get(fakeclient.ServiceClient(),
		"da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		"8989447062e04a818baf9e073fd04fa7").Extract()

	th.AssertNoErr(t, err)
	th.AssertNotNil(t, md)

	createdAt, err := time.Parse(time.RFC3339, "2013-11-26T07:21:21Z")
	th.AssertNoErr(t, err)

	updatedAt, err := time.Parse(time.RFC3339, "2013-11-26T07:21:21Z")
	th.AssertNoErr(t, err)

	th.AssertDeepEquals(t, ImageMember{
		CreatedAt: createdAt,
		ImageID:   "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		MemberID:  "8989447062e04a818baf9e073fd04fa7",
		Schema:    "/v2/schemas/member",
		Status:    "pending",
		UpdatedAt: updatedAt,
	}, *md)
}

func TestDeleteMember(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	counter := HandleImageMemberDeleteSuccessfully(t)

	result := Delete(fakeclient.ServiceClient(), "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		"8989447062e04a818baf9e073fd04fa7")
	th.AssertEquals(t, 1, counter.Counter)
	th.AssertNoErr(t, result.Err)
}

func TestDeleteMemberByNonOwner(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	counter := HandleImageMemberDeleteByNonOwner(t)

	result := Delete(fakeclient.ServiceClient(), "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		"8989447062e04a818baf9e073fd04fa7")
	th.AssertEquals(t, 1, counter.Counter)

	if result.Err == nil {
		t.Fatalf("Expected error in result defined (Err: %v)", result.Err)
	}

	message := result.Err.Error()
	if !strings.Contains(message, "You must be the owner of the specified image") {
		t.Fatalf("Wrong error message: %s", message)
	}
}

func TestMemberUpdateSuccessfully(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	counter := HandleImageMemberUpdate(t)
	im, err := Update(fakeclient.ServiceClient(), "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		"8989447062e04a818baf9e073fd04fa7",
		UpdateOpts{
			Status: "accepted",
		}).Extract()
	th.AssertEquals(t, 1, counter.Counter)
	th.AssertNoErr(t, err)

	createdAt, err := time.Parse(time.RFC3339, "2013-11-26T07:21:21Z")
	th.AssertNoErr(t, err)

	updatedAt, err := time.Parse(time.RFC3339, "2013-11-26T07:21:21Z")
	th.AssertNoErr(t, err)

	th.AssertDeepEquals(t, ImageMember{
		CreatedAt: createdAt,
		ImageID:   "da3b75d9-3f4a-40e7-8a2c-bfab23927dea",
		MemberID:  "8989447062e04a818baf9e073fd04fa7",
		Schema:    "/v2/schemas/member",
		Status:    "accepted",
		UpdatedAt: updatedAt,
	}, *im)

}

package webhooks

import (
	"testing"

	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

const (
	groupID  = "10eb3219-1b12-4b34-b1e4-e10ee4f24c65"
	policyID = "2b48d247-0282-4b9d-8775-5c4b67e8e649"
	firstID  = "2bd1822c-58c5-49fd-8b3d-ed44781a58d1" // FirstWebhook
	secondID = "76711c36-dfbe-4f5e-bea6-cded99690515" // SecondWebhook
)

func TestList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleWebhookListSuccessfully(t)

	pages := 0
	pager := List(client.ServiceClient(), groupID, policyID)

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		pages++

		webhooks, err := ExtractWebhooks(page)

		if err != nil {
			return false, err
		}

		if len(webhooks) != 2 {
			t.Fatalf("Expected 2 policies, got %d", len(webhooks))
		}

		th.CheckDeepEquals(t, FirstWebhook, webhooks[0])
		th.CheckDeepEquals(t, SecondWebhook, webhooks[1])

		return true, nil
	})

	th.AssertNoErr(t, err)

	if pages != 1 {
		t.Errorf("Expected 1 page, saw %d", pages)
	}
}

func TestCreate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleWebhookCreateSuccessfully(t)

	client := client.ServiceClient()
	opts := CreateOpts{
		{
			Name: "first hook",
		},
		{
			Name: "second hook",
			Metadata: map[string]string{
				"notes": "a note about this webhook",
			},
		},
	}

	webhooks, err := Create(client, groupID, policyID, opts).Extract()

	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, FirstWebhook, webhooks[0])
	th.CheckDeepEquals(t, SecondWebhook, webhooks[1])
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleWebhookGetSuccessfully(t)

	client := client.ServiceClient()

	webhook, err := Get(client, groupID, policyID, firstID).Extract()

	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, FirstWebhook, *webhook)
}

func TestUpdate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleWebhookUpdateSuccessfully(t)

	client := client.ServiceClient()
	opts := UpdateOpts{
		Name: "updated hook",
		Metadata: map[string]string{
			"new-key": "some data",
		},
	}

	err := Update(client, groupID, policyID, firstID, opts).ExtractErr()

	th.AssertNoErr(t, err)
}

func TestUpdateNoMetadata(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleWebhookUpdateSuccessfully(t)

	client := client.ServiceClient()
	opts := UpdateOpts{
		Name: "updated hook",
	}

	err := Update(client, groupID, policyID, firstID, opts).ExtractErr()

	th.AssertEquals(t, ErrNoMetadata, err)
}

func TestDelete(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandleWebhookDeleteSuccessfully(t)

	client := client.ServiceClient()
	err := Delete(client, groupID, policyID, firstID).ExtractErr()

	th.AssertNoErr(t, err)
}

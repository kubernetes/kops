package policies

import (
	"testing"
	"time"

	"github.com/rackspace/gophercloud/pagination"
	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

const (
	groupID         = "60b15dad-5ea1-43fa-9a12-a1d737b4da07"
	webhookPolicyID = "2b48d247-0282-4b9d-8775-5c4b67e8e649"
)

func TestList(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePolicyListSuccessfully(t)

	pages := 0
	pager := List(client.ServiceClient(), "60b15dad-5ea1-43fa-9a12-a1d737b4da07")

	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		pages++

		policies, err := ExtractPolicies(page)

		if err != nil {
			return false, err
		}

		if len(policies) != 3 {
			t.Fatalf("Expected 3 policies, got %d", len(policies))
		}

		th.CheckDeepEquals(t, WebhookPolicy, policies[0])
		th.CheckDeepEquals(t, OneTimePolicy, policies[1])
		th.CheckDeepEquals(t, SundayAfternoonPolicy, policies[2])

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
	HandlePolicyCreateSuccessfully(t)

	oneTime := time.Date(2020, time.April, 01, 23, 0, 0, 0, time.UTC)
	client := client.ServiceClient()
	opts := CreateOpts{
		{
			Name:            "webhook policy",
			Type:            Webhook,
			Cooldown:        300,
			AdjustmentType:  ChangePercent,
			AdjustmentValue: 3.3,
		},
		{
			Name:            "one time",
			Type:            Schedule,
			AdjustmentType:  Change,
			AdjustmentValue: -1,
			Schedule:        At(oneTime),
		},
		{
			Name:            "sunday afternoon",
			Type:            Schedule,
			AdjustmentType:  DesiredCapacity,
			AdjustmentValue: 2,
			Schedule:        Cron("59 15 * * 0"),
		},
	}

	policies, err := Create(client, groupID, opts).Extract()

	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, WebhookPolicy, policies[0])
	th.CheckDeepEquals(t, OneTimePolicy, policies[1])
	th.CheckDeepEquals(t, SundayAfternoonPolicy, policies[2])
}

func TestGet(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePolicyGetSuccessfully(t)

	client := client.ServiceClient()

	policy, err := Get(client, groupID, webhookPolicyID).Extract()

	th.AssertNoErr(t, err)
	th.CheckDeepEquals(t, WebhookPolicy, *policy)
}

func TestUpdate(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePolicyUpdateSuccessfully(t)

	client := client.ServiceClient()
	opts := UpdateOpts{
		Name:            "updated webhook policy",
		Type:            Webhook,
		Cooldown:        600,
		AdjustmentType:  ChangePercent,
		AdjustmentValue: 6.6,
	}

	err := Update(client, groupID, webhookPolicyID, opts).ExtractErr()

	th.AssertNoErr(t, err)
}

func TestDelete(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePolicyDeleteSuccessfully(t)

	client := client.ServiceClient()
	err := Delete(client, groupID, webhookPolicyID).ExtractErr()

	th.AssertNoErr(t, err)
}

func TestExecute(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()
	HandlePolicyExecuteSuccessfully(t)

	client := client.ServiceClient()
	err := Execute(client, groupID, webhookPolicyID).ExtractErr()

	th.AssertNoErr(t, err)
}

func TestValidateType(t *testing.T) {
	ok := validateType(Schedule)
	th.AssertEquals(t, true, ok)

	ok = validateType(Webhook)
	th.AssertEquals(t, true, ok)

	ok = validateType("BAD")
	th.AssertEquals(t, false, ok)
}

func TestValidateCooldown(t *testing.T) {
	ok := validateCooldown(0)
	th.AssertEquals(t, true, ok)

	ok = validateCooldown(86400)
	th.AssertEquals(t, true, ok)

	ok = validateCooldown(-1)
	th.AssertEquals(t, false, ok)

	ok = validateCooldown(172800)
	th.AssertEquals(t, false, ok)
}

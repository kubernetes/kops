// +build fixtures

package webhooks

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/rackspace/gophercloud"
	th "github.com/rackspace/gophercloud/testhelper"
	"github.com/rackspace/gophercloud/testhelper/client"
)

// WebhookListBody contains the canned body of a webhooks.List response.
const WebhookListBody = `
{
  "webhooks": [
    {
      "id": "2bd1822c-58c5-49fd-8b3d-ed44781a58d1",
      "name": "first hook",
      "links": [
        {
          "href": "https://dfw.autoscale.api.rackspacecloud.com/v1.0/123456/groups/60b15dad-5ea1-43fa-9a12-a1d737b4da07/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks/2bd1822c-58c5-49fd-8b3d-ed44781a58d1/",
          "rel": "self"
        },
        {
          "href": "https://dfw.autoscale.api.rackspacecloud.com/v1.0/execute/1/714c1c17c5e6ea5ef1e710d5ccc62e492575bab5216184d4c27dc0164db1bc06/",
          "rel": "capability"
        }
      ],
      "metadata": {}
    },
    {
      "id": "76711c36-dfbe-4f5e-bea6-cded99690515",
      "name": "second hook",
      "links": [
        {
          "href": "https://dfw.autoscale.api.rackspacecloud.com/v1.0/123456/groups/60b15dad-5ea1-43fa-9a12-a1d737b4da07/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks/76711c36-dfbe-4f5e-bea6-cded99690515/",
          "rel": "self"
        },
        {
          "href": "https://dfw.autoscale.api.rackspacecloud.com/v1.0/execute/1/982e24858723f9e8bc2afea42a73a3c357c8f518857735400a7f7d8b3f14ccdb/",
          "rel": "capability"
        }
      ],
      "metadata": {
        "notes": "a note about this webhook"
      }
    }
  ],
  "webhooks_links": []
}
`

// WebhookCreateBody contains the canned body of a webhooks.Create response.
const WebhookCreateBody = WebhookListBody

// WebhookCreateRequest contains the canned body of a webhooks.Create request.
const WebhookCreateRequest = `
[
  {
    "name": "first hook"
  },
  {
    "name": "second hook",
    "metadata": {
      "notes": "a note about this webhook"
    }
  }
]
`

// WebhookGetBody contains the canned body of a webhooks.Get response.
const WebhookGetBody = `
{
  "webhook": {
    "id": "2bd1822c-58c5-49fd-8b3d-ed44781a58d1",
    "name": "first hook",
    "links": [
      {
        "href": "https://dfw.autoscale.api.rackspacecloud.com/v1.0/123456/groups/60b15dad-5ea1-43fa-9a12-a1d737b4da07/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks/2bd1822c-58c5-49fd-8b3d-ed44781a58d1/",
        "rel": "self"
      },
      {
        "href": "https://dfw.autoscale.api.rackspacecloud.com/v1.0/execute/1/714c1c17c5e6ea5ef1e710d5ccc62e492575bab5216184d4c27dc0164db1bc06/",
        "rel": "capability"
      }
    ],
    "metadata": {}
  }
}
`

// WebhookUpdateRequest contains the canned body of a webhooks.Update request.
const WebhookUpdateRequest = `
{
  "name": "updated hook",
  "metadata": {
    "new-key": "some data"
  }
}
`

var (
	// FirstWebhook is a Webhook corresponding to the first result in WebhookListBody.
	FirstWebhook = Webhook{
		ID:   "2bd1822c-58c5-49fd-8b3d-ed44781a58d1",
		Name: "first hook",
		Links: []gophercloud.Link{
			{
				Href: "https://dfw.autoscale.api.rackspacecloud.com/v1.0/123456/groups/60b15dad-5ea1-43fa-9a12-a1d737b4da07/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks/2bd1822c-58c5-49fd-8b3d-ed44781a58d1/",
				Rel:  "self",
			},
			{
				Href: "https://dfw.autoscale.api.rackspacecloud.com/v1.0/execute/1/714c1c17c5e6ea5ef1e710d5ccc62e492575bab5216184d4c27dc0164db1bc06/",
				Rel:  "capability",
			},
		},
		Metadata: map[string]string{},
	}

	// SecondWebhook is a Webhook corresponding to the second result in WebhookListBody.
	SecondWebhook = Webhook{
		ID:   "76711c36-dfbe-4f5e-bea6-cded99690515",
		Name: "second hook",
		Links: []gophercloud.Link{
			{
				Href: "https://dfw.autoscale.api.rackspacecloud.com/v1.0/123456/groups/60b15dad-5ea1-43fa-9a12-a1d737b4da07/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks/76711c36-dfbe-4f5e-bea6-cded99690515/",
				Rel:  "self",
			},
			{
				Href: "https://dfw.autoscale.api.rackspacecloud.com/v1.0/execute/1/982e24858723f9e8bc2afea42a73a3c357c8f518857735400a7f7d8b3f14ccdb/",
				Rel:  "capability",
			},
		},
		Metadata: map[string]string{
			"notes": "a note about this webhook",
		},
	}
)

// HandleWebhookListSuccessfully sets up the test server to respond to a webhooks List request.
func HandleWebhookListSuccessfully(t *testing.T) {
	path := "/groups/10eb3219-1b12-4b34-b1e4-e10ee4f24c65/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks"

	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")

		fmt.Fprintf(w, WebhookListBody)
	})
}

// HandleWebhookCreateSuccessfully sets up the test server to respond to a webhooks Create request.
func HandleWebhookCreateSuccessfully(t *testing.T) {
	path := "/groups/10eb3219-1b12-4b34-b1e4-e10ee4f24c65/policies/2b48d247-0282-4b9d-8775-5c4b67e8e649/webhooks"

	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "POST")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)
		th.TestHeader(t, r, "Content-Type", "application/json")
		th.TestHeader(t, r, "Accept", "application/json")

		th.TestJSONRequest(t, r, WebhookCreateRequest)

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		fmt.Fprintf(w, WebhookCreateBody)
	})
}

// HandleWebhookGetSuccessfully sets up the test server to respond to a webhooks Get request.
func HandleWebhookGetSuccessfully(t *testing.T) {
	groupID := "10eb3219-1b12-4b34-b1e4-e10ee4f24c65"
	policyID := "2b48d247-0282-4b9d-8775-5c4b67e8e649"
	webhookID := "2bd1822c-58c5-49fd-8b3d-ed44781a58d1"

	path := fmt.Sprintf("/groups/%s/policies/%s/webhooks/%s", groupID, policyID, webhookID)

	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "GET")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.Header().Add("Content-Type", "application/json")

		fmt.Fprintf(w, WebhookGetBody)
	})
}

// HandleWebhookUpdateSuccessfully sets up the test server to respond to a webhooks Update request.
func HandleWebhookUpdateSuccessfully(t *testing.T) {
	groupID := "10eb3219-1b12-4b34-b1e4-e10ee4f24c65"
	policyID := "2b48d247-0282-4b9d-8775-5c4b67e8e649"
	webhookID := "2bd1822c-58c5-49fd-8b3d-ed44781a58d1"

	path := fmt.Sprintf("/groups/%s/policies/%s/webhooks/%s", groupID, policyID, webhookID)

	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "PUT")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		th.TestJSONRequest(t, r, WebhookUpdateRequest)

		w.WriteHeader(http.StatusNoContent)
	})
}

// HandleWebhookDeleteSuccessfully sets up the test server to respond to a webhooks Delete request.
func HandleWebhookDeleteSuccessfully(t *testing.T) {
	groupID := "10eb3219-1b12-4b34-b1e4-e10ee4f24c65"
	policyID := "2b48d247-0282-4b9d-8775-5c4b67e8e649"
	webhookID := "2bd1822c-58c5-49fd-8b3d-ed44781a58d1"

	path := fmt.Sprintf("/groups/%s/policies/%s/webhooks/%s", groupID, policyID, webhookID)

	th.Mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		th.TestMethod(t, r, "DELETE")
		th.TestHeader(t, r, "X-Auth-Token", client.TokenID)

		w.WriteHeader(http.StatusNoContent)
	})
}

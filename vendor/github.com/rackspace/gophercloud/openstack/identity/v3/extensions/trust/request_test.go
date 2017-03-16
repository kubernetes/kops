package trust

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/identity/v3/tokens"
	"github.com/rackspace/gophercloud/testhelper"
)

// authTokenPost verifies that providing certain AuthOptions and Scope results in an expected JSON structure.
func authTokenPost(t *testing.T, options gophercloud.AuthOptions, scope *tokens.Scope, requestJSON string) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	client := gophercloud.ServiceClient{
		ProviderClient: &gophercloud.ProviderClient{
			TokenID: options.TokenID,
		},
		Endpoint: testhelper.Endpoint(),
	}

	testhelper.Mux.HandleFunc("/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		testhelper.TestMethod(t, r, "POST")
		testhelper.TestHeader(t, r, "Content-Type", "application/json")
		testhelper.TestHeader(t, r, "Accept", "application/json")
		testhelper.TestJSONRequest(t, r, requestJSON)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{
			"token": {
				"expires_at": "2014-10-02T13:45:00.000000Z"
			}
		}`)
	})

	_, err := tokens.Create(&client, AuthOptionsExt{AuthOptions: tokens.AuthOptions{options}, TrustID: "123456"}, scope).Extract()
	if err != nil {
		t.Errorf("Create returned an error: %v", err)
	}
}

func authTokenPostErr(t *testing.T, options gophercloud.AuthOptions, scope *tokens.Scope, includeToken bool, expectedErr error) {
	testhelper.SetupHTTP()
	defer testhelper.TeardownHTTP()

	client := gophercloud.ServiceClient{
		ProviderClient: &gophercloud.ProviderClient{},
		Endpoint:       testhelper.Endpoint(),
	}
	if includeToken {
		client.TokenID = "abcdef123456"
	}

	_, err := tokens.Create(&client, AuthOptionsExt{AuthOptions: tokens.AuthOptions{options}, TrustID: "123456"}, scope).Extract()
	if err == nil {
		t.Errorf("Create did NOT return an error")
	}
	if err != expectedErr {
		t.Errorf("Create returned an unexpected error: wanted %v, got %v", expectedErr, err)
	}
}

func TestTrustIDTokenID(t *testing.T) {
	options := gophercloud.AuthOptions{TokenID: "trustee_token"}
	var scope *tokens.Scope
	authTokenPost(t, options, scope, `
		{
		  "auth": {
		    "identity": {
			      "methods": [
				"token"
			      ],
				"token": {
					"id": "trustee_token"
			      }
		    },
		    "scope": {
			      "OS-TRUST:trust": {
			        "id": "123456"
			      }
			    }
			}
		}

	`)
}

func TestFailurePassword(t *testing.T) {
	options := gophercloud.AuthOptions{TokenID: ""}
	//Service Client must have tokenId or password,
	//setting include tokenId to false
	//scope := &Scope{TrustID: "notenough"}
	var scope *tokens.Scope
	authTokenPostErr(t, options, scope, false, tokens.ErrMissingPassword)
}

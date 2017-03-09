package openstack

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/rackspace/gophercloud"
	th "github.com/rackspace/gophercloud/testhelper"
)

func TestAuthenticatedClientV3(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	const ID = "0123456789"

	th.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
			{
				"versions": {
					"values": [
						{
							"status": "stable",
							"id": "v3.0",
							"links": [
								{ "href": "%s", "rel": "self" }
							]
						},
						{
							"status": "stable",
							"id": "v2.0",
							"links": [
								{ "href": "%s", "rel": "self" }
							]
						}
					]
				}
			}
		`, th.Endpoint()+"v3/", th.Endpoint()+"v2.0/")
	})

	th.Mux.HandleFunc("/v3/auth/tokens", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("X-Subject-Token", ID)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{ "token": { "expires_at": "2013-02-02T18:30:59.000000Z" } }`)
	})

	options := gophercloud.AuthOptions{
		UserID:           "me",
		Password:         "secret",
		IdentityEndpoint: th.Endpoint(),
	}
	client, err := AuthenticatedClient(options)
	th.AssertNoErr(t, err)
	th.CheckEquals(t, ID, client.TokenID)
}

func TestAuthenticatedClientV2(t *testing.T) {
	th.SetupHTTP()
	defer th.TeardownHTTP()

	th.Mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
			{
				"versions": {
					"values": [
						{
							"status": "experimental",
							"id": "v3.0",
							"links": [
								{ "href": "%s", "rel": "self" }
							]
						},
						{
							"status": "stable",
							"id": "v2.0",
							"links": [
								{ "href": "%s", "rel": "self" }
							]
						}
					]
				}
			}
		`, th.Endpoint()+"v3/", th.Endpoint()+"v2.0/")
	})

	th.Mux.HandleFunc("/v2.0/tokens", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
			{
				"access": {
					"token": {
						"id": "01234567890",
						"expires": "2014-10-01T10:00:00.000000Z"
					},
					"serviceCatalog": [
						{
							"name": "Cloud Servers",
							"type": "compute",
							"endpoints": [
								{
									"tenantId": "t1000",
									"publicURL": "https://compute.north.host.com/v1/t1000",
									"internalURL": "https://compute.north.internal/v1/t1000",
									"region": "North",
									"versionId": "1",
									"versionInfo": "https://compute.north.host.com/v1/",
									"versionList": "https://compute.north.host.com/"
								},
								{
									"tenantId": "t1000",
									"publicURL": "https://compute.north.host.com/v1.1/t1000",
									"internalURL": "https://compute.north.internal/v1.1/t1000",
									"region": "North",
									"versionId": "1.1",
									"versionInfo": "https://compute.north.host.com/v1.1/",
									"versionList": "https://compute.north.host.com/"
								}
							],
							"endpoints_links": []
						},
						{
							"name": "Cloud Files",
							"type": "object-store",
							"endpoints": [
								{
									"tenantId": "t1000",
									"publicURL": "https://storage.north.host.com/v1/t1000",
									"internalURL": "https://storage.north.internal/v1/t1000",
									"region": "North",
									"versionId": "1",
									"versionInfo": "https://storage.north.host.com/v1/",
									"versionList": "https://storage.north.host.com/"
								},
								{
									"tenantId": "t1000",
									"publicURL": "https://storage.south.host.com/v1/t1000",
									"internalURL": "https://storage.south.internal/v1/t1000",
									"region": "South",
									"versionId": "1",
									"versionInfo": "https://storage.south.host.com/v1/",
									"versionList": "https://storage.south.host.com/"
								}
							]
						}
					]
				}
			}
		`)
	})

	options := gophercloud.AuthOptions{
		Username:         "me",
		Password:         "secret",
		IdentityEndpoint: th.Endpoint(),
	}
	client, err := AuthenticatedClient(options)
	th.AssertNoErr(t, err)
	th.CheckEquals(t, "01234567890", client.TokenID)
}

func TestNewClient(t *testing.T) {
	client, err := NewClient("https://example.com")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/", client.IdentityBase)

	client, err = NewClient("https://example.com/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/", client.IdentityBase)

	client, err = NewClient("https://example.com/v2.0")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/v2.0/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/", client.IdentityBase)

	client, err = NewClient("https://example.com/v2.0/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/v2.0/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/", client.IdentityBase)

	client, err = NewClient("https://example.com/foo/bar")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/foo/bar/", client.IdentityBase)

	client, err = NewClient("https://example.com/foo/bar/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/foo/bar/", client.IdentityBase)

	client, err = NewClient("https://example.com/foo/bar/v2.0")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/foo/bar/v2.0/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/foo/bar/", client.IdentityBase)

	client, err = NewClient("https://example.com/foo/bar/v2.0/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/foo/bar/v2.0/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/foo/bar/", client.IdentityBase)

	client, err = NewClient("https://example.com/foo/bar/v3/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/foo/bar/v3/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/foo/bar/", client.IdentityBase)

	client, err = NewClient("https://example.com/v3")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/v3/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/", client.IdentityBase)

	client, err = NewClient("https://example.com/v3/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://example.com/v3/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/", client.IdentityBase)

	_, err = NewClient("https://example.com/v2.3/")
	th.AssertErr(t, err)
	expected := fmt.Errorf("Invalid identity endpoint version %v. Supported versions: v2.0, v3", "v2.3")
	th.AssertEquals(t, expected.Error(), err.Error())

	_, err = NewClient("https://example.com/v2.0/foo")
	th.AssertErr(t, err)
	expected = fmt.Errorf("Path suffixes (after version) are not supported.")
	th.AssertEquals(t, expected.Error(), err.Error())

	// Does not match regexp, include to base
	client, err = NewClient("https://example.com/v2a0/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/v2a0/", client.IdentityBase)

	// Does not match regexp, include to base
	client, err = NewClient("https://example.com/v3api")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "", client.IdentityEndpoint)
	th.AssertEquals(t, "https://example.com/v3api/", client.IdentityBase)

	// Matches regexp, invalid version v20.
	_, err = NewClient("https://example.com/v20./")
	th.AssertErr(t, err)
	expected = fmt.Errorf("Invalid identity endpoint version %v. Supported versions: v2.0, v3", "v20.")
	th.AssertEquals(t, expected.Error(), err.Error())

	// domain contains version
	client, err = NewClient("https://v3.v2.0example.com/v3/")
	th.AssertNoErr(t, err)
	th.AssertEquals(t, "https://v3.v2.0example.com/v3/", client.IdentityEndpoint)
	th.AssertEquals(t, "https://v3.v2.0example.com/", client.IdentityBase)

}

package policies

import (
	"testing"

	"github.com/rackspace/gophercloud"
	th "github.com/rackspace/gophercloud/testhelper"
)

const endpoint = "http://localhost:57909/"

func endpointClient() *gophercloud.ServiceClient {
	return &gophercloud.ServiceClient{Endpoint: endpoint}
}

func TestListURL(t *testing.T) {
	actual := listURL(endpointClient(), "123")
	expected := endpoint + "groups/123/policies"
	th.CheckEquals(t, expected, actual)
}

func TestCreateURL(t *testing.T) {
	actual := createURL(endpointClient(), "123")
	expected := endpoint + "groups/123/policies"
	th.CheckEquals(t, expected, actual)
}

func TestGetURL(t *testing.T) {
	actual := getURL(endpointClient(), "123", "456")
	expected := endpoint + "groups/123/policies/456"
	th.CheckEquals(t, expected, actual)
}

func TestUpdateURL(t *testing.T) {
	actual := updateURL(endpointClient(), "123", "456")
	expected := endpoint + "groups/123/policies/456"
	th.CheckEquals(t, expected, actual)
}

func TestDeleteURL(t *testing.T) {
	actual := deleteURL(endpointClient(), "123", "456")
	expected := endpoint + "groups/123/policies/456"
	th.CheckEquals(t, expected, actual)
}

func TestExecuteURL(t *testing.T) {
	actual := executeURL(endpointClient(), "123", "456")
	expected := endpoint + "groups/123/policies/456/execute"
	th.CheckEquals(t, expected, actual)
}

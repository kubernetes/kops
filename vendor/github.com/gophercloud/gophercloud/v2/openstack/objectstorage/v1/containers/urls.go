package containers

import (
	"net/url"

	"github.com/gophercloud/gophercloud/v2"
	v1 "github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1"
)

func listURL(c *gophercloud.ServiceClient) string {
	return c.Endpoint
}

func createURL(c *gophercloud.ServiceClient, container string) (string, error) {
	if err := v1.CheckContainerName(container); err != nil {
		return "", err
	}
	return c.ServiceURL(url.PathEscape(container)), nil
}

func getURL(c *gophercloud.ServiceClient, container string) (string, error) {
	return createURL(c, container)
}

func deleteURL(c *gophercloud.ServiceClient, container string) (string, error) {
	return createURL(c, container)
}

func updateURL(c *gophercloud.ServiceClient, container string) (string, error) {
	return createURL(c, container)
}

func bulkDeleteURL(c *gophercloud.ServiceClient) string {
	return c.Endpoint + "?bulk-delete=true"
}

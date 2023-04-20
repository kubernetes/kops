package objects

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/objectstorage/v1/containers"
)

func listURL(c *gophercloud.ServiceClient, container string) (string, error) {
	if err := containers.CheckContainerName(container); err != nil {
		return "", err
	}
	return c.ServiceURL(container), nil
}

func copyURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	if err := containers.CheckContainerName(container); err != nil {
		return "", err
	}
	return c.ServiceURL(container, object), nil
}

func createURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	return copyURL(c, container, object)
}

func getURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	return copyURL(c, container, object)
}

func deleteURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	return copyURL(c, container, object)
}

func downloadURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	return copyURL(c, container, object)
}

func updateURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	return copyURL(c, container, object)
}

func bulkDeleteURL(c *gophercloud.ServiceClient) string {
	return c.Endpoint + "?bulk-delete=true"
}

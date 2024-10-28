package objects

import (
	"net/url"

	"github.com/gophercloud/gophercloud/v2"
	v1 "github.com/gophercloud/gophercloud/v2/openstack/objectstorage/v1"
)

// tempURL returns an unescaped virtual path to generate the HMAC signature.
// Names must not be URL-encoded in this case.
//
// See: https://docs.openstack.org/swift/latest/api/temporary_url_middleware.html#hmac-signature-for-temporary-urls
func tempURL(c *gophercloud.ServiceClient, container, object string) string {
	return c.ServiceURL(container, object)
}

func listURL(c *gophercloud.ServiceClient, container string) (string, error) {
	if err := v1.CheckContainerName(container); err != nil {
		return "", err
	}
	return c.ServiceURL(url.PathEscape(container)), nil
}

func copyURL(c *gophercloud.ServiceClient, container, object string) (string, error) {
	if err := v1.CheckContainerName(container); err != nil {
		return "", err
	}
	if err := v1.CheckObjectName(object); err != nil {
		return "", err
	}
	return c.ServiceURL(url.PathEscape(container), url.PathEscape(object)), nil
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

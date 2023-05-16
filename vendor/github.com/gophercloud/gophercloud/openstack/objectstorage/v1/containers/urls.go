package containers

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud"
)

const forbiddenContainerRunes = "/"

func CheckContainerName(s string) error {
	if strings.ContainsAny(s, forbiddenContainerRunes) {
		return ErrInvalidContainerName{}
	}

	// The input could (and should) already have been escaped. This cycle
	// checks for the escaped versions of the forbidden characters. Note
	// that a simple "contains" is sufficient, because Go's http library
	// won't accept invalid escape sequences (e.g. "%%2F").
	for _, r := range forbiddenContainerRunes {
		if strings.Contains(strings.ToLower(s), fmt.Sprintf("%%%x", r)) {
			return ErrInvalidContainerName{}
		}
	}
	return nil
}

func listURL(c *gophercloud.ServiceClient) string {
	return c.Endpoint
}

func createURL(c *gophercloud.ServiceClient, container string) (string, error) {
	if err := CheckContainerName(container); err != nil {
		return "", err
	}
	return c.ServiceURL(container), nil
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

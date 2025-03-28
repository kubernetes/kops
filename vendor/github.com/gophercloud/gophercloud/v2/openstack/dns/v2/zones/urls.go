package zones

import "github.com/gophercloud/gophercloud/v2"

// baseURL returns the base URL for zones.
func baseURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("zones")
}

// zoneURL returns the URL for a specific zone.
func zoneURL(c *gophercloud.ServiceClient, zoneID string) string {
	return c.ServiceURL("zones", zoneID)
}

// zoneShareURL returns the URL for sharing a zone.
func zoneShareURL(c *gophercloud.ServiceClient, zoneID string) string {
	return c.ServiceURL("zones", zoneID, "shares")
}

// zoneUnshareURL returns the URL for unsharing a zone.
func zoneUnshareURL(c *gophercloud.ServiceClient, zoneID, shareID string) string {
	return c.ServiceURL("zones", zoneID, "shares", shareID)
}

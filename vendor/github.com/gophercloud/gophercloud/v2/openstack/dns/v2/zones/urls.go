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

// sharesBaseURL returns the URL for shared zones.
func sharesBaseURL(c *gophercloud.ServiceClient, zoneID string) string {
	return c.ServiceURL("zones", zoneID, "shares")
}

// shareURL returns the URL for a shared zone.
func shareURL(c *gophercloud.ServiceClient, zoneID, sharedZoneID string) string {
	return c.ServiceURL("zones", zoneID, "shares", sharedZoneID)
}

package recordsets

import "github.com/gophercloud/gophercloud/v2"

func baseURL(c *gophercloud.ServiceClient, zoneID string) string {
	return c.ServiceURL("zones", zoneID, "recordsets")
}

func rrsetURL(c *gophercloud.ServiceClient, zoneID string, rrsetID string) string {
	return c.ServiceURL("zones", zoneID, "recordsets", rrsetID)
}

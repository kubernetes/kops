package policies

import "github.com/rackspace/gophercloud"

func listURL(c *gophercloud.ServiceClient, groupID string) string {
	return c.ServiceURL("groups", groupID, "policies")
}

func createURL(c *gophercloud.ServiceClient, groupID string) string {
	return c.ServiceURL("groups", groupID, "policies")
}

func getURL(c *gophercloud.ServiceClient, groupID, policyID string) string {
	return c.ServiceURL("groups", groupID, "policies", policyID)
}

func updateURL(c *gophercloud.ServiceClient, groupID, policyID string) string {
	return getURL(c, groupID, policyID)
}

func deleteURL(c *gophercloud.ServiceClient, groupID, policyID string) string {
	return getURL(c, groupID, policyID)
}

func executeURL(c *gophercloud.ServiceClient, groupID, policyID string) string {
	return c.ServiceURL("groups", groupID, "policies", policyID, "execute")
}

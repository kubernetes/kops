package volumeactions

import "github.com/rackspace/gophercloud"

func attachURL(c *gophercloud.ServiceClient, id string) string {
	return c.ServiceURL("volumes", id, "action")
}

func detachURL(c *gophercloud.ServiceClient, id string) string {
	return attachURL(c, id)
}

func reserveURL(c *gophercloud.ServiceClient, id string) string {
	return attachURL(c, id)
}

func unreserveURL(c *gophercloud.ServiceClient, id string) string {
	return attachURL(c, id)
}

func initializeConnectionURL(c *gophercloud.ServiceClient, id string) string {
	return attachURL(c, id)
}

func teminateConnectionURL(c *gophercloud.ServiceClient, id string) string {
	return attachURL(c, id)
}

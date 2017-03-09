package meters

import "github.com/rackspace/gophercloud"

func listURL(client *gophercloud.ServiceClient) string {
	return client.ServiceURL("v2", "meters")
}

func showURL(client *gophercloud.ServiceClient, name string) string {
	return client.ServiceURL("v2", "meters", name)
}

func createURL(client *gophercloud.ServiceClient, name string) string {
	return client.ServiceURL("v2", "meters", name)
}

func statisticsURL(client *gophercloud.ServiceClient, name string) string {
	return client.ServiceURL("v2", "meters", name, "statistics")
}

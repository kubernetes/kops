package scaleway

import (
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	scw "k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

const (
	resourceTypeDNSRecord    = "dns-record"
	resourceTypeLoadBalancer = "load-balancer"
	resourceTypeVolume       = "volume"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud scw.ScwCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listDNSRecords,
		listLoadBalancers,
		listServers,
		listVolumes,
	}

	for _, fn := range listFunctions {
		rt, err := fn(cloud, clusterName)
		if err != nil {
			return nil, err
		}
		for _, t := range rt {
			resourceTrackers[t.Type+":"+t.ID] = t
		}
	}

	return resourceTrackers, nil
}

func listDNSRecords(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	//TODO: implement this function

	//c := cloud.(scw.ScwCloud)
	resourcesTrackers := []*resources.Resource(nil)

	return resourcesTrackers, nil
}

func listLoadBalancers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	//TODO: implement this function

	//c := cloud.(scw.ScwCloud)
	resourcesTrackers := []*resources.Resource(nil)

	return resourcesTrackers, nil
}

func listServers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	//TODO: implement this function

	//c := cloud.(scw.ScwCloud)
	resourcesTrackers := []*resources.Resource(nil)

	return resourcesTrackers, nil
}

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	//TODO: implement this function

	//c := cloud.(scw.ScwCloud)
	resourcesTrackers := []*resources.Resource(nil)

	return resourcesTrackers, nil
}

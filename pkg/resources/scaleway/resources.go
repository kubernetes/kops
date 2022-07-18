package scaleway

import (
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	resourceTypeDNSRecord    = "dns-record"
	resourceTypeLoadBalancer = "load-balancer"
	resourceTypeVolume       = "volume"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud scw.ScalewayCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listVolumes,
		listDNSRecords,
		listLoadBalancers,
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

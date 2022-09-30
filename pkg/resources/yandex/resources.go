/*
Copyright 2022 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package yandex

import (
	"k8s.io/kops/upup/pkg/fi/cloudup/yandex"
	// TODO: yandex impement resources
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	/*
		resourceTypeSSHKey       = "ssh-key"
		resourceTypeNetwork      = "network"
		resourceTypeFirewall     = "firewall"
		resourceTypeLoadBalancer = "load-balancer"
		resourceTypeServer       = "server"
	*/
	resourceTypeVolume = "volume"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud yandex.YandexCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{ /*
			listSSHKeys,
			listNetworks,
			listFirewalls,
			listLoadBalancers,
			listServers, */
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

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	return resourceTrackers, nil
}

func deleteVolume(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting Volume: %s(%s)", r.Name, r.ID)
	return nil
}

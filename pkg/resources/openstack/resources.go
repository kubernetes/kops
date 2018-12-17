/*
Copyright 2018 The Kubernetes Authors.

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

package openstack

import (
	"fmt"
	"strings"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/subnets"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type openstackListFn func() ([]*resources.Resource, error)

const (
	typeSSHKey   = "SSHKey"
	typeRouterIF = "Router-IF"
	typeRouter   = "Router"
	typeSubnet   = "Subnet"
	typeNetwork  = "Network"
)

func ListResources(cloud openstack.OpenstackCloud, clusterName string) (map[string]*resources.Resource, error) {
	resources := make(map[string]*resources.Resource)

	os := &clusterDiscoveryOS{
		cloud:       cloud,
		osCloud:     cloud,
		clusterName: clusterName,
	}

	listFunctions := []openstackListFn{
		os.ListKeypairs,
		os.ListNetwork,
	}
	for _, fn := range listFunctions {
		resourceTrackers, err := fn()
		if err != nil {
			return nil, err
		}
		for _, t := range resourceTrackers {
			resources[t.Type+":"+t.ID] = t
		}
	}
	return resources, nil
}

type clusterDiscoveryOS struct {
	cloud       fi.Cloud
	osCloud     openstack.OpenstackCloud
	clusterName string

	zones []string
}

func openstackKeyPairName(org string) string {
	name := strings.Replace(org, ".", "-", -1)
	name = strings.Replace(name, ":", "_", -1)
	return name
}

func (os *clusterDiscoveryOS) ListNetwork() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource

	opt := networks.ListOpts{
		Name: os.clusterName,
	}
	networks, err := os.osCloud.ListNetworks(opt)
	if err != nil {
		return resourceTrackers, err
	}

	for _, network := range networks {
		optRouter := routers.ListOpts{
			Name: strings.Replace(os.clusterName, ".", "-", -1),
		}
		routers, err := os.osCloud.ListRouters(optRouter)
		if err != nil {
			return resourceTrackers, err
		}
		for _, router := range routers {
			resourceTracker := &resources.Resource{
				Name:    router.Name,
				ID:      router.ID,
				Type:    typeRouter,
				Deleter: DeleteRouter,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}

		optSubnet := subnets.ListOpts{
			NetworkID:  network.ID,
		}
		subnets, err := os.osCloud.ListSubnets(optSubnet)
		if err != nil {
			return resourceTrackers, err
		}
		for _, subnet := range subnets {
			// router interfaces
			for _, router := range routers {
				resourceTracker := &resources.Resource{
					Name:    router.ID,
					ID:      subnet.ID,
					Type:    typeRouterIF,
					Deleter: DeleteRouterIF,
				}
				resourceTrackers = append(resourceTrackers, resourceTracker)
			}
			resourceTracker := &resources.Resource{
				Name:    subnet.Name,
				ID:      subnet.ID,
				Type:    typeSubnet,
				Deleter: DeleteSubnet,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
		resourceTracker := &resources.Resource{
			Name:    network.Name,
			ID:      network.ID,
			Type:    typeNetwork,
			Deleter: DeleteNetwork,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}
	return resourceTrackers, nil
}

func DeleteRouterIF(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(openstack.OpenstackCloud)
	name := r.Name

	glog.V(2).Infof("Deleting Openstack router interface %q", name)

	opts := routers.RemoveInterfaceOpts{
		SubnetID: r.ID,
	}
	err := c.DeleteRouterInterface(name, opts)
	if err != nil {
		return fmt.Errorf("error deleting router interface %q: %v", name, err)
	}
	return nil
}

func DeleteRouter(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(openstack.OpenstackCloud)
	name := r.Name

	glog.V(2).Infof("Deleting Openstack router %q", name)

	err := c.DeleteRouter(r.ID)
	if err != nil {
		return fmt.Errorf("error deleting router %q: %v", name, err)
	}
	return nil
}

func DeleteSubnet(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(openstack.OpenstackCloud)
	name := r.Name

	glog.V(2).Infof("Deleting Openstack subnet %q", name)

	err := c.DeleteSubnet(r.ID)
	if err != nil {
		return fmt.Errorf("error deleting subnet %q: %v", name, err)
	}
	return nil
}

func DeleteNetwork(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(openstack.OpenstackCloud)
	name := r.Name

	glog.V(2).Infof("Deleting Openstack network %q", name)

	err := c.DeleteNetwork(r.ID)
	if err != nil {
		return fmt.Errorf("error deleting network %q: %v", name, err)
	}
	return nil
}

func (os *clusterDiscoveryOS) ListKeypairs() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	ks, err := os.osCloud.ListKeypairs()
	if err != nil {
		return resourceTrackers, err
	}

	for _, key := range ks {
		prefix := "kubernetes-" + openstackKeyPairName(os.clusterName)
		if strings.HasPrefix(key.Name, prefix) {
			resourceTracker := &resources.Resource{
				Name:    key.Name,
				ID:      key.Name,
				Type:    typeSSHKey,
				Deleter: DeleteSSHKey,
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}

func DeleteSSHKey(cloud fi.Cloud, r *resources.Resource) error {
	c := cloud.(openstack.OpenstackCloud)
	name := r.Name

	glog.V(2).Infof("Deleting Openstack Keypair %q", name)

	err := c.DeleteKeyPair(r.Name)
	if err != nil {
		return fmt.Errorf("error deleting KeyPair %q: %v", name, err)
	}
	return nil
}

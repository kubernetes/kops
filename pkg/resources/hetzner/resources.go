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

package hetzner

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
)

const (
	resourceTypeSSHKey       = "ssh-key"
	resourceTypeNetwork      = "network"
	resourceTypeFirewall     = "firewall"
	resourceTypeLoadBalancer = "load-balancer"
	resourceTypeServer       = "server"
	resourceTypeVolume       = "volume"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud hetzner.HetznerCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listSSHKeys,
		listNetworks,
		listFirewalls,
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

func listSSHKeys(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(hetzner.HetznerCloud)
	var resourceTrackers []*resources.Resource

	sshKeys, err := c.GetSSHKeys(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list ssh keys: %w", err)
	}

	for _, sshKey := range sshKeys {
		resourceTracker := &resources.Resource{
			Name:    sshKey.Name,
			ID:      strconv.Itoa(sshKey.ID),
			Type:    resourceTypeSSHKey,
			Deleter: deleteSSHKey,
			Obj:     sshKey,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listNetworks(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(hetzner.HetznerCloud)
	var resourceTrackers []*resources.Resource

	networks, err := c.GetNetworks(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	for _, network := range networks {
		resourceTracker := &resources.Resource{
			Name:    network.Name,
			ID:      strconv.Itoa(network.ID),
			Type:    resourceTypeNetwork,
			Deleter: deleteNetwork,
			Obj:     network,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listFirewalls(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(hetzner.HetznerCloud)
	var resourceTrackers []*resources.Resource

	firewalls, err := c.GetFirewalls(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list firewalls: %w", err)
	}

	for _, firewall := range firewalls {
		resourceTracker := &resources.Resource{
			Name:    firewall.Name,
			ID:      strconv.Itoa(firewall.ID),
			Type:    resourceTypeFirewall,
			Deleter: deleteFirewall,
			Obj:     firewall,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listLoadBalancers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(hetzner.HetznerCloud)
	var resourceTrackers []*resources.Resource

	loadBalancers, err := c.GetLoadBalancers(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list load balancers: %w", err)
	}

	for _, loadBalancer := range loadBalancers {
		resourceTracker := &resources.Resource{
			Name:    loadBalancer.Name,
			ID:      strconv.Itoa(loadBalancer.ID),
			Type:    resourceTypeLoadBalancer,
			Deleter: deleteLoadBalancer,
			Obj:     loadBalancer,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listServers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(hetzner.HetznerCloud)
	var resourceTrackers []*resources.Resource

	servers, err := c.GetServers(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}

	for _, server := range servers {
		resourceTracker := &resources.Resource{
			Name:    server.Name,
			ID:      strconv.Itoa(server.ID),
			Type:    resourceTypeServer,
			Deleter: deleteServer,
			Obj:     server,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(hetzner.HetznerCloud)
	var resourceTrackers []*resources.Resource

	volumes, err := c.GetVolumes(clusterName)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}

	for _, volume := range volumes {
		resourceTracker := &resources.Resource{
			Name:    volume.Name,
			ID:      strconv.Itoa(volume.ID),
			Type:    resourceTypeVolume,
			Deleter: deleteVolume,
			Obj:     volume,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteSSHKey(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting SSH Key: %s(%s)", r.Name, r.ID)

	c := cloud.(hetzner.HetznerCloud)
	client := c.SSHKeyClient()
	sshKey := r.Obj.(*hcloud.SSHKey)
	_, err := client.Delete(context.TODO(), sshKey)
	if err != nil {
		return fmt.Errorf("failed to delete ssh key %s(%s): %w", r.Name, r.ID, err)
	}

	return nil
}

func deleteNetwork(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting Network: %s(%s)", r.Name, r.ID)

	c := cloud.(hetzner.HetznerCloud)
	client := c.NetworkClient()
	network := r.Obj.(*hcloud.Network)
	_, err := client.Delete(context.TODO(), network)
	if err != nil {
		return fmt.Errorf("failed to delete network %s(%s): %w", r.Name, r.ID, err)
	}

	return nil
}

func deleteFirewall(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting Firewall: %s(%s)", r.Name, r.ID)

	c := cloud.(hetzner.HetznerCloud)
	client := c.FirewallClient()
	firewall := r.Obj.(*hcloud.Firewall)
	_, err := client.Delete(context.TODO(), firewall)
	if err != nil {
		return fmt.Errorf("failed to delete firewall %s(%s): %w", r.Name, r.ID, err)
	}

	return nil
}

func deleteLoadBalancer(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting LoadBalancer: %s(%s)", r.Name, r.ID)

	c := cloud.(hetzner.HetznerCloud)
	client := c.LoadBalancerClient()
	loadBalancer := r.Obj.(*hcloud.LoadBalancer)
	_, err := client.Delete(context.TODO(), loadBalancer)
	if err != nil {
		return fmt.Errorf("failed to delete load balancer %s(%s): %w", r.Name, r.ID, err)
	}

	return nil
}

func deleteServer(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting Server: %s(%s)", r.Name, r.ID)

	c := cloud.(hetzner.HetznerCloud)
	client := c.ServerClient()
	server := r.Obj.(*hcloud.Server)
	_, err := client.Delete(context.TODO(), server)
	if err != nil {
		return fmt.Errorf("failed to delete server %s(%s): %w", r.Name, r.ID, err)
	}

	return nil
}

func deleteVolume(cloud fi.Cloud, r *resources.Resource) error {
	klog.Infof("Deleting Volume: %s(%s)", r.Name, r.ID)

	c := cloud.(hetzner.HetznerCloud)
	client := c.VolumeClient()
	volume := r.Obj.(*hcloud.Volume)
	_, err := client.Delete(context.TODO(), volume)
	if err != nil {
		return fmt.Errorf("failed to delete volume %s(%s): %w", r.Name, r.ID, err)
	}

	return nil
}

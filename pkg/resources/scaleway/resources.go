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

package scaleway

import (
	"fmt"
	"strings"

	domain "github.com/scaleway/scaleway-sdk-go/api/domain/v2beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"

	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
)

const (
	resourceTypeDNSRecord    = "dns-record"
	resourceTypeLoadBalancer = "load-balancer"
	resourceTypeServer       = "server"
	resourceTypeServerIP     = "server-IP"
	resourceTypeSSHKey       = "ssh-key"
	resourceTypeVolume       = "volume"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud scaleway.ScwCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)
	clusterName := clusterInfo.Name

	listFunctions := []listFn{
		listLoadBalancers,
		listServers,
		listServerIPs,
		listSSHKeys,
		listVolumes,
	}
	if !strings.HasSuffix(clusterName, ".k8s.local") && !clusterInfo.UsesNoneDNS {
		listFunctions = append(listFunctions, listDNSRecords)
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
	c := cloud.(scaleway.ScwCloud)
	records, err := c.GetClusterDNSRecords(clusterName)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, record := range records {
		resourceTracker := &resources.Resource{
			Name: record.Name,
			ID:   record.ID,
			Type: resourceTypeDNSRecord,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteDNSRecord(cloud, tracker, clusterName)
			},
			Obj: record,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listLoadBalancers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)
	lbs, err := c.GetClusterLoadBalancers(clusterName)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, loadBalancer := range lbs {
		resourceTracker := &resources.Resource{
			Name:    loadBalancer.Name,
			ID:      loadBalancer.ID,
			Type:    resourceTypeLoadBalancer,
			Deleter: deleteLoadBalancer,
			Obj:     loadBalancer,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listServers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)
	servers, err := c.GetClusterServers(clusterName, nil)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, server := range servers {
		resourceTracker := &resources.Resource{
			Name:    server.Name,
			ID:      server.ID,
			Type:    resourceTypeServer,
			Deleter: deleteServer,
			Obj:     server,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listServerIPs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)

	ips, err := c.InstanceService().ListIPs(&instance.ListIPsRequest{
		Zone: scw.Zone(c.Zone()),
		Tags: []string{fmt.Sprintf("%s=%s", scaleway.TagClusterName, clusterName)},
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("listing IPs for deletion: %w", err)
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, ip := range ips.IPs {
		resourceTracker := &resources.Resource{
			Name:    ip.Address.String(),
			ID:      ip.ID,
			Type:    resourceTypeServerIP,
			Deleter: deleteServerIP,
			Obj:     ip,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listSSHKeys(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)
	sshkeys, err := c.GetClusterSSHKeys(clusterName)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, sshkey := range sshkeys {
		resourceTracker := &resources.Resource{
			Name:    sshkey.Name,
			ID:      sshkey.ID,
			Type:    resourceTypeSSHKey,
			Deleter: deleteSSHKey,
			Obj:     sshkey,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)
	volumes, err := c.GetClusterVolumes(clusterName)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, volume := range volumes {
		resourceTracker := &resources.Resource{
			Name:    volume.Name,
			ID:      volume.ID,
			Type:    resourceTypeVolume,
			Deleter: deleteVolume,
			Obj:     volume,
		}
		if volume.Server != nil {
			resourceTracker.Blocked = []string{resourceTypeServer + ":" + volume.Server.ID}
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteDNSRecord(cloud fi.Cloud, tracker *resources.Resource, domainName string) error {
	c := cloud.(scaleway.ScwCloud)
	record := tracker.Obj.(*domain.Record)

	return c.DeleteDNSRecord(record, domainName)
}

func deleteLoadBalancer(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	loadBalancer := tracker.Obj.(*lb.LB)

	return c.DeleteLoadBalancer(loadBalancer)
}

func deleteServer(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	server := tracker.Obj.(*instance.Server)

	return c.DeleteServer(server)
}

func deleteServerIP(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	ip := tracker.Obj.(*instance.IP)

	err := c.InstanceService().DeleteIP(&instance.DeleteIPRequest{
		Zone: ip.Zone,
		IP:   ip.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete instance IP %s: %w", ip.Address.String(), err)
	}

	return nil
}

func deleteSSHKey(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	sshkey := tracker.Obj.(*iam.SSHKey)

	return c.DeleteSSHKey(sshkey)
}

func deleteVolume(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	volume := tracker.Obj.(*instance.Volume)

	return c.DeleteVolume(volume)
}

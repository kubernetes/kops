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
	iam "github.com/scaleway/scaleway-sdk-go/api/iam/v1alpha1"
	"github.com/scaleway/scaleway-sdk-go/api/instance/v1"
	"github.com/scaleway/scaleway-sdk-go/api/lb/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpc/v1"
	"github.com/scaleway/scaleway-sdk-go/api/vpcgw/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
)

const (
	resourceTypeDNSRecord    = "dns-record"
	resourceTypeGateway      = "gateway"
	resourceTypeLoadBalancer = "load-balancer"
	resourceTypeServer       = "server"
	resourceTypeSSHKey       = "ssh-key"
	resourceTypeVolume       = "volume"
	resourceTypeVPC          = "vpc"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud scaleway.ScwCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)
	clusterName := clusterInfo.Name

	listFunctions := []listFn{
		listDNSRecords,
		listGateways,
		listLoadBalancers,
		listServers,
		listSSHKeys,
		listVolumes,
		listVPCs,
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

	if strings.HasSuffix(clusterName, ".k8s.local") {
		return nil, nil
	}

	names := strings.SplitN(clusterName, ".", 2)
	clusterNameShort := names[0]
	domainName := names[1]

	records, err := c.DomainService().ListDNSZoneRecords(&domain.ListDNSZoneRecordsRequest{
		DNSZone: domainName,
	}, scw.WithAllPages())
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %s", err)
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, record := range records.Records {
		if !strings.HasSuffix(record.Name, clusterNameShort) {
			continue
		}
		resourceTracker := &resources.Resource{
			Name: record.Name,
			ID:   record.ID,
			Type: resourceTypeDNSRecord,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteDNSRecord(cloud, tracker, domainName)
			},
			Obj: record,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listGateways(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)
	gws, err := c.GetClusterGateways(clusterName)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, gw := range gws {
		resourceTracker := &resources.Resource{
			Name: gw.Name,
			ID:   gw.ID,
			Type: resourceTypeGateway,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteGateway(cloud, tracker)
			},
			Obj: gw,
		}
		for _, gwNetwork := range gw.GatewayNetworks {
			resourceTracker.Blocks = append(resourceTracker.Blocks, resourceTypeVPC+":"+gwNetwork.PrivateNetworkID)
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
			Name: loadBalancer.Name,
			ID:   loadBalancer.ID,
			Type: resourceTypeLoadBalancer,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteLoadBalancer(cloud, tracker)
			},
			Obj: loadBalancer,
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
			Name: server.Name,
			ID:   server.ID,
			Type: resourceTypeServer,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteServer(cloud, tracker)
			},
			Obj: server,
		}
		for _, privateNic := range server.PrivateNics {
			resourceTracker.Blocks = append(resourceTracker.Blocks, resourceTypeVPC+":"+privateNic.PrivateNetworkID)
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
			Name: sshkey.Name,
			ID:   sshkey.ID,
			Type: resourceTypeSSHKey,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteSSHKey(cloud, tracker)
			},
			Obj: sshkey,
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
			Name: volume.Name,
			ID:   volume.ID,
			Type: resourceTypeVolume,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteVolume(cloud, tracker)
			},
			Obj: volume,
		}
		if volume.Server != nil {
			resourceTracker.Blocked = []string{resourceTypeServer + ":" + volume.Server.ID}
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listVPCs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(scaleway.ScwCloud)
	vpcs, err := c.GetClusterVPCs(clusterName)
	if err != nil {
		return nil, err
	}

	resourceTrackers := []*resources.Resource(nil)
	for _, vpc := range vpcs {
		resourceTracker := &resources.Resource{
			Name: vpc.Name,
			ID:   vpc.ID,
			Type: resourceTypeVPC,
			Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
				return deleteVPC(cloud, tracker)
			},
			Obj: vpc,
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func deleteDNSRecord(cloud fi.Cloud, tracker *resources.Resource, domainName string) error {
	c := cloud.(scaleway.ScwCloud)
	record := tracker.Obj.(*domain.Record)

	return c.DeleteRecord(record, domainName)
}

func deleteGateway(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	gateway := tracker.Obj.(*vpcgw.Gateway)

	return c.DeleteGateway(gateway)
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

func deleteVPC(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(scaleway.ScwCloud)
	privateNetwork := tracker.Obj.(*vpc.PrivateNetwork)

	return c.DeleteVPC(privateNetwork)
}

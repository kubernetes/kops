/*
Copyright 2019 The Kubernetes Authors.

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

package digitalocean

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	"k8s.io/klog/v2"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
)

const (
	resourceTypeDroplet      = "droplet"
	resourceTypeVolume       = "volume"
	resourceTypeDNSRecord    = "dns-record"
	resourceTypeLoadBalancer = "loadbalancer"
	resourceTypeVPC          = "vpc"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud do.DOCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listVolumes,
		listDroplets,
		listDNS,
		listLoadBalancers,
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

func listDroplets(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(do.DOCloud)
	var resourceTrackers []*resources.Resource

	clusterTag := "KubernetesCluster:" + strings.Replace(clusterName, ".", "-", -1)

	droplets, err := c.GetAllDropletsByTag(clusterTag)
	if err != nil {
		return nil, fmt.Errorf("failed to list droplets: %v", err)
	}

	for _, droplet := range droplets {
		resourceTracker := &resources.Resource{
			Name:    droplet.Name,
			ID:      strconv.Itoa(droplet.ID),
			Type:    resourceTypeDroplet,
			Deleter: deleteDroplet,
			Dumper:  dumpDroplet,
			Obj:     droplet,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(do.DOCloud)
	var resourceTrackers []*resources.Resource

	volumeMatch := strings.Replace(clusterName, ".", "-", -1)

	volumes, err := c.GetAllVolumesByRegion()
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %s", err)
	}

	for _, volume := range volumes {
		if strings.Contains(volume.Name, volumeMatch) {
			resourceTracker := &resources.Resource{
				Name:    volume.Name,
				ID:      volume.ID,
				Type:    resourceTypeVolume,
				Deleter: deleteVolume,
				Obj:     volume,
			}

			var blocks []string
			for _, dropletID := range volume.DropletIDs {
				blocks = append(blocks, "droplet:"+strconv.Itoa(dropletID))
			}

			resourceTracker.Blocks = blocks
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}

func listDNS(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(do.DOCloud)
	domains, _, err := c.DomainService().List(context.TODO(), &godo.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %s", err)
	}

	var domainName string
	for _, domain := range domains {
		if strings.HasSuffix(clusterName, domain.Name) {
			domainName = domain.Name
		}
	}

	if domainName == "" {
		if strings.HasSuffix(clusterName, ".k8s.local") {
			klog.Info("Domain Name is empty. Ok to have an empty domain name since cluster is configured as gossip cluster.")
			return nil, nil
		}

		return nil, fmt.Errorf("failed to find domain for cluster: %s", clusterName)
	}

	records, err := getAllRecordsByDomain(c, domainName)
	if err != nil {
		return nil, fmt.Errorf("failed to list records for domain %s: %s", domainName, err)
	}

	var resourceTrackers []*resources.Resource
	for _, record := range records {
		if !strings.HasSuffix(dns.EnsureDotSuffix(record.Name)+domainName, clusterName) {
			continue
		}

		// kops for digitalocean should only create A records
		// in the future that may change but for now this provides a safe filter
		// in case users assign NS records for the cluster subdomain
		if record.Type != "A" {
			continue
		}

		resourceTracker := &resources.Resource{
			Name: record.Name,
			ID:   strconv.Itoa(record.ID),
			Type: resourceTypeDNSRecord,
			Deleter: func(cloud fi.Cloud, resourceTracker *resources.Resource) error {
				return deleteRecord(cloud, domainName, resourceTracker)
			},
			Obj: record,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func getAllRecordsByDomain(cloud do.DOCloud, domain string) ([]godo.DomainRecord, error) {
	allRecords := []godo.DomainRecord{}

	opt := &godo.ListOptions{}
	for {
		records, resp, err := cloud.DomainService().Records(context.TODO(), domain, opt)
		if err != nil {
			return nil, err
		}

		allRecords = append(allRecords, records...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allRecords, nil
}

func listLoadBalancers(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(do.DOCloud)
	var resourceTrackers []*resources.Resource

	clusterTag := "KubernetesCluster-Master:" + strings.Replace(clusterName, ".", "-", -1)

	lbs, err := c.GetAllLoadBalancers()
	if err != nil {
		return nil, fmt.Errorf("failed to list lbs: %v", err)
	}

	for _, lb := range lbs {
		if strings.Contains(lb.Tag, clusterTag) {
			resourceTracker := &resources.Resource{
				Name:    lb.Name,
				ID:      lb.ID,
				Type:    resourceTypeLoadBalancer,
				Deleter: deleteLoadBalancer,
				Obj:     lb,
			}

			var blocks []string
			for _, dropletID := range lb.DropletIDs {
				blocks = append(blocks, "droplet:"+strconv.Itoa(dropletID))
			}

			resourceTracker.Blocks = blocks
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}

func deleteDroplet(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(do.DOCloud)
	dropletID, err := strconv.Atoi(t.ID)
	if err != nil {
		return fmt.Errorf("failed to convert droplet ID to int: %s", err)
	}

	_, err = c.DropletsService().Delete(context.TODO(), dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %d, err: %s", dropletID, err)
	}

	return nil
}

func deleteVPC(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(do.DOCloud)
	_, err := c.VPCsService().Delete(context.TODO(), t.ID)
	if err != nil {
		return fmt.Errorf("failed to delete VPC %s (ID %s): %s", t.Name, t.ID, err)
	}

	return nil
}

func deleteVolume(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(do.DOCloud)
	volume := t.Obj.(godo.Volume)
	for _, dropletID := range volume.DropletIDs {
		action, resp, err := c.VolumeActionService().DetachByDropletID(context.TODO(), volume.ID, dropletID)
		if err != nil {
			if resp != nil && resp.StatusCode == http.StatusNotFound {
				// Volume is already detached, nothing to do.
				continue
			}
			return fmt.Errorf("failed to detach volume %s: %s", volume.ID, err)
		}

		if err := waitForDetach(c, action); err != nil {
			return fmt.Errorf("error while waiting for volume %s to detach: %s", volume.ID, err)
		}
	}

	_, err := c.VolumeService().DeleteVolume(context.TODO(), t.ID)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %s, err: %s", t.ID, err)
	}

	return nil
}

func deleteRecord(cloud fi.Cloud, domain string, t *resources.Resource) error {
	c := cloud.(do.DOCloud)
	record := t.Obj.(godo.DomainRecord)

	_, err := c.DomainService().DeleteRecord(context.TODO(), domain, record.ID)
	if err != nil {
		return fmt.Errorf("failed to delete record for domain %s: %d", domain, record.ID)
	}

	return nil
}

func deleteLoadBalancer(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(do.DOCloud)
	lb := t.Obj.(godo.LoadBalancer)
	_, err := c.LoadBalancersService().Delete(context.TODO(), lb.ID)
	if err != nil {
		return fmt.Errorf("failed to delete load balancer with name %s %v", lb.Name, err)
	}

	return nil
}

func waitForDetach(cloud do.DOCloud, action *godo.Action) error {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return errors.New("timed out waiting for volume to detach")
		case <-ticker.C:
			updatedAction, _, err := cloud.ActionsService().Get(context.TODO(), action.ID)
			if err != nil {
				return err
			}

			if updatedAction.Status == godo.ActionCompleted {
				return nil
			}
		}
	}
}

func dumpDroplet(op *resources.DumpOperation, r *resources.Resource) error {
	data := make(map[string]interface{})
	data["id"] = r.ID
	data["type"] = godo.DropletResourceType
	data["raw"] = r.Obj
	op.Dump.Resources = append(op.Dump.Resources, data)

	droplet := r.Obj.(godo.Droplet)
	i := &resources.Instance{
		Name: r.ID,
	}
	if ip, err := droplet.PublicIPv4(); err == nil {
		i.PublicAddresses = append(i.PublicAddresses, ip)
	}
	if ip, err := droplet.PublicIPv6(); err == nil {
		i.PublicAddresses = append(i.PublicAddresses, ip)
	}
	if img := droplet.Image; img != nil {
		switch img.Distribution {
		case "Ubuntu":
			i.SSHUser = "root"
		default:
			klog.Warningf("unrecognized droplet image distribution: %v", img.Distribution)
		}
	}
	for _, tag := range droplet.Tags {
		if strings.HasPrefix(tag, "KubernetesCluster-Master") {
			i.Roles = []string{string(kops.InstanceGroupRoleMaster)}
			break
		}
	}
	if len(i.Roles) == 0 {
		i.Roles = []string{string(kops.InstanceGroupRoleNode)}
	}

	op.Dump.Instances = append(op.Dump.Instances, i)

	return nil
}

func listVPCs(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(do.DOCloud)
	var resourceTrackers []*resources.Resource

	clusterName = do.SafeClusterName(clusterName)
	vpcName := "vpc-" + clusterName

	vpcs, err := c.GetAllVPCs()
	if err != nil {
		return nil, fmt.Errorf("failed to list vpcs: %v", err)
	}

	for _, vpc := range vpcs {
		if vpc.Name == vpcName {
			resourceTracker := &resources.Resource{
				Name:    vpc.Name,
				ID:      vpc.ID,
				Type:    resourceTypeVPC,
				Deleter: deleteVPC,
				Obj:     vpc,
			}

			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}

	return resourceTrackers, nil
}

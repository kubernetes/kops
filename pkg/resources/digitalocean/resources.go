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
	"strconv"
	"strings"
	"time"

	"github.com/digitalocean/godo"

	"k8s.io/klog"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	resourceTypeDroplet      = "droplet"
	resourceTypeVolume       = "volume"
	resourceTypeDNSRecord    = "dns-record"
	resourceTypeLoadBalancer = "loadbalancer"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud *Cloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listVolumes,
		listDroplets,
		listDNS,
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

func listDroplets(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(*Cloud)
	var resourceTrackers []*resources.Resource

	clusterTag := "KubernetesCluster:" + strings.Replace(clusterName, ".", "-", -1)

	droplets, err := getAllDropletsByTag(c, clusterTag)
	if err != nil {
		return nil, fmt.Errorf("failed to list droplets: %v", err)
	}

	for _, droplet := range droplets {
		resourceTracker := &resources.Resource{
			Name:    droplet.Name,
			ID:      strconv.Itoa(droplet.ID),
			Type:    resourceTypeDroplet,
			Deleter: deleteDroplet,
			Obj:     droplet,
		}

		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func getAllDropletsByTag(cloud *Cloud, tag string) ([]godo.Droplet, error) {
	allDroplets := []godo.Droplet{}

	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := cloud.Droplets().ListByTag(context.TODO(), tag, opt)
		if err != nil {
			return nil, err
		}

		allDroplets = append(allDroplets, droplets...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allDroplets, nil
}

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(*Cloud)
	var resourceTrackers []*resources.Resource

	volumeMatch := strings.Replace(clusterName, ".", "-", -1)

	volumes, err := getAllVolumesByRegion(c, c.Region())
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

func getAllVolumesByRegion(cloud *Cloud, region string) ([]godo.Volume, error) {
	allVolumes := []godo.Volume{}

	opt := &godo.ListOptions{}
	for {
		volumes, resp, err := cloud.Volumes().ListVolumes(context.TODO(), &godo.ListVolumeParams{
			Region:      region,
			ListOptions: opt,
		})

		if err != nil {
			return nil, err
		}

		allVolumes = append(allVolumes, volumes...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allVolumes, nil
}

func listDNS(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(*Cloud)

	domains, _, err := c.Client.Domains.List(context.TODO(), &godo.ListOptions{})
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

func getAllRecordsByDomain(cloud *Cloud, domain string) ([]godo.DomainRecord, error) {
	allRecords := []godo.DomainRecord{}

	opt := &godo.ListOptions{}
	for {
		records, resp, err := cloud.Client.Domains.Records(context.TODO(), domain, opt)
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
	c := cloud.(*Cloud)
	var resourceTrackers []*resources.Resource

	clusterTag := "KubernetesCluster-Master:" + strings.Replace(clusterName, ".", "-", -1)

	lbs, err := getAllLoadBalancers(c)
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

func getAllLoadBalancers(cloud *Cloud) ([]godo.LoadBalancer, error) {
	allLoadBalancers := []godo.LoadBalancer{}

	opt := &godo.ListOptions{}
	for {
		lbs, resp, err := cloud.LoadBalancers().List(context.TODO(), opt)
		if err != nil {
			return nil, err
		}

		allLoadBalancers = append(allLoadBalancers, lbs...)

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		opt.Page = page + 1
	}

	return allLoadBalancers, nil
}

func deleteDroplet(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(*Cloud)

	dropletID, err := strconv.Atoi(t.ID)
	if err != nil {
		return fmt.Errorf("failed to convert droplet ID to int: %s", err)
	}

	_, err = c.Droplets().Delete(context.TODO(), dropletID)
	if err != nil {
		return fmt.Errorf("failed to delete droplet: %d, err: %s", dropletID, err)
	}

	return nil
}

func deleteVolume(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(*Cloud)

	volume := t.Obj.(godo.Volume)
	for _, dropletID := range volume.DropletIDs {
		action, _, err := c.VolumeActions().DetachByDropletID(context.TODO(), volume.ID, dropletID)
		if err != nil {
			return fmt.Errorf("failed to detach volume: %s, err: %s", volume.ID, err)
		}
		if err := waitForDetach(c, action); err != nil {
			return fmt.Errorf("error while waiting for volume %s to detach: %s", volume.ID, err)
		}
	}

	_, err := c.Volumes().DeleteVolume(context.TODO(), t.ID)
	if err != nil {
		return fmt.Errorf("failed to delete volume: %s, err: %s", t.ID, err)
	}

	return nil
}

func deleteRecord(cloud fi.Cloud, domain string, t *resources.Resource) error {
	c := cloud.(*Cloud)
	record := t.Obj.(godo.DomainRecord)

	_, err := c.Client.Domains.DeleteRecord(context.TODO(), domain, record.ID)
	if err != nil {
		return fmt.Errorf("failed to delete record for domain %s: %d", domain, record.ID)
	}

	return nil
}

func deleteLoadBalancer(cloud fi.Cloud, t *resources.Resource) error {
	c := cloud.(*Cloud)
	lb := t.Obj.(godo.LoadBalancer)
	_, err := c.Client.LoadBalancers.Delete(context.TODO(), lb.ID)

	if err != nil {
		return fmt.Errorf("failed to delete load balancer with name %s %v", lb.Name, err)
	}

	return nil
}

func waitForDetach(cloud *Cloud, action *godo.Action) error {
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeout:
			return errors.New("timed out waiting for volume to detach")
		case <-ticker.C:
			updatedAction, _, err := cloud.Client.Actions.Get(context.TODO(), action.ID)
			if err != nil {
				return err
			}

			if updatedAction.Status == godo.ActionCompleted {
				return nil
			}
		}
	}
}

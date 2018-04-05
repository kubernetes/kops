/*
Copyright 2016 The Kubernetes Authors.

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

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

const (
	resourceTypeDroplet = "droplet"
	resourceTypeVolume  = "volume"
)

type listFn func(fi.Cloud, string) ([]*resources.Resource, error)

func ListResources(cloud *Cloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listVolumes,
		listDroplets,
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

	droplets, _, err := c.Droplets().ListByTag(context.TODO(), clusterTag, &godo.ListOptions{})
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

func listVolumes(cloud fi.Cloud, clusterName string) ([]*resources.Resource, error) {
	c := cloud.(*Cloud)
	var resourceTrackers []*resources.Resource

	volumeMatch := strings.Replace(clusterName, ".", "-", -1)

	volumes, _, err := c.Volumes().ListVolumes(context.TODO(), &godo.ListVolumeParams{
		Region: c.Region,
	})

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

func waitForDetach(cloud *Cloud, action *godo.Action) error {
	timeout := time.After(10 * time.Second)
	tick := time.Tick(500 * time.Millisecond)
	for {
		select {
		case <-timeout:
			return errors.New("timed out waiting for volume to detach")
		case <-tick:
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

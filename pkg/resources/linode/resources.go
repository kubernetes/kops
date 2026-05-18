/*
Copyright 2026 The Kubernetes Authors.

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

package linode

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/linode/linodego"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	cloudlinode "k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const (
	resourceTypeInstance     = "instance"
	resourceTypeSSHKey       = "ssh-key"
	resourceTypeVolume       = "volume"
	resourceTypeNodeBalancer = "nodebalancer"
)

type listFn func(fi.Cloud, resources.ClusterInfo) ([]*resources.Resource, error)

func parseTrackerIntID(tracker *resources.Resource) (int, error) {
	id, err := strconv.Atoi(tracker.ID)
	if err != nil {
		return 0, fmt.Errorf("error parsing Linode (Akamai) %s ID %q: %w", tracker.Type, tracker.ID, err)
	}
	return id, nil
}

// ListResources collects Linode (Akamai) cloud resources owned by the cluster.
func ListResources(cloud cloudlinode.LinodeCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listInstances,
		listVolumes,
		listSSHKeys,
		listNodeBalancers,
	}

	for _, fn := range listFunctions {
		trackers, err := fn(cloud, clusterInfo)
		if err != nil {
			return nil, err
		}

		for _, tracker := range trackers {
			resourceTrackers[tracker.Type+":"+tracker.ID] = tracker
		}
	}

	return resourceTrackers, nil
}

// listInstances lists Linode (Akamai) instances that are tagged as belonging to the cluster.
func listInstances(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	instances, err := c.Client().ListInstances(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) instances: %w", err)
	}

	clusterTag := cloudlinode.BuildLinodeTag(kops.LabelClusterName, clusterInfo.Name)
	var resourceTrackers []*resources.Resource
	for _, instance := range instances {
		if !slices.Contains(instance.Tags, clusterTag) {
			continue
		}

		resourceTrackers = append(resourceTrackers, &resources.Resource{
			Name:    instance.Label,
			ID:      strconv.Itoa(instance.ID),
			Type:    resourceTypeInstance,
			Deleter: deleteInstance,
			Obj:     instance,
		})
	}

	return resourceTrackers, nil
}

// listSSHKeys lists Linode (Akamai) SSH keys that are tagged as belonging to the cluster.
func listSSHKeys(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	keys, err := c.Client().ListSSHKeys(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) SSH keys: %w", err)
	}

	// Precompute match criteria once for the entire key list.
	var matchFn func(label string) bool
	if clusterInfo.LinodeSSHKeyName != "" {
		normalizedName := cloudlinode.NormalizeLinodeSSHKeyLabel(clusterInfo.LinodeSSHKeyName)
		matchFn = func(label string) bool {
			return label == clusterInfo.LinodeSSHKeyName || label == normalizedName
		}
	} else {
		rawPrefix := "kubernetes." + clusterInfo.Name + "-"
		normalizedPrefix := cloudlinode.NormalizeLinodeSSHKeyLabel(rawPrefix)
		matchFn = func(label string) bool {
			return strings.HasPrefix(label, rawPrefix) || strings.HasPrefix(label, normalizedPrefix)
		}
	}

	var resourceTrackers []*resources.Resource
	for _, key := range keys {
		if !matchFn(key.Label) {
			continue
		}

		resourceTrackers = append(resourceTrackers, &resources.Resource{
			Name:    key.Label,
			ID:      strconv.Itoa(key.ID),
			Type:    resourceTypeSSHKey,
			Deleter: deleteSSHKey,
			Obj:     key,
		})
	}

	return resourceTrackers, nil
}

// listVolumes lists Linode (Akamai) volumes that are tagged as belonging to the cluster.
func listVolumes(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	volumes, err := c.Client().ListVolumes(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) volumes: %w", err)
	}

	clusterTag := cloudlinode.BuildLinodeTag(kops.LabelClusterName, clusterInfo.Name)
	var resourceTrackers []*resources.Resource
	for _, volume := range volumes {
		if !slices.Contains(volume.Tags, clusterTag) {
			continue
		}

		resourceTrackers = append(resourceTrackers, &resources.Resource{
			Name:    volume.Label,
			ID:      strconv.Itoa(volume.ID),
			Type:    resourceTypeVolume,
			Deleter: deleteVolume,
			Obj:     volume,
		})
	}

	return resourceTrackers, nil
}

// listNodeBalancers lists Linode (Akamai) NodeBalancers that are tagged as belonging to the cluster.
func listNodeBalancers(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	nodeBalancers, err := c.Client().ListNodeBalancers(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) node balancers: %w", err)
	}

	clusterTag := cloudlinode.BuildLinodeTag(kops.LabelClusterName, clusterInfo.Name)
	var resourceTrackers []*resources.Resource
	for _, nodeBalancer := range nodeBalancers {
		if !slices.Contains(nodeBalancer.Tags, clusterTag) {
			continue
		}

		resourceTrackers = append(resourceTrackers, &resources.Resource{
			Name:    fi.ValueOf(nodeBalancer.Label),
			ID:      strconv.Itoa(nodeBalancer.ID),
			Type:    resourceTypeNodeBalancer,
			Deleter: deleteNodeBalancer,
			Obj:     nodeBalancer,
		})
	}

	return resourceTrackers, nil
}

// deleteInstance deletes a Linode (Akamai) instance.
func deleteInstance(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	instance := &cloudinstances.CloudInstance{ID: tracker.ID}
	if err := c.DeleteInstance(instance); err != nil {
		return fmt.Errorf("error deleting Linode (Akamai) instance %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteSSHKey deletes a Linode (Akamai) SSH key.
func deleteSSHKey(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	keyID, err := parseTrackerIntID(tracker)
	if err != nil {
		return err
	}

	if err := c.Client().DeleteSSHKey(context.Background(), keyID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Linode (Akamai) SSH key %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteVolume deletes a Linode (Akamai) volume.
func deleteVolume(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	volumeID, err := parseTrackerIntID(tracker)
	if err != nil {
		return err
	}

	if err := c.Client().DeleteVolume(context.Background(), volumeID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Linode (Akamai) volume %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteNodeBalancer deletes a Linode (Akamai) NodeBalancer.
func deleteNodeBalancer(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	nodeBalancerID, err := parseTrackerIntID(tracker)
	if err != nil {
		return err
	}

	if err := c.Client().DeleteNodeBalancer(context.Background(), nodeBalancerID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Linode (Akamai) node balancer %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

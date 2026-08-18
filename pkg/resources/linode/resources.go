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
	"sort"
	"strconv"
	"strings"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	cloudlinode "k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

type listFn func(fi.Cloud, resources.ClusterInfo) ([]*resources.Resource, error)

const (
	resourceTypeVPC      = "vpc"
	resourceTypeSubnet   = "subnet"
	resourceTypeSSHKey   = "ssh-key"
	resourceTypeInstance = "instance"
	resourceTypeVolume   = "volume"
)

// parseTrackerIntID parses the tracker's string ID into an integer, which is used for Akamai (Linode) resource IDs.
func parseTrackerIntID(tracker *resources.Resource) (int, error) {
	id, err := strconv.Atoi(tracker.ID)
	if err != nil {
		return 0, fmt.Errorf("error parsing Akamai (Linode) %s ID %q: %w", tracker.Type, tracker.ID, err)
	}
	return id, nil
}

// ListResources collects Akamai (Linode) cloud resources owned by the cluster.
func ListResources(cloud cloudlinode.LinodeCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	listFunctions := []listFn{
		listVPCs,
		listSubnets,
		listInstances,
		listVolumes,
		listSSHKeys,
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

// listInstances lists Akamai (Linode) instances owned by the cluster.
func listInstances(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	listOptions, err := cloudlinode.ListOptionsForTags(fmt.Sprintf("%s:%s", cloudlinode.TagKubernetesClusterName, clusterInfo.Name))
	if err != nil {
		return nil, err
	}

	instances, err := c.Client().ListInstances(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) instances: %w", err)
	}

	var resourceTrackers []*resources.Resource
	for _, instance := range instances {
		blocks, err := instanceSubnetBlocks(c, instance)
		if err != nil {
			return nil, err
		}

		resourceTrackers = append(resourceTrackers, &resources.Resource{
			Name:    instance.Label,
			ID:      strconv.Itoa(instance.ID),
			Type:    resourceTypeInstance,
			Deleter: deleteInstance,
			Blocks:  blocks,
			Obj:     instance,
		})
	}

	return resourceTrackers, nil
}

// listVolumes lists Akamai (Linode) block storage volumes owned by the cluster.
func listVolumes(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	clusterTag := cloudlinode.NormalizeLinodeLabel(clusterInfo.Name)
	listOptions, err := cloudlinode.ListOptionsForTags(fmt.Sprintf("%s:%s", cloudlinode.TagKubernetesClusterName, clusterTag))
	if err != nil {
		return nil, err
	}

	volumes, err := c.Client().ListVolumes(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) volumes: %w", err)
	}

	resourceTrackers := make([]*resources.Resource, 0, len(volumes))
	for _, volume := range volumes {
		resourceTracker := &resources.Resource{
			Name:    volume.Label,
			ID:      strconv.Itoa(volume.ID),
			Type:    resourceTypeVolume,
			Deleter: deleteVolume,
			Obj:     volume,
		}
		if volume.LinodeID != nil {
			resourceTracker.Blocked = []string{resourceTypeInstance + ":" + strconv.Itoa(*volume.LinodeID)}
		}
		resourceTrackers = append(resourceTrackers, resourceTracker)
	}

	return resourceTrackers, nil
}

func instanceSubnetBlocks(cloud cloudlinode.LinodeCloud, instance linodego.Instance) ([]string, error) {
	interfaces, err := cloud.Client().ListInterfaces(context.Background(), instance.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) interfaces for instance %s(%d): %w", instance.Label, instance.ID, err)
	}

	blockSet := make(map[string]struct{})
	for _, iface := range interfaces {
		if iface.VPC == nil {
			continue
		}
		blockSet[resourceTypeSubnet+":"+strconv.Itoa(iface.VPC.SubnetID)] = struct{}{}
	}

	blocks := make([]string, 0, len(blockSet))
	for block := range blockSet {
		blocks = append(blocks, block)
	}
	sort.Strings(blocks)

	return blocks, nil
}

// findClusterVPCs finds Akamai (Linode) VPCs with the cluster's deterministic VPC label.
func findClusterVPCs(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]linodego.VPC, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	vpcLabel := cloudlinode.NormalizeLinodeLabel(clusterInfo.Name)
	listOptions, err := cloudlinode.ListOptionsForLabel(vpcLabel)
	if err != nil {
		return nil, err
	}

	vpcs, err := c.Client().ListVPCs(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) VPCs: %w", err)
	}

	region := c.Region()

	var clusterVPCs []linodego.VPC
	for _, vpc := range vpcs {
		if vpc.Label != vpcLabel {
			continue
		}
		if region != "" && vpc.Region != region {
			continue
		}

		clusterVPCs = append(clusterVPCs, vpc)
	}

	return clusterVPCs, nil
}

// listSSHKeys lists Akamai (Linode) SSH keys that were generated for the cluster.
func listSSHKeys(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	keys, err := c.Client().ListSSHKeys(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Akamai (Linode) SSH keys: %w", err)
	}

	keyLabelPrefix := cloudlinode.NormalizeLinodeLabel("kubernetes."+clusterInfo.Name) + "-"
	var resourceTrackers []*resources.Resource
	for _, key := range keys {
		if !strings.HasPrefix(key.Label, keyLabelPrefix) {
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

// listVPCs lists Akamai (Linode) VPC resources owned by the cluster.
func listVPCs(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	vpcs, err := findClusterVPCs(cloud, clusterInfo)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, vpc := range vpcs {
		resourceTrackers = append(resourceTrackers, &resources.Resource{
			Name:    vpc.Label,
			ID:      strconv.Itoa(vpc.ID),
			Type:    resourceTypeVPC,
			Deleter: deleteVPC,
			Obj:     vpc,
		})
	}

	return resourceTrackers, nil
}

// listSubnets lists Akamai (Linode) VPC subnets attached to the cluster's managed VPC.
func listSubnets(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	vpcs, err := findClusterVPCs(cloud, clusterInfo)
	if err != nil {
		return nil, err
	}

	var resourceTrackers []*resources.Resource
	for _, vpc := range vpcs {
		subnets, err := c.Client().ListVPCSubnets(context.Background(), vpc.ID, nil)
		if err != nil {
			return nil, fmt.Errorf("error listing Akamai (Linode) VPC subnets for VPC %s(%d): %w", vpc.Label, vpc.ID, err)
		}

		for _, subnet := range subnets {
			resourceTrackers = append(resourceTrackers, &resources.Resource{
				Name: subnet.Label,
				ID:   strconv.Itoa(subnet.ID),
				Type: resourceTypeSubnet,
				Deleter: func(cloud fi.Cloud, tracker *resources.Resource) error {
					return deleteSubnet(vpc.ID, cloud, tracker)
				},
				Blocks: []string{resourceTypeVPC + ":" + strconv.Itoa(vpc.ID)},
				Obj:    subnet,
			})
		}
	}

	return resourceTrackers, nil
}

// deleteSSHKey deletes an Akamai (Linode) SSH key.
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
		return fmt.Errorf("error deleting Akamai (Linode) SSH key %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteVPC deletes an Akamai (Linode) VPC.
func deleteVPC(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	vpcID, err := parseTrackerIntID(tracker)
	if err != nil {
		return err
	}

	if err := c.Client().DeleteVPC(context.Background(), vpcID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Akamai (Linode) VPC %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteSubnet deletes an Akamai (Linode) VPC subnet.
func deleteSubnet(vpcID int, cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	subnetID, err := parseTrackerIntID(tracker)
	if err != nil {
		return err
	}

	if err := c.Client().DeleteVPCSubnet(context.Background(), vpcID, subnetID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Akamai (Linode) subnet %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteInstance deletes an Akamai (Linode) instance.
func deleteInstance(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	instanceID, err := parseTrackerIntID(tracker)
	if err != nil {
		return err
	}

	if err := c.Client().DeleteInstance(context.Background(), instanceID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Akamai (Linode) instance %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteVolume deletes an Akamai (Linode) block storage volume.
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
		return fmt.Errorf("error deleting Akamai (Linode) volume %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

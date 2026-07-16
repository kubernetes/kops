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
	"strconv"
	"strings"

	"github.com/linode/linodego/v2"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	cloudlinode "k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

type listFn func(fi.Cloud, resources.ClusterInfo) ([]*resources.Resource, error)

const (
	resourceTypeVPC    = "vpc"
	resourceTypeSubnet = "subnet"
	resourceTypeSSHKey = "ssh-key"
)

// parseTrackerIntID parses the tracker's string ID into an integer, which is used for Linode (Akamai) resource IDs.
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
		listVPCs,
		listSubnets,
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

// findClusterVPCs finds Linode (Akamai) VPCs with the cluster's deterministic VPC label.
func findClusterVPCs(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]linodego.VPC, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	vpcLabel := cloudlinode.NormalizeLinodeLabel(clusterInfo.Name)
	listOptions, err := cloudlinode.ListOptionsForLabel(vpcLabel)
	if err != nil {
		return nil, err
	}

	vpcs, err := c.Client().ListVPCs(context.Background(), listOptions)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) VPCs: %w", err)
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

// listSSHKeys lists Linode (Akamai) SSH keys that were generated for the cluster.
func listSSHKeys(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	keys, err := c.Client().ListSSHKeys(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) SSH keys: %w", err)
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

// listVPCs lists Linode (Akamai) VPC resources owned by the cluster.
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

// listSubnets lists Linode (Akamai) VPC subnets attached to the cluster's managed VPC.
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
			return nil, fmt.Errorf("error listing Linode (Akamai) VPC subnets for VPC %s(%d): %w", vpc.Label, vpc.ID, err)
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

// deleteVPC deletes a Linode (Akamai) VPC.
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
		return fmt.Errorf("error deleting Linode (Akamai) VPC %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

// deleteSubnet deletes a Linode (Akamai) VPC subnet.
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
		return fmt.Errorf("error deleting Linode (Akamai) subnet %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

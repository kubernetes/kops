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

	"github.com/linode/linodego"
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	cloudlinode "k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

const resourceTypeVPC = "vpc"

// ListResources collects Linode (Akamai) cloud resources owned by the cluster.
func ListResources(cloud cloudlinode.LinodeCloud, clusterInfo resources.ClusterInfo) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	trackers, err := listVPCs(cloud, clusterInfo)
	if err != nil {
		return nil, err
	}
	for _, tracker := range trackers {
		resourceTrackers[tracker.Type+":"+tracker.ID] = tracker
	}

	return resourceTrackers, nil
}

// listVPCs lists Linode (Akamai) VPCs with the cluster's deterministic VPC label.
func listVPCs(cloud fi.Cloud, clusterInfo resources.ClusterInfo) ([]*resources.Resource, error) {
	c := cloud.(cloudlinode.LinodeCloud)
	vpcs, err := c.Client().ListVPCs(context.Background(), nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) VPCs: %w", err)
	}

	vpcLabel := cloudlinode.NormalizeLinodeVPCLabel(clusterInfo.Name)
	region := c.Region()
	var resourceTrackers []*resources.Resource
	for _, vpc := range vpcs {
		if vpc.Label != vpcLabel {
			continue
		}
		if region != "" && vpc.Region != region {
			continue
		}

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

// deleteVPC deletes a Linode (Akamai) VPC.
func deleteVPC(cloud fi.Cloud, tracker *resources.Resource) error {
	c := cloud.(cloudlinode.LinodeCloud)
	vpcID, err := strconv.Atoi(tracker.ID)
	if err != nil {
		return fmt.Errorf("error parsing Linode (Akamai) %s ID %q: %w", tracker.Type, tracker.ID, err)
	}

	if err := c.Client().DeleteVPC(context.Background(), vpcID); err != nil {
		if linodego.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("error deleting Linode (Akamai) VPC %s(%s): %w", tracker.Name, tracker.ID, err)
	}

	return nil
}

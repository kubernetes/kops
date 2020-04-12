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

package commands

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/validation"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

// UpdateCluster writes the updated cluster to the state store, after performing validation
func UpdateCluster(ctx context.Context, clientset simple.Clientset, cluster *kops.Cluster, instanceGroups []*kops.InstanceGroup) error {
	err := cloudup.PerformAssignments(cluster)
	if err != nil {
		return fmt.Errorf("error populating configuration: %v", err)
	}

	assetBuilder := assets.NewAssetBuilder(cluster, "")
	fullCluster, err := cloudup.PopulateClusterSpec(clientset, cluster, assetBuilder)
	if err != nil {
		return err
	}

	err = validation.DeepValidate(fullCluster, instanceGroups, true)
	if err != nil {
		return err
	}

	// Retrieve the current status of the cluster.  This will eventually be part of the cluster object.
	statusDiscovery := &CloudDiscoveryStatusStore{}
	status, err := statusDiscovery.FindClusterStatus(cluster)
	if err != nil {
		return err
	}

	// Note we perform as much validation as we can, before writing a bad config
	_, err = clientset.UpdateCluster(ctx, cluster, status)
	if err != nil {
		return err
	}

	return nil
}

// ReadAllInstanceGroups reads all the instance groups for the cluster
func ReadAllInstanceGroups(ctx context.Context, clientset simple.Clientset, cluster *kops.Cluster) ([]*kops.InstanceGroup, error) {
	list, err := clientset.InstanceGroupsFor(cluster).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var instanceGroups []*kops.InstanceGroup
	for i := range list.Items {
		instanceGroups = append(instanceGroups, &list.Items[i])
	}
	return instanceGroups, nil
}

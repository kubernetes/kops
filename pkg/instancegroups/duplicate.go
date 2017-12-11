/*
Copyright 2017 The Kubernetes Authors.

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

package instancegroups

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	api "k8s.io/kops/pkg/apis/kops"
)

// duplicateProvider session data
type duplicateProvider struct {
	GroupUpdate *RollingUpdateInstanceGroup
}

// NewDuplicateRollout creates and returns a drain provider
func NewDuplicateRollout(update *RollingUpdateInstanceGroup) Rollout {
	return &duplicateProvider{GroupUpdate: update}
}

// RollingUpdate performs a duplicate rolling update of an instance group
func (p *duplicateProvider) RollingUpdate(ctx context.Context, list *api.InstanceGroupList) error {
	group := p.GroupUpdate.CloudGroup
	ig := group.InstanceGroup
	name := group.InstanceGroup.Name
	strategy := ig.Spec.Strategy
	update := p.GroupUpdate.Update

	update.Infof("rolling out with strategy: %s, instancegroups: %s, batch: %d", strategy.Name, name, strategy.Batch)

	// @check the instancegroup is not already a duplicate
	if _, found := ig.GetAnnotations()[DuplicateParentAnnotation]; found {
		update.Infof("ignoring instancegroup: %s as we found parent duplicate annotation", ig.Name)
		return nil
	}

	// @step: attempt to duplicate the instance group
	newName := fmt.Sprintf("%s-%d", group.InstanceGroup.Name, time.Now().Unix())
	update.Infof("creating new instancegroup: %s, from parent: %s", newName, name)

	duplicate, err := p.GroupUpdate.DuplicateGroup(newName)
	if err != nil {
		return update.Errorf("unable to duplicate the instancegroup: %s, error: %v", name, err)
	}

	// @step: we need to update from cloud knowlegde, not from the spec as runtime might have changed
	min, max := int32(p.GroupUpdate.CloudGroup.Size()), int32(group.MaxSize)
	duplicate.Spec.MinSize = &min
	duplicate.Spec.MaxSize = &max

	// @step: create the new instancegroup and wait for the cluster to validate
	update.Infof("reconfiguring the cluster with the new instancegroup: %s", newName)
	if err = update.UpdateCluster(ctx); err != nil {
		return update.Errorf("unable to update the cluster, error: %v", err)
	}

	// @step: wait for the instancegroup
	update.Infof("waiting for instancegroup: %s to scale", newName)
	if err := p.GroupUpdate.WaitForGroupSize(ctx, newName, int(min), update.ScaleTimeout); err != nil {
		return update.Errorf("instancegroup: %s has not reached size: %d within: %s", newName, min, update.ScaleTimeout)
	}

	// @step: get an updated list of instancegroups
	newList, err := update.Clientset.InstanceGroupsFor(update.Cluster).List(v1.ListOptions{})
	if err != nil {
		return update.Errorf("unable to get update list of instancegroups, error: %v", err)
	}

	// @step: validate the cluster again post creation of new instancegroup
	update.Infof("validating the cluster post creation of new instancegroup: %s", newName)
	if err = p.GroupUpdate.ValidateClusterWithTimeout(ctx, newList, update.FailOnValidateTimeout); err != nil {
		return update.Errorf("unable to validate cluster after %s", update.FailOnValidateTimeout)
	}

	// @step: if not cloudonly we drain all the nodes and then delete the instance group
	if !update.CloudOnly && strategy.Drain {
		options := &DrainOptions{
			Batch:             strategy.Batch,
			CloudOnly:         update.CloudOnly,
			Delete:            false,
			DrainPods:         strategy.Drain,
			FailOnValidation:  update.FailOnValidate,
			Interval:          strategy.Interval.Duration,
			PostDelay:         strategy.PostDrainDelay.Duration,
			Timeout:           strategy.DrainTimeout.Duration,
			ValidateCluster:   !update.CloudOnly,
			ValidationTimeout: update.FailOnValidateTimeout,
		}
		// @step: drain the instancegroup of the pods
		if err = p.GroupUpdate.DrainGroup(ctx, options, newList); err != nil {
			return update.Errorf("unable to drain parent instancegroup: %s, error: %v", name, err)
		}
		// @step: validate the cluster once more
		if err = p.GroupUpdate.ValidateClusterWithTimeout(ctx, newList, strategy.Interval.Duration); err != nil {
			return update.Errorf("unable to validate cluster post drain after %s, error: %v", update.FailOnValidateTimeout, err)
		}
	}

	// @step: delete the instancegroup and rename the other one
	update.Infof("deleting the parent instancegroup: %s from kops cluster", name)
	if err = update.Cloud.DeleteGroup(group); err != nil {
		return update.Errorf("unable to delete group: %s, error: %v", name, err)
	}

	// @step: delete the instance group from configuration
	update.Infof("deleting the parent instancegroup: %s configuration from kops cluster specification", name)
	if err = update.Clientset.InstanceGroupsFor(update.Cluster).Delete(name, &v1.DeleteOptions{}); err != nil {
		return update.Errorf("unable to delete instancegroup: %s from configuration, error: %v", name, err)
	}

	// @step: remove the child annotation from the new instancegroup
	update.Infof("updating the duplicate instancegroup: %s with source name: %s", newName, name)
	delete(duplicate.ObjectMeta.Annotations, DuplicateParentAnnotation)
	if _, err = update.Clientset.InstanceGroupsFor(update.Cluster).Create(duplicate); err != nil {
		return update.Errorf("unable to update name of new instancegroup: %s in configuration, error: %v", name, err)
	}

	// @step: we need to revert the min size of the group from the desired size back to the min
	min = int32(group.MinSize)
	duplicate.Spec.MinSize = &min
	if _, err = update.Clientset.InstanceGroupsFor(update.Cluster).Update(duplicate); err != nil {
		return update.Errorf("unable to delete instancegroup: %s from configuration, error: %v", name, err)
	}
	if err = update.UpdateCluster(ctx); err != nil {
		return update.Errorf("unable to update the cluster, error: %v", err)
	}

	return nil
}

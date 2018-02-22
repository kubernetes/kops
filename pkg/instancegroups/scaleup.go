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

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// scaleProvider session data
type scaleProvider struct {
	GroupUpdate *RollingUpdateInstanceGroup
}

// NewScaleUpRollout creates and returns a scale provider
func NewScaleUpRollout(update *RollingUpdateInstanceGroup) Rollout {
	return &scaleProvider{GroupUpdate: update}
}

// RollingUpdate performs a duplicate rolling update of an instance group
func (p *scaleProvider) RollingUpdate(ctx context.Context, list *api.InstanceGroupList) error {
	group := p.GroupUpdate.CloudGroup.InstanceGroup
	fullname := p.GroupUpdate.Name
	name := group.Name
	strategy := group.Spec.Strategy
	update := p.GroupUpdate.Update

	// @step: increase the size of the instancegroup by x percent
	var newMinSize, newMaxSize, oldMinSize, oldMaxSize int32
	total := int32(p.GroupUpdate.CloudGroup.Size())
	oldMinSize = int32(p.GroupUpdate.CloudGroup.MinSize)
	oldMaxSize = int32(p.GroupUpdate.CloudGroup.MaxSize)
	newMinSize = int32(p.GroupUpdate.CloudGroup.Size() * 2)
	newMaxSize = int32(p.GroupUpdate.CloudGroup.MaxSize)
	if newMinSize > newMaxSize {
		newMaxSize = newMinSize
	}
	update.Infof("strategy: %s, adjusting size setting on instancegroup: %s, min/max (%d/%d)",
		strategy.Name, name, newMinSize, newMaxSize)

	update.Infof("updating the configuration with adjusted instancegroup settings for: %s", name)

	// @step: grab the configuration from source
	config, err := update.Clientset.InstanceGroupsFor(update.Cluster).Get(name, v1.GetOptions{})
	if err != nil {
		return update.Errorf("unable to retrieve configuration for instancegroup: %s, error: %v", name, err)
	}
	config.Spec.MinSize = &newMinSize
	config.Spec.MaxSize = &newMaxSize

	// @step: attempt to adjust the size of the instancegroup
	if _, err := update.Clientset.InstanceGroupsFor(update.Cluster).Update(config); err != nil {
		return update.Errorf("unable to update instancegroup: %s configuration, error: %v", name, err)
	}

	// @step: we need to update the termination policy to ensure if there is an error the new nodes are the
	// are the ones which get kicked out first
	update.Infof("updating termination policy on the instancegroup: %s to newest first", name)
	if err := update.Cloud.SetTerminationPolicy(update.Cluster, fullname, cloudinstances.TerminateNewestPolicy); err != nil {
		return update.Errorf("unable to update termination policy on instancegroup: %s, error: %v", name, err)
	}

	// @step: wrap the functionality for a rollback
	result := func() error {
		// @step: attempt to update the cluster
		if err := update.UpdateCluster(ctx); err != nil {
			return fmt.Errorf("unable to update the cluster, error: %v", err)
		}
		update.Infof("successfully update the configuration for instancegroup: %s", name)

		// @step: we need to wait for a node interval and then attempt to validate cluster
		update.Infof("waiting to %s while the instancegroup: %s to rescale", update.ScaleTimeout, name)
		if err := p.GroupUpdate.WaitForGroupSize(ctx, name, int(newMinSize), update.ScaleTimeout); err != nil {
			return err
		}
		update.Infof("instancegroup: %s has successfully rescaled", name)

		// @TODO look at a better way of doing, we should probably just get the instancegroup count and wait for
		// the same count of nodes to be registered
		// @step: we need to wait for a node interval for the instances to register otherwise a cluster validation will
		// pass successfully as the nodes haven't yet some into cluster
		update.Infof("waiting to %s for new nodes to register", strategy.Interval.Duration)
		if err := p.GroupUpdate.WaitFor(ctx, strategy.Interval.Duration); err != nil {
			return err
		}

		// @step: attempt to validate the cluster
		update.Infof("attempting to validate the cluster: %s", update.ClusterName)
		if err := p.GroupUpdate.ValidateClusterWithTimeout(ctx, list, update.FailOnValidateTimeout); err != nil {
			return err
		}

		// @check if we need to drain the nodes
		if !update.CloudOnly && strategy.Drain {
			update.Infof("attempting to drain the nodes from instancegroup: %s, batch: %d", name, strategy.Batch)
			options := &DrainOptions{
				Batch:             strategy.Batch,
				CloudOnly:         update.CloudOnly,
				Delete:            false,
				DrainPods:         strategy.Drain,
				Interval:          strategy.Interval.Duration,
				PostDelay:         strategy.PostDrainDelay.Duration,
				Timeout:           strategy.DrainTimeout.Duration,
				ValidateCluster:   !update.CloudOnly,
				ValidationTimeout: update.FailOnValidateTimeout,
			}
			if err := p.GroupUpdate.DrainGroup(ctx, options, list); err != nil {
				return fmt.Errorf("unable to drain the old nodes on instancegroup: %s, error: %v", name, err)
			}
		}

		// @step: attempt to validate the cluster
		update.Infof("attempting to validate post drain on instancegroup: %s", name)
		if err := p.GroupUpdate.ValidateClusterWithTimeout(ctx, list, update.FailOnValidateTimeout); err != nil {
			return err
		}

		// @step: change the termination policy back to oldest first, enuring the old nodes are kicked out first
		update.Infof("changing the termination policy of intstancegroup: %s to oldest first", name)
		if err := update.Cloud.SetTerminationPolicy(update.Cluster, fullname, cloudinstances.TerminateOldestPolicy); err != nil {
			return fmt.Errorf("unable to revert termination policy on instancegroup: %s, error: %s", name, err)
		}

		// @step: everything is looking good we need to shrink the instancegroup back to the original size by
		// fiddling with the max size
		{
			update.Infof("adjusting the size on instancegroup: %s back to original min/max (%d/%d)", name, oldMinSize, oldMaxSize)
			config.Spec.MinSize = &oldMinSize
			config.Spec.MaxSize = &total

			if _, err := update.Clientset.InstanceGroupsFor(update.Cluster).Update(config); err != nil {
				return fmt.Errorf("unable to update instancegroup: %s configuration, error: %v", name, err)
			}
			if err := update.UpdateCluster(ctx); err != nil {
				return fmt.Errorf("unable to update the cluster, error: %v", err)
			}
			update.Infof("successfully reverted the size of instancegroup: %s", name)
		}

		// @step: we need to revert the max size back to the original and then update the configuration
		// Note: this is all due to us not having control of the DesiredSize
		{
			config.Spec.MinSize = &oldMinSize
			config.Spec.MaxSize = &oldMaxSize

			if _, err := update.Clientset.InstanceGroupsFor(update.Cluster).Update(config); err != nil {
				return fmt.Errorf("unable to update instancegroup: %s configuration, error: %v", name, err)
			}
			if err := update.UpdateCluster(ctx); err != nil {
				return fmt.Errorf("unable to update the cluster, error: %v", err)
			}
		}

		{
			// @step: we need to wait for a node interval and then attempt to validate cluster
			update.Infof("waiting to %s while the instancegroup: %s is reverted", strategy.Interval.Duration, name)
			if err = p.GroupUpdate.WaitFor(ctx, strategy.Interval.Duration); err != nil {
				return err
			}

			// @step: attempt to validate the cluster
			update.Infof("attempting to validate post drain on instancegroup: %s", name)
			if err = p.GroupUpdate.ValidateClusterWithTimeout(ctx, list, update.FailOnValidateTimeout); err != nil {
				return err
			}
		}

		return nil
	}()
	if result != nil {
		// @question: should we rollout the instancegroup size here? ... perhaps have and option to choice??
		// we have adjusted the termnation policy to cope with this
		return update.Errorf(result.Error())
	}

	return nil
}

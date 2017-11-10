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

	api "k8s.io/kops/pkg/apis/kops"
)

// defaultProvider is the standard provider
type defaultProvider struct {
	GroupUpdate *RollingUpdateInstanceGroup
}

// NewDefaultRollout returns a default provider
func NewDefaultRollout(update *RollingUpdateInstanceGroup) Rollout {
	return &defaultProvider{GroupUpdate: update}
}

// RollingUpdate is responsible for performing the rollout
func (p *defaultProvider) RollingUpdate(ctx context.Context, list *api.InstanceGroupList) error {
	group := p.GroupUpdate.CloudGroup.InstanceGroup
	name := group.Name
	strategy := group.Spec.Strategy
	update := p.GroupUpdate.Update

	interval := strategy.Interval.Duration.String()

	p.GroupUpdate.Infof("using stratergy: %s, instancegroup: %s, batch: %d, interval: %s",
		strategy.Rollout, name, strategy.Batch, interval)

	return p.GroupUpdate.DrainGroup(ctx, &DrainOptions{
		Batch:             strategy.Batch,
		CloudOnly:         update.CloudOnly,
		Delete:            true,
		DrainPods:         strategy.Drain,
		Interval:          strategy.Interval.Duration,
		PostDelay:         strategy.PostDrainDelay.Duration,
		Timeout:           strategy.DrainTimeout.Duration,
		ValidateCluster:   !update.CloudOnly,
		ValidationTimeout: update.FailOnValidateTimeout,
	}, list)
}

/*

err := func() error {
	for i, node := range group.NeedUpdate {
		// @step: wait for either our turn to operate or a stop signal from the gods above
		select {
		case <-ctx.Done():
			return ErrRolloutCancelled
		case err := <- errorsCh:
			return err
		case <-batch:
		}

		// @check if we have a rollout count and we are over it, we can break out
		count := p.GroupUpdate.Update.Count
		if count > 0 && i >= count {
			skipping := len(group.NeedUpdate) - count
			//p.GroupUpdate.("reached instance count: %d in instancegroup: %s, skipping: %d", count, name, skipping)
			return nil
		}

		wg.Add(1)
		go func(member *cloudinstances.CloudInstanceGroupMember) {
			defer func() {
				batch <- struct{}{}
				wg.Done()
			}()

			if err := p.performRollout(member, list); err != nil {
				p.GroupUpdate.Errorf("failed to rollout member: %s in  instancegroup: %s, error: %s", member.ID, groupName, err)
				stopCh <- struct{}{}
			}
		}(node)
	}
}()
// performRollout is responsible for handling the actual deployment
func (p *defaultProvider) performRollout(member *cloudinstances.CloudInstanceGroupMember, list *api.InstanceGroupList) error {
	update := p.GroupUpdate.Update
	group := p.GroupUpdate.CloudGroup
	groupName := group.HumanName

	strategy := group.InstanceGroup.Spec.Strategy
	drainTimeout := strategy.DrainTimeout.Duration
	nodeInterval := strategy.Interval.Duration
	postDrainDelay := strategy.PostDrainDelay.Duration
	failOnValidationTimeout := update.FailOnValidateTimeout

	status := p.GroupUpdate.Status

	// @step: get the node name, or default to the member id
	nodeName := member.ID
	if member.Node != nil {
		nodeName = member.Node.Name
	}

	// @check if we are cloudonly
	if update.CloudOnly {
		status.Event("not draining member: %s as rollout is cloudonly", nodeName)
	}

	if !*strategy.Drain {
		status.Event("not draining member: %s as feature not enabled", nodeName)
	}

	// @check if cloudonly is false and drain is true, lets drain the member
	if !update.CloudOnly && *strategy.Drain {
		// @check if this is a registered member instance
		if member.Node == nil {
			status.Event("skipping drain of member: %s, because it is not registered in kubernetes", nodeName)
		} else {
			status.Event("draining member: %s, instancegroups: %s, timeout: %s", nodeName, groupName, drainTimeout.String())

			if err := p.GroupUpdate.DrainNode(member, drainTimeout, postDrainDelay); err != nil {
				if update.FailOnDrainError {
					return status.Error("failed to drain member: %s, error: %v", nodeName, err)
				}
				status.Event("ignoring drain error on member: %s, error: %v", nodeName, err)
			}
		}
	}
	status.Event("deleting the member: %s from instance group: %s", nodeName, groupName)

	// @step: terminate the instance from the group
	if err := update.Cloud.DeleteInstance(member); err != nil {
		return status.Error("unable to delete group member: %s, error: %v", nodeName, err)
	}

	// @step: wait for the new instance to enter into the group
	if nodeInterval > 0 {
		status.Event("waiting for %s post instance termination of: %s", nodeInterval.String(), nodeName)
		time.Sleep(nodeInterval)
	}

	// @step: if were not cloudonly lets attempt to validate the cluster
	if !update.CloudOnly {
		interval := time.Second * 10

		status.Event("attempting to validate the cluster, timeout: %s", failOnValidationTimeout.String())

		// @step: check the cluster validates
		if err := p.GroupUpdate.ValidateClusterWithTimeout(list, failOnValidationTimeout, interval); err != nil {
			if update.FailOnValidate {
				return status.Error("failed validating after removing member, error: %v", err)
			}

			status.Event("cluster validation failed but skipping since fail-on-validate is false: %v", err)
		}
	}

	return nil
}
*/

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

package instancegroups

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/upup/pkg/fi/cloudup"
)

// RollingUpdate performs a rolling update on the kubernetes cluster
func (c *RollingUpdateCluster) RollingUpdate(ctx context.Context, options *RollingUpdateOptions) ResultCh {
	doneCh := make(ResultCh, 0)

	// @step: create a anonymous roles and rates struct
	instanceGroupRate := []struct {
		Rate int
		Role api.InstanceGroupRole
	}{
		{Rate: 1, Role: api.InstanceGroupRoleBastion},
		{Rate: 1, Role: api.InstanceGroupRoleMaster},
		{Rate: c.NodeBatch, Role: api.InstanceGroupRoleNode},
	}
	// @step: get a sorted list of instancegroups by roles
	instancegroups := SortInstanceGroupByRole(options.InstanceGroups)

	// @step: iterate the roles in order and kick off a deployment on each of the instancegroups within the role.
	go func() {
		err := func() error {
			for _, x := range instanceGroupRate {
				// get the groups for this role and fire off a deployment to the groups
				groups, found := instancegroups[x.Role]
				if !found {
					continue
				}
				// @step: kick off a deployment for this role
				resultCh := c.invokeInstanceGroupUpdate(ctx, groups, options.List, x.Rate)
				// @step: wait for either a termination from above or a result from the rollout
				select {
				case <-ctx.Done():
					c.Infof("recieved termination signal, waiting for rollout to cancel")
					return <-resultCh
				case err := <-resultCh:
					if err != nil {
						return fmt.Errorf("encountered an error on deployment: %s", err)
					}
				}
			}

			return nil
		}()
		doneCh <- err
	}()

	return doneCh
}

// invokeInstanceGroupUpdate is responsible for handling the rollout of instancegroups for a particular role
func (c *RollingUpdateCluster) invokeInstanceGroupUpdate(ctx context.Context, groups []*cloudinstances.CloudInstanceGroup, list *api.InstanceGroupList, batch int) ResultCh {
	// used to handle the concurrency of the rollouts
	bucket := newRateBucket(batch)
	// used to recieve the errors from each rollout
	errorCh := make(ResultCh, batch)
	// doneCh is used to internal signal an end of work
	doneCh := make(SignalCh, 0)
	// a channel used to hand pass result upstream
	resultCh := make(ResultCh, 0)
	// a wait group used to wait for child routines to finish
	worker := sync.WaitGroup{}
	// a cascaded context
	rctx, cancel := context.WithCancel(context.Background())

	// @NOTES:
	// - Since we are not passing the errors upstream we have a local context used to cancel the rollout on
	//   any errors or from a cancellation above.
	// - The control routine below listens for a cancellation, and error from the worker routines or a done signal
	// - The worker simply iterates the groups in the role passes the context to the group rollout implementation.
	//   On any errors the error is as indicated above picked up by the controller routine; whom immediately cancels
	//   the context. Everything underneath cancels on the context and records and error.
	// - On exit of the loop, we sit and wait for the worker jobs to finish, either by finishing the work or being
	//   cancel via context.

	// @step: create a controller routine
	go func() {
		err := func() error {
			select {
			case <-ctx.Done():
				cancel()
				return ErrRolloutCancelled
			case err := <-errorCh:
				cancel()
				return err
			case <-doneCh:
				return nil
			}
		}()
		resultCh <- err
	}()

	// @step: we iterate the groups within the role and kick off a rollout if required
	go func() {
		if err := func() error {
			for _, x := range groups {
				// @check if the instancegroup is being filtered out
				name := x.InstanceGroup.Name
				if !c.IsGroupUpdating(name) {
					continue
				}
				// @step: wait for a slot to operate
				<-bucket

				// @step: determine the update strategy for this instancegroup
				if err := c.DetermineGroupStratergy(x); err != nil {
					return c.Errorf("unable to determine rollout strategy on instancegroup: %s, error: %v", name, err)
				}

				// @check we have instances inside the group which require updating
				if len(x.NeedUpdate) == 0 {
					c.Infof("skipping instancegroup: %s as no members have pending updates", name)
					continue
				}

				worker.Add(1)
				go func(group *cloudinstances.CloudInstanceGroup) {
					defer func() {
						worker.Done()
						bucket <- Signal
					}()
					update := &RollingUpdateInstanceGroup{Update: c, CloudGroup: group}
					// @step: perform a rollout on the group
					if err := update.RollingUpdate(rctx, list); err != nil {
						// return the error for the next iteration to pick up; so the channel is
						// buffered so non-blocking here
						errorCh <- err
					}
				}(x)
			}

			return nil
		}(); err != nil {
			errorCh <- err
			return
		}
		// @step: wait for all the routines to finish
		worker.Wait()
		doneCh <- Signal
	}()

	return resultCh
}

// UpdateCluster is responsible for updating the kops cluster
func (c *RollingUpdateCluster) UpdateCluster(ctx context.Context) error {
	// @NOTE i'm not sure there is a way of cancelling this operation??
	applyCmd := &cloudup.ApplyClusterCmd{
		Clientset:       c.Clientset,
		Cluster:         c.Cluster,
		DryRun:          false,
		InstanceGroups:  nil,
		MaxTaskDuration: cloudup.DefaultMaxTaskDuration,
		Models:          nil,
		OutDir:          ".",
		TargetName:      cloudup.TargetDirect,
	}

	return applyCmd.Run()
}

// Infof is used to provide details about the on-going rollout
func (c *RollingUpdateCluster) Infof(message string, opts ...interface{}) {
	glog.Infof(fmt.Sprintf(message, opts...))
}

// Errorf provides an error log for the rolling update
func (c *RollingUpdateCluster) Errorf(message string, opts ...interface{}) error {
	msg := fmt.Sprintf(message, opts...)
	glog.Errorf(msg)

	return errors.New(msg)
}

// DetermineGroupStratergy is responsible for strategy for this instancegroup
func (c *RollingUpdateCluster) DetermineGroupStratergy(group *cloudinstances.CloudInstanceGroup) error {
	groupName := group.InstanceGroup.Name
	role := group.InstanceGroup.Spec.Role
	strategy := group.InstanceGroup.Spec.Strategy

	// @check if nil and if so give it a default stratergy
	if strategy == nil {
		strategy = &api.UpdateStrategy{
			Batch: 1,
			Drain: false,
			Name:  api.DefaultRollout,
		}
		group.InstanceGroup.Spec.Strategy = strategy
	}
	if strategy.DrainTimeout == nil {
		strategy.DrainTimeout = &metav1.Duration{Duration: c.DrainTimeout}
	}
	if strategy.PostDrainDelay == nil {
		strategy.PostDrainDelay = &metav1.Duration{Duration: c.PostDrainDelay}
	}
	// @check if rollout options overrides the ig strategy
	if c.Strategy != "" {
		c.Infof("using rollout strategy: %s on instancegroup: %s", strategy.Name, groupName)
		strategy.Name = c.Strategy
	}
	// @check is rollout options override post delay
	if c.PostDrainDelay > 0 {
		strategy.PostDrainDelay = &metav1.Duration{Duration: c.PostDrainDelay}
	}
	// @check if we forcing a update and update
	if c.Force {
		group.NeedUpdate = append(group.NeedUpdate, group.Ready...)
		group.Ready = make([]*cloudinstances.CloudInstanceGroupMember, 0)
	}
	// @check if rollout strategy is nothing an default
	if strategy.Name == "" {
		strategy.Name = api.DefaultRollout
	}
	// @check if the batch size is at least one
	if strategy.Batch <= 0 {
		strategy.Batch = 1
	}
	// @check if the batch size has been overrided by options
	if c.Batch > 0 {
		strategy.Batch = c.Batch
	}
	// @check if cloudonly and adject the drain options
	if c.CloudOnly {
		strategy.Drain = false
	}
	if c.Drain != nil {
		strategy.Drain = *c.Drain
	}

	// @step: work out the interval for this role
	if strategy.Interval == nil {
		switch role {
		case api.InstanceGroupRoleBastion:
			strategy.Interval = &metav1.Duration{Duration: c.BastionInterval}
		case api.InstanceGroupRoleMaster:
			strategy.Interval = &metav1.Duration{Duration: c.MasterInterval}
		case api.InstanceGroupRoleNode:
			strategy.Interval = &metav1.Duration{Duration: c.NodeInterval}
		}
	}

	// @step: check the stratergy is compatible with the role
	switch role {
	case api.InstanceGroupRoleMaster, api.InstanceGroupRoleBastion:
		if strategy.Name != api.DefaultRollout {
			return fmt.Errorf("rollout strategy: %s is not supported for role: %s", strategy.Name, role)
		}
	}

	// @check the rollout stratergy
	switch strategy.Name {
	case api.DuplicateRollout, api.ScaleUpRollout:
		if role != api.InstanceGroupRoleNode {
			return c.Errorf("rollout strategy: %s is only supported on node instancegroups", api.DuplicateRollout)
		}
		if len(group.NeedUpdate) > 0 {
			c.Infof("toggling the force flag given we are using a %s strategy", strategy.Name)
			c.Force = true
		}
	}

	// @check try to ensure that if everything requires a update and the batch size is
	// bigger than the group, we leave an instance active
	if len(group.Ready) == 0 {
		if strategy.Batch > 1 && strategy.Batch >= len(group.NeedUpdate) {
			strategy.Batch = strategy.Batch - 1
			c.Infof("adjusting batch size to: %d to keep availability Ready/Update (0/%d)",
				strategy.Batch, len(group.NeedUpdate))
		}
	}

	return nil
}

// IsGroupUpdating checks to see if the instancegroup is going to rollout
func (c *RollingUpdateCluster) IsGroupUpdating(name string) bool {
	if len(c.InstanceGroups) <= 0 {
		return true
	}

	for _, x := range c.InstanceGroups {
		if x == name {
			return true
		}
	}

	return false
}

// SortInstanceGroupByRole is responsible for slicing up the instance group by role
func SortInstanceGroupByRole(groups map[string]*cloudinstances.CloudInstanceGroup) map[api.InstanceGroupRole][]*cloudinstances.CloudInstanceGroup {
	list := make(map[api.InstanceGroupRole][]*cloudinstances.CloudInstanceGroup, 0)

	for _, x := range groups {
		switch x.InstanceGroup.Spec.Role {
		case api.InstanceGroupRoleNode:
			list[api.InstanceGroupRoleNode] = append(list[api.InstanceGroupRoleNode], x)
		case api.InstanceGroupRoleMaster:
			list[api.InstanceGroupRoleMaster] = append(list[api.InstanceGroupRoleMaster], x)
		case api.InstanceGroupRoleBastion:
			list[api.InstanceGroupRoleBastion] = append(list[api.InstanceGroupRoleBastion], x)
		}
	}

	return list
}

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
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/cloudinstances"
	"k8s.io/kops/pkg/validation"
	"k8s.io/kops/upup/pkg/fi/cloudup"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/conversion"
	"k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	// DuplicateParentAnnotation is a group annotation name for the parent used when duplicating a cluster instance group
	DuplicateParentAnnotation = "kops.alpha.kubernetes.io/instancegroup/parent"
)

var (
	// validationRetryInterval is the time duration between cluster validation retires
	validationRetryInterval = time.Second * 3
	// validationCheckInterval is the time duration between attempts to valdated cluster
	validationCheckInterval = time.Second * 15
)

// RollingUpdateInstanceGroup is the AWS ASG backing an InstanceGroup.
type RollingUpdateInstanceGroup struct {
	// Update is the reference to the cluster update
	Update *RollingUpdateCluster
	// CloudGroup is the cloud instanceGroup we are updating
	CloudGroup *cloudinstances.CloudInstanceGroup

	// TODO should remove the need to have rollingupdate struct and add:
	// TODO - the kubernetes client
	// TODO - the cluster name
	// TODO - the client config
	// TODO - fail on validate
	// TODO - fail on drain
	// TODO - cloudonly
}

// RollingUpdate performs a rolling update on a instanceGroup
func (r *RollingUpdateInstanceGroup) RollingUpdate(ctx context.Context, list *api.InstanceGroupList) error {
	update := r.Update

	// @step: if this is non-cloudonly rolling update we should validate the cluster before update
	if !update.CloudOnly && !r.IsRole(api.InstanceGroupRoleBastion) {
		if err := r.ValidateClusterWithRetries(ctx, list, 3); err != nil {
			if update.FailOnValidate {
				return update.Errorf("cluster has failed validation: %v", err)
			}
			update.Infof("ignoring cluster validate error: %v as fail-on-validation is false", err)
		}
	}

	// @step: kick off a deployment to the instancegroup, passing the context
	return r.NewRollout().RollingUpdate(ctx, list)
}

// DrainGroup is responsible for draining the instance group and deleting if required
func (r *RollingUpdateInstanceGroup) DrainGroup(ctx context.Context, options *DrainOptions, list *api.InstanceGroupList) error {
	bucket := newRateBucket(options.Batch)
	errorCh := make(ResultCh, options.Batch)
	groupName := r.CloudGroup.InstanceGroup.Name
	worker := &sync.WaitGroup{}

	if options.CloudOnly {
		r.Infof("performing a cloudonly rollout on instancegroup: %s", groupName)
	}

	err := func() error {
		for i, x := range r.CloudGroup.NeedUpdate {
			select {
			case <-ctx.Done():
				return ErrRolloutCancelled
			case err := <-errorCh:
				return err
			case <-bucket:
			}

			// @check if we are only rolling out x instances
			if options.Count > 0 && i > options.Count {
				r.Infof("stopping rollout in instancegroup: %s, max count: %d, skipped: %d",
					groupName, options.Count, len(r.CloudGroup.NeedUpdate)-options.Count)
				return nil
			}

			worker.Add(1)
			go func(node *cloudinstances.CloudInstanceGroupMember) {
				defer func() {
					worker.Done()
					bucket <- Signal
				}()
				// @step: set the nodename
				nodeName := node.ID
				if node.Node != nil {
					nodeName = node.Node.Name
				}

				// add a convient wrapper to handle any errors
				err := func() error {
					// @check if we are draining the node of pods
					if options.DrainPods {
						// @check this is a known kubernetes node
						if node.Node == nil {
							r.Infof("unknown node: %s found in instancegroup: %s, skipping the drain", x.ID, groupName)
						} else {
							r.Infof("draining node: %s from instancegroup: %s", nodeName, groupName)
							if err := r.DrainNode(ctx, node, options); err != nil {
								return r.Errorf("failed to drain the node: %s in instancegroup: %s", nodeName, groupName)
							}

							// @step: should be add a delay post the drain?
							if options.PostDelay > 0 {
								r.Infof("waiting on pods to stabilize in instancegroup: %s, waiting: %s", groupName, options.PostDelay.String())
								if err := r.WaitFor(ctx, options.PostDelay); err != nil {
									return err
								}
							}
						}
					}

					// @check if we should delete this node
					if options.Delete {
						r.Infof("deleting the node: %s from instancegroup: %s", nodeName, groupName)
						if err := r.DeleteInstance(ctx, node); err != nil {
							return r.Errorf("failed to delete node: %s, instancegroup: %s, error: %v", nodeName, groupName, err)
						}
					}

					// @check if we should wait for a certain time before moving on
					if options.Interval > 0 {
						r.Infof("waiting for %s before moving to next instance in instancegroup: %s", options.Interval.String(), groupName)
						if err := r.WaitFor(ctx, options.Interval); err != nil {
							return err
						}
					}

					// @check if we should validate the cluster
					if options.ValidateCluster {
						r.Infof("validating cluster post update on node: %s, timeout: %s, fail-on-error: %t",
							nodeName, options.ValidationTimeout.String(), options.FailOnValidation)
						if err := r.ValidateClusterWithTimeout(ctx, list, options.ValidationTimeout); err != nil {
							if options.FailOnValidation {
								return r.Errorf("failed validating after removing member, error: %v", err)
							}

							r.Errorf("cluster validation failed but skipping since fail-on-validate is false: %v", err)
						}
					}

					return nil
				}()
				if err != nil {
					// push the error in the channel, this is a buffered channel so the operation
					// is non-blocking
					errorCh <- err
				}
			}(x)
		}

		return nil
	}()
	// @step: wait for everything to checkin and return
	worker.Wait()

	return err
}

// DrainNode is responsible for cordoning a node and draining the pods
func (r *RollingUpdateInstanceGroup) DrainNode(ctx context.Context, u *cloudinstances.CloudInstanceGroupMember, options *DrainOptions) error {
	f := cmdutil.NewFactory(r.Update.ClientConfig)
	out := ioutil.Discard
	errOut := ioutil.Discard

	cmdOptions := &cmd.DrainOptions{
		DeleteLocalData:  true,
		ErrOut:           errOut,
		Factory:          f,
		Force:            true,
		IgnoreDaemonsets: true,
		Out:              out,
		Timeout:          options.Timeout,
	}

	cmd := &cobra.Command{Use: "cordon NODE"}
	args := []string{u.Node.Name}

	if err := cmdOptions.SetupDrain(cmd, args); err != nil {
		return fmt.Errorf("error setting up drain: %v", err)
	}

	if err := cmdOptions.RunCordonOrUncordon(true); err != nil {
		return fmt.Errorf("error cordoning node node: %v", err)
	}

	if err := cmdOptions.RunDrain(); err != nil {
		return fmt.Errorf("error draining node: %v", err)
	}

	return nil
}

// DuplicateGroup creates a copy of the InstanceGroup cluster instance group for the provided cluster.
// The cluster instance groups are updated with annotations that denote a child an parent relationship.
func (r *RollingUpdateInstanceGroup) DuplicateGroup(name string) (*api.InstanceGroup, error) {
	clientset := r.Update.Clientset
	cluster := r.Update.Cluster
	group := r.CloudGroup.InstanceGroup

	obj, err := conversion.NewCloner().DeepCopy(group)
	if err != nil {
		return nil, fmt.Errorf("unable to clone instance group: %v", err)
	}

	ig, ok := obj.(*api.InstanceGroup)
	if !ok {
		return nil, fmt.Errorf("unexpected object type: %T", obj)
	}

	// DeepCopy does not get the maps need to copy those through
	ig.ObjectMeta.Annotations = group.ObjectMeta.Annotations
	ig.ObjectMeta.Labels = group.ObjectMeta.Labels
	ig.ObjectMeta.Name = name
	ig.Spec.CloudLabels = group.Spec.CloudLabels
	ig.Spec.NodeLabels = group.Spec.NodeLabels
	ig.Spec.Role = group.Spec.Role

	if ig.ObjectMeta.Annotations == nil {
		ig.ObjectMeta.Annotations = make(map[string]string)
	}

	ig.ObjectMeta.Annotations[DuplicateParentAnnotation] = group.ObjectMeta.Name

	if err := cloudup.CreateInstanceGroup(ig, cluster, clientset); err != nil {
		return nil, fmt.Errorf("unable to create new instance group: %v", err)
	}

	return ig, nil
}

// ValidateCluster runs our validation methods on the K8s Cluster.
func (r *RollingUpdateInstanceGroup) ValidateCluster(list *api.InstanceGroupList) error {
	if _, err := validation.ValidateCluster(r.Update.ClusterName, list, r.Update.Client); err != nil {
		return fmt.Errorf("cluster: %s did not pass validation: %s", r.Update.ClusterName, err)
	}

	return nil
}

// ValidateClusterWithRetries is responsible for attempting to validate the cluster
func (r *RollingUpdateInstanceGroup) ValidateClusterWithRetries(ctx context.Context, list *api.InstanceGroupList, retries int) error {
	// @step: try to validate cluster at least once, this will handle durations that are lower than our tick time
	if r.tryValidateCluster(list) {
		return nil
	}
	// @TODO: fix this as technically do one more than the retires specified
	ticker := time.NewTicker(validationRetryInterval)
	for i := 1; i < retries; i++ {
		select {
		case <-ctx.Done():
			return ErrRolloutCancelled
		case <-ticker.C:
			if r.tryValidateCluster(list) {
				r.Update.Infof("successfully validated cluster: %s", r.Update.ClusterName)
				return nil
			}
			r.Errorf("cluster has failed validation, attempt: %d, retries: %d", i, retries)
		}
	}

	return fmt.Errorf("cluster has failed validation after %d retries", retries)
}

// ValidateClusterWithTimeout runs validation.ValidateCluster until either we get positive result or the timeout expires
func (r *RollingUpdateInstanceGroup) ValidateClusterWithTimeout(ctx context.Context, list *api.InstanceGroupList, waitTime time.Duration) error {
	// @step: try to validate cluster at least once, this will handle durations that are lower than our tick time
	if r.tryValidateCluster(list) {
		return nil
	}
	expires := time.Now().Add(waitTime)
	tick := time.NewTicker(validationCheckInterval)
	timeout := time.After(waitTime)

	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return ErrRolloutCancelled
		case <-timeout:
			return fmt.Errorf("cluster did not pass validation within: %s", waitTime.String())
		case <-tick.C:
			if r.tryValidateCluster(list) {
				r.Infof("cluster validation successfully")
				return nil
			}
			r.Infof("cluster has not passed validation yet, expiration: %s", expires.Sub(time.Now()).String())
		}
	}
}

// tryValidateCluster attempts to validate to the cluster
func (r *RollingUpdateInstanceGroup) tryValidateCluster(list *api.InstanceGroupList) bool {
	if _, err := validation.ValidateCluster(r.Update.ClusterName, list, r.Update.Client); err != nil {
		return false
	}

	return true
}

// WaitFor is responsible for convenience method used to wait for a certain time or cancel
func (r *RollingUpdateInstanceGroup) WaitFor(ctx context.Context, waitTime time.Duration) error {
	select {
	case <-ctx.Done():
		return ErrRolloutCancelled
	case <-time.After(waitTime):
	}

	return nil
}

// Delete is responsible for deleting a cloudinstanceGroup
func (r *RollingUpdateInstanceGroup) Delete() error {
	return r.Update.Cloud.DeleteGroup(r.CloudGroup)
}

// DeleteInstance is responsible for deleting a instancegroup member
func (r *RollingUpdateInstanceGroup) DeleteInstance(ctx context.Context, u *cloudinstances.CloudInstanceGroupMember) error {
	id := u.ID
	nodeName := ""
	groupName := r.CloudGroup.InstanceGroup.Name

	if u.Node != nil {
		nodeName = u.Node.Name
	}
	if nodeName != "" {
		r.Infof("stopping instance %s, node %s, instancegroup: %s", id, nodeName, groupName)
	} else {
		r.Infof("stopping instance %s, in instancegroup %s", id, groupName)
	}

	if err := r.Update.Cloud.DeleteInstance(u); err != nil {
		if nodeName != "" {
			return r.Errorf("error deleting instance %q, node %q: %v", id, nodeName, err)
		}

		return r.Errorf("error deleting instance %q: %v", id, err)
	}

	return nil
}

// IsRole checks the role of the instancegroup
func (r *RollingUpdateInstanceGroup) IsRole(role api.InstanceGroupRole) bool {
	return r.CloudGroup.InstanceGroup.Spec.Role == role
}

// Role returns the role of the instance group
func (r *RollingUpdateInstanceGroup) Role() api.InstanceGroupRole {
	return r.CloudGroup.InstanceGroup.Spec.Role
}

// NewRollout creates and returns a rollout provider
func (r *RollingUpdateInstanceGroup) NewRollout() Rollout {
	switch name := r.CloudGroup.InstanceGroup.Spec.Strategy.Rollout; name {
	case api.DefaultRollout:
		return NewDefaultRollout(r)
	case api.DuplicateRollout:
		return NewDuplicateRollout(r)
	case api.ScaleUpRollout:
		return NewScaleUpRollout(r)
	default:
		return NewDefaultRollout(r)
	}
}

// Infof is used to provide details about the on-going rollout
func (r *RollingUpdateInstanceGroup) Infof(message string, opts ...interface{}) {
	r.Update.Infof(message, opts...)
}

// Errorf provides an error log for the rolling update
func (r *RollingUpdateInstanceGroup) Errorf(message string, opts ...interface{}) error {
	return r.Update.Errorf(message, opts...)
}

// newRateBucket returns a bucket of x size
func newRateBucket(size int) chan struct{} {
	c := make(chan struct{}, size)
	for i := 0; i < size; i++ {
		c <- struct{}{}
	}

	return c
}

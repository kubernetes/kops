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

	// @step: increase the size of the instancegroup by x percent

	// @step: drain the old

	/*
		ig := p.GroupUpdate.CloudGroup.InstanceGroup.Spec
		strategy := ig.Strategy

		// @step: iterate the member channel and handle the default strategy
		memberCh := make(chan *cloudinstances.CloudInstanceGroupMember, 0)
		errorCh := make(chan error, 0)
		go func() {

		}()

		// @logic
		// - we divide the needsupdate into a series of iterations.
		// - we increase the
		// -

		// @step: iterate the nodes from the previous instanceGroup, drain if required and move one by one of batches
		batch := newRateBucket(strategy.Batch)
		for {
			select {
			// wait for a token to
			case <-batch:

			case err := <-errorCh:
				return err

			}

		}
	*/

	return nil
}

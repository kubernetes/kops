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
	interval := strategy.Interval.Duration

	p.GroupUpdate.Infof("using stratergy: %s, instancegroup: %s, batch: %d, interval: %s",
		strategy.Name, name, strategy.Batch, interval)

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

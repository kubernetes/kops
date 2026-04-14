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

package linodetasks

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/linode"
)

// +kops:fitask
type LoadBalancerBackends struct {
	Name      *string
	Lifecycle fi.Lifecycle

	LoadBalancer *LoadBalancer
}

var _ fi.CloudupTask = &LoadBalancerBackends{}
var _ fi.HasLifecycle = &LoadBalancerBackends{}
var _ fi.HasName = &LoadBalancerBackends{}

func (l *LoadBalancerBackends) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask

	if l.LoadBalancer != nil {
		deps = append(deps, l.LoadBalancer)
	}

	return deps
}

func (l *LoadBalancerBackends) GetLifecycle() fi.Lifecycle {
	return l.Lifecycle
}

func (l *LoadBalancerBackends) SetLifecycle(lifecycle fi.Lifecycle) {
	l.Lifecycle = lifecycle
}

func (l *LoadBalancerBackends) GetName() *string {
	return l.Name
}

func (l *LoadBalancerBackends) String() string {
	return fi.CloudupTaskAsString(l)
}

func (l *LoadBalancerBackends) Run(c *fi.CloudupContext) error {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	// The fi framework does not expose a first-class dry-run predicate, so we
	// type-assert against the concrete dry-run target to avoid blocking a
	// preview run on resources that do not yet exist.
	_, isDryRun := c.Target.(*fi.CloudupDryRunTarget)

	nodebalancerID := fi.ValueOf(l.LoadBalancer.ID)
	if nodebalancerID == 0 {
		if isDryRun {
			return nil
		}
		return fi.NewTryAgainLaterError("waiting for NodeBalancer to be created")
	}

	lbName := fi.ValueOf(l.LoadBalancer.Name)
	backends, err := linodeDiscoverControlPlaneBackends(cloud.Client(), l.LoadBalancer.Tags)
	if err != nil {
		return err
	}
	if len(backends) == 0 {
		if isDryRun {
			return nil
		}
		return fi.NewTryAgainLaterError("waiting for backend instances to be ready")
	}

	return ensureLoadBalancerConfigs(cloud.Client(), nodebalancerID, lbName, backends)
}

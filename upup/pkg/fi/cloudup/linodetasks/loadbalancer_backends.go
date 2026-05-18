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
	"fmt"

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

func (l *LoadBalancerBackends) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(l, c)
}

func (l *LoadBalancerBackends) Find(c *fi.CloudupContext) (*LoadBalancerBackends, error) {
	cloud := c.T.Cloud.(linode.LinodeCloud)

	nodebalancerID := fi.ValueOf(l.LoadBalancer.ID)
	if nodebalancerID == 0 {
		return nil, nil
	}

	configs, err := cloud.Client().ListNodeBalancerConfigs(c.Context(), nodebalancerID, nil)
	if err != nil {
		return nil, fmt.Errorf("error listing Linode (Akamai) NodeBalancer configs: %w", err)
	}
	if len(configs) == 0 {
		return nil, nil
	}

	return &LoadBalancerBackends{
		Name:         l.Name,
		Lifecycle:    l.Lifecycle,
		LoadBalancer: l.LoadBalancer,
	}, nil
}

func (_ *LoadBalancerBackends) CheckChanges(actual, expected, changes *LoadBalancerBackends) error {
	return nil
}

func (_ *LoadBalancerBackends) RenderLinode(t *linode.APITarget, actual, expected, changes *LoadBalancerBackends) error {
	nodebalancerID := fi.ValueOf(expected.LoadBalancer.ID)
	if nodebalancerID == 0 {
		return fi.NewTryAgainLaterError("waiting for NodeBalancer to be created")
	}

	lbName := fi.ValueOf(expected.LoadBalancer.Name)
	backends, err := linodeDiscoverControlPlaneBackends(t.Cloud.Client(), expected.LoadBalancer.Tags)
	if err != nil {
		return err
	}
	if len(backends) == 0 {
		return fi.NewTryAgainLaterError("waiting for backend instances to be ready")
	}

	return ensureLoadBalancerConfigs(t.Cloud.Client(), nodebalancerID, lbName, backends)
}

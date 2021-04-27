/*
Copyright 2021 The Kubernetes Authors.

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

package mockcompute

import (
	"context"
	"fmt"
	"sync"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type forwardingRuleClient struct {
	// forwardingRules are forwardingRules keyed by project, region, and forwardingRule name.
	forwardingRules map[string]map[string]map[string]*compute.ForwardingRule
	sync.Mutex
}

var _ gce.ForwardingRuleClient = &forwardingRuleClient{}

func newForwardingRuleClient() *forwardingRuleClient {
	return &forwardingRuleClient{
		forwardingRules: map[string]map[string]map[string]*compute.ForwardingRule{},
	}
}

func (c *forwardingRuleClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, regions := range c.forwardingRules {
		for _, frs := range regions {
			for n, fr := range frs {
				m[n] = fr
			}
		}
	}
	return m
}

func (c *forwardingRuleClient) Insert(project, region string, fr *compute.ForwardingRule) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.forwardingRules[project]
	if !ok {
		regions = map[string]map[string]*compute.ForwardingRule{}
		c.forwardingRules[project] = regions
	}
	frs, ok := regions[region]
	if !ok {
		frs = map[string]*compute.ForwardingRule{}
		regions[region] = frs
	}
	fr.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/forwardingRules/%s", project, region, fr.Name)
	frs[fr.Name] = fr
	return doneOperation(), nil
}

func (c *forwardingRuleClient) Delete(project, region, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.forwardingRules[project]
	if !ok {
		return nil, notFoundError()
	}
	frs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := frs[name]; !ok {
		return nil, notFoundError()
	}
	delete(frs, name)
	return doneOperation(), nil
}

func (c *forwardingRuleClient) Get(project, region, name string) (*compute.ForwardingRule, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.forwardingRules[project]
	if !ok {
		return nil, notFoundError()
	}
	frs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	fr, ok := frs[name]
	if !ok {
		return nil, notFoundError()
	}
	return fr, nil
}

func (c *forwardingRuleClient) List(ctx context.Context, project, region string) ([]*compute.ForwardingRule, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.forwardingRules[project]
	if !ok {
		return nil, nil
	}
	frs, ok := regions[region]
	if !ok {
		return nil, nil
	}
	var l []*compute.ForwardingRule
	for _, fr := range frs {
		l = append(l, fr)
	}
	return l, nil
}

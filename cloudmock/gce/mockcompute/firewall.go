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

type firewallClient struct {
	// firewalls are firewalls keyed by project and firewall name.
	firewalls map[string]map[string]*compute.Firewall
	sync.Mutex
}

var _ gce.FirewallClient = &firewallClient{}

func newFirewallClient() *firewallClient {
	return &firewallClient{
		firewalls: map[string]map[string]*compute.Firewall{},
	}
}

func (c *firewallClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, fws := range c.firewalls {
		for n, fw := range fws {
			m[n] = fw
		}
	}
	return m
}

func (c *firewallClient) Insert(project string, firewall *compute.Firewall) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	firewalls, ok := c.firewalls[project]
	if !ok {
		firewalls = map[string]*compute.Firewall{}
		c.firewalls[project] = firewalls
	}
	firewall.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/firewalls/%s", project, firewall.Name)
	firewalls[firewall.Name] = firewall
	return doneOperation(), nil
}

func (c *firewallClient) Delete(project, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	firewalls, ok := c.firewalls[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := firewalls[name]; !ok {
		return nil, notFoundError()
	}
	delete(firewalls, name)
	return doneOperation(), nil
}

func (c *firewallClient) Update(project, name string, fw *compute.Firewall) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	firewalls, ok := c.firewalls[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := firewalls[name]; !ok {
		return nil, notFoundError()
	}
	firewalls[name] = fw
	return doneOperation(), nil
}

func (c *firewallClient) Get(project, name string) (*compute.Firewall, error) {
	c.Lock()
	defer c.Unlock()
	firewalls, ok := c.firewalls[project]
	if !ok {
		return nil, notFoundError()
	}
	firewall, ok := firewalls[name]
	if !ok {
		return nil, notFoundError()
	}
	return firewall, nil
}

func (c *firewallClient) List(ctx context.Context, project string) ([]*compute.Firewall, error) {
	c.Lock()
	defer c.Unlock()
	firewalls, ok := c.firewalls[project]
	if !ok {
		return nil, nil
	}
	var l []*compute.Firewall
	for _, fw := range firewalls {
		l = append(l, fw)
	}
	return l, nil
}

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

type subnetworkClient struct {
	// subnetworks are subnetworks keyed by project, region, and subnetwork name.
	subnetworks map[string]map[string]map[string]*compute.Subnetwork
	sync.Mutex
}

var _ gce.SubnetworkClient = &subnetworkClient{}

func newSubnetworkClient() *subnetworkClient {
	return &subnetworkClient{
		subnetworks: map[string]map[string]map[string]*compute.Subnetwork{},
	}
}

func (c *subnetworkClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, regions := range c.subnetworks {
		for _, subs := range regions {
			for n, sub := range subs {
				m[n] = sub
			}
		}
	}
	return m
}

func (c *subnetworkClient) Insert(project, region string, sub *compute.Subnetwork) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.subnetworks[project]
	if !ok {
		regions = map[string]map[string]*compute.Subnetwork{}
		c.subnetworks[project] = regions
	}
	subs, ok := regions[region]
	if !ok {
		subs = map[string]*compute.Subnetwork{}
		regions[region] = subs
	}
	sub.Region = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s", project, region)
	sub.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/subnetworks/%s", project, region, sub.Name)
	subs[sub.Name] = sub
	return doneOperation(), nil
}

func (c *subnetworkClient) Patch(project, region, name string, sub *compute.Subnetwork) (*compute.Operation, error) {
	// Insert does the locking here
	// c.Lock()
	// defer c.Unlock()
	return c.Insert(project, region, sub)
}

func (c *subnetworkClient) Delete(project, region, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.subnetworks[project]
	if !ok {
		return nil, notFoundError()
	}
	subs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := subs[name]; !ok {
		return nil, notFoundError()
	}
	delete(subs, name)
	return doneOperation(), nil
}

func (c *subnetworkClient) Get(project, region, name string) (*compute.Subnetwork, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.subnetworks[project]
	if !ok {
		return nil, notFoundError()
	}
	subs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	sub, ok := subs[name]
	if !ok {
		return nil, notFoundError()
	}
	return sub, nil
}

func (c *subnetworkClient) List(ctx context.Context, project, region string) ([]*compute.Subnetwork, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.subnetworks[project]
	if !ok {
		return nil, nil
	}
	subs, ok := regions[region]
	if !ok {
		return nil, nil
	}
	var l []*compute.Subnetwork
	for _, s := range subs {
		l = append(l, s)
	}
	return l, nil
}

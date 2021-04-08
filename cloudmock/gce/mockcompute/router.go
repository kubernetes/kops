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

type routerClient struct {
	// routers are routers keyed by project, region, and router name.
	routers map[string]map[string]map[string]*compute.Router
	sync.Mutex
}

var _ gce.RouterClient = &routerClient{}

func newRouterClient() *routerClient {
	return &routerClient{
		routers: map[string]map[string]map[string]*compute.Router{},
	}
}

func (c *routerClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, regions := range c.routers {
		for _, rs := range regions {
			for n, r := range rs {
				m[n] = r
			}
		}
	}
	return m
}

func (c *routerClient) Insert(project, region string, r *compute.Router) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.routers[project]
	if !ok {
		regions = map[string]map[string]*compute.Router{}
		c.routers[project] = regions
	}
	rs, ok := regions[region]
	if !ok {
		rs = map[string]*compute.Router{}
		regions[region] = rs
	}
	r.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/routers/%s", project, region, r.Name)
	rs[r.Name] = r
	return doneOperation(), nil
}

func (c *routerClient) Delete(project, region, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.routers[project]
	if !ok {
		return nil, notFoundError()
	}
	rs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := rs[name]; !ok {
		return nil, notFoundError()
	}
	delete(rs, name)
	return doneOperation(), nil
}

func (c *routerClient) Get(project, region, name string) (*compute.Router, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.routers[project]
	if !ok {
		return nil, notFoundError()
	}
	rs, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	r, ok := rs[name]
	if !ok {
		return nil, notFoundError()
	}
	return r, nil
}

func (c *routerClient) List(ctx context.Context, project, region string) ([]*compute.Router, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.routers[project]
	if !ok {
		return nil, nil
	}
	rs, ok := regions[region]
	if !ok {
		return nil, nil
	}
	var l []*compute.Router
	for _, r := range rs {
		l = append(l, r)
	}
	return l, nil
}

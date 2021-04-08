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
	"sync"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type routeClient struct {
	// routes are routes keyed by project and route name.
	routes map[string]map[string]*compute.Route
	sync.Mutex
}

var _ gce.RouteClient = &routeClient{}

func newRouteClient() *routeClient {
	return &routeClient{
		routes: map[string]map[string]*compute.Route{},
	}
}

func (c *routeClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, routes := range c.routes {
		for n, r := range routes {
			m[n] = r
		}
	}
	return m
}

func (c *routeClient) Delete(project, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	routes, ok := c.routes[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := routes[name]; !ok {
		return nil, notFoundError()
	}
	delete(routes, name)
	return doneOperation(), nil
}

func (c *routeClient) List(ctx context.Context, project string) ([]*compute.Route, error) {
	c.Lock()
	defer c.Unlock()
	routes, ok := c.routes[project]
	if !ok {
		return nil, nil
	}
	var l []*compute.Route
	for _, fw := range routes {
		l = append(l, fw)
	}
	return l, nil
}

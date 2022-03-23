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

type backendServiceClient struct {
	// backendServices are backendServices keyed by project and backendService name.
	backendServices map[string]map[string]*compute.BackendService
	sync.Mutex
}

var _ gce.RegionBackendServiceClient = &backendServiceClient{}

func newBackendServiceClient() *backendServiceClient {
	return &backendServiceClient{
		backendServices: map[string]map[string]*compute.BackendService{},
	}
}

func (c *backendServiceClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, nws := range c.backendServices {
		for n, nw := range nws {
			m[n] = nw
		}
	}
	return m
}

func (c *backendServiceClient) Insert(project, region string, backendService *compute.BackendService) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	backendServices, ok := c.backendServices[project]
	if !ok {
		backendServices = map[string]*compute.BackendService{}
		c.backendServices[project] = backendServices
	}
	backendService.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/backendServices/%s", project, region, backendService.Name)
	backendServices[backendService.Name] = backendService
	return doneOperation(), nil
}

func (c *backendServiceClient) Delete(project, _, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	backendServices, ok := c.backendServices[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := backendServices[name]; !ok {
		return nil, notFoundError()
	}
	delete(backendServices, name)
	return doneOperation(), nil
}

func (c *backendServiceClient) Get(project, _, name string) (*compute.BackendService, error) {
	c.Lock()
	defer c.Unlock()
	backendServices, ok := c.backendServices[project]
	if !ok {
		return nil, notFoundError()
	}
	backendService, ok := backendServices[name]
	if !ok {
		return nil, notFoundError()
	}
	return backendService, nil
}

func (c *backendServiceClient) List(_ context.Context, project, _ string) ([]*compute.BackendService, error) {
	c.Lock()
	defer c.Unlock()
	backendServices, ok := c.backendServices[project]
	if !ok {
		return nil, notFoundError()
	}
	var backendServiceList []*compute.BackendService
	for _, backendService := range backendServices {
		backendServiceList = append(backendServiceList, backendService)
	}
	return backendServiceList, nil
}

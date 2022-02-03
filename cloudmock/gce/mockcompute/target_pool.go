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

type targetPoolClient struct {
	// targetPools are targetPools keyed by project, region, and targetPool name.
	targetPools map[string]map[string]map[string]*compute.TargetPool
	sync.Mutex
}

var _ gce.TargetPoolClient = &targetPoolClient{}

func newTargetPoolClient() *targetPoolClient {
	return &targetPoolClient{
		targetPools: map[string]map[string]map[string]*compute.TargetPool{},
	}
}

func (c *targetPoolClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, regions := range c.targetPools {
		for _, tps := range regions {
			for n, tp := range tps {
				m[n] = tp
			}
		}
	}
	return m
}

func (c *targetPoolClient) Insert(project, region string, tp *compute.TargetPool) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.targetPools[project]
	if !ok {
		regions = map[string]map[string]*compute.TargetPool{}
		c.targetPools[project] = regions
	}
	tps, ok := regions[region]
	if !ok {
		tps = map[string]*compute.TargetPool{}
		regions[region] = tps
	}
	tp.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/targetPools/%s", project, region, tp.Name)
	tps[tp.Name] = tp
	return doneOperation(), nil
}

func (c *targetPoolClient) Delete(project, region, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.targetPools[project]
	if !ok {
		return nil, notFoundError()
	}
	tps, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := tps[name]; !ok {
		return nil, notFoundError()
	}
	delete(tps, name)
	return doneOperation(), nil
}

func (c *targetPoolClient) Get(project, region, name string) (*compute.TargetPool, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.targetPools[project]
	if !ok {
		return nil, notFoundError()
	}
	tps, ok := regions[region]
	if !ok {
		return nil, notFoundError()
	}
	tp, ok := tps[name]
	if !ok {
		return nil, notFoundError()
	}
	return tp, nil
}

func (c *targetPoolClient) List(ctx context.Context, project, region string) ([]*compute.TargetPool, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.targetPools[project]
	if !ok {
		return nil, nil
	}
	tps, ok := regions[region]
	if !ok {
		return nil, nil
	}
	var l []*compute.TargetPool
	for _, tp := range tps {
		l = append(l, tp)
	}
	return l, nil
}

func (c *targetPoolClient) List(ctx context.Context, project, region string) ([]*compute.TargetPool, error) {
	c.Lock()
	defer c.Unlock()
	regions, ok := c.targetPools[project]
	if !ok {
		return nil, nil
	}
	tps, ok := regions[region]
	if !ok {
		return nil, nil
	}
	var l []*compute.TargetPool
	for _, tp := range tps {
		l = append(l, tp)
	}
	return l, nil
}

func (c *targetPoolClient) AddHealthCheck(project, region, name string, req *compute.TargetPoolsAddHealthCheckRequest) (*compute.Operation, error) {
	// TODO: AddHealthCheck test
	return doneOperation(), nil
}

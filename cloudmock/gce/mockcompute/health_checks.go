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

type healthCheckClient struct {
	// healthChecks are healthChecks keyed by project and healthCheck name.
	healthChecks map[string]map[string]*compute.HealthCheck
	sync.Mutex
}

var _ gce.RegionHealthChecksClient = &healthCheckClient{}

func newHealthCheckClient() *healthCheckClient {
	return &healthCheckClient{
		healthChecks: map[string]map[string]*compute.HealthCheck{},
	}
}

func (c *healthCheckClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, nws := range c.healthChecks {
		for n, nw := range nws {
			m[n] = nw
		}
	}
	return m
}

func (c *healthCheckClient) Insert(project, region string, healthCheck *compute.HealthCheck) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	healthChecks, ok := c.healthChecks[project]
	if !ok {
		healthChecks = map[string]*compute.HealthCheck{}
		c.healthChecks[project] = healthChecks
	}
	healthCheck.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/healthChecks/%s", project, region, healthCheck.Name)
	healthChecks[healthCheck.Name] = healthCheck
	return doneOperation(), nil
}

func (c *healthCheckClient) Delete(project, _, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	healthChecks, ok := c.healthChecks[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := healthChecks[name]; !ok {
		return nil, notFoundError()
	}
	delete(healthChecks, name)
	return doneOperation(), nil
}

func (c *healthCheckClient) Get(project, _, name string) (*compute.HealthCheck, error) {
	c.Lock()
	defer c.Unlock()
	healthChecks, ok := c.healthChecks[project]
	if !ok {
		return nil, notFoundError()
	}
	healthCheck, ok := healthChecks[name]
	if !ok {
		return nil, notFoundError()
	}
	return healthCheck, nil
}

func (c *healthCheckClient) List(_ context.Context, project, _ string) ([]*compute.HealthCheck, error) {
	c.Lock()
	defer c.Unlock()
	healthChecks, ok := c.healthChecks[project]
	if !ok {
		return nil, notFoundError()
	}
	var healthCheckList []*compute.HealthCheck
	for _, healthCheck := range healthChecks {
		healthCheckList = append(healthCheckList, healthCheck)
	}
	return healthCheckList, nil
}

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

type instanceGroupManagerClient struct {
	// instanceGroupManagers are instanceGroupManagers keyed by project, zone, and name.
	instanceGroupManagers map[string]map[string]map[string]*compute.InstanceGroupManager
	sync.Mutex
}

var _ gce.InstanceGroupManagerClient = &instanceGroupManagerClient{}

func newInstanceGroupManagerClient() *instanceGroupManagerClient {
	return &instanceGroupManagerClient{
		instanceGroupManagers: map[string]map[string]map[string]*compute.InstanceGroupManager{},
	}
}

func (c *instanceGroupManagerClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, zones := range c.instanceGroupManagers {
		for _, igms := range zones {
			for n, igm := range igms {
				m[n] = igm
			}
		}
	}
	return m
}

func (c *instanceGroupManagerClient) Insert(project, zone string, igm *compute.InstanceGroupManager) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.instanceGroupManagers[project]
	if !ok {
		zones = map[string]map[string]*compute.InstanceGroupManager{}
		c.instanceGroupManagers[project] = zones
	}
	igms, ok := zones[zone]
	if !ok {
		igms = map[string]*compute.InstanceGroupManager{}
		zones[zone] = igms
	}
	igm.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/zones/%s/instanceGroupManagers/%s", project, zone, igm.Name)
	igms[igm.Name] = igm
	return doneOperation(), nil
}

func (c *instanceGroupManagerClient) Delete(project, zone, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.instanceGroupManagers[project]
	if !ok {
		return nil, notFoundError()
	}
	igms, ok := zones[zone]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := igms[name]; !ok {
		return nil, notFoundError()
	}
	delete(igms, name)
	return doneOperation(), nil
}

func (c *instanceGroupManagerClient) Get(project, zone, name string) (*compute.InstanceGroupManager, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.instanceGroupManagers[project]
	if !ok {
		return nil, notFoundError()
	}
	igms, ok := zones[zone]
	if !ok {
		return nil, notFoundError()
	}
	igm, ok := igms[name]
	if !ok {
		return nil, notFoundError()
	}
	return igm, nil
}

func (c *instanceGroupManagerClient) List(ctx context.Context, project, zone string) ([]*compute.InstanceGroupManager, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.instanceGroupManagers[project]
	if !ok {
		return nil, nil
	}
	igms, ok := zones[zone]
	if !ok {
		return nil, nil
	}
	var l []*compute.InstanceGroupManager
	for _, d := range igms {
		l = append(l, d)
	}
	return l, nil
}

func (c *instanceGroupManagerClient) ListManagedInstances(ctx context.Context, project, zone, name string) ([]*compute.ManagedInstance, error) {
	var instances []*compute.ManagedInstance
	return instances, nil
}

func (c *instanceGroupManagerClient) RecreateInstances(project, zone, name, id string) (*compute.Operation, error) {
	return doneOperation(), nil
}

func (c *instanceGroupManagerClient) SetTargetPools(project, zone, name string, targetPools []string) (*compute.Operation, error) {
	return doneOperation(), nil
}

func (c *instanceGroupManagerClient) SetInstanceTemplate(project, zone, name, instanceTemplateURL string) (*compute.Operation, error) {
	return doneOperation(), nil
}

func (c *instanceGroupManagerClient) Resize(project, zone, name string, newSize int64) (*compute.Operation, error) {
	return doneOperation(), nil
}

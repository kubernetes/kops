/*
Copyright The Kubernetes Authors.

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

type instanceClient struct {
	// instances are instances keyed by project, zone, and name.
	instances map[string]map[string]map[string]*compute.Instance
	sync.Mutex
}

var _ gce.InstanceClient = &instanceClient{}

func newInstanceClient() *instanceClient {
	return &instanceClient{
		instances: map[string]map[string]map[string]*compute.Instance{},
	}
}

func (c *instanceClient) All() map[string]interface{} {
	return nil
}

func (c *instanceClient) Insert(project, zone string, instance *compute.Instance) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.instances[project]
	if !ok {
		zones = map[string]map[string]*compute.Instance{}
		c.instances[project] = zones
	}
	instances, ok := zones[zone]
	if !ok {
		instances = map[string]*compute.Instance{}
		zones[zone] = instances
	}
	instance.SelfLink = instance.Name
	instances[instance.Name] = instance

	return doneOperation(), nil
}

func (c *instanceClient) Delete(project, zone, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()

	zones, ok := c.instances[project]
	if !ok {
		return nil, notFoundError()
	}
	instances, ok := zones[zone]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := instances[name]; !ok {
		return nil, notFoundError()
	}
	delete(instances, name)
	return doneOperation(), nil
}

func (c *instanceClient) Get(project, zone, name string) (*compute.Instance, error) {
	c.Lock()
	defer c.Unlock()
	zones, ok := c.instances[project]
	if !ok {
		return nil, notFoundError()
	}
	res, ok := zones[zone]
	if !ok {
		return nil, notFoundError()
	}
	igm, ok := res[name]
	if !ok {
		return nil, notFoundError()
	}
	return igm, nil
}

func (c *instanceClient) List(ctx context.Context, project, zone string) ([]*compute.Instance, error) {
	c.Lock()
	defer c.Unlock()

	zones, ok := c.instances[project]
	if !ok {
		return nil, nil
	}
	instances, ok := zones[zone]
	if !ok {
		return nil, nil
	}

	var l []*compute.Instance
	for _, instance := range instances {
		l = append(l, instance)
	}
	return l, nil
}

func (c *instanceClient) SetMetadata(project, zone, name string, metadata *compute.Metadata) (*compute.Operation, error) {
	return nil, fmt.Errorf("setmetadata unimplemented")
}

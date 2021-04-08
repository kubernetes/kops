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

type instanceTemplateClient struct {
	// instanceTemplates are instanceTemplates keyed by project and instanceTemplate name.
	instanceTemplates map[string]map[string]*compute.InstanceTemplate
	sync.Mutex
}

var _ gce.InstanceTemplateClient = &instanceTemplateClient{}

func newInstanceTemplateClient() *instanceTemplateClient {
	return &instanceTemplateClient{
		instanceTemplates: map[string]map[string]*compute.InstanceTemplate{},
	}
}

func (c *instanceTemplateClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, ts := range c.instanceTemplates {
		for n, t := range ts {
			m[n] = t
		}
	}
	return m
}

func (c *instanceTemplateClient) Insert(project string, t *compute.InstanceTemplate) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	ts, ok := c.instanceTemplates[project]
	if !ok {
		ts = map[string]*compute.InstanceTemplate{}
		c.instanceTemplates[project] = ts
	}
	t.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/instanceTemplates/%s", project, t.Name)
	ts[t.Name] = t
	return doneOperation(), nil
}

func (c *instanceTemplateClient) Delete(project, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	ts, ok := c.instanceTemplates[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := ts[name]; !ok {
		return nil, notFoundError()
	}
	delete(ts, name)
	return doneOperation(), nil
}

func (c *instanceTemplateClient) List(ctx context.Context, project string) ([]*compute.InstanceTemplate, error) {
	c.Lock()
	defer c.Unlock()
	ts, ok := c.instanceTemplates[project]
	if !ok {
		return nil, nil
	}
	var l []*compute.InstanceTemplate
	for _, t := range ts {
		l = append(l, t)
	}
	return l, nil
}

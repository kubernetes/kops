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

type httpHealthChecksClient struct {
	// httpHealthchecks are httpHealthchecks keyed by project, and httpHealthcheck name.
	httpHealthchecks map[string]map[string]*compute.HttpHealthCheck
	sync.Mutex
}

var _ gce.HttpHealthChecksClient = &httpHealthChecksClient{}

func newHttpHealthChecksClient() *httpHealthChecksClient {
	return &httpHealthChecksClient{
		httpHealthchecks: map[string]map[string]*compute.HttpHealthCheck{},
	}
}

func (c *httpHealthChecksClient) All() map[string]interface{} {
	c.Lock()
	defer c.Unlock()
	m := map[string]interface{}{}
	for _, hcs := range c.httpHealthchecks {
		for n, hc := range hcs {
			m[n] = hc
		}
	}
	return m
}

func (c *httpHealthChecksClient) Insert(project string, hc *compute.HttpHealthCheck) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	hcs, ok := c.httpHealthchecks[project]
	if !ok {
		hcs = map[string]*compute.HttpHealthCheck{}
		c.httpHealthchecks[project] = hcs
	}
	hc.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/httpHealthChecks/%s", project, hc.Name)
	hcs[hc.Name] = hc
	return doneOperation(), nil
}

func (c *httpHealthChecksClient) Delete(project, name string) (*compute.Operation, error) {
	c.Lock()
	defer c.Unlock()
	hcs, ok := c.httpHealthchecks[project]
	if !ok {
		return nil, notFoundError()
	}
	if _, ok := hcs[name]; !ok {
		return nil, notFoundError()
	}
	delete(hcs, name)
	return doneOperation(), nil
}

func (c *httpHealthChecksClient) Get(project, name string) (*compute.HttpHealthCheck, error) {
	c.Lock()
	defer c.Unlock()
	hcs, ok := c.httpHealthchecks[project]
	if !ok {
		return nil, notFoundError()
	}
	hc, ok := hcs[name]
	if !ok {
		return nil, notFoundError()
	}
	return hc, nil
}

func (c *httpHealthChecksClient) List(ctx context.Context, project string) ([]*compute.HttpHealthCheck, error) {
	c.Lock()
	defer c.Unlock()
	hcs, ok := c.httpHealthchecks[project]
	if !ok {
		return nil, nil
	}
	var l []*compute.HttpHealthCheck
	for _, hc := range hcs {
		l = append(l, hc)
	}
	return l, nil
}

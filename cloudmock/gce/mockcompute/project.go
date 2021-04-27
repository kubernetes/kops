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
	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
)

type projectClient struct {
	projects map[string]*compute.Project
}

var _ gce.ProjectClient = &projectClient{}

func newProjectClient(project string) *projectClient {
	return &projectClient{
		projects: map[string]*compute.Project{
			project: {
				Name: project,
			},
		},
	}
}

func (c *projectClient) All() map[string]interface{} {
	m := map[string]interface{}{}
	for n, p := range c.projects {
		m[n] = p
	}
	return m
}

func (c *projectClient) Get(project string) (*compute.Project, error) {
	p, ok := c.projects[project]
	if !ok {
		return nil, notFoundError()
	}
	return p, nil
}

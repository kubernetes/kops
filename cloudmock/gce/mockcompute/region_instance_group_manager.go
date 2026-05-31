/*
Copyright 2026 The Kubernetes Authors.

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

type regionInstanceGroupManagerClient struct {
// regionInstanceGroupManagers keyed by project, region, and name.
regionInstanceGroupManagers map[string]map[string]map[string]*compute.InstanceGroupManager
sync.Mutex
}

var _ gce.RegionInstanceGroupManagerClient = &regionInstanceGroupManagerClient{}

func newRegionInstanceGroupManagerClient() *regionInstanceGroupManagerClient {
return &regionInstanceGroupManagerClient{
regionInstanceGroupManagers: map[string]map[string]map[string]*compute.InstanceGroupManager{},
}
}

func (c *regionInstanceGroupManagerClient) All() map[string]interface{} {
c.Lock()
defer c.Unlock()
m := map[string]interface{}{}
for _, regions := range c.regionInstanceGroupManagers {
for _, igms := range regions {
for n, igm := range igms {
m[n] = igm
}
}
}
return m
}

func (c *regionInstanceGroupManagerClient) Insert(project, region string, igm *compute.InstanceGroupManager) (*compute.Operation, error) {
c.Lock()
defer c.Unlock()
igmRegions, ok := c.regionInstanceGroupManagers[project]
if !ok {
igmRegions = map[string]map[string]*compute.InstanceGroupManager{}
c.regionInstanceGroupManagers[project] = igmRegions
}
igms, ok := igmRegions[region]
if !ok {
igms = map[string]*compute.InstanceGroupManager{}
igmRegions[region] = igms
}
igm.SelfLink = fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/instanceGroupManagers/%s", project, region, igm.Name)
igms[igm.Name] = igm
return doneOperation(), nil
}

func (c *regionInstanceGroupManagerClient) Delete(project, region, name string) (*compute.Operation, error) {
c.Lock()
defer c.Unlock()
regions, ok := c.regionInstanceGroupManagers[project]
if !ok {
return nil, notFoundError()
}
igms, ok := regions[region]
if !ok {
return nil, notFoundError()
}
if _, ok := igms[name]; !ok {
return nil, notFoundError()
}
delete(igms, name)
return doneOperation(), nil
}

func (c *regionInstanceGroupManagerClient) Get(project, region, name string) (*compute.InstanceGroupManager, error) {
c.Lock()
defer c.Unlock()
regions, ok := c.regionInstanceGroupManagers[project]
if !ok {
return nil, notFoundError()
}
igms, ok := regions[region]
if !ok {
return nil, notFoundError()
}
igm, ok := igms[name]
if !ok {
return nil, notFoundError()
}
return igm, nil
}

func (c *regionInstanceGroupManagerClient) List(ctx context.Context, project, region string) ([]*compute.InstanceGroupManager, error) {
c.Lock()
defer c.Unlock()
regions, ok := c.regionInstanceGroupManagers[project]
if !ok {
return nil, nil
}
igms, ok := regions[region]
if !ok {
return nil, nil
}
var l []*compute.InstanceGroupManager
for _, d := range igms {
l = append(l, d)
}
return l, nil
}

func (c *regionInstanceGroupManagerClient) SetTargetPools(project, region, name string, targetPools []string) (*compute.Operation, error) {
return doneOperation(), nil
}

func (c *regionInstanceGroupManagerClient) SetInstanceTemplate(project, region, name, instanceTemplateURL string) (*compute.Operation, error) {
return doneOperation(), nil
}

func (c *regionInstanceGroupManagerClient) Resize(project, region, name string, newSize int64) (*compute.Operation, error) {
return doneOperation(), nil
}

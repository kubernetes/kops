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

package gcetasks

import (
"fmt"
"reflect"

compute "google.golang.org/api/compute/v1"
"k8s.io/kops/upup/pkg/fi"
"k8s.io/kops/upup/pkg/fi/cloudup/gce"
"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type RegionInstanceGroupManager struct {
Name      *string
Lifecycle fi.Lifecycle

Region                      *string
BaseInstanceName            *string
InstanceTemplate            *InstanceTemplate
ListManagedInstancesResults string
TargetSize                  *int64
UpdatePolicy                *UpdatePolicy
DistributionPolicyZones     []string

TargetPools []*TargetPool
}

var _ fi.CompareWithID = (*RegionInstanceGroupManager)(nil)

func (e *RegionInstanceGroupManager) CompareWithID() *string {
return e.Name
}

func (e *RegionInstanceGroupManager) Find(c *fi.CloudupContext) (*RegionInstanceGroupManager, error) {
cloud := c.T.Cloud.(gce.GCECloud)

r, err := cloud.Compute().RegionInstanceGroupManagers().Get(cloud.Project(), *e.Region, *e.Name)
if err != nil {
if gce.IsNotFound(err) {
return nil, nil
}
return nil, fmt.Errorf("error getting RegionInstanceGroupManager: %v", err)
}

actual := &RegionInstanceGroupManager{}
actual.Name = &r.Name
actual.Region = fi.PtrTo(lastComponent(r.Region))
actual.BaseInstanceName = &r.BaseInstanceName
actual.TargetSize = e.TargetSize
actual.InstanceTemplate = &InstanceTemplate{ID: fi.PtrTo(lastComponent(r.InstanceTemplate))}
actual.ListManagedInstancesResults = r.ListManagedInstancesResults

if policy := r.UpdatePolicy; policy != nil {
actual.UpdatePolicy = &UpdatePolicy{MinimalAction: policy.MinimalAction, Type: policy.Type}
}

if r.DistributionPolicy != nil {
for _, zone := range r.DistributionPolicy.Zones {
actual.DistributionPolicyZones = append(actual.DistributionPolicyZones, lastComponent(zone.Zone))
}
}

for _, targetPool := range r.TargetPools {
actual.TargetPools = append(actual.TargetPools, &TargetPool{
Name: fi.PtrTo(lastComponent(targetPool)),
})
}

// Ignore "system" fields
actual.Lifecycle = e.Lifecycle

return actual, nil
}

func (e *RegionInstanceGroupManager) Run(c *fi.CloudupContext) error {
return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *RegionInstanceGroupManager) CheckChanges(a, e, changes *RegionInstanceGroupManager) error {
return nil
}

func (_ *RegionInstanceGroupManager) RenderGCE(t *gce.GCEAPITarget, a, e, changes *RegionInstanceGroupManager) error {
project := t.Cloud.Project()

instanceTemplateURL, err := e.InstanceTemplate.URL(project)
if err != nil {
return err
}

i := &compute.InstanceGroupManager{
Name:                        *e.Name,
Region:                      *e.Region,
BaseInstanceName:            *e.BaseInstanceName,
TargetSize:                  *e.TargetSize,
InstanceTemplate:            instanceTemplateURL,
ListManagedInstancesResults: e.ListManagedInstancesResults,
}

if len(e.DistributionPolicyZones) > 0 {
dp := &compute.DistributionPolicy{}
for _, zone := range e.DistributionPolicyZones {
dp.Zones = append(dp.Zones, &compute.DistributionPolicyZoneConfiguration{
Zone: fmt.Sprintf("zones/%s", zone),
})
}
i.DistributionPolicy = dp
}

if policy := e.UpdatePolicy; policy != nil {
i.UpdatePolicy = &compute.InstanceGroupManagerUpdatePolicy{
MinimalAction: policy.MinimalAction,
Type:          policy.Type,
}
}

for _, targetPool := range e.TargetPools {
i.TargetPools = append(i.TargetPools, targetPool.URL(t.Cloud))
}

if a == nil {
op, err := t.Cloud.Compute().RegionInstanceGroupManagers().Insert(project, *e.Region, i)
if err != nil {
return fmt.Errorf("error creating RegionInstanceGroupManager: %v", err)
}
if err := t.Cloud.WaitForOp(op); err != nil {
return fmt.Errorf("error creating RegionInstanceGroupManager: %v", err)
}
} else {
if changes.TargetPools != nil {
op, err := t.Cloud.Compute().RegionInstanceGroupManagers().SetTargetPools(project, *e.Region, i.Name, i.TargetPools)
if err != nil {
return fmt.Errorf("error updating TargetPools for RegionInstanceGroupManager: %v", err)
}
if err := t.Cloud.WaitForOp(op); err != nil {
return fmt.Errorf("error updating TargetPools for RegionInstanceGroupManager: %v", err)
}
changes.TargetPools = nil
}

if changes.InstanceTemplate != nil {
op, err := t.Cloud.Compute().RegionInstanceGroupManagers().SetInstanceTemplate(project, *e.Region, i.Name, instanceTemplateURL)
if err != nil {
return fmt.Errorf("error updating InstanceTemplate for RegionInstanceGroupManager: %v", err)
}
if err := t.Cloud.WaitForOp(op); err != nil {
return fmt.Errorf("error updating InstanceTemplate for RegionInstanceGroupManager: %v", err)
}
changes.InstanceTemplate = nil
}

if changes.TargetSize != nil {
newSize := int64(0)
if i.TargetSize != 0 {
newSize = int64(i.TargetSize)
}
op, err := t.Cloud.Compute().RegionInstanceGroupManagers().Resize(project, *e.Region, i.Name, newSize)
if err != nil {
return fmt.Errorf("error resizing RegionInstanceGroupManager: %v", err)
}
if err := t.Cloud.WaitForOp(op); err != nil {
return fmt.Errorf("error resizing RegionInstanceGroupManager: %v", err)
}
changes.TargetSize = nil
}

empty := &RegionInstanceGroupManager{}
if !reflect.DeepEqual(empty, changes) {
return fmt.Errorf("cannot apply changes to RegionInstanceGroupManager: %v", changes)
}
}

return nil
}

type terraformRegionInstanceGroupManager struct {
Lifecycle                   *terraform.Lifecycle       `cty:"lifecycle"`
Name                        *string                    `cty:"name"`
Region                      *string                    `cty:"region"`
BaseInstanceName            *string                    `cty:"base_instance_name"`
ListManagedInstancesResults string                     `cty:"list_managed_instances_results"`
Version                     *terraformVersion          `cty:"version"`
TargetSize                  *int64                     `cty:"target_size"`
UpdatePolicy                *terraformUpdatePolicy     `cty:"update_policy"`
TargetPools                 []*terraformWriter.Literal `cty:"target_pools"`
DistributionPolicyZones     []string                   `cty:"distribution_policy_zones"`
}

func (_ *RegionInstanceGroupManager) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *RegionInstanceGroupManager) error {
tf := &terraformRegionInstanceGroupManager{
Name:                        e.Name,
Region:                      e.Region,
BaseInstanceName:            e.BaseInstanceName,
TargetSize:                  e.TargetSize,
ListManagedInstancesResults: e.ListManagedInstancesResults,
DistributionPolicyZones:     e.DistributionPolicyZones,
}
tf.Lifecycle = &terraform.Lifecycle{
IgnoreChanges: []*terraformWriter.Literal{{String: "target_size"}},
}
if policy := e.UpdatePolicy; policy != nil {
tf.UpdatePolicy = &terraformUpdatePolicy{
MinimalAction: policy.MinimalAction,
Type:          policy.Type,
}
}
tf.Version = &terraformVersion{
InstanceTemplate: e.InstanceTemplate.TerraformLink(),
}

for _, targetPool := range e.TargetPools {
tf.TargetPools = append(tf.TargetPools, targetPool.TerraformLink())
}

return t.RenderResource("google_compute_region_instance_group_manager", *e.Name, tf)
}

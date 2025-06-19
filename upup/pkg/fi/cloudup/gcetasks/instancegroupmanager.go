/*
Copyright 2019 The Kubernetes Authors.

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
type InstanceGroupManager struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Zone                        *string
	BaseInstanceName            *string
	InstanceTemplate            *InstanceTemplate
	ListManagedInstancesResults string
	TargetSize                  *int64

	TargetPools []*TargetPool
}

var _ fi.CompareWithID = &InstanceGroupManager{}

func (e *InstanceGroupManager) CompareWithID() *string {
	return e.Name
}

func (e *InstanceGroupManager) Find(c *fi.CloudupContext) (*InstanceGroupManager, error) {
	cloud := c.T.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().InstanceGroupManagers().Get(cloud.Project(), *e.Zone, *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
	}

	actual := &InstanceGroupManager{}
	actual.Name = &r.Name
	actual.Zone = fi.PtrTo(lastComponent(r.Zone))
	actual.BaseInstanceName = &r.BaseInstanceName
	actual.TargetSize = e.TargetSize
	actual.InstanceTemplate = &InstanceTemplate{ID: fi.PtrTo(lastComponent(r.InstanceTemplate))}
	actual.ListManagedInstancesResults = r.ListManagedInstancesResults

	for _, targetPool := range r.TargetPools {
		actual.TargetPools = append(actual.TargetPools, &TargetPool{
			Name: fi.PtrTo(lastComponent(targetPool)),
		})
	}
	// TODO: Sort by name

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *InstanceGroupManager) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *InstanceGroupManager) CheckChanges(a, e, changes *InstanceGroupManager) error {
	return nil
}

func (_ *InstanceGroupManager) RenderGCE(t *gce.GCEAPITarget, a, e, changes *InstanceGroupManager) error {
	project := t.Cloud.Project()

	instanceTemplateURL, err := e.InstanceTemplate.URL(project)
	if err != nil {
		return err
	}

	i := &compute.InstanceGroupManager{
		Name:                        *e.Name,
		Zone:                        *e.Zone,
		BaseInstanceName:            *e.BaseInstanceName,
		TargetSize:                  *e.TargetSize,
		InstanceTemplate:            instanceTemplateURL,
		ListManagedInstancesResults: e.ListManagedInstancesResults,
	}

	for _, targetPool := range e.TargetPools {
		i.TargetPools = append(i.TargetPools, targetPool.URL(t.Cloud))
	}

	if a == nil {
		if i.TargetSize == 0 {
			// TargetSize 0 will normally be omitted by the marshaling code; we need to force it
			i.ForceSendFields = append(i.ForceSendFields, "TargetSize")
		}
		op, err := t.Cloud.Compute().InstanceGroupManagers().Insert(t.Cloud.Project(), *e.Zone, i)
		if err != nil {
			return fmt.Errorf("error creating InstanceGroupManager: %v", err)
		}

		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error creating InstanceGroupManager: %v", err)
		}
	} else {
		if changes.TargetPools != nil {
			op, err := t.Cloud.Compute().InstanceGroupManagers().SetTargetPools(t.Cloud.Project(), *e.Zone, i.Name, i.TargetPools)
			if err != nil {
				return fmt.Errorf("error updating TargetPools for InstanceGroupManager: %v", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("error updating TargetPools for InstanceGroupManager: %v", err)
			}

			changes.TargetPools = nil
		}

		if changes.InstanceTemplate != nil {
			op, err := t.Cloud.Compute().InstanceGroupManagers().SetInstanceTemplate(t.Cloud.Project(), *e.Zone, i.Name, instanceTemplateURL)
			if err != nil {
				return fmt.Errorf("error updating InstanceTemplate for InstanceGroupManager: %v", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("error updating InstanceTemplate for InstanceGroupManager: %v", err)
			}

			changes.InstanceTemplate = nil
		}

		if changes.TargetSize != nil {
			newSize := int64(0)
			if i.TargetSize != 0 {
				newSize = int64(i.TargetSize)
			}
			op, err := t.Cloud.Compute().InstanceGroupManagers().Resize(t.Cloud.Project(), *e.Zone, i.Name, newSize)
			if err != nil {
				return fmt.Errorf("error resizing InstanceGroupManager: %v", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("error resizing InstanceGroupManager: %v", err)
			}

			changes.TargetSize = nil
		}

		empty := &InstanceGroupManager{}
		if !reflect.DeepEqual(empty, changes) {
			return fmt.Errorf("cannot apply changes to InstanceGroupManager: %v", changes)
		}
	}

	return nil
}

type terraformInstanceGroupManager struct {
	Lifecycle                   *terraform.Lifecycle       `cty:"lifecycle"`
	Name                        *string                    `cty:"name"`
	Zone                        *string                    `cty:"zone"`
	BaseInstanceName            *string                    `cty:"base_instance_name"`
	ListManagedInstancesResults string                     `cty:"list_managed_instances_results"`
	Version                     *terraformVersion          `cty:"version"`
	TargetSize                  *int64                     `cty:"target_size"`
	TargetPools                 []*terraformWriter.Literal `cty:"target_pools"`
}

type terraformVersion struct {
	InstanceTemplate *terraformWriter.Literal `cty:"instance_template"`
}

func (_ *InstanceGroupManager) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InstanceGroupManager) error {
	tf := &terraformInstanceGroupManager{
		Name:                        e.Name,
		Zone:                        e.Zone,
		BaseInstanceName:            e.BaseInstanceName,
		TargetSize:                  e.TargetSize,
		ListManagedInstancesResults: e.ListManagedInstancesResults,
	}
	tf.Lifecycle = &terraform.Lifecycle{
		IgnoreChanges: []*terraformWriter.Literal{{String: "target_size"}},
	}
	tf.Version = &terraformVersion{
		InstanceTemplate: e.InstanceTemplate.TerraformLink(),
	}

	for _, targetPool := range e.TargetPools {
		tf.TargetPools = append(tf.TargetPools, targetPool.TerraformLink())
	}

	return t.RenderResource("google_compute_instance_group_manager", *e.Name, tf)
}

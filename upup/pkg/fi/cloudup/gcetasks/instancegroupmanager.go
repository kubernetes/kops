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
)

//go:generate fitask -type=InstanceGroupManager
type InstanceGroupManager struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Zone             *string
	BaseInstanceName *string
	InstanceTemplate *InstanceTemplate
	TargetSize       *int64

	TargetPools []*TargetPool
}

var _ fi.CompareWithID = &InstanceGroupManager{}

func (e *InstanceGroupManager) CompareWithID() *string {
	return e.Name
}

func (e *InstanceGroupManager) Find(c *fi.Context) (*InstanceGroupManager, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().InstanceGroupManagers.Get(cloud.Project(), *e.Zone, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing InstanceGroupManagers: %v", err)
	}

	actual := &InstanceGroupManager{}
	actual.Name = &r.Name
	actual.Zone = fi.String(lastComponent(r.Zone))
	actual.BaseInstanceName = &r.BaseInstanceName
	actual.TargetSize = &r.TargetSize
	actual.InstanceTemplate = &InstanceTemplate{ID: fi.String(lastComponent(r.InstanceTemplate))}

	for _, targetPool := range r.TargetPools {
		actual.TargetPools = append(actual.TargetPools, &TargetPool{
			Name: fi.String(lastComponent(targetPool)),
		})
	}
	// TODO: Sort by name

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *InstanceGroupManager) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
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
		Name:             *e.Name,
		Zone:             *e.Zone,
		BaseInstanceName: *e.BaseInstanceName,
		TargetSize:       *e.TargetSize,
		InstanceTemplate: instanceTemplateURL,
	}

	for _, targetPool := range e.TargetPools {
		i.TargetPools = append(i.TargetPools, targetPool.URL(t.Cloud))
	}

	if a == nil {
		if i.TargetSize == 0 {
			// TargetSize 0 will normally be omitted by the marshaling code; we need to force it
			i.ForceSendFields = append(i.ForceSendFields, "TargetSize")
		}
		op, err := t.Cloud.Compute().InstanceGroupManagers.Insert(t.Cloud.Project(), *e.Zone, i).Do()
		if err != nil {
			return fmt.Errorf("error creating InstanceGroupManager: %v", err)
		}

		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error creating InstanceGroupManager: %v", err)
		}
	} else {
		if changes.TargetPools != nil {
			request := &compute.InstanceGroupManagersSetTargetPoolsRequest{
				TargetPools: i.TargetPools,
			}
			op, err := t.Cloud.Compute().InstanceGroupManagers.SetTargetPools(t.Cloud.Project(), *e.Zone, i.Name, request).Do()
			if err != nil {
				return fmt.Errorf("error updating TargetPools for InstanceGroupManager: %v", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("error updating TargetPools for InstanceGroupManager: %v", err)
			}

			changes.TargetPools = nil
		}

		if changes.InstanceTemplate != nil {
			request := &compute.InstanceGroupManagersSetInstanceTemplateRequest{
				InstanceTemplate: instanceTemplateURL,
			}
			op, err := t.Cloud.Compute().InstanceGroupManagers.SetInstanceTemplate(t.Cloud.Project(), *e.Zone, i.Name, request).Do()
			if err != nil {
				return fmt.Errorf("error updating InstanceTemplate for InstanceGroupManager: %v", err)
			}

			if err := t.Cloud.WaitForOp(op); err != nil {
				return fmt.Errorf("error updating InstanceTemplate for InstanceGroupManager: %v", err)
			}

			changes.InstanceTemplate = nil
		}

		if changes.TargetSize != nil {

			req := t.Cloud.Compute().InstanceGroupManagers.Resize(t.Cloud.Project(), *e.Zone, i.Name, i.TargetSize)
			resp, err := req.Do()
			if err != nil {
				return fmt.Errorf("error resizing ManagedInstances in %s: %v", i.Name, err)
			}
			if err := t.Cloud.WaitForOp(resp); err != nil {
				return fmt.Errorf("error resizing ManagedInstances in %s: %v", i.Name, err)
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
	Name             *string              `json:"name"`
	Zone             *string              `json:"zone"`
	BaseInstanceName *string              `json:"base_instance_name"`
	Version          *terraformVersion    `json:"version"`
	TargetSize       *int64               `json:"target_size"`
	TargetPools      []*terraform.Literal `json:"target_pools,omitempty"`
}

type terraformVersion struct {
	InstanceTemplate *terraform.Literal `json:"instance_template"`
}

func (_ *InstanceGroupManager) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InstanceGroupManager) error {
	tf := &terraformInstanceGroupManager{
		Name:             e.Name,
		Zone:             e.Zone,
		BaseInstanceName: e.BaseInstanceName,
		TargetSize:       e.TargetSize,
	}
	tf.Version = &terraformVersion{
		InstanceTemplate: e.InstanceTemplate.TerraformLink(),
	}

	for _, targetPool := range e.TargetPools {
		tf.TargetPools = append(tf.TargetPools, targetPool.TerraformLink())
	}

	return t.RenderResource("google_compute_instance_group_manager", *e.Name, tf)
}

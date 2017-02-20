/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/golang/glog"
	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"time"
)

//go:generate fitask -type=ManagedInstanceGroup
type ManagedInstanceGroup struct {
	Name *string

	Zone             *string
	BaseInstanceName *string
	InstanceTemplate *InstanceTemplate
	TargetSize       *int64
}

var _ fi.CompareWithID = &ManagedInstanceGroup{}

func (e *ManagedInstanceGroup) CompareWithID() *string {
	return e.Name
}

func (e *ManagedInstanceGroup) Find(c *fi.Context) (*ManagedInstanceGroup, error) {
	cloud := c.Cloud.(*gce.GCECloud)

	r, err := cloud.Compute.InstanceGroupManagers.Get(cloud.Project, *e.Zone, *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing ManagedInstanceGroups: %v", err)
	}

	actual := &ManagedInstanceGroup{}
	actual.Name = &r.Name
	actual.Zone = fi.String(lastComponent(r.Zone))
	actual.BaseInstanceName = &r.BaseInstanceName
	actual.TargetSize = &r.TargetSize
	actual.InstanceTemplate = &InstanceTemplate{Name: fi.String(lastComponent(r.InstanceTemplate))}

	return actual, nil
}

func (e *ManagedInstanceGroup) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *ManagedInstanceGroup) CheckChanges(a, e, changes *ManagedInstanceGroup) error {
	return nil
}

func BuildInstanceTemplateURL(project, name string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/instanceTemplates/%s", project, name)
}

func (_ *ManagedInstanceGroup) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ManagedInstanceGroup) error {
	project := t.Cloud.Project

	i := &compute.InstanceGroupManager{
		Name:             *e.Name,
		Zone:             *e.Zone,
		BaseInstanceName: *e.BaseInstanceName,
		TargetSize:       *e.TargetSize,
		InstanceTemplate: BuildInstanceTemplateURL(project, *e.InstanceTemplate.Name),
	}

	if a == nil {
		for {
			_, err := t.Cloud.Compute.InstanceGroupManagers.Insert(t.Cloud.Project, *e.Zone, i).Do()
			if err != nil {
				if gce.IsNotReady(err) {
					glog.Infof("Found resourceNotReady error - sleeping before retry: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}
				return fmt.Errorf("error creating ManagedInstanceGroup: %v", err)
			} else {
				break
			}
		}
	} else {
		return fmt.Errorf("Cannot apply changes to ManagedInstanceGroup: %v", changes)
	}

	return nil
}

type terraformInstanceGroupManager struct {
	Name             *string            `json:"name"`
	Zone             *string            `json:"zone"`
	BaseInstanceName *string            `json:"base_instance_name"`
	InstanceTemplate *terraform.Literal `json:"instance_template"`
	TargetSize       *int64             `json:"target_size"`
}

func (_ *ManagedInstanceGroup) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ManagedInstanceGroup) error {
	tf := &terraformInstanceGroupManager{
		Name:             e.Name,
		Zone:             e.Zone,
		BaseInstanceName: e.BaseInstanceName,
		InstanceTemplate: e.InstanceTemplate.TerraformLink(),
		TargetSize:       e.TargetSize,
	}

	return t.RenderResource("google_compute_instance_group_manager", *e.Name, tf)
}

/*
Copyright 2017 The Kubernetes Authors.

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

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// TargetPool represents a GCE TargetPool
//go:generate fitask -type=TargetPool
type TargetPool struct {
	Name      *string
	Lifecycle *fi.Lifecycle
}

var _ fi.CompareWithID = &TargetPool{}

func (e *TargetPool) CompareWithID() *string {
	return e.Name
}

func (e *TargetPool) Find(c *fi.Context) (*TargetPool, error) {
	cloud := c.Cloud.(gce.GCECloud)
	name := fi.StringValue(e.Name)

	r, err := cloud.Compute().TargetPools.Get(cloud.Project(), cloud.Region(), name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting TargetPool %q: %v", name, err)
	}

	actual := &TargetPool{}
	actual.Name = fi.String(r.Name)

	return actual, nil
}

func (e *TargetPool) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *TargetPool) CheckChanges(a, e, changes *TargetPool) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (e *TargetPool) URL(cloud gce.GCECloud) string {
	name := fi.StringValue(e.Name)

	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/targetPools/%s", cloud.Project(), cloud.Region(), name)
}

func (_ *TargetPool) RenderGCE(t *gce.GCEAPITarget, a, e, changes *TargetPool) error {
	name := fi.StringValue(e.Name)

	o := &compute.TargetPool{
		Name: name,
	}

	if a == nil {
		klog.V(4).Infof("Creating TargetPool %q", o.Name)

		op, err := t.Cloud.Compute().TargetPools.Insert(t.Cloud.Project(), t.Cloud.Region(), o).Do()
		if err != nil {
			return fmt.Errorf("error creating TargetPool %q: %v", name, err)
		}

		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error creating TargetPool: %v", err)
		}
	} else {
		return fmt.Errorf("cannot apply changes to TargetPool: %v", changes)
	}

	return nil
}

type terraformTargetPool struct {
	Name            string   `json:"name" cty:"name"`
	Description     string   `json:"description,omitempty" cty:"description"`
	HealthChecks    []string `json:"health_checks,omitempty" cty:"health_checks"`
	Instances       []string `json:"instances,omitempty" cty:"instances"`
	SessionAffinity string   `json:"session_affinity,omitempty" cty:"session_affinity"`
}

func (_ *TargetPool) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *TargetPool) error {
	name := fi.StringValue(e.Name)

	tf := &terraformTargetPool{
		Name: name,
	}

	return t.RenderResource("google_compute_target_pool", name, tf)
}

func (e *TargetPool) TerraformLink() *terraform.Literal {
	name := fi.StringValue(e.Name)

	return terraform.LiteralSelfLink("google_compute_target_pool", name)
}

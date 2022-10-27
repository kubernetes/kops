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

	compute "google.golang.org/api/compute/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// TargetPool represents a GCE TargetPool
// +kops:fitask
type TargetPool struct {
	Name      *string
	Lifecycle fi.Lifecycle
}

var _ fi.CompareWithID = &TargetPool{}

func (e *TargetPool) CompareWithID() *string {
	return e.Name
}

func (e *TargetPool) Find(c *fi.CloudContext) (*TargetPool, error) {
	cloud := c.Cloud.(gce.GCECloud)
	name := fi.StringValue(e.Name)

	r, err := cloud.Compute().TargetPools().Get(cloud.Project(), cloud.Region(), name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting TargetPool %q: %v", name, err)
	}

	actual := &TargetPool{}
	actual.Name = fi.String(r.Name)

	// Avoid spurious changes
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *TargetPool) Run(c *fi.CloudContext) error {
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

		op, err := t.Cloud.Compute().TargetPools().Insert(t.Cloud.Project(), t.Cloud.Region(), o)
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
	Name            string   `cty:"name"`
	Description     string   `cty:"description"`
	HealthChecks    []string `cty:"health_checks"`
	Instances       []string `cty:"instances"`
	SessionAffinity string   `cty:"session_affinity"`
}

func (_ *TargetPool) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *TargetPool) error {
	name := fi.StringValue(e.Name)

	tf := &terraformTargetPool{
		Name: name,
	}

	return t.RenderResource("google_compute_target_pool", name, tf)
}

func (e *TargetPool) TerraformLink() *terraformWriter.Literal {
	name := fi.StringValue(e.Name)

	return terraformWriter.LiteralSelfLink("google_compute_target_pool", name)
}

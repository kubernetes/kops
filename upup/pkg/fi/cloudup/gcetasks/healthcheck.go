/*
Copyright 2022 The Kubernetes Authors.

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
)

// Healthcheck represents a GCE Healthcheck
// +kops:fitask
type Healthcheck struct {
	Name      *string
	Lifecycle fi.Lifecycle

	SelfLink string
	Port     *int64
}

var _ fi.CompareWithID = &Healthcheck{}

func (e *Healthcheck) CompareWithID() *string {
	return e.Name
}

func (e *Healthcheck) Find(c *fi.Context) (*Healthcheck, error) {
	cloud := c.Cloud.(gce.GCECloud)
	name := fi.StringValue(e.Name)
	r, err := cloud.Compute().HTTPHealthChecks().Get(cloud.Project(), name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting HealthCheck %q: %v", name, err)
	}
	actual := &Healthcheck{
		Name:     fi.String(r.Name),
		Port:     fi.Int64(r.Port),
		SelfLink: r.SelfLink,
	}
	// System fields
	actual.Lifecycle = e.Lifecycle
	e.SelfLink = r.SelfLink
	return actual, nil
}

func (e *Healthcheck) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *Healthcheck) CheckChanges(a, e, changes *Healthcheck) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (h *Healthcheck) RenderGCE(t *gce.GCEAPITarget, a, e, changes *Healthcheck) error {
	if a == nil {
		o := &compute.HttpHealthCheck{
			Name:        fi.StringValue(e.Name),
			Port:        fi.Int64Value(e.Port),
			RequestPath: "/healthz",
		}

		klog.V(4).Infof("Creating Healthcheck %q", o.Name)
		r, err := t.Cloud.Compute().HTTPHealthChecks().Insert(t.Cloud.Project(), o)
		if err != nil {
			return fmt.Errorf("error creating Healthcheck %q: %v", o.Name, err)
		}
		if err := t.Cloud.WaitForOp(r); err != nil {
			return fmt.Errorf("error creating Healthcheck: %v", err)
		}
		h.SelfLink = r.TargetLink
	}
	return nil
}

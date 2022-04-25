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
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
// HealthCheck represents a GCE "healthcheck" type - this is the
// non-deprecated new-style HC, which combines the deprecated HTTPHealthCheck
// and HTTPSHealthCheck.  Those HCs are still needed for some types, so both
// are implemented in kops, but this one should be preferred when possible.
type HealthCheck struct {
	Name      *string
	Port      int64
	Lifecycle fi.Lifecycle
}

var _ fi.CompareWithID = &HealthCheck{}

func (e *HealthCheck) CompareWithID() *string {
	return e.Name
}

func (e *HealthCheck) Find(c *fi.Context) (*HealthCheck, error) {
	actual, err := e.find(c.Cloud.(gce.GCECloud))
	if actual != nil && err == nil {
		// Ignore system fields
		actual.Lifecycle = e.Lifecycle
	}
	return actual, err
}

func (e *HealthCheck) URL(cloud gce.GCECloud) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/healthChecks/%s",
		cloud.Project(),
		cloud.Region(),
		*e.Name)
}

func (e *HealthCheck) find(cloud gce.GCECloud) (*HealthCheck, error) {
	r, err := cloud.Compute().RegionHealthChecks().Get(cloud.Project(), cloud.Region(), *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing Health Checks: %v", err)
	}

	actual := &HealthCheck{}
	actual.Name = &r.Name
	if r.TcpHealthCheck != nil {
		actual.Port = r.TcpHealthCheck.Port
	}

	return actual, nil
}

func (e *HealthCheck) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *HealthCheck) CheckChanges(a, e, changes *HealthCheck) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if e.Port != a.Port {
			return fi.CannotChangeField("Port")
		}
	}
	return nil
}

func (_ *HealthCheck) RenderGCE(t *gce.GCEAPITarget, a, e, changes *HealthCheck) error {
	cloud := t.Cloud
	hc := &compute.HealthCheck{
		Name: *e.Name,
		TcpHealthCheck: &compute.TCPHealthCheck{
			Port: e.Port,
		},
		Type: "TCP",

		Region: cloud.Region(),
	}

	if a == nil {
		klog.Infof("GCE creating healthcheck: %q", hc.Name)

		op, err := cloud.Compute().RegionHealthChecks().Insert(cloud.Project(), cloud.Region(), hc)
		if err != nil {
			return fmt.Errorf("error creating healthcheck: %v", err)
		}

		if err := cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for healthcheck: %v", err)
		}
	} else {
		return fmt.Errorf("cannot apply changes to healthcheck: %v", changes)
	}

	return nil
}

type terraformTCPBlock struct {
	Port int64 `cty:"port"`
}

type terraformHealthCheck struct {
	Name           string            `cty:"name"`
	TCPHealthCheck terraformTCPBlock `cty:"tcp_health_check"`
}

func (_ *HealthCheck) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *HealthCheck) error {
	tf := &terraformHealthCheck{
		Name: *e.Name,
		TCPHealthCheck: terraformTCPBlock{
			Port: e.Port,
		},
	}
	return t.RenderResource("google_compute_health_check", *e.Name, tf)
}

func (e *HealthCheck) TerraformAddress() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("google_compute_health_check", *e.Name, "id")
}

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
	Name        *string
	Port        *int64
	Lifecycle   fi.Lifecycle
	Region      string
	RequestPath *string
}

var _ fi.CompareWithID = &HealthCheck{}

func (e *HealthCheck) CompareWithID() *string {
	return e.Name
}

func (e *HealthCheck) Find(c *fi.CloudupContext) (*HealthCheck, error) {
	actual, err := e.find(c.T.Cloud.(gce.GCECloud))
	if actual != nil && err == nil {
		// Ignore system fields
		actual.Lifecycle = e.Lifecycle
	}
	return actual, err
}

func (e *HealthCheck) URL(cloud gce.GCECloud, region string) string {
	if region == "" {
		return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/healthChecks/%s",
			cloud.Project(),
			*e.Name)
	}

	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/regions/%s/healthChecks/%s",
		cloud.Project(),
		cloud.Region(),
		*e.Name)
}

func (e *HealthCheck) find(cloud gce.GCECloud) (*HealthCheck, error) {
	r, err := cloud.Compute().HealthChecks().Get(cloud.Project(), e.Region, *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}

		return nil, fmt.Errorf("error listing Health Checks: %v", err)
	}

	actual := &HealthCheck{}
	actual.Name = &r.Name
	if r.HttpHealthCheck != nil {
		actual.Port = fi.PtrTo(r.TcpHealthCheck.Port)
	}

	return actual, nil
}

func (e *HealthCheck) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (*HealthCheck) CheckChanges(a, e, changes *HealthCheck) error {
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

func (*HealthCheck) RenderGCE(t *gce.GCEAPITarget, a, e, changes *HealthCheck) error {
	cloud := t.Cloud
	hc := &compute.HealthCheck{
		Name: *e.Name,
		HttpHealthCheck: &compute.HTTPHealthCheck{
			Port:        fi.ValueOf(e.Port),
			RequestPath: fi.ValueOf(e.RequestPath),
		},
		Type: "HTTP",
	}

	if a == nil {
		klog.V(2).Infof("Creating HealthCheck %q", hc.Name)

		op, err := cloud.Compute().HealthChecks().Insert(cloud.Project(), e.Region, hc)
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

type terraformHTTPBlock struct {
	Port        int64  `cty:"port"`
	RequestPath string `cty:"request_path"`
}

type terraformHealthCheck struct {
	Name            string             `cty:"name"`
	Region          string             `cty:"region"`
	HTTPHealthCheck terraformHTTPBlock `cty:"http_health_check"`
}

func (*HealthCheck) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *HealthCheck) error {
	tf := &terraformHealthCheck{
		Name: *e.Name,
		HTTPHealthCheck: terraformHTTPBlock{
			Port:        fi.ValueOf(e.Port),
			RequestPath: fi.ValueOf(e.RequestPath),
		},
	}
	if e.Region == "" {
		return t.RenderResource("google_compute_health_check", *e.Name, tf)
	}
	tf.Region = e.Region
	return t.RenderResource("google_compute_region_health_check", *e.Name, tf)
}

func (e *HealthCheck) TerraformAddress() *terraformWriter.Literal {
	if e.Region == "" {
		return terraformWriter.LiteralProperty("google_compute_region_health_check", *e.Name, "id")
	}
	return terraformWriter.LiteralProperty("google_compute_health_check", *e.Name, "id")
}

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

// HealthCheckProtocol is the protocol used for a health check.
type HealthCheckProtocol string

const (
	HealthCheckProtocolTCP   HealthCheckProtocol = "TCP"
	HealthCheckProtocolSSL   HealthCheckProtocol = "SSL"
	HealthCheckProtocolHTTPS HealthCheckProtocol = "HTTPS"
)

// +kops:fitask
// HealthCheck represents a GCE "healthcheck" type - this is the
// non-deprecated new-style HC, which combines the deprecated HTTPHealthCheck
// and HTTPSHealthCheck.  Those HCs are still needed for some types, so both
// are implemented in kops, but this one should be preferred when possible.
type HealthCheck struct {
	Name     *string
	Port     int64
	Protocol HealthCheckProtocol
	// RequestPath is the path requested by HTTPS health checks.
	RequestPath *string
	Lifecycle   fi.Lifecycle
}

var _ fi.CompareWithID = (*HealthCheck)(nil)

func (e *HealthCheck) CompareWithID() *string {
	return e.Name
}

// protocol returns the effective protocol, defaulting to TCP.
func (e *HealthCheck) protocol() HealthCheckProtocol {
	if e.Protocol == "" {
		return HealthCheckProtocolTCP
	}
	return e.Protocol
}

func (e *HealthCheck) Find(c *fi.CloudupContext) (*HealthCheck, error) {
	actual, err := e.find(c.T.Cloud.(gce.GCECloud))
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
	switch r.Type {
	case "SSL":
		actual.Protocol = HealthCheckProtocolSSL
		if r.SslHealthCheck != nil {
			actual.Port = r.SslHealthCheck.Port
		}
	case "HTTPS":
		actual.Protocol = HealthCheckProtocolHTTPS
		if r.HttpsHealthCheck != nil {
			actual.Port = r.HttpsHealthCheck.Port
			actual.RequestPath = &r.HttpsHealthCheck.RequestPath
		}
	default:
		actual.Protocol = HealthCheckProtocolTCP
		if r.TcpHealthCheck != nil {
			actual.Port = r.TcpHealthCheck.Port
		}
	}

	return actual, nil
}

func (e *HealthCheck) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (_ *HealthCheck) CheckChanges(a, e, changes *HealthCheck) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if e.Port != a.Port {
			return fi.CannotChangeField("Port")
		}
		if e.protocol() != a.protocol() {
			// GCE does not allow changing the type of an existing health check.
			return fi.CannotChangeField("Protocol")
		}
	}
	return nil
}

func (_ *HealthCheck) RenderGCE(t *gce.GCEAPITarget, a, e, changes *HealthCheck) error {
	cloud := t.Cloud
	hc := &compute.HealthCheck{
		Name:   *e.Name,
		Region: cloud.Region(),
	}

	switch e.protocol() {
	case HealthCheckProtocolSSL:
		hc.Type = "SSL"
		hc.SslHealthCheck = &compute.SSLHealthCheck{
			Port: e.Port,
		}
	case HealthCheckProtocolHTTPS:
		hc.Type = "HTTPS"
		hc.HttpsHealthCheck = &compute.HTTPSHealthCheck{
			Port:        e.Port,
			RequestPath: fi.ValueOf(e.RequestPath),
		}
	default:
		hc.Type = "TCP"
		hc.TcpHealthCheck = &compute.TCPHealthCheck{
			Port: e.Port,
		}
	}

	if a == nil {
		klog.V(2).Infof("Creating HealthCheck %q", hc.Name)

		op, err := cloud.Compute().RegionHealthChecks().Insert(cloud.Project(), cloud.Region(), hc)
		if err != nil {
			return fmt.Errorf("error creating healthcheck: %v", err)
		}

		if err := cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for healthcheck: %v", err)
		}
	} else if changes.RequestPath != nil {
		klog.V(2).Infof("Updating HealthCheck %q", hc.Name)

		op, err := cloud.Compute().RegionHealthChecks().Update(cloud.Project(), cloud.Region(), hc.Name, hc)
		if err != nil {
			return fmt.Errorf("error updating healthcheck: %v", err)
		}

		if err := cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error waiting for healthcheck: %v", err)
		}
	} else {
		return fmt.Errorf("cannot apply changes to healthcheck: %v", changes)
	}

	return nil
}

type terraformHealthCheckBlock struct {
	Port int64 `cty:"port"`
}

type terraformHTTPSHealthCheckBlock struct {
	Port        int64   `cty:"port"`
	RequestPath *string `cty:"request_path"`
}

type terraformHealthCheck struct {
	Name             string                          `cty:"name"`
	TCPHealthCheck   *terraformHealthCheckBlock      `cty:"tcp_health_check"`
	SSLHealthCheck   *terraformHealthCheckBlock      `cty:"ssl_health_check"`
	HTTPSHealthCheck *terraformHTTPSHealthCheckBlock `cty:"https_health_check"`
}

func (_ *HealthCheck) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *HealthCheck) error {
	tf := &terraformHealthCheck{
		Name: *e.Name,
	}

	switch e.protocol() {
	case HealthCheckProtocolSSL:
		tf.SSLHealthCheck = &terraformHealthCheckBlock{Port: e.Port}
	case HealthCheckProtocolHTTPS:
		tf.HTTPSHealthCheck = &terraformHTTPSHealthCheckBlock{Port: e.Port, RequestPath: e.RequestPath}
	default:
		tf.TCPHealthCheck = &terraformHealthCheckBlock{Port: e.Port}
	}

	return t.RenderResource("google_compute_region_health_check", *e.Name, tf)
}

func (e *HealthCheck) TerraformAddress() *terraformWriter.Literal {
	return terraformWriter.LiteralProperty("google_compute_region_health_check", *e.Name, "id")
}

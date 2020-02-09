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
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// ForwardingRule represents a GCE ForwardingRule
//go:generate fitask -type=ForwardingRule
type ForwardingRule struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	PortRange  string
	TargetPool *TargetPool
	IPAddress  *Address
	IPProtocol string
}

var _ fi.CompareWithID = &ForwardingRule{}

func (e *ForwardingRule) CompareWithID() *string {
	return e.Name
}

func (e *ForwardingRule) Find(c *fi.Context) (*ForwardingRule, error) {
	cloud := c.Cloud.(gce.GCECloud)
	name := fi.StringValue(e.Name)

	r, err := cloud.Compute().ForwardingRules.Get(cloud.Project(), cloud.Region(), name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
	}

	actual := &ForwardingRule{
		Name:       fi.String(r.Name),
		PortRange:  r.PortRange,
		IPProtocol: r.IPProtocol,
	}
	if r.Target != "" {
		actual.TargetPool = &TargetPool{
			Name: fi.String(lastComponent(r.Target)),
		}
	}
	if r.IPAddress != "" {
		address, err := findAddressByIP(cloud, r.IPAddress)
		if err != nil {
			return nil, fmt.Errorf("error finding Address with IP=%q: %v", r.IPAddress, err)
		}
		actual.IPAddress = address
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *ForwardingRule) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
}

func (_ *ForwardingRule) CheckChanges(a, e, changes *ForwardingRule) error {
	if fi.StringValue(e.Name) == "" {
		return fi.RequiredField("Name")
	}
	return nil
}

func (_ *ForwardingRule) RenderGCE(t *gce.GCEAPITarget, a, e, changes *ForwardingRule) error {
	name := fi.StringValue(e.Name)

	o := &compute.ForwardingRule{
		Name:       name,
		PortRange:  e.PortRange,
		IPProtocol: e.IPProtocol,
	}

	if e.TargetPool != nil {
		o.Target = e.TargetPool.URL(t.Cloud)
	}

	if e.IPAddress != nil {
		o.IPAddress = fi.StringValue(e.IPAddress.IPAddress)
		if o.IPAddress == "" {
			addr, err := e.IPAddress.find(t.Cloud)
			if err != nil {
				return fmt.Errorf("error finding Address %q: %v", e.IPAddress, err)
			}
			if addr == nil {
				return fmt.Errorf("Address %q was not found", e.IPAddress)
			}

			o.IPAddress = fi.StringValue(addr.IPAddress)
			if o.IPAddress == "" {
				return fmt.Errorf("Address had no IP: %v", e.IPAddress)
			}
		}
	}

	if a == nil {
		klog.V(4).Infof("Creating ForwardingRule %q", o.Name)

		_, err := t.Cloud.Compute().ForwardingRules.Insert(t.Cloud.Project(), t.Cloud.Region(), o).Do()
		if err != nil {
			return fmt.Errorf("error creating ForwardingRule %q: %v", o.Name, err)
		}

	} else {
		return fmt.Errorf("cannot apply changes to ForwardingRule: %v", changes)
	}

	return nil
}

type terraformForwardingRule struct {
	Name       string             `json:"name"`
	PortRange  string             `json:"port_range,omitempty"`
	Target     *terraform.Literal `json:"target,omitempty"`
	IPAddress  *terraform.Literal `json:"ip_address,omitempty"`
	IPProtocol string             `json:"ip_protocol,omitempty"`
}

func (_ *ForwardingRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ForwardingRule) error {
	name := fi.StringValue(e.Name)

	tf := &terraformForwardingRule{
		Name:       name,
		PortRange:  e.PortRange,
		IPProtocol: e.IPProtocol,
	}

	if e.TargetPool != nil {
		tf.Target = e.TargetPool.TerraformLink()
	}

	if e.IPAddress != nil {
		tf.IPAddress = e.IPAddress.TerraformAddress()
	}

	return t.RenderResource("google_compute_forwarding_rule", name, tf)
}

func (e *ForwardingRule) TerraformLink() *terraform.Literal {
	name := fi.StringValue(e.Name)

	return terraform.LiteralSelfLink("google_compute_forwarding_rule", name)
}

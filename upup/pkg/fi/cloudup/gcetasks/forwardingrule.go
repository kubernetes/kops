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

// ForwardingRule represents a GCE ForwardingRule
// +kops:fitask
type ForwardingRule struct {
	Name      *string
	Lifecycle fi.Lifecycle

	PortRange  *string
	Ports      []string
	TargetPool *TargetPool
	// An IP address can be specified either in dotted decimal
	// or by reference to an address object.  The following two
	// fields are mutually exclusive.
	IPAddress     *Address
	RuleIPAddress *string

	IPProtocol          string
	LoadBalancingScheme *string
	Network             *Network
	Subnetwork          *Subnet
	BackendService      *BackendService
}

var _ fi.CompareWithID = &ForwardingRule{}

func (e *ForwardingRule) CompareWithID() *string {
	return e.Name
}

func (e *ForwardingRule) Find(c *fi.Context) (*ForwardingRule, error) {
	cloud := c.Cloud.(gce.GCECloud)
	name := fi.StringValue(e.Name)

	r, err := cloud.Compute().ForwardingRules().Get(cloud.Project(), cloud.Region(), name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error getting ForwardingRule %q: %v", name, err)
	}

	actual := &ForwardingRule{
		Name:       fi.String(r.Name),
		IPProtocol: r.IPProtocol,
	}
	if r.PortRange != "" {
		actual.PortRange = &r.PortRange
	}
	if len(r.Ports) > 0 {
		actual.Ports = r.Ports
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
		IPProtocol: e.IPProtocol,
	}
	if e.PortRange != nil {
		o.PortRange = *e.PortRange
	}
	if len(e.Ports) > 0 {
		o.Ports = e.Ports
	}

	if e.LoadBalancingScheme != nil {
		o.LoadBalancingScheme = *e.LoadBalancingScheme
	}

	if e.TargetPool != nil {
		o.Target = e.TargetPool.URL(t.Cloud)
	}

	if e.BackendService != nil {
		if o.Target != "" {
			return fmt.Errorf("cannot specify both %q and %q for forwarding rule target.", o.Target, e.BackendService)
		}
		o.BackendService = e.BackendService.URL(t.Cloud)
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
	if o.IPAddress != "" && e.RuleIPAddress != nil {
		return fmt.Errorf("Specified both IP Address and rule-managed IP address: %v, %v", e.IPAddress, *e.RuleIPAddress)
	}
	if e.RuleIPAddress != nil {
		o.IPAddress = *e.RuleIPAddress
	}

	if e.Network != nil {
		project := t.Cloud.Project()
		if e.Network.Project != nil {
			project = *e.Network.Project
		}
		o.Network = e.Network.URL(project)
	}

	if e.Subnetwork != nil {
		project := t.Cloud.Project()
		if e.Network.Project != nil {
			project = *e.Network.Project
		}
		o.Subnetwork = e.Subnetwork.URL(project, t.Cloud.Region())
	}

	if a == nil {
		klog.V(4).Infof("Creating ForwardingRule %q", o.Name)

		op, err := t.Cloud.Compute().ForwardingRules().Insert(t.Cloud.Project(), t.Cloud.Region(), o)
		if err != nil {
			return fmt.Errorf("error creating ForwardingRule %q: %v", o.Name, err)
		}

		if err := t.Cloud.WaitForOp(op); err != nil {
			return fmt.Errorf("error creating forwarding rule: %v", err)
		}

	} else {
		return fmt.Errorf("cannot apply changes to ForwardingRule: %v", changes)
	}

	return nil
}

type terraformForwardingRule struct {
	Name                string                   `cty:"name"`
	PortRange           *string                  `cty:"port_range"`
	Ports               []string                 `cty:"ports"`
	Target              *terraformWriter.Literal `cty:"target"`
	IPAddress           *terraformWriter.Literal `cty:"ip_address"`
	IPProtocol          string                   `cty:"ip_protocol"`
	LoadBalancingScheme *string                  `cty:"load_balancing_scheme"`
	Network             *terraformWriter.Literal `cty:"network"`
	Subnetwork          *terraformWriter.Literal `cty:"subnetwork"`
	BackendService      *terraformWriter.Literal `cty:"backend_service"`
}

func (_ *ForwardingRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *ForwardingRule) error {
	name := fi.StringValue(e.Name)

	tf := &terraformForwardingRule{
		Name:                name,
		IPProtocol:          e.IPProtocol,
		LoadBalancingScheme: e.LoadBalancingScheme,
		Ports:               e.Ports,
		PortRange:           e.PortRange,
	}

	if e.TargetPool != nil {
		tf.Target = e.TargetPool.TerraformLink()
	}

	if e.Network != nil {
		tf.Network = e.Network.TerraformLink()
	}

	if e.Subnetwork != nil {
		tf.Subnetwork = e.Subnetwork.TerraformLink()
	}

	if e.BackendService != nil {
		tf.BackendService = e.BackendService.TerraformAddress()
	}

	if e.IPAddress != nil {
		tf.IPAddress = e.IPAddress.TerraformAddress()
	} else if e.RuleIPAddress != nil {
		tf.IPAddress = terraformWriter.LiteralFromStringValue(*e.RuleIPAddress)
	}

	return t.RenderResource("google_compute_forwarding_rule", name, tf)
}

func (e *ForwardingRule) TerraformLink() *terraformWriter.Literal {
	name := fi.StringValue(e.Name)

	return terraformWriter.LiteralSelfLink("google_compute_forwarding_rule", name)
}

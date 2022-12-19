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
	"net"
	"strings"

	compute "google.golang.org/api/compute/v1"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// FirewallRule represents a GCE firewall rules
// +kops:fitask
type FirewallRule struct {
	Name      *string
	Lifecycle fi.Lifecycle

	Network      *Network
	SourceTags   []string
	SourceRanges []string
	TargetTags   []string
	Allowed      []string

	// Disabled: Denotes whether the firewall rule is disabled. When set to
	// true, the firewall rule is not enforced and the network behaves as if
	// it did not exist. If this is unspecified, the firewall rule will be
	// enabled.
	Disabled bool
}

var _ fi.CompareWithID = &FirewallRule{}
var _ fi.CloudupTaskNormalize = &FirewallRule{}

func (e *FirewallRule) CompareWithID() *string {
	return e.Name
}

func (e *FirewallRule) Find(c *fi.CloudupContext) (*FirewallRule, error) {
	cloud := c.T.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Firewalls().Get(cloud.Project(), *e.Name)
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing FirewallRules: %v", err)
	}

	actual := &FirewallRule{}
	actual.Name = &r.Name
	actual.Network = &Network{Name: fi.PtrTo(lastComponent(r.Network))}
	actual.TargetTags = r.TargetTags
	actual.SourceRanges = r.SourceRanges
	actual.SourceTags = r.SourceTags
	actual.Disabled = r.Disabled
	for _, a := range r.Allowed {
		actual.Allowed = append(actual.Allowed, serializeFirewallAllowed(a))
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *FirewallRule) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

// Normalize applies some validation that isn't technically required,
// but avoids some problems with surprising behaviours.
func (e *FirewallRule) Normalize(c *fi.CloudupContext) error {
	if !e.Disabled {
		// Treat it as an error if SourceRanges _and_ SourceTags empty with Disabled=false
		// this is interpreted as SourceRanges="0.0.0.0/0", which is likely not what was intended.
		if len(e.SourceRanges) == 0 && len(e.SourceTags) == 0 {
			return fmt.Errorf("either SourceRanges or SourceTags should be specified when Disabled is false")
		}
	}

	// Treat it as an error if SourceRanges _and_ SourceTags both set;
	// this is interpreted as OR, not AND, which is likely not what was intended.
	if len(e.SourceRanges) != 0 && len(e.SourceTags) != 0 {
		return fmt.Errorf("SourceRanges and SourceTags should not both be specified")
	}

	name := fi.ValueOf(e.Name)

	// Make sure we've split the ipv4 / ipv6 addresses.
	// A single firewall rule can't mix ipv4 and ipv6 addresses, so we split them into two rules.
	for _, sourceRange := range e.SourceRanges {
		_, cidr, err := net.ParseCIDR(sourceRange)
		if err != nil {
			return fmt.Errorf("sourceRange %q is not valid: %w", sourceRange, err)
		}
		if cidr.IP.To4() != nil {
			// IPv4
			if strings.Contains(name, "-ipv6") {
				return fmt.Errorf("ipv4 ranges should not be in a ipv6-named rule (found %s in %s)", sourceRange, name)
			}
		} else {
			// IPv6
			if !strings.Contains(name, "-ipv6") {
				return fmt.Errorf("ipv6 ranges should be in a ipv6-named rule (found %s in %s)", sourceRange, name)
			}
		}
	}

	return nil
}

func (_ *FirewallRule) CheckChanges(a, e, changes *FirewallRule) error {
	if e.Network == nil {
		return fi.RequiredField("Network")
	}
	return nil
}

func parseFirewallAllowed(rule string) (*compute.FirewallAllowed, error) {
	o := &compute.FirewallAllowed{}

	tokens := strings.Split(rule, ":")
	if len(tokens) < 1 || len(tokens) > 2 {
		return nil, fmt.Errorf("expected protocol[:portspec] in firewall rule %q", rule)
	}

	o.IPProtocol = tokens[0]
	if len(tokens) == 1 {
		return o, nil
	}

	o.Ports = []string{tokens[1]}
	return o, nil
}

func serializeFirewallAllowed(r *compute.FirewallAllowed) string {
	if len(r.Ports) == 0 {
		return r.IPProtocol
	}

	var tokens []string
	for _, ports := range r.Ports {
		tokens = append(tokens, r.IPProtocol+":"+ports)
	}

	return strings.Join(tokens, ",")
}

func (e *FirewallRule) mapToGCE(project string) (*compute.Firewall, error) {
	var allowed []*compute.FirewallAllowed
	if e.Allowed != nil {
		for _, a := range e.Allowed {
			p, err := parseFirewallAllowed(a)
			if err != nil {
				return nil, err
			}
			allowed = append(allowed, p)
		}
	}
	firewall := &compute.Firewall{
		Name:         *e.Name,
		Network:      e.Network.URL(project),
		SourceTags:   e.SourceTags,
		SourceRanges: e.SourceRanges,
		TargetTags:   e.TargetTags,
		Allowed:      allowed,
		Disabled:     e.Disabled,
	}
	return firewall, nil
}

func (_ *FirewallRule) RenderGCE(t *gce.GCEAPITarget, a, e, changes *FirewallRule) error {
	cloud := t.Cloud
	firewall, err := e.mapToGCE(cloud.Project())
	if err != nil {
		return err
	}

	if a == nil {
		_, err := t.Cloud.Compute().Firewalls().Insert(t.Cloud.Project(), firewall)
		if err != nil {
			return fmt.Errorf("error creating FirewallRule: %v", err)
		}
	} else {
		_, err := t.Cloud.Compute().Firewalls().Update(t.Cloud.Project(), *e.Name, firewall)
		if err != nil {
			return fmt.Errorf("error creating FirewallRule: %v", err)
		}
	}

	return nil
}

type terraformAllow struct {
	Protocol string   `cty:"protocol"`
	Ports    []string `cty:"ports"`
}

type terraformFirewall struct {
	Name    string                   `cty:"name"`
	Network *terraformWriter.Literal `cty:"network"`

	Allowed []*terraformAllow `cty:"allow"`

	SourceTags []string `cty:"source_tags"`

	SourceRanges []string `cty:"source_ranges"`
	TargetTags   []string `cty:"target_tags"`

	Disabled bool `cty:"disabled"`
}

func (_ *FirewallRule) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *FirewallRule) error {
	g, err := e.mapToGCE(t.Project)
	if err != nil {
		return err
	}

	var allowed []*terraformAllow
	if g.Allowed != nil {
		for _, ga := range g.Allowed {
			a := &terraformAllow{
				Protocol: ga.IPProtocol,
				Ports:    ga.Ports,
			}

			allowed = append(allowed, a)
		}
	}
	tf := &terraformFirewall{
		Name:         g.Name,
		SourceRanges: g.SourceRanges,
		TargetTags:   g.TargetTags,
		SourceTags:   g.SourceTags,
		Allowed:      allowed,
		Disabled:     g.Disabled,
	}

	tf.Network = e.Network.TerraformLink()

	return t.RenderResource("google_compute_firewall", *e.Name, tf)
}

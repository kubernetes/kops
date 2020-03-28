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
	"strings"

	compute "google.golang.org/api/compute/v0.beta"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// FirewallRule represents a GCE firewall rules
//go:generate fitask -type=FirewallRule
type FirewallRule struct {
	Name      *string
	Lifecycle *fi.Lifecycle

	Network      *Network
	SourceTags   []string
	SourceRanges []string
	TargetTags   []string
	Allowed      []string
}

var _ fi.CompareWithID = &FirewallRule{}

func (e *FirewallRule) CompareWithID() *string {
	return e.Name
}

func (e *FirewallRule) Find(c *fi.Context) (*FirewallRule, error) {
	cloud := c.Cloud.(gce.GCECloud)

	r, err := cloud.Compute().Firewalls.Get(cloud.Project(), *e.Name).Do()
	if err != nil {
		if gce.IsNotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("error listing FirewallRules: %v", err)
	}

	actual := &FirewallRule{}
	actual.Name = &r.Name
	actual.Network = &Network{Name: fi.String(lastComponent(r.Network))}
	actual.TargetTags = r.TargetTags
	actual.SourceRanges = r.SourceRanges
	actual.SourceTags = r.SourceTags
	for _, a := range r.Allowed {
		actual.Allowed = append(actual.Allowed, serializeFirewallAllowed(a))
	}

	// Ignore "system" fields
	actual.Lifecycle = e.Lifecycle

	return actual, nil
}

func (e *FirewallRule) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(e, c)
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
		_, err := t.Cloud.Compute().Firewalls.Insert(t.Cloud.Project(), firewall).Do()
		if err != nil {
			return fmt.Errorf("error creating FirewallRule: %v", err)
		}
	} else {
		_, err := t.Cloud.Compute().Firewalls.Update(t.Cloud.Project(), *e.Name, firewall).Do()
		if err != nil {
			return fmt.Errorf("error creating FirewallRule: %v", err)
		}
	}

	return nil
}

type terraformAllow struct {
	Protocol string   `json:"protocol,omitempty" cty:"protocol"`
	Ports    []string `json:"ports,omitempty" cty:"ports"`
}

type terraformFirewall struct {
	Name    string             `json:"name" cty:"name"`
	Network *terraform.Literal `json:"network" cty:"network"`

	Allowed []*terraformAllow `json:"allow,omitempty" cty:"allow"`

	SourceTags []string `json:"source_tags,omitempty" cty:"source_tags"`

	SourceRanges []string `json:"source_ranges,omitempty" cty:"source_ranges"`
	TargetTags   []string `json:"target_tags,omitempty" cty:"target_tags"`
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
	}

	// TODO: This doesn't seem right, but it looks like a TF problem
	tf.Network = e.Network.TerraformName()

	return t.RenderResource("google_compute_firewall", *e.Name, tf)
}

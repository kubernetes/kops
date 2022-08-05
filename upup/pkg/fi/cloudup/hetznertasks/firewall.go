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

package hetznertasks

import (
	"context"
	"net"
	"strconv"

	"github.com/hetznercloud/hcloud-go/hcloud"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
)

// +kops:fitask
type Firewall struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID       *int
	Selector string
	Rules    []*FirewallRule

	Labels map[string]string
}

var _ fi.CompareWithID = &Firewall{}

func (v *Firewall) CompareWithID() *string {
	return fi.String(strconv.Itoa(fi.IntValue(v.ID)))
}

func (v *Firewall) Find(c *fi.Context) (*Firewall, error) {
	cloud := c.Cloud.(hetzner.HetznerCloud)
	client := cloud.FirewallClient()

	// TODO(hakman): Find using label selector
	firewalls, err := client.All(context.TODO())
	if err != nil {
		return nil, err
	}

	for _, firewall := range firewalls {
		if firewall.Name == fi.StringValue(v.Name) {
			matches := &Firewall{
				Lifecycle: v.Lifecycle,
				Name:      fi.String(firewall.Name),
				ID:        fi.Int(firewall.ID),
				Labels:    firewall.Labels,
			}
			for _, rule := range firewall.Rules {
				firewallRule := FirewallRule{
					Direction: string(rule.Direction),
					SourceIPs: rule.SourceIPs,
					Protocol:  string(rule.Protocol),
					Port:      rule.Port,
				}
				matches.Rules = append(matches.Rules, &firewallRule)
			}
			for _, selector := range firewall.AppliedTo {
				if selector.Type == hcloud.FirewallResourceTypeLabelSelector && selector.LabelSelector != nil {
					matches.Selector = selector.LabelSelector.Selector
				}
			}
			v.ID = matches.ID
			return matches, nil
		}
	}

	return nil, nil
}

func (v *Firewall) Run(c *fi.Context) error {
	return fi.DefaultDeltaRunMethod(v, c)
}

func (_ *Firewall) CheckChanges(a, e, changes *Firewall) error {
	if a != nil {
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Selector != "" && a.Selector != "" {
			return fi.CannotChangeField("Selector")
		}
	} else {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
		if e.Selector == "" {
			return fi.RequiredField("Selector")
		}
	}
	return nil
}

func (_ *Firewall) RenderHetzner(t *hetzner.HetznerAPITarget, a, e, changes *Firewall) error {
	client := t.Cloud.FirewallClient()
	if a == nil {
		opts := hcloud.FirewallCreateOpts{
			Name: fi.StringValue(e.Name),
			ApplyTo: []hcloud.FirewallResource{
				{
					Type:          hcloud.FirewallResourceTypeLabelSelector,
					LabelSelector: &hcloud.FirewallResourceLabelSelector{Selector: e.Selector},
				},
			},
			Labels: e.Labels,
		}
		for _, rule := range e.Rules {
			firewallRule := hcloud.FirewallRule{
				Direction: hcloud.FirewallRuleDirection(rule.Direction),
				SourceIPs: rule.SourceIPs,
				Protocol:  hcloud.FirewallRuleProtocol(rule.Protocol),
				Port:      rule.Port,
			}
			opts.Rules = append(opts.Rules, firewallRule)
		}
		_, _, err := client.Create(context.TODO(), opts)
		if err != nil {
			return err
		}

	} else {
		firewall, _, err := client.Get(context.TODO(), fi.StringValue(e.Name))
		if err != nil {
			return err
		}

		// Update the labels
		if changes.Name != nil || len(changes.Labels) != 0 {
			_, _, err := client.Update(context.TODO(), firewall, hcloud.FirewallUpdateOpts{
				Name:   fi.StringValue(e.Name),
				Labels: e.Labels,
			})
			if err != nil {
				return err
			}
		}

		// Update the rules
		if len(changes.Rules) > 0 {
			var firewallRules []hcloud.FirewallRule
			for _, rule := range e.Rules {
				firewallRule := hcloud.FirewallRule{
					Direction: hcloud.FirewallRuleDirection(rule.Direction),
					SourceIPs: rule.SourceIPs,
					Protocol:  hcloud.FirewallRuleProtocol(rule.Protocol),
					Port:      rule.Port,
				}
				firewallRules = append(firewallRules, firewallRule)
			}
			_, _, err = client.SetRules(context.TODO(), firewall, hcloud.FirewallSetRulesOpts{
				Rules: firewallRules,
			})
			if err != nil {
				return err
			}
		}

		// Update the selector
		if a.Selector == "" {
			firewallResources := []hcloud.FirewallResource{
				{
					Type:          hcloud.FirewallResourceTypeLabelSelector,
					LabelSelector: &hcloud.FirewallResourceLabelSelector{Selector: e.Selector},
				},
			}
			_, _, err = client.ApplyResources(context.TODO(), firewall, firewallResources)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// FirewallRule represents a Firewall's rules.
type FirewallRule struct {
	Direction string
	SourceIPs []net.IPNet
	Protocol  string
	Port      *string
}

var _ fi.HasDependencies = &FirewallRule{}

func (e *FirewallRule) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	return nil
}

type terraformFirewall struct {
	Name    *string                     `cty:"name"`
	ApplyTo []*terraformFirewallApplyTo `cty:"apply_to"`
	Rules   []*terraformFirewallRule    `cty:"rule"`
	Labels  map[string]string           `cty:"labels"`
}

type terraformFirewallApplyTo struct {
	LabelSelector *string `cty:"label_selector"`
}

type terraformFirewallRule struct {
	Direction *string   `cty:"direction"`
	SourceIPs []*string `cty:"source_ips"`
	Protocol  *string   `cty:"protocol"`
	Port      *string   `cty:"port"`
}

func (_ *Firewall) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *Firewall) error {
	{
		tf := &terraformFirewall{
			Name: e.Name,
			ApplyTo: []*terraformFirewallApplyTo{
				{
					LabelSelector: fi.String(e.Selector),
				},
			},
			Labels: e.Labels,
		}
		for _, rule := range e.Rules {
			tfr := &terraformFirewallRule{
				Direction: fi.String(string(rule.Direction)),
				Protocol:  fi.String(string(rule.Protocol)),
				Port:      rule.Port,
			}
			for _, ip := range rule.SourceIPs {
				tfr.SourceIPs = append(tfr.SourceIPs, fi.String(ip.String()))
			}
			tf.Rules = append(tf.Rules, tfr)
		}

		err := t.RenderResource("hcloud_firewall", *e.Name, tf)
		if err != nil {
			return err
		}

	}

	return nil
}

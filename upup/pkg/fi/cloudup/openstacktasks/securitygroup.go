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

package openstacktasks

import (
	"fmt"
	"strconv"
	"strings"

	sg "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	sgr "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=SecurityGroup
type SecurityGroup struct {
	ID               *string
	Name             *string
	Description      *string
	RemoveExtraRules []string
	RemoveGroup      bool
	Lifecycle        *fi.Lifecycle
}

var _ fi.CompareWithID = &SecurityGroup{}

func (s *SecurityGroup) CompareWithID() *string {
	return s.ID
}

func (s *SecurityGroup) Find(context *fi.Context) (*SecurityGroup, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	// avoid creating new group if it has removegroup flag
	if s.RemoveGroup {
		return s, nil
	}
	return getSecurityGroupByName(s, cloud)
}

func getSecurityGroupByName(s *SecurityGroup, cloud openstack.OpenstackCloud) (*SecurityGroup, error) {
	opt := sg.ListOpts{
		Name: fi.StringValue(s.Name),
	}
	gs, err := cloud.ListSecurityGroups(opt)
	if err != nil {
		return nil, err
	}
	n := len(gs)
	if n == 0 {
		return nil, nil
	} else if n != 1 {
		return nil, fmt.Errorf("found multiple SecurityGroups with name: %s", fi.StringValue(s.Name))
	}
	g := gs[0]
	actual := &SecurityGroup{
		ID:          fi.String(g.ID),
		Name:        fi.String(g.Name),
		Description: fi.String(g.Description),
		Lifecycle:   s.Lifecycle,
	}
	actual.RemoveExtraRules = s.RemoveExtraRules
	actual.RemoveGroup = s.RemoveGroup
	s.ID = actual.ID
	return actual, nil
}

func (s *SecurityGroup) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *SecurityGroup) CheckChanges(a, e, changes *SecurityGroup) error {
	if a == nil {
		if e.Name == nil {
			return fi.RequiredField("Name")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Name != nil {
			return fi.CannotChangeField("Name")
		}
	}
	return nil
}

func (_ *SecurityGroup) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *SecurityGroup) error {
	if a == nil {
		klog.V(2).Infof("Creating SecurityGroup with Name:%q", fi.StringValue(e.Name))

		opt := sg.CreateOpts{
			Name:        fi.StringValue(e.Name),
			Description: fi.StringValue(e.Description),
		}

		g, err := t.Cloud.CreateSecurityGroup(opt)
		if err != nil {
			return fmt.Errorf("error creating SecurityGroup: %v", err)
		}

		e.ID = fi.String(g.ID)
		return nil
	}

	klog.V(2).Infof("Openstack task SecurityGroup::RenderOpenstack did nothing")
	return nil
}

func (s *SecurityGroup) FindDeletions(c *fi.Context) ([]fi.Deletion, error) {
	var removals []fi.Deletion

	if len(s.RemoveExtraRules) == 0 && !s.RemoveGroup {
		return nil, nil
	}

	cloud := c.Cloud.(openstack.OpenstackCloud)
	if s.RemoveGroup {
		sg, err := getSecurityGroupByName(s, cloud)
		if err != nil {
			return nil, err
		}
		if sg != nil {
			removals = append(removals, &deleteSecurityGroup{
				securityGroup: sg,
			})
		}
	}

	var rules []RemovalRule
	for _, r := range s.RemoveExtraRules {
		rule, err := ParseRemovalRule(r)
		if err != nil {
			return nil, fmt.Errorf("cannot parse rule %q: %v", r, err)
		}
		rules = append(rules, rule)
	}

	sg, err := getSecurityGroupByName(s, cloud)
	if err != nil {
		return nil, err
	}
	if sg == nil {
		return nil, nil
	}

	sgRules, err := cloud.ListSecurityGroupRules(sgr.ListOpts{
		SecGroupID: fi.StringValue(sg.ID),
	})
	if err != nil {
		return nil, err
	}

	for _, permission := range sgRules {
		match := false
		for _, rule := range rules {
			if rule.Matches(permission) {
				klog.V(2).Infof("permission matches rule %s: %v", rule, permission)
				match = true
				break
			}
		}
		if !match {
			klog.V(4).Infof("Ignoring security group permission %q (did not match removal rules)", permission)
			continue
		}

		found := false
		for _, t := range c.AllTasks() {
			er, ok := t.(*SecurityGroupRule)
			if !ok {
				continue
			}

			if matches(er, permission) {
				found = true
			}
		}
		if !found {
			removals = append(removals, &deleteSecurityGroupRule{
				rule:          permission,
				securityGroup: s,
			})
		}
	}
	return removals, nil
}

func matches(t *SecurityGroupRule, perm sgr.SecGroupRule) bool {
	if fi.IntValue(t.PortRangeMin) != perm.PortRangeMin {
		return false
	}

	if fi.IntValue(t.PortRangeMax) != perm.PortRangeMax {
		return false
	}

	if perm.Protocol != "tcp" {
		return false
	}

	if perm.RemoteIPPrefix != fi.StringValue(t.RemoteIPPrefix) {
		return false
	}

	return true
}

type deleteSecurityGroup struct {
	securityGroup *SecurityGroup
}

var _ fi.Deletion = &deleteSecurityGroup{}

func (d *deleteSecurityGroup) Delete(t fi.Target) error {
	klog.V(2).Infof("deleting security group: %v", fi.DebugAsJsonString(d.securityGroup.Name))

	os, ok := t.(*openstack.OpenstackAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}
	err := os.Cloud.DeleteSecurityGroup(fi.StringValue(d.securityGroup.ID))
	if err != nil {
		return fmt.Errorf("error revoking SecurityGroup: %v", err)
	}
	return nil
}

func (d *deleteSecurityGroup) TaskName() string {
	return "SecurityGroup"
}

func (d *deleteSecurityGroup) Item() string {
	s := fmt.Sprintf("securitygroup=%s", fi.StringValue(d.securityGroup.Name))
	return s
}

type deleteSecurityGroupRule struct {
	rule          sgr.SecGroupRule
	securityGroup *SecurityGroup
}

var _ fi.Deletion = &deleteSecurityGroupRule{}

func (d *deleteSecurityGroupRule) Delete(t fi.Target) error {
	klog.V(2).Infof("deleting security group permission: %v", fi.DebugAsJsonString(d.rule))

	os, ok := t.(*openstack.OpenstackAPITarget)
	if !ok {
		return fmt.Errorf("unexpected target type for deletion: %T", t)
	}
	err := os.Cloud.DeleteSecurityGroupRule(d.rule.ID)
	if err != nil {
		return fmt.Errorf("error revoking SecurityGroupRule: %v", err)
	}
	return nil
}

func (d *deleteSecurityGroupRule) TaskName() string {
	return "SecurityGroupRule"
}

func (d *deleteSecurityGroupRule) Item() string {
	s := ""
	if d.rule.PortRangeMin != 0 {
		s += fmt.Sprintf(" port=%d", d.rule.PortRangeMin)
		if d.rule.PortRangeMin != d.rule.PortRangeMax {
			s += fmt.Sprintf("-%d", d.rule.PortRangeMax)
		}
	}
	s += " protocol=tcp"
	s += fmt.Sprintf(" ip=%s", d.rule.RemoteIPPrefix)
	s += fmt.Sprintf(" securitygroup=%s", fi.StringValue(d.securityGroup.Name))
	return s
}

// RemovalRule is a rule that filters the permissions we should remove
type RemovalRule interface {
	Matches(sgr.SecGroupRule) bool
}

// ParseRemovalRule parses our removal rule DSL into a RemovalRule
func ParseRemovalRule(rule string) (RemovalRule, error) {
	rule = strings.TrimSpace(rule)
	tokens := strings.Split(rule, "=")

	// Simple little language:
	//   port=N matches rules that filter (only) by port=N
	//
	// Note this language is internal, so isn't required to be stable

	if len(tokens) == 2 {
		if tokens[0] == "port" {
			port, err := strconv.Atoi(tokens[1])
			if err != nil {
				return nil, fmt.Errorf("cannot parse rule %q", rule)
			}

			return &PortRemovalRule{Port: port}, nil
		} else {
			return nil, fmt.Errorf("cannot parse rule %q", rule)
		}
	}
	return nil, fmt.Errorf("cannot parse rule %q", rule)
}

type PortRemovalRule struct {
	Port int
}

var _ RemovalRule = &PortRemovalRule{}

func (r *PortRemovalRule) String() string {
	return fi.DebugAsJsonString(r)
}

func (r *PortRemovalRule) Matches(rule sgr.SecGroupRule) bool {
	// Check if port matches
	if rule.PortRangeMin != r.Port {
		return false
	}
	if rule.PortRangeMax != r.Port {
		return false
	}
	return true
}

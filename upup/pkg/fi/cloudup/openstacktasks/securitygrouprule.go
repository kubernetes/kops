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

package openstacktasks

import (
	"fmt"

	sgr "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/rules"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

func Int(v int) *int {
	return &v
}

func IntValue(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

type SecurityGroupRule struct {
	ID             *string
	Direction      *string
	EtherType      *string
	SecGroup       *SecurityGroup
	PortRangeMin   *int
	PortRangeMax   *int
	Protocol       *string
	RemoteIPPrefix *string
	RemoteGroup    *SecurityGroup
	Lifecycle      *fi.Lifecycle
}

// GetDependencies returns the dependencies of the Instance task
func (e *SecurityGroupRule) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task
	for _, task := range tasks {
		if _, ok := task.(*SecurityGroup); ok {
			deps = append(deps, task)
		}
	}
	return deps
}

var _ fi.CompareWithID = &SecurityGroupRule{}

func (r *SecurityGroupRule) CompareWithID() *string {
	return r.ID
}

func (r *SecurityGroupRule) Find(context *fi.Context) (*SecurityGroupRule, error) {
	if r.SecGroup == nil || r.SecGroup.ID == nil {
		return nil, nil
	}

	cloud := context.Cloud.(openstack.OpenstackCloud)

	opt := sgr.ListOpts{
		Direction:      fi.StringValue(r.Direction),
		EtherType:      fi.StringValue(r.EtherType),
		PortRangeMax:   IntValue(r.PortRangeMax),
		PortRangeMin:   IntValue(r.PortRangeMin),
		Protocol:       fi.StringValue(r.Protocol),
		RemoteIPPrefix: fi.StringValue(r.RemoteIPPrefix),
		SecGroupID:     fi.StringValue(r.SecGroup.ID),
	}
	if r.RemoteGroup != nil {
		opt.RemoteGroupID = fi.StringValue(r.RemoteGroup.ID)
	}
	rs, err := cloud.ListSecurityGroupRules(opt)
	if err != nil {
		return nil, err
	}
	n := len(rs)
	if n == 0 {
		return nil, nil
	} else if n != 1 {
		return nil, fmt.Errorf("found multiple SecurityGroupRules")
	}
	rule := rs[0]
	actual := &SecurityGroupRule{
		ID:             fi.String(rule.ID),
		Direction:      fi.String(rule.Direction),
		EtherType:      fi.String(rule.EtherType),
		PortRangeMax:   Int(rule.PortRangeMax),
		PortRangeMin:   Int(rule.PortRangeMin),
		Protocol:       fi.String(rule.Protocol),
		RemoteIPPrefix: fi.String(rule.RemoteIPPrefix),
		RemoteGroup: &SecurityGroup{
			ID: fi.String(rule.RemoteGroupID),
		},
		SecGroup:  &SecurityGroup{ID: fi.String(rule.SecGroupID)},
		Lifecycle: r.Lifecycle,
	}
	r.ID = actual.ID
	return actual, nil
}

func (r *SecurityGroupRule) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(r, context)
}

func (_ *SecurityGroupRule) CheckChanges(a, e, changes *SecurityGroupRule) error {
	if a == nil {
		if e.Direction == nil {
			return fi.RequiredField("Direction")
		}
		if e.EtherType == nil {
			return fi.RequiredField("EtherType")
		}
		if e.SecGroup == nil {
			return fi.RequiredField("SecGroup")
		}
	} else {
		if changes.ID != nil {
			return fi.CannotChangeField("ID")
		}
		if changes.Direction != nil {
			return fi.CannotChangeField("Direction")
		}
		if changes.EtherType != nil {
			return fi.CannotChangeField("EtherType")
		}
		if changes.SecGroup != nil {
			return fi.CannotChangeField("SecGroup")
		}
	}
	return nil
}

func (_ *SecurityGroupRule) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *SecurityGroupRule) error {
	if a == nil {
		klog.V(2).Infof("Creating SecurityGroupRule")

		opt := sgr.CreateOpts{
			Direction:      sgr.RuleDirection(fi.StringValue(e.Direction)),
			EtherType:      sgr.RuleEtherType(fi.StringValue(e.EtherType)),
			SecGroupID:     fi.StringValue(e.SecGroup.ID),
			PortRangeMax:   IntValue(e.PortRangeMax),
			PortRangeMin:   IntValue(e.PortRangeMin),
			Protocol:       sgr.RuleProtocol(fi.StringValue(e.Protocol)),
			RemoteIPPrefix: fi.StringValue(e.RemoteIPPrefix),
		}
		if e.RemoteGroup != nil {
			opt.RemoteGroupID = fi.StringValue(e.RemoteGroup.ID)
		}

		r, err := t.Cloud.CreateSecurityGroupRule(opt)
		if err != nil {
			return fmt.Errorf("error creating SecurityGroupRule in SG %s: %v", fi.StringValue(e.SecGroup.GetName()), err)
		}

		e.ID = fi.String(r.ID)
		return nil
	}

	klog.V(2).Infof("Openstack task SecurityGroupRule::RenderOpenstack did nothing")
	return nil
}

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

	sg "github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/security/groups"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=SecurityGroup
type SecurityGroup struct {
	ID          *string
	Name        *string
	Description *string
	Lifecycle   *fi.Lifecycle
}

var _ fi.CompareWithID = &SecurityGroup{}

func (s *SecurityGroup) CompareWithID() *string {
	return s.ID
}

func (s *SecurityGroup) Find(context *fi.Context) (*SecurityGroup, error) {
	cloud := context.Cloud.(openstack.OpenstackCloud)
	return s.getSecurityGroupByName(cloud)
}

func (s *SecurityGroup) getSecurityGroupByName(cloud openstack.OpenstackCloud) (*SecurityGroup, error) {
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

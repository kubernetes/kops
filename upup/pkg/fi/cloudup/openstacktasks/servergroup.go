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

	"github.com/golang/glog"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=ServerGroup
type ServerGroup struct {
	ID        *string
	Name      *string
	Members   []string
	Policies  []string
	Lifecycle *fi.Lifecycle
}

var _ fi.CompareWithID = &ServerGroup{}

func (s *ServerGroup) CompareWithID() *string {
	return s.ID
}

func (s *ServerGroup) Find(context *fi.Context) (*ServerGroup, error) {
	if s == nil || s.ID == nil {
		return nil, nil
	}
	id := *(s.ID)
	cloud := context.Cloud.(openstack.OpenstackCloud)
	g, err := servergroups.Get(cloud.ComputeClient(), id).Extract()
	if err != nil {
		return nil, err
	}

	a := &ServerGroup{
		ID:        fi.String(g.ID),
		Name:      fi.String(g.Name),
		Members:   g.Members,
		Policies:  g.Policies,
		Lifecycle: s.Lifecycle,
	}
	return a, nil
}

func (s *ServerGroup) Run(context *fi.Context) error {
	return fi.DefaultDeltaRunMethod(s, context)
}

func (_ *ServerGroup) CheckChanges(a, e, changes *ServerGroup) error {
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

func (_ *ServerGroup) RenderOpenstack(t *openstack.OpenstackAPITarget, a, e, changes *ServerGroup) error {
	if a == nil {
		glog.V(2).Infof("Creating ServerGroup with Name:%q", fi.StringValue(e.Name))

		opt := servergroups.CreateOpts{
			Name:     fi.StringValue(e.Name),
			Policies: e.Policies,
		}

		g, err := t.Cloud.CreateServerGroup(opt)
		if err != nil {
			return fmt.Errorf("error creating ServerGroup: %v", err)
		}

		e.ID = fi.String(g.ID)
		return nil
	}

	glog.V(2).Infof("Openstack task ServerGroup::RenderOpenstack did nothing")
	return nil
}

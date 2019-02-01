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
	if s == nil || s.Name == nil {
		return nil, nil
	}
	cloud := context.Cloud.(openstack.OpenstackCloud)
	//TODO: move to cloud, add vfs backoff

	page, err := servergroups.List(cloud.ComputeClient()).AllPages()
	if err != nil {
		return nil, fmt.Errorf("Failed to list server groups: %v", err)
	}
	serverGroups, err := servergroups.ExtractServerGroups(page)
	if err != nil {
		return nil, fmt.Errorf("Failed to extract server groups: %v", err)
	}
	var actual *ServerGroup
	for _, serverGroup := range serverGroups {
		if serverGroup.Name == *s.Name {
			if actual != nil {
				return nil, fmt.Errorf("Found multiple server groups with name %s", fi.StringValue(s.Name))
			}
			actual = &ServerGroup{
				Name:      fi.String(serverGroup.Name),
				ID:        fi.String(serverGroup.ID),
				Members:   serverGroup.Members,
				Lifecycle: s.Lifecycle,
				Policies:  serverGroup.Policies,
			}
		}
	}
	if actual == nil {
		return nil, nil
	}
	s.ID = actual.ID
	s.Members = actual.Members
	return actual, nil
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

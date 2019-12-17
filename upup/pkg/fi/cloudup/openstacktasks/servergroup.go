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
	"strings"

	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/servergroups"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"k8s.io/klog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

//go:generate fitask -type=ServerGroup
type ServerGroup struct {
	ID          *string
	Name        *string
	ClusterName *string
	IGName      *string
	Members     []string
	Policies    []string
	MaxSize     *int32
	Lifecycle   *fi.Lifecycle
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
				Name:        fi.String(serverGroup.Name),
				ClusterName: s.ClusterName,
				IGName:      s.IGName,
				ID:          fi.String(serverGroup.ID),
				Members:     serverGroup.Members,
				Lifecycle:   s.Lifecycle,
				Policies:    serverGroup.Policies,
				MaxSize:     fi.Int32(int32(len(serverGroup.Members))),
			}
		}
	}
	if actual == nil {
		return nil, nil
	}

	// ignore if IG is scaled up, this is handled in instancetasks
	if fi.Int32Value(actual.MaxSize) < fi.Int32Value(s.MaxSize) {
		s.MaxSize = actual.MaxSize
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
		klog.V(2).Infof("Creating ServerGroup with Name:%q", fi.StringValue(e.Name))

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
	} else if changes.MaxSize != nil && fi.Int32Value(a.MaxSize) > fi.Int32Value(changes.MaxSize) {
		currentLastIndex := fi.Int32Value(a.MaxSize)

		for currentLastIndex > fi.Int32Value(changes.MaxSize) {
			iName := strings.ToLower(fmt.Sprintf("%s-%d.%s", fi.StringValue(a.IGName), currentLastIndex, fi.StringValue(a.ClusterName)))
			instanceName := strings.Replace(iName, ".", "-", -1)
			opts := servers.ListOpts{
				Name: fmt.Sprintf("^%s$", instanceName),
			}
			instances, err := t.Cloud.ListInstances(opts)
			if err != nil {
				return fmt.Errorf("error fetching instance list: %v", err)
			}

			if len(instances) == 1 {
				klog.V(2).Infof("Openstack task ServerGroup scaling down instance %s", instanceName)
				err := t.Cloud.DeleteInstanceWithID(instances[0].ID)
				if err != nil {
					return fmt.Errorf("Could not delete instance %s: %v", instanceName, err)
				}
			} else {
				return fmt.Errorf("found %d instances with name: %s", len(instances), instanceName)
			}
			currentLastIndex -= 1
		}
	}

	klog.V(2).Infof("Openstack task ServerGroup::RenderOpenstack did nothing")
	return nil
}

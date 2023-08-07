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
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

// +kops:fitask
type ServerGroup struct {
	ID          *string
	Name        *string
	ClusterName *string
	IGMap       map[string]*int32
	Policies    []string
	Lifecycle   fi.Lifecycle
}

var _ fi.CompareWithID = &ServerGroup{}

func (s *ServerGroup) CompareWithID() *string {
	return s.ID
}

func (s *ServerGroup) Find(context *fi.CloudupContext) (*ServerGroup, error) {
	if s == nil || s.Name == nil {
		return nil, nil
	}
	cloud := context.T.Cloud.(openstack.OpenstackCloud)

	serverGroups, err := cloud.ListServerGroups(servergroups.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("Failed to list server groups: %v", err)
	}

	serverList, err := cloud.ListInstances(servers.ListOpts{})
	if err != nil {
		return nil, fmt.Errorf("Failed to list servers: %v", err)
	}

	serverMap := make(map[string]servers.Server)
	for _, server := range serverList {
		val, ok := server.Metadata["k8s"]
		if !ok || val != fi.ValueOf(s.ClusterName) {
			continue
		}
		serverMap[server.ID] = server
	}

	var actual *ServerGroup
	for _, serverGroup := range serverGroups {
		if serverGroup.Name == *s.Name {
			if actual != nil {
				return nil, fmt.Errorf("Found multiple server groups with name %s", fi.ValueOf(s.Name))
			}
			igMap := make(map[string]*int32)
			for _, serverID := range serverGroup.Members {
				server, ok := serverMap[serverID]
				if !ok {
					return nil, fmt.Errorf("Could not find Server with id %s which is part of ServerGroup %s members", serverID, serverGroup.Name)
				}
				igName, ok := server.Metadata[openstack.TagKopsInstanceGroup]
				if !ok {
					klog.Warningf("Could not find instancegroup metadata tag for server %s", serverID)
					continue
				}

				val, ok := igMap[igName]
				if !ok {
					igMap[igName] = fi.PtrTo(int32(1))
				} else {
					igMap[igName] = fi.PtrTo(fi.ValueOf(val) + 1)
				}
			}
			actual = &ServerGroup{
				Name:        fi.PtrTo(serverGroup.Name),
				ClusterName: s.ClusterName,
				IGMap:       igMap,
				ID:          fi.PtrTo(serverGroup.ID),
				Lifecycle:   s.Lifecycle,
				Policies:    serverGroup.Policies,
			}
		}
	}
	if actual == nil {
		return nil, nil
	}

	// ignore if IG is scaled up, this is handled in instancetasks
	for name, maxSize := range s.IGMap {
		if actual.IGMap[name] != nil && fi.ValueOf(actual.IGMap[name]) < fi.ValueOf(maxSize) {
			s.IGMap[name] = actual.IGMap[name]
		} else if actual.IGMap[name] == nil {
			delete(s.IGMap, name)
		}
	}

	s.ID = actual.ID
	return actual, nil
}

func (s *ServerGroup) Run(context *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(s, context)
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
		klog.V(2).Infof("Creating ServerGroup with Name:%q", fi.ValueOf(e.Name))

		opt := servergroups.CreateOpts{
			Name:     fi.ValueOf(e.Name),
			Policies: e.Policies,
		}

		g, err := t.Cloud.CreateServerGroup(opt)
		if err != nil {
			return fmt.Errorf("error creating ServerGroup: %v", err)
		}
		e.ID = fi.PtrTo(g.ID)
		return nil
	} else if changes.IGMap != nil {
		for igName, maxSize := range changes.IGMap {
			actualIG := a.IGMap[igName]
			if fi.ValueOf(actualIG) > fi.ValueOf(maxSize) {
				currentLastIndex := fi.ValueOf(actualIG)

				for currentLastIndex > fi.ValueOf(maxSize) {
					iName := strings.ToLower(fmt.Sprintf("%s-%d.%s", igName, currentLastIndex, fi.ValueOf(a.ClusterName)))
					instanceName := strings.Replace(iName, ".", "-", -1)
					opts := servers.ListOpts{
						Name: fmt.Sprintf("^%s", igName),
					}
					allInstances, err := t.Cloud.ListInstances(opts)
					if err != nil {
						return fmt.Errorf("error fetching instance list: %v", err)
					}

					instances := []servers.Server{}
					for _, server := range allInstances {
						val, ok := server.Metadata["k8s"]
						if !ok || val != fi.ValueOf(a.ClusterName) {
							continue
						}
						metadataName := ""
						val, ok = server.Metadata[openstack.TagKopsName]
						if ok {
							metadataName = val
						}
						// name or metadata tag should match to instance name
						// this is needed for backwards compatibility
						if server.Name == instanceName || metadataName == instanceName {
							instances = append(instances, server)
						}
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
		}
		return nil
	}

	klog.V(2).Infof("Openstack task ServerGroup::RenderOpenstack did nothing")
	return nil
}

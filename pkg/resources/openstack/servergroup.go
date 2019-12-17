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

package openstack

import (
	"strings"

	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

const (
	typeServerGroup = "ServerGroup"
)

func (os *clusterDiscoveryOS) ListServerGroups() ([]*resources.Resource, error) {
	var resourceTrackers []*resources.Resource
	servergroups, err := os.osCloud.ListServerGroups()
	if err != nil {
		return resourceTrackers, err
	}

	for _, servergroup := range servergroups {
		if strings.HasPrefix(servergroup.Name, os.clusterName) {
			resourceTracker := &resources.Resource{
				Name: servergroup.Name,
				ID:   servergroup.ID,
				Type: typeServerGroup,
				Deleter: func(cloud fi.Cloud, r *resources.Resource) error {
					return cloud.(openstack.OpenstackCloud).DeleteServerGroup(r.ID)
				},
			}
			resourceTrackers = append(resourceTrackers, resourceTracker)
		}
	}
	return resourceTrackers, nil
}

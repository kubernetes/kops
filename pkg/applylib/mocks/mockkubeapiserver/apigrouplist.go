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

package mockkubeapiserver

import (
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// apiGroupList is a request for api discovery, such as GET /apis
type apiGroupList struct {
	baseRequest
}

// Run serves the GET /apis endpoint
func (r *apiGroupList) Run(s *MockKubeAPIServer) error {
	groupMap := make(map[string]*metav1.APIGroup)
	for _, resource := range s.schema.resources {
		group := groupMap[resource.Group]
		if group == nil {
			group = &metav1.APIGroup{Name: resource.Group}
			groupMap[resource.Group] = group
		}

		foundVersion := false
		for _, version := range group.Versions {
			if version.Version == resource.Version {
				foundVersion = true
			}
		}
		if !foundVersion {
			group.Versions = append(group.Versions, metav1.GroupVersionForDiscovery{
				GroupVersion: resource.Group + "/" + resource.Version,
				Version:      resource.Version,
			})
		}
	}

	for _, group := range groupMap {
		sort.Slice(group.Versions, func(i, j int) bool {
			return group.Versions[i].Version < group.Versions[j].Version
		})
	}

	var groupKeys []string
	for key := range groupMap {
		groupKeys = append(groupKeys, key)
	}
	sort.Strings(groupKeys)

	response := &metav1.APIGroupList{}
	response.Kind = "APIGroupList"
	response.APIVersion = "v1"
	for _, groupKey := range groupKeys {
		group := groupMap[groupKey]
		// Assume preferred version is newest
		group.PreferredVersion = group.Versions[len(group.Versions)-1]
		response.Groups = append(response.Groups, *group)
	}
	return r.writeResponse(response)
}

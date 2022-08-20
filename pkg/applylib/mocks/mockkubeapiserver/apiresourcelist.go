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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// apiResourceList is a request for api discovery, such as GET /apis/resourcemanager.cnrm.cloud.google.com/v1beta1
type apiResourceList struct {
	baseRequest

	Group   string
	Version string
}

// Run serves the http request
func (req *apiResourceList) Run(s *MockKubeAPIServer) error {
	gv := schema.GroupVersion{
		Group:   req.Group,
		Version: req.Version,
	}
	response := &metav1.APIResourceList{}
	response.Kind = "APIResourceList"
	response.APIVersion = "v1"
	response.GroupVersion = gv.String()
	for _, resource := range s.schema.resources {
		if resource.Group != req.Group || resource.Version != req.Version {
			continue
		}
		response.APIResources = append(response.APIResources, resource.APIResource)
	}
	return req.writeResponse(response)
}

/*
Copyright 2018 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
)

type listFn func(openstack.OpenstackCloud, string) ([]*resources.Resource, error)

func ListResources(cloud openstack.OpenstackCloud, clusterName string) (map[string]*resources.Resource, error) {
	resourceTrackers := make(map[string]*resources.Resource)

	// TODO(lmb): Implement resource list

	return resourceTrackers, nil
}

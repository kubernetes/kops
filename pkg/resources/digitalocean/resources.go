/*
Copyright 2016 The Kubernetes Authors.

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

package digitalocean

import (
	"k8s.io/kops/pkg/resources"
	"k8s.io/kops/upup/pkg/fi"
)

type Resources struct {
	Cloud       fi.Cloud
	ClusterName string
}

// ListResources fetches all digitalocean resources into tracker.Resources
func (r *Resources) ListResources() (map[string]*resources.Resource, error) {
	return nil, nil
}

// DeleteResources deletes all resources passed in the form in tracker.Resources
func (r *Resources) DeleteResources(resources map[string]*resources.Resource) error {
	return nil
}

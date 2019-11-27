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

package registry

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	// Path for the user-specified cluster spec
	PathCluster = "config"
	// Path for completed cluster spec in the state store
	PathClusterCompleted = "cluster.spec"
)

func ConfigBase(c *api.Cluster) (vfs.Path, error) {
	if c.Spec.ConfigBase == "" {
		return nil, field.Required(field.NewPath("Spec", "ConfigBase"), "")
	}
	configBase, err := vfs.Context.BuildVfsPath(c.Spec.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("error parsing ConfigBase %q: %v", c.Spec.ConfigBase, err)
	}
	return configBase, nil
}

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

package registry

import (
	"fmt"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"strings"
)

// Path for the user-specified cluster spec
const PathCluster = "config"

// Path for completed cluster spec in the state store
const PathClusterCompleted = "cluster.spec"

func DeleteAllClusterState(basePath vfs.Path) error {
	paths, err := basePath.ReadTree()
	if err != nil {
		return fmt.Errorf("error listing files in state store: %v", err)
	}

	for _, path := range paths {
		relativePath, err := vfs.RelativePath(basePath, path)
		if err != nil {
			return err
		}
		if relativePath == "config" || relativePath == "cluster.spec" {
			continue
		}
		if strings.HasPrefix(relativePath, "addons/") {
			continue
		}
		if strings.HasPrefix(relativePath, "pki/") {
			continue
		}
		if strings.HasPrefix(relativePath, "secrets/") {
			continue
		}
		if strings.HasPrefix(relativePath, "instancegroup/") {
			continue
		}

		return fmt.Errorf("refusing to delete: unknown file found: %s", path)
	}

	for _, path := range paths {
		err = path.Remove()
		if err != nil {
			return fmt.Errorf("error deleting cluster file %s: %v", path, err)
		}
	}

	return nil
}

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

func SecretStore(c *api.Cluster) (fi.SecretStore, error) {
	configBase, err := ConfigBase(c)
	if err != nil {
		return nil, err
	}
	basedir := configBase.Join("secrets")
	return secrets.NewVFSSecretStore(basedir), nil
}

func KeyStore(c *api.Cluster) (fi.CAStore, error) {
	configBase, err := ConfigBase(c)
	if err != nil {
		return nil, err
	}
	basedir := configBase.Join("pki")
	return fi.NewVFSCAStore(basedir), nil
}

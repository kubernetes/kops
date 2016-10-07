package registry

import (
	"fmt"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/util/validation/field"
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

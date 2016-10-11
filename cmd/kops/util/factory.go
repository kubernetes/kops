package util

import (
	"fmt"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/util/validation/field"
)

type FactoryOptions struct {
	RegistryPath string
}

type Factory struct {
	options   *FactoryOptions
	clientset simple.Clientset
}

func NewFactory(options *FactoryOptions) *Factory {
	return &Factory{
		options: options,
	}
}

func (f *Factory) Clientset() (simple.Clientset, error) {
	if f.clientset == nil {
		registryPath := f.options.RegistryPath
		if registryPath == "" {
			return nil, field.Required(field.NewPath("RegistryPath"), "")
		}
		basePath, err := vfs.Context.BuildVfsPath(registryPath)
		if err != nil {
			return nil, fmt.Errorf("error building path for %q: %v", registryPath, err)
		}

		if !vfs.IsClusterReadable(basePath) {
			return nil, field.Invalid(field.NewPath("RegistryPath"), registryPath, "Not cloud-reachable - please use an S3 bucket")
		}

		f.clientset = vfsclientset.NewVFSClientset(basePath)
	}

	return f.clientset, nil
}

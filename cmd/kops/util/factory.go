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

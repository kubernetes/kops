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
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/util/pkg/vfs"

	"k8s.io/client-go/rest"
	kopsclient "k8s.io/kops/pkg/client/clientset_generated/clientset"

	// Register our APIs
	"github.com/golang/glog"
	_ "k8s.io/kops/pkg/apis/kops/install"
	"net/url"
	"strings"
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

const (
	STATE_ERROR = `Please set the --state flag or export KOPS_STATE_STORE.
A valid value follows the format s3://<bucket>.
A s3 bucket is required to store cluster state information.`

	INVALID_STATE_ERROR = `Unable to read state store s3 bucket.
Please use a valid s3 bucket uri when setting --state or KOPS_STATE_STORE evn var.
A valid value follows the format s3://<bucket>.
Trailing slash will be trimmed.`
)

func (f *Factory) Clientset() (simple.Clientset, error) {
	if f.clientset == nil {
		registryPath := f.options.RegistryPath
		if registryPath == "" {
			return nil, field.Required(field.NewPath("State Store"), STATE_ERROR)
		}

		// We recognize a `k8s` scheme; this might change in future so we won't document it yet
		// In practice nobody is going to hit this accidentally, so I don't think we need a feature flag.
		if strings.HasPrefix(registryPath, "k8s://") {
			u, err := url.Parse(registryPath)
			if err != nil {
				return nil, fmt.Errorf("Invalid kops server url: %q", registryPath)
			}

			config := &rest.Config{
				Host: u.Scheme + "://" + u.Host,
			}

			glog.Warning("Using insecure TLS")
			config.Insecure = true

			kopsClient, err := kopsclient.NewForConfig(config)
			if err != nil {
				return nil, fmt.Errorf("error building kops API client: %v", err)
			}

			f.clientset = &simple.RESTClientset{
				BaseURL: &url.URL{
					Scheme: "k8s",
					Host:   u.Host,
				},
				KopsClient: kopsClient.Kops(),
			}
		} else {
			basePath, err := vfs.Context.BuildVfsPath(registryPath)
			if err != nil {
				return nil, fmt.Errorf("error building path for %q: %v", registryPath, err)
			}

			if !vfs.IsClusterReadable(basePath) {
				return nil, field.Invalid(field.NewPath("State Store"), registryPath, INVALID_STATE_ERROR)
			}

			f.clientset = vfsclientset.NewVFSClientset(basePath)
		}
	}

	return f.clientset, nil
}

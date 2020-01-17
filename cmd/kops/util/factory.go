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

package util

import (
	"fmt"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	gceacls "k8s.io/kops/pkg/acls/gce"
	s3acls "k8s.io/kops/pkg/acls/s3"
	kopsclient "k8s.io/kops/pkg/client/clientset_generated/clientset"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/client/simple/api"
	"k8s.io/kops/pkg/client/simple/vfsclientset"
	"k8s.io/kops/util/pkg/vfs"
)

type FactoryOptions struct {
	RegistryPath string
}

type Factory struct {
	options   *FactoryOptions
	clientset simple.Clientset
}

func NewFactory(options *FactoryOptions) *Factory {
	gceacls.Register()
	s3acls.Register()

	return &Factory{
		options: options,
	}
}

const (
	STATE_ERROR = `Please set the --state flag or export KOPS_STATE_STORE.
For example, a valid value follows the format s3://<bucket>.
You can find the supported stores in https://kops.sigs.k8s.io/state.`

	INVALID_STATE_ERROR = `Unable to read state store.
Please use a valid state store when setting --state or KOPS_STATE_STORE env var.
For example, a valid value follows the format s3://<bucket>.
Trailing slash will be trimmed.`
)

func (f *Factory) Clientset() (simple.Clientset, error) {
	if f.clientset == nil {
		registryPath := f.options.RegistryPath
		klog.V(2).Infof("state store %s", registryPath)
		if registryPath == "" {
			return nil, field.Required(field.NewPath("State Store"), STATE_ERROR)
		}

		// We recognize a `k8s` scheme; this might change in future so we won't document it yet
		// In practice nobody is going to hit this accidentally, so I don't think we need a feature flag.
		if strings.HasPrefix(registryPath, "k8s://") {
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()

			configOverrides := &clientcmd.ConfigOverrides{}

			if registryPath == "k8s://" {
			} else {
				u, err := url.Parse(registryPath)
				if err != nil {
					return nil, fmt.Errorf("invalid kops server url: %q", registryPath)
				}
				configOverrides.CurrentContext = u.Host
			}

			kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
			config, err := kubeConfig.ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("error loading kubeconfig for %q", registryPath)
			}

			kopsClient, err := kopsclient.NewForConfig(config)
			if err != nil {
				return nil, fmt.Errorf("error building kops API client: %v", err)
			}

			f.clientset = &api.RESTClientset{
				BaseURL: &url.URL{
					Scheme: "k8s",
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

			// For kops CLI / controller, we do allow vfs list (unlike nodeup!)
			allowVFSList := true

			f.clientset = vfsclientset.NewVFSClientset(basePath, allowVFSList)
		}
		if strings.HasPrefix(registryPath, "file://") {
			klog.Warning("The local filesystem state store is not functional for running clusters")
		}
	}

	return f.clientset, nil
}

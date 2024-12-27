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
	"net/http"
	"net/url"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	gceacls "k8s.io/kops/pkg/acls/gce"
	"k8s.io/kops/pkg/apis/kops"
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

	vfsContext *vfs.VFSContext

	// mutex protects access to the clusters map
	mutex sync.Mutex
	// clusters holds REST connection configuration for connecting to clusters
	clusters map[string]*clusterInfo
}

// clusterInfo holds REST connection configuration for connecting to a cluster
type clusterInfo struct {
	clusterName string

	cachedHTTPClient    *http.Client
	cachedRESTConfig    *rest.Config
	cachedDynamicClient dynamic.Interface
}

func NewFactory(options *FactoryOptions) *Factory {
	gceacls.Register()

	if options == nil {
		options = &FactoryOptions{}
	}

	return &Factory{
		options:  options,
		clusters: make(map[string]*clusterInfo),
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

func (f *Factory) KopsClient() (simple.Clientset, error) {
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

			f.clientset = api.NewRESTClientset(
				f.VFSContext(),
				&url.URL{
					Scheme: "k8s",
				},
				kopsClient.Kops(),
			)
		} else {
			basePath, err := f.VFSContext().BuildVfsPath(registryPath)
			if err != nil {
				return nil, fmt.Errorf("error building path for %q: %v", registryPath, err)
			}

			if !vfs.IsClusterReadable(basePath) {
				return nil, field.Invalid(field.NewPath("State Store"), registryPath, INVALID_STATE_ERROR)
			}

			f.clientset = vfsclientset.NewVFSClientset(f.VFSContext(), basePath)
		}
		if strings.HasPrefix(registryPath, "file://") {
			klog.Warning("The local filesystem state store is not functional for running clusters")
		}
	}

	return f.clientset, nil
}

// KopsStateStore returns the configured KOPS_STATE_STORE in use
func (f *Factory) KopsStateStore() string {
	return f.options.RegistryPath
}

func (f *Factory) getClusterInfo(clusterName string) *clusterInfo {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if clusterInfo, ok := f.clusters[clusterName]; ok {
		return clusterInfo
	}
	clusterInfo := &clusterInfo{}
	f.clusters[clusterName] = clusterInfo
	return clusterInfo
}

func (f *Factory) RESTConfig(cluster *kops.Cluster) (*rest.Config, error) {
	clusterInfo := f.getClusterInfo(cluster.ObjectMeta.Name)
	return clusterInfo.RESTConfig()
}

func (f *clusterInfo) RESTConfig() (*rest.Config, error) {
	if f.cachedRESTConfig == nil {
		// Get the kubeconfig from the context

		clientGetter := genericclioptions.NewConfigFlags(true)
		if f.clusterName != "" {
			contextName := f.clusterName
			clientGetter.Context = &contextName
		}

		restConfig, err := clientGetter.ToRESTConfig()
		if err != nil {
			return nil, fmt.Errorf("loading kubecfg settings for %q: %w", f.clusterName, err)
		}

		restConfig.UserAgent = "kops"
		restConfig.Burst = 50
		restConfig.QPS = 20
		f.cachedRESTConfig = restConfig
	}
	return f.cachedRESTConfig, nil
}

func (f *Factory) HTTPClient(cluster *kops.Cluster) (*http.Client, error) {
	clusterInfo := f.getClusterInfo(cluster.ObjectMeta.Name)
	return clusterInfo.HTTPClient()
}

func (f *clusterInfo) HTTPClient() (*http.Client, error) {
	if f.cachedHTTPClient == nil {
		restConfig, err := f.RESTConfig()
		if err != nil {
			return nil, err
		}
		httpClient, err := rest.HTTPClientFor(restConfig)
		if err != nil {
			return nil, fmt.Errorf("building http client: %w", err)
		}
		f.cachedHTTPClient = httpClient
	}
	return f.cachedHTTPClient, nil
}

// DynamicClient returns a dynamic client
func (f *Factory) DynamicClient(clusterName string) (dynamic.Interface, error) {
	clusterInfo := f.getClusterInfo(clusterName)
	return clusterInfo.DynamicClient()
}

func (f *clusterInfo) DynamicClient() (dynamic.Interface, error) {
	if f.cachedDynamicClient == nil {
		restConfig, err := f.RESTConfig()
		if err != nil {
			return nil, err
		}

		httpClient, err := f.HTTPClient()
		if err != nil {
			return nil, err
		}

		dynamicClient, err := dynamic.NewForConfigAndClient(restConfig, httpClient)
		if err != nil {
			return nil, fmt.Errorf("building dynamic client: %w", err)
		}
		f.cachedDynamicClient = dynamicClient
	}
	return f.cachedDynamicClient, nil
}

func (f *Factory) VFSContext() *vfs.VFSContext {
	if f.vfsContext == nil {
		// TODO vfs.NewVFSContext()
		f.vfsContext = vfs.Context
	}
	return f.vfsContext
}

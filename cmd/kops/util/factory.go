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
	"context"
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
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi/cloudup"
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
	factory *Factory
	cluster *kops.Cluster

	cachedHTTPClient    *http.Client
	cachedRESTConfig    *rest.Config
	cachedDynamicClient dynamic.Interface
	kubeconfig.CreateKubecfgOptions
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

func (f *Factory) getClusterInfo(cluster *kops.Cluster, options kubeconfig.CreateKubecfgOptions) *clusterInfo {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	key := cluster.ObjectMeta.Name
	if clusterInfo, ok := f.clusters[key]; ok {
		return clusterInfo
	}
	clusterInfo := &clusterInfo{
		factory:              f,
		cluster:              cluster,
		CreateKubecfgOptions: options,
	}
	f.clusters[key] = clusterInfo
	return clusterInfo
}

func (f *Factory) RESTConfig(ctx context.Context, cluster *kops.Cluster, options kubeconfig.CreateKubecfgOptions) (*rest.Config, error) {
	clusterInfo := f.getClusterInfo(cluster, options)
	return clusterInfo.RESTConfig(ctx)
}

func (f *clusterInfo) RESTConfig(ctx context.Context) (*rest.Config, error) {
	if f.cachedRESTConfig == nil {
		restConfig, err := f.factory.buildRESTConfig(ctx, f.cluster, f.CreateKubecfgOptions)
		if err != nil {
			return nil, err
		}

		configureRESTConfig(restConfig)

		f.cachedRESTConfig = restConfig
	}
	return f.cachedRESTConfig, nil
}

func configureRESTConfig(restConfig *rest.Config) {
	restConfig.UserAgent = "kops"
	restConfig.Burst = 50
	restConfig.QPS = 20
}

func (f *Factory) HTTPClient(restConfig *rest.Config) (*http.Client, error) {
	return rest.HTTPClientFor(restConfig)
}

func (f *clusterInfo) HTTPClient(restConfig *rest.Config) (*http.Client, error) {
	if f.cachedHTTPClient == nil {
		httpClient, err := rest.HTTPClientFor(restConfig)
		if err != nil {
			return nil, fmt.Errorf("building http client: %w", err)
		}
		f.cachedHTTPClient = httpClient
	}
	return f.cachedHTTPClient, nil
}

// DynamicClient returns a dynamic client
func (f *Factory) DynamicClient(ctx context.Context, cluster *kops.Cluster, options kubeconfig.CreateKubecfgOptions) (dynamic.Interface, error) {
	clusterInfo := f.getClusterInfo(cluster, options)
	restConfig, err := clusterInfo.RESTConfig(ctx)
	if err != nil {
		return nil, err
	}
	return clusterInfo.DynamicClient(restConfig)
}

func (f *clusterInfo) DynamicClient(restConfig *rest.Config) (dynamic.Interface, error) {
	if f.cachedDynamicClient == nil {
		httpClient, err := f.HTTPClient(restConfig)
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

func (f *Factory) buildRESTConfig(ctx context.Context, cluster *kops.Cluster, options kubeconfig.CreateKubecfgOptions) (*rest.Config, error) {
	clientset, err := f.KopsClient()
	if err != nil {
		return nil, err
	}

	keyStore, err := clientset.KeyStore(cluster)
	if err != nil {
		return nil, err
	}

	secretStore, err := clientset.SecretStore(cluster)
	if err != nil {
		return nil, err
	}

	cloud, err := cloudup.BuildCloud(cluster)
	if err != nil {
		return nil, err
	}

	// backwards compatibility
	if options.Admin == 0 {
		options.Admin = kubeconfig.DefaultKubecfgAdminLifetime
	}

	if options.UseKubeconfig {
		// Get the kubeconfig from the context
		klog.Infof("--use-kubeconfig is set; loading connectivity information from kubeconfig (instead of generating it)")

		clusterName := cluster.ObjectMeta.Name

		clientGetter := genericclioptions.NewConfigFlags(true)
		contextName := clusterName
		clientGetter.Context = &contextName

		restConfig, err := clientGetter.ToRESTConfig()
		if err != nil {
			return nil, fmt.Errorf("loading kubecfg settings for %q: %w", clusterName, err)
		}

		configureRESTConfig(restConfig)

		if options.OverrideAPIServer != "" {
			klog.Infof("overriding API server with %q", options.OverrideAPIServer)
			restConfig.Host = options.OverrideAPIServer
		}

		return restConfig, nil
	}

	conf, err := kubeconfig.BuildKubecfg(
		ctx,
		cluster,
		keyStore,
		secretStore,
		cloud,
		options,
		f.KopsStateStore())
	if err != nil {
		return nil, err
	}

	return conf.ToRESTConfig()
}

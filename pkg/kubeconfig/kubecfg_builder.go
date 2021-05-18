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

package kubeconfig

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
)

// Config writes cluster configuration to a kubeconfig file, typically ~/.kube/config
type Config struct {
	// CurrentContext is the new CurrentContext to set.
	CurrentContext string

	// Contexts contains Context data to insert.
	// Context data replaces any existing values with the same key.
	Contexts map[string]*clientcmdapi.Context

	// Clusters contains Cluster data to insert.
	// Cluster data is merged into any existing values with the same key.
	Clusters map[string]*clientcmdapi.Cluster

	// AuthInfos contains AuthInfo data to insert.
	// AuthInfo data replaces any existing values with the same key.
	AuthInfos map[string]*clientcmdapi.AuthInfo
}

// DeleteClusterConfig removes the configuration for a cluster, done to cleanup ~/.kube/config after kops delete cluster.
func DeleteClusterConfig(configAccess clientcmd.ConfigAccess, context string) error {
	if context != "" {
		return fmt.Errorf("context must be provided")
	}

	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("error loading kubeconfig: %w", err)
	}

	if config == nil || clientcmdapi.IsConfigEmpty(config) {
		klog.V(2).Info("kubeconfig is empty")
		return nil
	}

	delete(config.Clusters, context)
	delete(config.AuthInfos, context)
	delete(config.AuthInfos, fmt.Sprintf("%s-basic-auth", context))
	delete(config.Contexts, context)

	if config.CurrentContext == context {
		config.CurrentContext = ""
	}

	if err := clientcmd.ModifyConfig(configAccess, *config, false); err != nil {
		return fmt.Errorf("error writing kubeconfig: %w", err)
	}

	fmt.Printf("Deleted kubectl config for %s\n", context)
	return nil
}

// WriteKubecfg adds the configuration to the kube configuration specified by configAccess, typically ~/.kube/config
func (b *Config) WriteKubecfg(configAccess clientcmd.ConfigAccess) error {
	config, err := configAccess.GetStartingConfig()
	if err != nil {
		return fmt.Errorf("error reading kubeconfig: %w", err)
	}

	if config == nil {
		config = &clientcmdapi.Config{}
	}

	for k, want := range b.Clusters {
		cluster := config.Clusters[k]
		if cluster == nil {
			cluster = clientcmdapi.NewCluster()
		}
		cluster.Server = want.Server
		cluster.CertificateAuthorityData = want.CertificateAuthorityData

		if config.Clusters == nil {
			config.Clusters = make(map[string]*clientcmdapi.Cluster)
		}
		config.Clusters[k] = cluster
	}

	for k, authInfo := range b.AuthInfos {
		if config.AuthInfos == nil {
			config.AuthInfos = make(map[string]*clientcmdapi.AuthInfo)
		}
		config.AuthInfos[k] = authInfo

		// If we have a bearer token, also create a credential entry with basic auth
		// so that it is easy to discover the basic auth password for your cluster
		// to use in a web browser.
		// This is deprecated behaviour along with the deprecation of basic auth.
		// We can likely remove when we no longer support basic auth.
		if k == b.CurrentContext && authInfo.Username != "" && authInfo.Password != "" {
			baiName := k + "-basic-auth"
			bai := config.AuthInfos[baiName]

			bai.Username = authInfo.Username
			bai.Password = authInfo.Password

			config.AuthInfos[baiName] = bai
		}
	}

	for k, context := range b.Contexts {
		// Verify AuthInfo, in particular for manually passed user configs.
		if context.AuthInfo != "" && config.AuthInfos[context.AuthInfo] == nil {
			return fmt.Errorf("could not find user %q", context.AuthInfo)
		}

		if config.Contexts == nil {
			config.Contexts = make(map[string]*clientcmdapi.Context)
		}
		config.Contexts[k] = context
	}

	if b.CurrentContext == "" {
		config.CurrentContext = b.CurrentContext
	}

	if err := clientcmd.ModifyConfig(configAccess, *config, true); err != nil {
		return err
	}

	fmt.Printf("kops has set your kubectl context to %s\n", b.CurrentContext)
	return nil
}

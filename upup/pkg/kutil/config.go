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

package kutil

import (
	"fmt"

	"github.com/golang/glog"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func NewClientConfig(clientConfig *restclient.Config, namespace string) clientcmd.ClientConfig {
	c := &SimpleClientConfig{clientConfig: clientConfig, namespace: namespace}
	return c
}

// SimpleClientConfig is used to make it easy to get an api server client
type SimpleClientConfig struct {
	clientConfig *restclient.Config
	namespace    string
}

var _ clientcmd.ClientConfig = &SimpleClientConfig{}

// RawConfig returns the merged result of all overrides
func (c *SimpleClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, fmt.Errorf("SimpleClientConfig::RawConfig not implemented")
}

// ClientConfig returns a complete client config
func (c *SimpleClientConfig) ClientConfig() (*restclient.Config, error) {
	return c.clientConfig, nil
}

// Namespace returns the namespace resulting from the merged
// result of all overrides and a boolean indicating if it was
// overridden
func (c *SimpleClientConfig) Namespace() (string, bool, error) {
	return c.namespace, false, nil
}

// ConfigAccess returns the rules for loading/persisting the config.
func (c *SimpleClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	glog.Fatalf("SimpleClientConfig::RawConfig not implemented")
	return nil
}

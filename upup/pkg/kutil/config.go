package kutil

import (
	"k8s.io/kubernetes/pkg/client/restclient"
	"fmt"
	"k8s.io/kubernetes/pkg/client/unversioned/clientcmd"
	"github.com/golang/glog"
	clientcmdapi "k8s.io/kubernetes/pkg/client/unversioned/clientcmd/api"
)

func NewClientConfig(clientConfig *restclient.Config, namespace string) clientcmd.ClientConfig {
	c := &SimpleClientConfig{clientConfig: clientConfig, namespace: namespace}
	return c
}

// ClientConfig is used to make it easy to get an api server client
type SimpleClientConfig struct {
	clientConfig *restclient.Config
	namespace string
}

var _ clientcmd.ClientConfig = &SimpleClientConfig{}

// RawConfig returns the merged result of all overrides
func (c*SimpleClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, fmt.Errorf("SimpleClientConfig::RawConfig not implemented")
}

// ClientConfig returns a complete client config
func (c*SimpleClientConfig) ClientConfig() (*restclient.Config, error) {
	return c.clientConfig, nil
}
// Namespace returns the namespace resulting from the merged
// result of all overrides and a boolean indicating if it was
// overridden
func (c*SimpleClientConfig) Namespace() (string, bool, error) {
	return c.namespace, false, nil
}
// ConfigAccess returns the rules for loading/persisting the config.
func (c*SimpleClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	glog.Fatalf("SimpleClientConfig::RawConfig not implemented")
	return nil
}
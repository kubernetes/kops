/*
Copyright 2021 The Kubernetes Authors.

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

package model

import (
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KonnectivityConfigPath is the path where we store the kube-apiserver egress-selector-configuration file when using konnectivity
const KonnectivityConfigPath = "/etc/kubernetes/konnectivity-server/egress-selector-configuration.yaml"

// KonnectivityServerBuilder installs konnectivity-agent
type KonnectivityServerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KonnectivityServerBuilder{}

const EgressSelectorConfiguration = `
apiVersion: apiserver.k8s.io/v1beta1
kind: EgressSelectorConfiguration
egressSelections:
# Since we want to control the egress traffic to the cluster, we use the
# "cluster" as the name. Other supported values are "etcd", and "master".
- name: cluster
  connection:
    # This controls the protocol between the API Server and the Konnectivity
    # server. Supported values are "GRPC" and "HTTPConnect". There is no
    # end user visible difference between the two modes. You need to set the
    # Konnectivity server to work in the same mode.
    proxyProtocol: GRPC
    transport:
      # This controls what transport the API Server uses to communicate with the
      # Konnectivity server. UDS is recommended if the Konnectivity server
      # locates on the same machine as the API Server. You need to configure the
      # Konnectivity server to listen on the same UDS socket.
      # The other supported transport is "tcp". You will need to set up TLS 
      # config to secure the TCP transport.
      uds:
        udsName: /etc/kubernetes/konnectivity-server/uds/konnectivity-server.socket
`

// Build is responsible for building the kube-proxy manifest
// @TODO we should probably change this to a daemonset in the future and follow the kubeadm path
func (b *KonnectivityServerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.HasAPIServer {
		return nil
	}

	// var kubeAPIServer kops.KubeAPIServerConfig
	// if b.NodeupConfig.APIServerConfig.KubeAPIServer != nil {
	// 	kubeAPIServer = *b.NodeupConfig.APIServerConfig.KubeAPIServer
	// }

	if b.UseKonnectivity() {
		c.AddTask(&nodetasks.File{
			Path:     KonnectivityConfigPath,
			Contents: fi.NewStringResource(EgressSelectorConfiguration),
			Type:     nodetasks.FileType_File,
		})
	}

	return nil
}

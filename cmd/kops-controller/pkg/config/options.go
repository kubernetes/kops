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

package config

import (
	"k8s.io/kops/cmd/kops-controller/pkg/nodebootstrap"
)

type Options struct {
	Cloud      string `json:"cloud,omitempty"`
	ConfigBase string `json:"configBase,omitempty"`

	// GRPC configures our GRPC endpoint for in-cluster services
	GRPC *GRPCOptions `json:"grpc,omitempty"`

	// NodeBootstrapService configures how we verify nodes for joining
	NodeBootstrapService *nodebootstrap.Options `json:"nodeBootstrapService,omitempty"`
}

func (o *Options) PopulateDefaults() {
	o.GRPC = &GRPCOptions{}
	o.GRPC.PopulateDefaults()
	o.NodeBootstrapService = &nodebootstrap.Options{}
	o.NodeBootstrapService.PopulateDefaults()
}

type GRPCOptions struct {
	// Listen in the network endpoint (ip and port) we should listen on
	Listen string `json:"listen,omitempty"`

	// ServerKeyPath is the path to our TLS serving private key
	ServerKeyPath string `json:"serverKeyPath,omitempty"`
	// ServerCertificatePath is the path to our TLS serving certificate
	ServerCertificatePath string `json:"serverCertificatePath,omitempty"`
}

func (o *GRPCOptions) PopulateDefaults() {
}

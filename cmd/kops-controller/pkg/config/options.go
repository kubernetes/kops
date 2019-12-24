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

type Options struct {
	Cloud      string          `json:"cloud,omitempty"`
	ConfigBase string          `json:"configBase,omitempty"`
	Cluster    *ClusterOptions `json:"cluster,omitempty"`

	// GRPC holds the options for the grpc server, used for node configuration
	GRPC *GRPCOptions `json:"grpc,omitempty"`
}

type ClusterOptions struct {
	Enabled bool `json:"enabled,omitempty"`
}

type GRPCOptions struct {
	Listen string `json:"listen,omitempty"`

	ClientEndpoint string `json:"clientEndpoint,omitempty"`
	CACertPath     string `json:"caCert,omitempty"`

	ServerKeyPath  string `json:"serverKey,omitempty"`
	ServerCertPath string `json:"serverCert,omitempty"`
}

func (o *GRPCOptions) PopulateDefaults() {
}

func (o *Options) PopulateDefaults() {
	if o.GRPC == nil {
		o.GRPC = &GRPCOptions{}
	}
	o.GRPC.PopulateDefaults()
}

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

package client

import "time"

// Config is the configuration for the service
type Config struct {
	// Authorizer is the name of the verifier to use
	Authorizer string
	// Interval is the pause between failed attempts
	Interval time.Duration
	// KubeAPI is the url for the kubernetes api
	KubeAPI string
	// KubeConfigPath is the location to write the bootstrap token config
	KubeConfigPath string
	// NodeURL is the url for the node authozier service
	NodeURL string
	// Timeout is the time will are willing to wait
	Timeout time.Duration
	// TLSCertPath is the path to the server TLS certificate
	TLSCertPath string
	// TLSClientCAPath is the path to a certificate authority
	TLSClientCAPath string
	// TLSPrivateKeyPath is the path to the private key
	TLSPrivateKeyPath string
	// Verbose indicate verbose logging
	Verbose bool
}

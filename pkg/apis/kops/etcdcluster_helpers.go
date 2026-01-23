/*
Copyright 2026 The Kubernetes Authors.

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

package kops

// IsClientTLSEnabled returns whether client TLS is enabled for this etcd cluster.
// Defaults to true for security.
func (e *EtcdClusterSpec) IsClientTLSEnabled() bool {
	if e.ClientTLSEnabled == nil {
		return true // Default to HTTPS for security
	}
	return *e.ClientTLSEnabled
}

// GetClientScheme returns the URL scheme (http or https) for client connections to etcd.
func (e *EtcdClusterSpec) GetClientScheme() string {
	if e.IsClientTLSEnabled() {
		return "https"
	}
	return "http"
}

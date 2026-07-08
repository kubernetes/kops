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

package dns

const (
	// PlaceholderIP is the IP kOps writes into DNS records before the control plane comes up.
	// Clients treat a record resolving to this IP as not yet available.
	// It is from TEST-NET-3: https://en.wikipedia.org/wiki/Reserved_IP_addresses
	PlaceholderIP = "203.0.113.123"
	// PlaceholderIPv6 is the IPv6 equivalent of PlaceholderIP.
	PlaceholderIPv6 = "fd00:dead:add::"
)

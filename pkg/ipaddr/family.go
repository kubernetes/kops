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

package ipaddr

import (
	"fmt"
	"net"
	"strings"
)

// Family is the IP address family (type) for an address; typically IPV4 or IPV6
type Family string

const (
	FamilyIPV4 Family = "ipv4"
	FamilyIPV6 Family = "ipv6"
)

// GetFamily returns the family for an IP.
// IPv4-in-IPv6 addresses are treated as IPv6.
func GetFamily(ip string) (Family, error) {
	// Verify it is an IP (and not a host:port)
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return "", fmt.Errorf("cannot parse IP %q", ip)
	}

	// Check for ipv6 using colons; this is hacky but means we classify IPv4-in-IPv6 as v6
	// Context: https://github.com/golang/go/issues/37921
	family := FamilyIPV4
	if strings.Contains(ip, ":") {
		family = FamilyIPV6
	}

	return family, nil
}

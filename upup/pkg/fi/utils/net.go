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

package utils

import (
	"fmt"
	"math/big"
	"net"
	"regexp"
	"strconv"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
)

// IsIPv4IP checks if a string is a valid IPv4 IP.
func IsIPv4IP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}

	// Must convert to IPv4
	if ip.To4() == nil {
		return false
	}
	// Must NOT contain ":"
	if strings.Contains(s, ":") {
		return false
	}

	return true
}

// IsIPv6IP checks if a string is a valid IPv6 IP.
func IsIPv6IP(s string) bool {
	ip := net.ParseIP(s)
	if ip == nil {
		return false
	}

	// Must NOT convert to IPv4
	if ip.To4() != nil {
		return false
	}

	return true
}

// IsIPv4CIDR checks if a string is a valid IPv4 CIDR.
func IsIPv4CIDR(s string) bool {
	ip, _, err := net.ParseCIDR(s)
	if err != nil {
		return false
	}

	// Must convert to IPv4
	if ip.To4() == nil {
		return false
	}
	// Must NOT contain ":"
	if strings.Contains(s, ":") {
		return false
	}

	return true
}

// IsIPv6CIDR checks if a string is a valid IPv6 CIDR.
func IsIPv6CIDR(s string) bool {
	ip, _, err := net.ParseCIDR(s)
	if err != nil {
		return false
	}

	// Must NOT convert to IPv4
	if ip.To4() != nil {
		return false
	}

	return true
}

// ParseCIDRNotation parses a string in the format "/<newSize>#<netNum>"
// and returns the values of <newSize> and <netNum>.
func ParseCIDRNotation(subnet string) (int, int64, error) {
	re := regexp.MustCompile(`^/(\d+)#([a-f0-9]+)$`)
	s := re.FindStringSubmatch(subnet)
	if len(s) != 3 {
		return 0, 0, fmt.Errorf("unable to parse CIDR subnet string %q using %q", subnet, re)
	}

	newSize, err := strconv.Atoi(s[1])
	if err != nil {
		return 0, 0, fmt.Errorf("unable to convert CIDR subnet new bits to int: %q: %v", s[1], err)
	}

	netNum, err := strconv.ParseInt(s[2], 16, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to convert CIDR subnet net num to int: %q: %v", s[2], err)
	}

	return newSize, netNum, nil
}

// CIDRSubnet calculates a subnet address within given IP network address prefix.
// Inspired by the Terraform implementation of the "cidrsubnet" function
// https://www.terraform.io/docs/language/functions/cidrsubnet.html
func CIDRSubnet(prefix string, newSize int, netNum int64) (string, error) {
	_, baseCIDR, err := net.ParseCIDR(prefix)
	if err != nil {
		return "", fmt.Errorf("unable to parse CIDR for %q: %v", prefix, err)
	}

	oldSize, totalSize := baseCIDR.Mask.Size()
	if oldSize == 0 && totalSize == 0 {
		return "", fmt.Errorf("unable to calculate CIDR mask size for %q: %q", prefix, baseCIDR.Mask)
	}

	newNetwork, err := cidr.SubnetBig(baseCIDR, newSize-oldSize, big.NewInt(netNum))
	if err != nil {
		return "", fmt.Errorf("unable to calculate subnet CIDR for %q -> /%d#%d : %v", prefix, newSize, netNum, err)
	}

	return newNetwork.String(), nil
}

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

package awsbootstrap

import "strings"

// PrivateDNSName returns the DNS name that EC2 generates for an IP-named instance from its primary
// private IPv4 address: ip-a-b-c-d.ec2.internal in us-east-1 and
// ip-a-b-c-d.<region>.compute.internal in all other regions. With IP-based node names, both nodeup
// and kops-controller derive the node name with this formula, guaranteeing that the name the node
// registers with matches the name its certificates are issued for.
func PrivateDNSName(privateIPv4, region string) string {
	domain := region + ".compute.internal"
	if region == "us-east-1" {
		domain = "ec2.internal"
	}
	return "ip-" + strings.ReplaceAll(privateIPv4, ".", "-") + "." + domain
}

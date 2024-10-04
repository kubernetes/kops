/*
Copyright 2024 The Kubernetes Authors.

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

package networking

import (
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

// AmazonVPCRoutedENIBuilder writes the Amazon VPC CNI configuration
type AmazonVPCRoutedENIBuilder struct {
	*model.NodeupModelContext
}

var _ fi.NodeupModelBuilder = &AmazonVPCRoutedENIBuilder{}

// Build is responsible for configuring the network cni
func (b *AmazonVPCRoutedENIBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if b.NodeupConfig.Networking.AmazonVPC == nil {
		return nil
	}

	if b.Distribution == distributions.DistributionAmazonLinux2023 {
		// Mask udev triggers installed by amazon-ec2-net-utils package
		// Create an empty file 99-vpc-policy-routes.rules
		c.AddTask(&nodetasks.File{
			Path:     "/etc/udev/rules.d/99-vpc-policy-routes.rules",
			Contents: fi.NewStringResource(""),
			Type:     nodetasks.FileType_File,
			OnChangeExecute: [][]string{
				{"udevadm", "control", "--reload-rules"},
				{"udevadm", "trigger"},
			},
		})

		// Make systemd-networkd ignore foreign settings, else it may
		// unexpectedly delete IP rules and routes added by CNI
		contents := `
# Do not clobber any routes or rules added by CNI.
[Network]
ManageForeignRoutes=no
ManageForeignRoutingPolicyRules=no
`
		c.AddTask(&nodetasks.File{
			Path:            "/usr/lib/systemd/networkd.conf.d/80-release.conf",
			Contents:        fi.NewStringResource(contents),
			Type:            nodetasks.FileType_File,
			OnChangeExecute: [][]string{{"systemctl", "restart", "systemd-networkd"}},
		})
	}

	// Running Amazon VPC CNI on Ubuntu 22.04 or any version of al2023 requires
	// setting MACAddressPolicy to `none` (ref: https://github.com/aws/amazon-vpc-cni-k8s/issues/2103
	// & https://github.com/kubernetes/kops/issues/16255)
	if (b.Distribution.IsUbuntu() && b.Distribution.Version() == 22.04) ||
		b.Distribution == distributions.DistributionAmazonLinux2023 {
		contents := `
[Match]
OriginalName=*
[Link]
NamePolicy=keep kernel database onboard slot path
AlternativeNamesPolicy=database onboard slot path
MACAddressPolicy=none
`

		// Copy all the relevant entries and replace the one that contains MACAddressPolicy= with MACAddressPolicy=none
		c.AddTask(&nodetasks.File{
			Path:            "/etc/systemd/network/99-default.link",
			Contents:        fi.NewStringResource(contents),
			Type:            nodetasks.FileType_File,
			OnChangeExecute: [][]string{{"systemctl", "restart", "systemd-networkd"}},
		})

	}
	return nil
}

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
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

// maskEC2NetUtilsUdevRules creates an empty /etc/udev/rules.d/99-vpc-policy-routes.rules
// to shadow the system udev rules, preventing policy-routes@ services from running.
// The policy-routes@ service adds secondary IPs (including CNI-allocated pod IPs) as
// /32 addresses on the host interface, populating the kernel's local routing table
// and breaking pod networking.
//
// Masking the udev rules alone is not sufficient: by the time nodeup runs,
// policy-routes@ens5.service is already active (started at boot for the primary ENI)
// and its refresh-policy-routes@ timer re-creates ec2net_alias.conf drop-ins every
// ~60 seconds with fresh secondary IPs from IMDS. We must also stop the running
// services, disable the timers, remove the drop-in files that add pod IPs as /32
// addresses, and restart systemd-networkd to flush the stale address assignments.
//
// AL2023 only.
func maskEC2NetUtilsUdevRules(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if dist != distributions.DistributionAmazonLinux2023 {
		return
	}

	c.AddTask(&nodetasks.File{
		Path:     "/etc/udev/rules.d/99-vpc-policy-routes.rules",
		Contents: fi.NewStringResource(""),
		Type:     nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			// Reload udev rules so the empty mask file takes effect for future ENI attach events.
			{"udevadm", "control", "--reload-rules"},
			{"udevadm", "trigger"},
			// Stop already-running policy-routes services and their refresh timers.
			// These were started at boot (before nodeup) by the system udev rules for
			// the primary ENI and would otherwise continue adding pod IPs to interfaces.
			{"bash", "-c", "systemctl stop 'policy-routes@*.service' 'refresh-policy-routes@*.service' 'refresh-policy-routes@*.timer' 2>/dev/null; true"},
			{"bash", "-c", "systemctl disable 'policy-routes@*.service' 'refresh-policy-routes@*.timer' 2>/dev/null; true"},
			// Remove ec2net_alias.conf drop-ins that added secondary IPs (pod IPs) as
			// /32 addresses on host interfaces, which populated the kernel's local routing
			// table and caused local delivery instead of forwarding through lxc veths.
			{"bash", "-c", "rm -f /run/systemd/network/*/ec2net_alias.conf"},
			// Restart systemd-networkd to flush the /32 address assignments and local
			// routing table entries that were applied from the now-removed drop-ins.
			{"systemctl", "restart", "systemd-networkd"},
		},
	})
}

// disableManageForeignRoutes configures systemd-networkd to not remove foreign routes/rules
// added by CNI. Without this, systemd-networkd may unexpectedly delete IP rules and routes.
// AL2023, Ubuntu 22.04+, and Debian 12+.
func disableManageForeignRoutes(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if !((dist.IsUbuntu() && dist.Version() >= 22.04) ||
		(dist.IsDebian() && dist.Version() >= 12) ||
		dist == distributions.DistributionAmazonLinux2023) {
		return
	}

	contents := `
# Do not clobber any routes or rules added by CNI.
[Network]
ManageForeignRoutes=no
ManageForeignRoutingPolicyRules=no
`
	c.AddTask(&nodetasks.File{
		Path:            "/usr/lib/systemd/networkd.conf.d/40-disable-manage-foreign-routes.conf",
		Contents:        fi.NewStringResource(contents),
		Type:            nodetasks.FileType_File,
		OnChangeExecute: [][]string{{"systemctl", "restart", "systemd-networkd"}},
	})
}

// setMACAddressPolicyNone prevents systemd-networkd from assigning predictable MAC-based
// names to ENIs, which can interfere with CNI interface management.
// AL2023, Ubuntu 22.04+, and Debian 12+.
// ref: https://github.com/aws/amazon-vpc-cni-k8s/issues/2103
// ref: https://github.com/aws/amazon-vpc-cni-k8s/issues/2839
// ref: https://github.com/kubernetes/kops/issues/16255
func setMACAddressPolicyNone(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if !((dist.IsUbuntu() && dist.Version() >= 22.04) ||
		(dist.IsDebian() && dist.Version() >= 12) ||
		dist == distributions.DistributionAmazonLinux2023) {
		return
	}

	contents := `
[Match]
OriginalName=*
[Link]
NamePolicy=keep kernel database onboard slot path
AlternativeNamesPolicy=database onboard slot path
MACAddressPolicy=none
`
	c.AddTask(&nodetasks.File{
		Path:            "/etc/systemd/network/99-default.link",
		Contents:        fi.NewStringResource(contents),
		Type:            nodetasks.FileType_File,
		OnChangeExecute: [][]string{{"systemctl", "restart", "systemd-networkd"}},
	})
}

// markSecondaryENIsUnmanaged tells systemd-networkd to ignore secondary ENIs (ens6+).
// Without this, systemd-networkd fully manages secondary ENIs via DHCP, creating
// competing routes that interfere with CNI networking.
// AL2023 and Debian 12+.
// ref: https://github.com/aws/amazon-vpc-cni-k8s/issues/3524
func markSecondaryENIsUnmanaged(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if !(dist == distributions.DistributionAmazonLinux2023 ||
		(dist.IsDebian() && dist.Version() >= 12)) {
		return
	}

	contents := `
[Match]
Name=ens[6-9]* ens[1-9][0-9]*

[Link]
Unmanaged=yes
`
	c.AddTask(&nodetasks.File{
		Path:            "/etc/systemd/network/10-eni-secondary.network",
		Contents:        fi.NewStringResource(contents),
		Type:            nodetasks.FileType_File,
		OnChangeExecute: [][]string{{"systemctl", "restart", "systemd-networkd"}},
	})
}

// disableCloudInitNetworkHotplug prevents cloud-init from reconfiguring the network
// when ENIs are attached, which breaks CNI networking.
// Ubuntu 24.04+ and Debian 12+.
// ref: https://github.com/kubernetes/kops/issues/17881
func disableCloudInitNetworkHotplug(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if !((dist.IsUbuntu() && dist.Version() >= 24.04) ||
		(dist.IsDebian() && dist.Version() >= 12)) {
		return
	}

	contents := `# Disable cloud-init network hotplug to prevent interference with CNI ENI management.
# See: https://github.com/kubernetes/kops/issues/17881
updates:
  network:
    when: [boot-new-instance]
`
	c.AddTask(&nodetasks.File{
		Path:     "/etc/cloud/cloud.cfg.d/99-disable-network-hotplug.cfg",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})
}

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
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"

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

// markSecondaryENIsUnmanaged causes systemd-networkd to ignore the secondary ENIs that use the
// "ena" driver. Without this file, systemd-networkd starts DHCP on these ENIs and makes routes
// that do not agree with the CNI routing.
// AL2023 and Debian 12+.
// ref: https://github.com/aws/amazon-vpc-cni-k8s/issues/3524
//
// It is not possible to find the secondary ENIs by their names. The names change with the
// instance family and the systemd naming scheme. Examples: the primary network interface is
// "ens5" on most Nitro instances, "ens34" on Graviton4 instances (c8g etc.), and "enp39s0" on
// 8th-generation Intel instances (c8i etc.). Thus the file finds all the interfaces that use
// the "ena" driver, but does not include the primary network interface, which nodeup finds at
// boot time. The file uses the udev property "INTERFACE" for this, because a negated "Name="
// test also agrees with the alternative names of an interface.
func markSecondaryENIsUnmanaged(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) error {
	if !(dist == distributions.DistributionAmazonLinux2023 ||
		(dist.IsDebian() && dist.Version() >= 12)) {
		return nil
	}

	primary, err := primaryInterfaceName(c.Context())
	if err != nil {
		// Do not make the file if the primary network interface is not known. A match that
		// includes the primary network interface causes systemd-networkd to ignore it, and
		// then systemd-resolved has no DNS servers for it.
		return fmt.Errorf("finding primary network interface: %w", err)
	}

	contents := fmt.Sprintf(`
[Match]
Driver=ena
Property=!INTERFACE=%s

[Link]
Unmanaged=yes
`, primary)

	c.AddTask(&nodetasks.File{
		Path:            "/etc/systemd/network/10-eni-secondary.network",
		Contents:        fi.NewStringResource(contents),
		Type:            nodetasks.FileType_File,
		OnChangeExecute: [][]string{{"systemctl", "restart", "systemd-networkd"}},
	})
	return nil
}

// primaryInterfaceName gives the name of the primary network interface. It gets the MAC address
// of the primary ENI (device-number 0) from the IMDS item "mac". Then it compares this MAC
// address with the physical network interfaces in sysfs.
func primaryInterfaceName(ctx context.Context) (string, error) {
	config, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("loading AWS config: %w", err)
	}
	metadata := imds.NewFromConfig(config)
	resp, err := metadata.GetMetadata(ctx, &imds.GetMetadataInput{Path: "mac"})
	if err != nil {
		return "", fmt.Errorf("getting primary MAC address from ec2 metadata: %w", err)
	}
	defer resp.Content.Close()
	mac, err := io.ReadAll(resp.Content)
	if err != nil {
		return "", fmt.Errorf("reading primary MAC address from ec2 metadata: %w", err)
	}

	return findPhysicalInterfaceByMAC("/sys/class/net", strings.TrimSpace(string(mac)))
}

// findPhysicalInterfaceByMAC gives the name of the physical network interface that has the
// specified MAC address. The function ignores the virtual interfaces (veths, bridges, VLANs),
// because a virtual interface can have the same MAC address as a physical interface. The
// function gives an error if it does not find exactly one physical interface with this MAC
// address.
func findPhysicalInterfaceByMAC(sysClassNet string, mac string) (string, error) {
	entries, err := os.ReadDir(sysClassNet)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", sysClassNet, err)
	}

	var matches []string
	for _, entry := range entries {
		name := entry.Name()
		// The scan uses the "device" symlink in sysfs to know if an interface is physical.
		if _, err := os.Stat(filepath.Join(sysClassNet, name, "device")); err != nil {
			continue
		}
		address, err := os.ReadFile(filepath.Join(sysClassNet, name, "address"))
		if err != nil {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(string(address)), mac) {
			matches = append(matches, name)
		}
	}

	switch len(matches) {
	case 1:
		return matches[0], nil
	case 0:
		return "", fmt.Errorf("no physical network interface found with MAC address %q", mac)
	default:
		return "", fmt.Errorf("multiple physical network interfaces found with MAC address %q: %v", mac, matches)
	}
}

// narrowCloudIfupdownHelperRule rewrites Debian 11's
// /etc/udev/rules.d/75-cloud-ifupdown.rules to exclude AWS VPC CNI veths.
// The package-shipped rule matches ENV{INTERFACE}=="eth*|en*", which catches
// real ENIs (ens*) and CNI veths (eni*) alike. For each new netdev,
// /etc/network/cloud-ifupdown-helper generates a DHCP ifupdown stanza and
// starts ifup@$IFACE.service. On CNI veths DHCP times out, ifdown then takes
// the veth DOWN, and pod networking is broken.
//
// The rule and helper are written by cloud-init at first boot and are not
// owned by any dpkg package, so overwriting the file is safe.
//
// Debian 11 only.
func narrowCloudIfupdownHelperRule(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if dist != distributions.DistributionDebian11 {
		return
	}

	contents := `# Handle allow-hotplug interfaces.
# kops: ENV{INTERFACE} narrowed from "eth*|en*" to "eth*|ens*" so AWS VPC CNI
# veths (eni*) are not hijacked by cloud-ifupdown-helper.
SUBSYSTEM=="net", ACTION=="add", ENV{INTERFACE}=="eth*|ens*", RUN+="/etc/network/cloud-ifupdown-helper"
`
	c.AddTask(&nodetasks.File{
		Path:     "/etc/udev/rules.d/75-cloud-ifupdown.rules",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			{"udevadm", "control", "--reload-rules"},
		},
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

// disableNMCloudSetup masks NetworkManager's nm-cloud-setup service and timer.
// nm-cloud-setup polls IMDS, sees the secondary IPs assigned to ENIs (the pod
// IPs), and tells NetworkManager to install per-IP source-routing rules in
// reserved tables 30200/30201/30400/30401 with priority 30200-30401. Those
// priorities are lower (= higher precedence) than the AWS VPC CNI rules at
// 32765, so pod traffic is routed through tables that don't have the routes
// the CNI needs for the service CIDR or IMDS — pods can't reach 100.64.0.1
// or 169.254.169.254 and cluster validation fails.
// ref: https://github.com/aws/amazon-vpc-cni-k8s/blob/master/docs/troubleshooting.md
// RHEL 9 only.
func disableNMCloudSetup(c *fi.NodeupModelBuilderContext, dist distributions.Distribution) {
	if dist != distributions.DistributionRhel9 {
		return
	}

	// Use a marker file so we run the disable steps once and idempotently. The
	// File task triggers OnChangeExecute the first time we land it.
	c.AddTask(&nodetasks.File{
		Path:     "/etc/kops/nm-cloud-setup-disabled",
		Contents: fi.NewStringResource("# Marker: nm-cloud-setup disabled by kops to avoid breaking AWS VPC CNI pod routing.\n"),
		Type:     nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			// Stop and disable the unit + timer so they can't run again.
			{"bash", "-c", "systemctl disable --now nm-cloud-setup.service nm-cloud-setup.timer 2>/dev/null; true"},
			// Mask so future package updates / preset re-enables can't bring them back.
			{"systemctl", "mask", "nm-cloud-setup.service", "nm-cloud-setup.timer"},
			// Drop any rules/routes nm-cloud-setup already installed before we
			// got here. NetworkManager owns these via "device-reapply" with
			// proto=static; tearing the connections down and back up clears the
			// per-IP rules in tables 30200/30201/30400/30401.
			{"bash", "-c", "for c in $(nmcli -t -f NAME connection show --active 2>/dev/null); do nmcli connection down \"$c\" 2>/dev/null; nmcli connection up \"$c\" 2>/dev/null; done; true"},
		},
	})
}

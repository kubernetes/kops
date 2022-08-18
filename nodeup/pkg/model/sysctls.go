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

package model

import (
	"fmt"
	"net"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

// SysctlBuilder set up our sysctls
type SysctlBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &SysctlBuilder{}

// Build is responsible for configuring sysctl settings
func (b *SysctlBuilder) Build(c *fi.ModelBuilderContext) error {
	var sysctls []string

	// Common settings
	{
		sysctls = append(sysctls,
			"# Kubernetes Settings",
			"")

		// A higher vm.max_map_count is great for elasticsearch, mongo, or other mmap users
		// See https://github.com/kubernetes/kops/issues/1340
		sysctls = append(sysctls, "vm.max_map_count = 262144",
			"")

		// See https://github.com/kubernetes/kubernetes/pull/38001
		sysctls = append(sysctls,
			"kernel.softlockup_panic = 1",
			"kernel.softlockup_all_cpu_backtrace = 1",
			"")

		// See https://github.com/kubernetes/kops/issues/6342
		portRange := b.Cluster.Spec.KubeAPIServer.ServiceNodePortRange
		if portRange == "" {
			portRange = "30000-32767" // Default kube-apiserver ServiceNodePortRange
		}
		sysctls = append(sysctls, "net.ipv4.ip_local_reserved_ports = "+portRange,
			"")

		// See https://github.com/kubernetes/kube-deploy/issues/261
		// and https://github.com/kubernetes/kops/issues/10206
		sysctls = append(sysctls,
			"# Increase the number of connections",
			"net.core.somaxconn = 32768",
			"",

			"# Maximum Socket Receive Buffer",
			"net.core.rmem_max = 16777216",
			"",

			"# Maximum Socket Send Buffer",
			"net.core.wmem_max = 16777216",
			"",

			"# Increase the maximum total buffer-space allocatable",
			"net.ipv4.tcp_wmem = 4096 87380 16777216",
			"net.ipv4.tcp_rmem = 4096 87380 16777216",
			"",

			"# Increase the number of outstanding syn requests allowed",
			"net.ipv4.tcp_max_syn_backlog = 8096",
			"",

			"# For persistent HTTP connections",
			"net.ipv4.tcp_slow_start_after_idle = 0",
			"",

			"# Allow to reuse TIME_WAIT sockets for new connections",
			"# when it is safe from protocol viewpoint",
			"net.ipv4.tcp_tw_reuse = 1",
			"",

			// We can't change the local_port_range without changing the NodePort range
			//"# Allowed local port range",
			//"net.ipv4.ip_local_port_range = 10240 65535",
			//"",

			"# Max number of packets that can be queued on interface input",
			"# If kernel is receiving packets faster than can be processed",
			"# this queue increases",
			"net.core.netdev_max_backlog = 16384",
			"",

			"# Increase size of file handles and inode cache",
			"fs.file-max = 2097152",
			"",

			"# Max number of inotify instances and watches for a user",
			"# Since dockerd runs as a single user, the default instances value of 128 per user is too low",
			"# e.g. uses of inotify: nginx ingress controller, kubectl logs -f",
			"fs.inotify.max_user_instances = 8192",
			"fs.inotify.max_user_watches = 524288",

			"# Additional sysctl flags that kubelet expects",
			"vm.overcommit_memory = 1",
			"kernel.panic = 10",
			"kernel.panic_on_oops = 1",
			"",
		)
	}

	if b.CloudProvider == kops.CloudProviderAWS {
		sysctls = append(sysctls,
			"# AWS settings",
			"",
			"# Issue #23395",
			"net.ipv4.neigh.default.gc_thresh1=0",
			"")
	}

	// Running Flannel on Amazon Linux 2 needs custom settings
	if b.Cluster.Spec.Networking.Flannel != nil && b.Distribution == distributions.DistributionAmazonLinux2 {
		proxyMode := b.Cluster.Spec.KubeProxy.ProxyMode
		if proxyMode == "" || proxyMode == "iptables" {
			sysctls = append(sysctls,
				"# Flannel settings on Amazon Linux 2",
				"# Issue https://github.com/coreos/flannel/issues/902",
				"net.bridge.bridge-nf-call-ip6tables=1",
				"net.bridge.bridge-nf-call-iptables=1",
				"")
		}
	}

	if b.Cluster.Spec.IsIPv6Only() {
		if b.Distribution == distributions.DistributionDebian11 {
			// Accepting Router Advertisements must be enabled for each existing network interface to take effect.
			// net.ipv6.conf.all.accept_ra takes effect only for newly created network interfaces.
			// https://bugzilla.kernel.org/show_bug.cgi?id=11655
			sysctls = append(sysctls, "# Enable Router Advertisements to get the default IPv6 route")
			ifaces, err := net.Interfaces()
			if err != nil {
				return err
			}
			for _, iface := range ifaces {
				// Accept Router Advertisements for ethernet network interfaces with slot position.
				// https://www.freedesktop.org/software/systemd/man/systemd.net-naming-scheme.html
				if strings.HasPrefix(iface.Name, "ens") {
					sysctls = append(sysctls, fmt.Sprintf("net.ipv6.conf.%s.accept_ra=2", iface.Name))
				}
			}
		}
		sysctls = append(sysctls,
			"# Enable IPv6 forwarding for network plugins that don't do it themselves",
			"net.ipv6.conf.all.forwarding=1",
			"")
	} else {
		sysctls = append(sysctls,
			"# Prevent docker from changing iptables: https://github.com/kubernetes/kubernetes/issues/40182",
			"net.ipv4.ip_forward=1",
			"")
	}

	if b.Cluster.Spec.Networking.Cilium != nil {
		sysctls = append(sysctls,
			"# Depending on systemd version, cloud and distro, rp_filters may be enabled.",
			"# Cilium requires this to be disabled. See https://github.com/cilium/cilium/issues/10645",
			"net.ipv4.conf.all.rp_filter=0",
			"")
	}

	if params := b.NodeupConfig.SysctlParameters; len(params) > 0 {
		sysctls = append(sysctls,
			"# Custom sysctl parameters from instance group spec",
			"")
		for _, param := range params {
			if !strings.ContainsRune(param, '=') {
				return fmt.Errorf("invalid SysctlParameter: expected %q to contain '='", param)
			}
			sysctls = append(sysctls, param)
		}
	}

	if params := b.Cluster.Spec.SysctlParameters; len(params) > 0 {
		sysctls = append(sysctls,
			"# Custom sysctl parameters from cluster spec",
			"")
		for _, param := range params {
			if !strings.ContainsRune(param, '=') {
				return fmt.Errorf("invalid SysctlParameter: expected %q to contain '='", param)
			}
			sysctls = append(sysctls, param)
		}
	}

	c.AddTask(&nodetasks.File{
		Path:            "/etc/sysctl.d/99-k8s-general.conf",
		Contents:        fi.NewStringResource(strings.Join(sysctls, "\n")),
		Type:            nodetasks.FileType_File,
		OnChangeExecute: [][]string{{"sysctl", "--system"}},
	})

	return nil
}

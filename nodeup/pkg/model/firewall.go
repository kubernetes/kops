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
	"k8s.io/klog"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// FirewallBuilder configures the firewall (iptables)
type FirewallBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &FirewallBuilder{}

// Build is responsible for generating any node firewall rules
func (b *FirewallBuilder) Build(c *fi.ModelBuilderContext) error {
	// We need forwarding enabled (https://github.com/kubernetes/kubernetes/issues/40182)
	c.AddTask(b.buildFirewallScript())
	c.AddTask(b.buildSystemdService())

	return nil
}

func (b *FirewallBuilder) buildSystemdService() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Configure iptables for kubernetes")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kops")
	manifest.Set("Unit", "Before", "network.target")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "RemainAfterExit", "yes")
	manifest.Set("Service", "ExecStart", "/opt/kops/bin/iptables-setup")
	manifest.Set("Install", "WantedBy", "basic.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "kubernetes-iptables-setup", manifestString)

	service := &nodetasks.Service{
		Name:       "kubernetes-iptables-setup.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *FirewallBuilder) buildFirewallScript() *nodetasks.File {
	// TODO: Do we want to rely on running nodeup on every boot, or do we want to install systemd units?

	// TODO: The if statement in the script doesn't make it idempotent

	// This is borrowed from gce/gci/configure-helper.sh
	script := `#!/bin/bash
# Built by kops - do not edit

# The GCI image has host firewall which drop most inbound/forwarded packets.
# We need to add rules to accept all TCP/UDP/ICMP packets.
if iptables -L INPUT | grep "Chain INPUT (policy DROP)" > /dev/null; then
echo "Add rules to accept all inbound TCP/UDP/ICMP packets"
iptables -A INPUT -w -p TCP -j ACCEPT
iptables -A INPUT -w -p UDP -j ACCEPT
iptables -A INPUT -w -p ICMP -j ACCEPT
fi
if iptables -L FORWARD | grep "Chain FORWARD (policy DROP)" > /dev/null; then
echo "Add rules to accept all forwarded TCP/UDP/ICMP packets"
iptables -A FORWARD -w -p TCP -j ACCEPT
iptables -A FORWARD -w -p UDP -j ACCEPT
iptables -A FORWARD -w -p ICMP -j ACCEPT
fi
`
	return &nodetasks.File{
		Path:     "/opt/kops/bin/iptables-setup",
		Contents: fi.NewStringResource(script),
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	}
}

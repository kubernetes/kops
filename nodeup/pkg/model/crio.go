/*
Copyright 2020 The Kubernetes Authors.

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
	"path/filepath"
	"regexp"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// CrioBuilder installs crio
type CrioBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &CrioBuilder{}

// Build configures the crio daemon
func (b *CrioBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.skipInstall() {
		klog.Infof("SkipInstall is set to true; won't install crio")
		return nil
	}

	// Add Apache2 license
	{
		t := &nodetasks.File{
			Path:     "/usr/share/doc/crio/apache.txt",
			Contents: fi.NewStringResource(resources.CrioLicense),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add container signature policy.json
	{
		t := &nodetasks.File{
			Path:     "/etc/containers/policy.json",
			Contents: fi.NewStringResource(fi.StringValue(b.Cluster.Spec.Crio.ContainerPolicyOverride)),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add registries.conf
	{
		t := &nodetasks.File{
			Path:     "/etc/containers/registries.conf",
			Contents: fi.NewStringResource(fi.StringValue(b.Cluster.Spec.Crio.ContainerRegistriesOverride)),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add config file
	{
		t := &nodetasks.File{
			Path:     "/etc/crio/crio.conf",
			Contents: fi.NewStringResource(fi.StringValue(b.Cluster.Spec.Crio.ConfigOverride)),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add binaries from assets
	if b.Cluster.Spec.ContainerRuntime == "crio" {
		f := b.Assets.FindMatches(regexp.MustCompile(`bin/(crio|conmon|pinns|crio-status|crun|runc)`))
		//f := b.Assets.FindMatches(regexp.MustCompile(`(bin/crio|bin/conmon|bin/pinns|bin/crio-status)`))
		if len(f) == 0 {
			return fmt.Errorf("unable to find any crio binaries in assets")
		}
		for k, v := range f {
			fileTask := &nodetasks.File{
				Path:     filepath.Join("/usr/bin", k),
				Contents: v,
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("0755"),
			}
			c.AddTask(fileTask)
		}
	}

	c.AddTask(b.buildSystemdService())

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

func (b *CrioBuilder) buildSystemdService() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "crio container runtime")
	manifest.Set("Unit", "Documentation", "https://github.com/cri-o/cri-o")
	manifest.Set("Unit", "Wants", "network.target local-fs.target")
	manifest.Set("Unit", "After", "network.target local-fs.target")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/crio")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")
	manifest.Set("Service", "ExecStartPre", "-/sbin/modprobe overlay")
	manifest.Set("Service", "ExecStart", "/usr/bin/crio $CRIO_CONFIG_OPTIONS")

	manifest.Set("Service", "Type", "notify")
	// set delegate yes so that systemd does not reset the cgroups of crio containers
	manifest.Set("Service", "Delegate", "yes")
	// kill only the crio process, not all processes in the cgroup
	manifest.Set("Service", "KillMode", "process")

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "5")

	manifest.Set("Service", "LimitNPROC", "infinity")
	manifest.Set("Service", "LimitCORE", "infinity")
	manifest.Set("Service", "LimitNOFILE", "infinity")
	manifest.Set("Service", "TasksMax", "infinity")

	// make killing of processes of this unit under memory pressure very unlikely
	manifest.Set("Service", "OOMScoreAdjust", "-999")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "crio", manifestString)

	service := &nodetasks.Service{
		Name:       "crio.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *CrioBuilder) buildSysconfig(c *fi.ModelBuilderContext) error {
	var crio kops.CrioConfig
	if b.Cluster.Spec.Crio != nil {
		crio = *b.Cluster.Spec.Crio
	}

	flagsString, err := flagbuilder.BuildFlags(&crio)
	if err != nil {
		return fmt.Errorf("error building containerd flags: %v", err)
	}

	lines := []string{
		"CRIO_CONFIG_OPTIONS=" + flagsString,
	}
	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		Path:     "/etc/sysconfig/crio",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})

	return nil
}

// skipInstall determines if kops should skip the installation and configuration of crio
func (b *CrioBuilder) skipInstall() bool {
	d := b.Cluster.Spec.Crio

	// don't skip install if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.SkipInstall
}

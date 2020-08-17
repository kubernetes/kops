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
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog/v2"
)

// UpdateServiceBuilder enables/disables the OS automatic updates.
type UpdateServiceBuilder struct {
	*NodeupModelContext
}

const flatcarServiceName = "update-service"
const debianPackageName = "unattended-upgrades"

var _ fi.ModelBuilder = &UpdateServiceBuilder{}

// Build is responsible for configuring automatic updates based on the OS.
func (b *UpdateServiceBuilder) Build(c *fi.ModelBuilderContext) error {

	if b.Distribution == distros.DistributionFlatcar {
		b.buildFlatcarSystemdService(c)
	} else if b.Distribution.IsDebianFamily() {
		b.buildDebianPackage(c)
	}

	return nil
}

func (b *UpdateServiceBuilder) buildFlatcarSystemdService(c *fi.ModelBuilderContext) {
	if b.Cluster.Spec.UpdatePolicy == nil || *b.Cluster.Spec.UpdatePolicy != kops.UpdatePolicyExternal {
		klog.Infof("UpdatePolicy not set in Cluster Spec; skipping creation of %s", flatcarServiceName)
		return
	}

	for _, spec := range [][]kops.HookSpec{b.InstanceGroup.Spec.Hooks, b.Cluster.Spec.Hooks} {
		for _, hook := range spec {
			if hook.Name == flatcarServiceName || hook.Name == flatcarServiceName+".service" {
				klog.Infof("Detected kops Hook for '%s'; skipping creation", flatcarServiceName)
				return
			}
		}
	}

	klog.Infof("Detected OS %s; building %s service to disable update scheduler", b.Distribution, flatcarServiceName)

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Disable OS Update Scheduler")

	manifest.Set("Unit", "Before", "locksmithd.service")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "ExecStart", "/usr/bin/systemctl mask --now locksmithd.service")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", flatcarServiceName, manifestString)

	service := &nodetasks.Service{
		Name:       flatcarServiceName + ".service",
		Definition: s(manifestString),
	}

	service.InitDefaults()
	c.AddTask(service)
}

func (b *UpdateServiceBuilder) buildDebianPackage(c *fi.ModelBuilderContext) {
	if b.Cluster.Spec.UpdatePolicy != nil && *b.Cluster.Spec.UpdatePolicy == kops.UpdatePolicyExternal {
		klog.Infof("UpdatePolicy is External; skipping installation of %s", debianPackageName)
		return
	}

	klog.Infof("Detected OS %s; installing %s package", b.Distribution, debianPackageName)

	contents := `APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Unattended-Upgrade "1";

APT::Periodic::AutocleanInterval "7";
`
	c.AddTask(&nodetasks.File{
		Path:     "/etc/apt/apt.conf.d/20auto-upgrades",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})

	c.AddTask(&nodetasks.Package{Name: debianPackageName})
}

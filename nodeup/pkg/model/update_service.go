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

	"k8s.io/klog"
)

// UpdateServiceBuilder disables the OS automatic updates
type UpdateServiceBuilder struct {
	*NodeupModelContext
}

// ServiceName is the name given to the service to be created
const ServiceName = "update-service"

var _ fi.ModelBuilder = &UpdateServiceBuilder{}

// Build is responsible for creating the relevant systemd service based on OS
func (b *UpdateServiceBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Cluster.Spec.UpdatePolicy == nil || *b.Cluster.Spec.UpdatePolicy != kops.UpdatePolicyExternal {
		klog.Infof("UpdatePolicy not set in Cluster Spec; skipping creation of %s", ServiceName)
		return nil
	}

	for _, spec := range [][]kops.HookSpec{b.InstanceGroup.Spec.Hooks, b.Cluster.Spec.Hooks} {
		for _, hook := range spec {
			if hook.Name == ServiceName || hook.Name == ServiceName+".service" {
				klog.Infof("Detected kops Hook for '%s'; skipping creation", ServiceName)
				return nil
			}
		}
	}

	if b.Distribution == distros.DistributionCoreOS {
		klog.Infof("Detected OS %s; building %s service to disable update scheduler", ServiceName, b.Distribution)
		c.AddTask(b.buildCoreOSSystemdService())
	}

	if b.Distribution == distros.DistributionFlatcar {
		klog.Infof("Detected OS %s; building %s service to disable update scheduler", ServiceName, b.Distribution)
		c.AddTask(b.buildFlatcarSystemdService())
	}

	return nil
}

func (b *UpdateServiceBuilder) buildCoreOSSystemdService() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Disable OS Update Scheduler")

	manifest.Set("Unit", "Before", "locksmithd.service")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "ExecStart", "/usr/bin/systemctl mask --now locksmithd.service")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", ServiceName, manifestString)

	service := &nodetasks.Service{
		Name:       ServiceName + ".service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *UpdateServiceBuilder) buildFlatcarSystemdService() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Disable OS Update Scheduler")

	manifest.Set("Unit", "Before", "locksmithd.service")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "ExecStart", "/usr/bin/systemctl mask --now locksmithd.service")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", ServiceName, manifestString)

	service := &nodetasks.Service{
		Name:       ServiceName + ".service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

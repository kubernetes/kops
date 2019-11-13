/*
Copyright 2017 The Kubernetes Authors.

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
	"errors"
	"fmt"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"

	"k8s.io/klog"
)

// HookBuilder configures the hooks
type HookBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &HookBuilder{}

// Build is responsible for implementing the cluster hook
func (h *HookBuilder) Build(c *fi.ModelBuilderContext) error {
	// we keep a list of hooks name so we can allow local instanceGroup hooks override the cluster ones
	hookNames := make(map[string]bool)
	for i, spec := range []*[]kops.HookSpec{&h.InstanceGroup.Spec.Hooks, &h.Cluster.Spec.Hooks} {
		for j, hook := range *spec {
			isInstanceGroup := i == 0
			// filter roles if required
			if len(hook.Roles) > 0 && !containsRole(h.InstanceGroup.Spec.Role, hook.Roles) {
				continue
			}

			// I don't want to affect those whom are already using the hooks, so I'm going to try to keep the name for now
			// i.e. use the default naming convention - kops-hook-<index>, only those using the Name or hooks in IG should alter
			var name string
			switch hook.Name {
			case "":
				name = fmt.Sprintf("kops-hook-%d", j)
				if isInstanceGroup {
					name += "-ig"
				}
			default:
				name = hook.Name
			}

			if _, found := hookNames[name]; found {
				klog.V(2).Infof("Skipping the hook: %v as we've already processed a similar service name", name)
				continue
			}
			hookNames[name] = true

			// are we disabling the service?
			if hook.Disabled {
				enabled := false
				managed := true
				c.AddTask(&nodetasks.Service{
					Name:        h.EnsureSystemdSuffix(name),
					ManageState: &managed,
					Enabled:     &enabled,
					Running:     &enabled,
				})
				continue
			}

			service, err := h.buildSystemdService(name, &hook)
			if err != nil {
				return err
			}

			if service != nil {
				c.AddTask(service)
			}
		}
	}

	return nil
}

// buildSystemdService is responsible for generating the service
func (h *HookBuilder) buildSystemdService(name string, hook *kops.HookSpec) (*nodetasks.Service, error) {
	// perform some basic validation
	if hook.ExecContainer == nil && hook.Manifest == "" {
		klog.Warningf("hook: %s has neither a raw unit or exec image configured", name)
		return nil, nil
	}
	if hook.ExecContainer != nil {
		if err := isValidExecContainerAction(hook.ExecContainer); err != nil {
			klog.Warningf("invalid hook action, name: %s, error: %v", name, err)
			return nil, nil
		}
	}
	// build the base unit file
	var definition *string
	if hook.UseRawManifest {
		definition = s(hook.Manifest)
	} else {
		unit := &systemd.Manifest{}
		unit.Set("Unit", "Description", "Kops Hook "+name)

		// add any service dependencies to the unit
		for _, x := range hook.Requires {
			unit.Set("Unit", "Requires", x)
		}
		for _, x := range hook.Before {
			unit.Set("Unit", "Before", x)
		}

		// are we a raw unit file or a docker exec?
		switch hook.ExecContainer {
		case nil:
			unit.SetSection("Service", hook.Manifest)
		default:
			if err := h.buildDockerService(unit, hook); err != nil {
				return nil, err
			}
		}
		definition = s(unit.Render())
	}

	service := &nodetasks.Service{
		Name:       h.EnsureSystemdSuffix(name),
		Definition: definition,
	}

	service.InitDefaults()

	return service, nil
}

// buildDockerService is responsible for generating a docker exec unit file
func (h *HookBuilder) buildDockerService(unit *systemd.Manifest, hook *kops.HookSpec) error {
	dockerArgs := []string{
		"/usr/bin/docker", "run",
		"-v", "/:/rootfs/",
		"-v", "/var/run/dbus:/var/run/dbus",
		"-v", "/run/systemd:/run/systemd",
		"--net=host",
		"--privileged",
	}
	dockerArgs = append(dockerArgs, buildDockerEnvironmentVars(hook.ExecContainer.Environment)...)
	dockerArgs = append(dockerArgs, hook.ExecContainer.Image)
	dockerArgs = append(dockerArgs, hook.ExecContainer.Command...)

	dockerRunCommand := systemd.EscapeCommand(dockerArgs)
	dockerPullCommand := systemd.EscapeCommand([]string{"/usr/bin/docker", "pull", hook.ExecContainer.Image})

	unit.Set("Unit", "Requires", "docker.service")
	unit.Set("Service", "ExecStartPre", dockerPullCommand)
	unit.Set("Service", "ExecStart", dockerRunCommand)
	unit.Set("Service", "Type", "oneshot")
	unit.Set("Install", "WantedBy", "multi-user.target")

	return nil
}

// isValidExecContainerAction checks the validity of the execContainer - personally i think this validation
// should be done high up the chain, but
func isValidExecContainerAction(action *kops.ExecContainerAction) error {
	action.Image = strings.TrimSpace(action.Image)
	if action.Image == "" {
		return errors.New("the image for the hook exec action not set")
	}

	return nil
}

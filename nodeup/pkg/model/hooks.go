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

	"github.com/golang/glog"
)

// HookBuilder configures the hooks
type HookBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &HookBuilder{}

// Build is responsible for implementing the cluster hook
func (h *HookBuilder) Build(c *fi.ModelBuilderContext) error {
	// iterate the hooks and render the systemd units
	for i := range h.Cluster.Spec.Hooks {
		hook := &h.Cluster.Spec.Hooks[i]

		// filter out on master and node flags if required
		if (hook.MasterOnly && !h.IsMaster) || (hook.NodeOnly && h.IsMaster) {
			continue
		}

		// are we disabling the service?
		if hook.Disabled {
			enabled := false
			managed := true
			c.AddTask(&nodetasks.Service{
				Name:        hook.Name,
				ManageState: &managed,
				Enabled:     &enabled,
				Running:     &enabled,
			})
			continue
		}

		// use the default naming convention - kops-hook-<index>
		name := fmt.Sprintf("kops-hook-%d", i)
		if hook.Name != "" {
			name = hook.Name
		}

		// generate the systemd service
		service, err := h.buildSystemdService(name, hook)
		if err != nil {
			return err
		}

		if service != nil {
			c.AddTask(service)
		}
	}

	return nil
}

// buildSystemdService is responsible for generating the service
func (h *HookBuilder) buildSystemdService(name string, hook *kops.HookSpec) (*nodetasks.Service, error) {
	// perform some basic validation
	if hook.ExecContainer == nil && hook.Manifest == "" {
		glog.Warningf("hook: %s has neither a raw unit or exec image configured", name)
		return nil, nil
	}
	if hook.ExecContainer != nil {
		if err := isValidExecContainerAction(hook.ExecContainer); err != nil {
			glog.Warningf("invalid hook action, name: %s, error: %v", name, err)
			return nil, nil
		}
	}
	// build the base unit file
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

	service := &nodetasks.Service{
		Name:       name,
		Definition: s(unit.Render()),
	}

	service.InitDefaults()

	return service, nil
}

// buildDockerService is responsible for generating a docker exec unit file
func (h *HookBuilder) buildDockerService(unit *systemd.Manifest, hook *kops.HookSpec) error {
	// the docker command line
	dockerArgs := []string{
		"/usr/bin/docker",
		"run",
		"-v", "/:/rootfs/",
		"-v", "/var/run/dbus:/var/run/dbus",
		"-v", "/run/systemd:/run/systemd",
		"--net=host",
		"--privileged",
		hook.ExecContainer.Image,
	}
	dockerArgs = append(dockerArgs, hook.ExecContainer.Command...)

	dockerRunCommand := systemd.EscapeCommand(dockerArgs)
	dockerPullCommand := systemd.EscapeCommand([]string{"/usr/bin/docker", "pull", hook.ExecContainer.Image})

	unit.Set("Service", "ExecStartPre", dockerPullCommand)
	unit.Set("Service", "ExecStart", dockerRunCommand)
	unit.Set("Service", "Type", "oneshot")
	unit.Set("Install", "WantedBy", "multi-user.target")

	return nil
}

// isValidExecContainerAction checks the validatity of the execContainer - personally i think this validation
// should be done high up the chain, but
func isValidExecContainerAction(action *kops.ExecContainerAction) error {
	action.Image = strings.TrimSpace(action.Image)
	if action.Image == "" {
		return errors.New("the image for the hook exec action not set")
	}

	return nil
}

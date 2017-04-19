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
	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"strconv"
	"strings"
)

// HooksBuilder configures the hooks
type HookBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &HookBuilder{}

func (b *HookBuilder) Build(c *fi.ModelBuilderContext) error {
	for i := range b.Cluster.Spec.Hooks {
		hook := &b.Cluster.Spec.Hooks[i]

		// TODO: Allow (alphanumeric?) names
		name := strconv.Itoa(i + 1)

		service, err := b.buildSystemdService(name, hook)
		if err != nil {
			return err
		}
		c.AddTask(service)
	}

	return nil
}

func (b *HookBuilder) buildSystemdService(name string, hook *kops.HookSpec) (*nodetasks.Service, error) {
	// We could give the container a kubeconfig, but we would probably do better to have a real pod / daemonset / job at that point
	execContainer := hook.ExecContainer
	if execContainer == nil {
		glog.Warningf("No ExecContainer found for hook: %v", hook)
		return nil, nil
	}

	image := strings.TrimSpace(execContainer.Image)
	if image == "" {
		glog.Warningf("No Image found for hook: %v", hook)
		return nil, nil
	}

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

	dockerArgs = append(dockerArgs, execContainer.Command...)

	dockerRunCommand := systemd.EscapeCommand(dockerArgs)
	dockerPullCommand := systemd.EscapeCommand([]string{"/usr/bin/docker", "pull", hook.ExecContainer.Image})

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kops Hook "+name)

	manifest.Set("Service", "ExecStartPre", dockerPullCommand)
	manifest.Set("Service", "ExecStart", dockerRunCommand)
	manifest.Set("Service", "Type", "oneshot")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built systemd manifest %q\n%s", name, manifestString)

	service := &nodetasks.Service{
		Name:       "kops-hook-" + name + ".service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service, nil
}

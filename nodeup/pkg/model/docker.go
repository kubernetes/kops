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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blang/semver/v4"

	"k8s.io/klog/v2"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
)

// DockerBuilder install docker (just the packages at the moment)
type DockerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DockerBuilder{}

func (b *DockerBuilder) dockerVersion() (string, error) {
	dockerVersion := ""
	if b.Cluster.Spec.Docker != nil {
		dockerVersion = fi.StringValue(b.Cluster.Spec.Docker.Version)
	}
	if dockerVersion == "" {
		return "", fmt.Errorf("error finding Docker version")
	}
	return dockerVersion, nil
}

// Build is responsible for configuring the docker daemon
func (b *DockerBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.skipInstall() {
		klog.Infof("SkipInstall is set to true; won't install Docker")
		return nil
	}

	// @check: neither flatcar nor containeros need provision docker.service, just the docker daemon options
	switch b.Distribution {
	case distributions.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distributions.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil
	}

	c.AddTask(b.buildDockerGroup())
	c.AddTask(b.buildSystemdSocket())

	// Add binaries from assets
	{
		f := b.Assets.FindMatches(regexp.MustCompile(`^docker/`))
		if len(f) == 0 {
			return fmt.Errorf("unable to find any Docker binaries in assets")
		}
		for k, v := range f {
			klog.V(4).Infof("Found matching Docker asset: %q", k)
			c.AddTask(&nodetasks.File{
				Path:     filepath.Join("/usr/bin", k),
				Contents: v,
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("0755"),
			})
		}
	}

	// Add Apache2 license
	{
		t := &nodetasks.File{
			Path:     "/usr/share/doc/docker/apache.txt",
			Contents: fi.NewStringResource(resources.DockerApache2License),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	dockerVersion, err := b.dockerVersion()
	if err != nil {
		return err
	}

	v, err := semver.ParseTolerant(dockerVersion)
	if err != nil {
		return fmt.Errorf("error parsing docker version %q: %v", dockerVersion, err)
	}
	c.AddTask(b.buildSystemdService(v))

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	// Enable health-check
	if b.healthCheck() || (b.IsKubernetesLT("1.18") && b.Distribution.IsDebianFamily()) {
		c.AddTask(b.buildSystemdHealthCheckScript())
		c.AddTask(b.buildSystemdHealthCheckService())
		c.AddTask(b.buildSystemdHealthCheckTimer())
	}

	return nil
}

// buildDockerGroup creates the docker group, which owns the docker.socket
func (b *DockerBuilder) buildDockerGroup() *nodetasks.GroupTask {
	return &nodetasks.GroupTask{
		Name:   "docker",
		System: true,
	}
}

// buildSystemdSocket creates docker.socket, for when we're not installing from a package
func (b *DockerBuilder) buildSystemdSocket() *nodetasks.Service {
	// Based on https://github.com/docker/docker-ce-packaging/blob/master/systemd/docker.socket

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Docker Socket for the API")
	manifest.Set("Unit", "PartOf", "docker.service")

	manifest.Set("Socket", "ListenStream", "/var/run/docker.sock")
	manifest.Set("Socket", "SocketMode", "0660")
	manifest.Set("Socket", "SocketUser", "root")
	manifest.Set("Socket", "SocketGroup", "docker")

	manifest.Set("Install", "WantedBy", "sockets.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built docker.socket manifest\n%s", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.socket",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdService(dockerVersion semver.Version) *nodetasks.Service {
	// Based on https://github.com/docker/docker-ce-packaging/blob/master/systemd/docker.service

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Docker Application Container Engine")
	manifest.Set("Unit", "Documentation", "https://docs.docker.com")
	if dockerVersion.GTE(semver.MustParse("18.9.0")) {
		manifest.Set("Unit", "BindsTo", "containerd.service")
		manifest.Set("Unit", "After", "network-online.target firewalld.service containerd.service")
	} else {
		manifest.Set("Unit", "After", "network-online.target firewalld.service")
	}
	manifest.Set("Unit", "Wants", "network-online.target")
	manifest.Set("Unit", "Requires", "docker.socket")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/docker")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")

	// Restore the default SELinux security contexts for the Docker binaries
	if b.Distribution.IsRHELFamily() && b.Cluster.Spec.Docker != nil && fi.BoolValue(b.Cluster.Spec.Docker.SelinuxEnabled) {
		manifest.Set("Service", "ExecStartPre", "/bin/sh -c 'restorecon -v /usr/bin/docker*'")
	}

	// the default is not to use systemd for cgroups because the delegate issues still
	// exists and systemd currently does not support the cgroup feature set required
	// for containers run by docker
	manifest.Set("Service", "Type", "notify")
	manifest.Set("Service", "ExecStart", "/usr/bin/dockerd -H fd:// \"$DOCKER_OPTS\"")
	manifest.Set("Service", "ExecReload", "/bin/kill -s HUP $MAINPID")
	manifest.Set("Service", "TimeoutSec", "0")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "Restart", "always")

	// Note that StartLimit* options were moved from "Service" to "Unit" in systemd 229.
	// Both the old, and new location are accepted by systemd 229 and up, so using the old location
	// to make them work for either version of systemd.
	manifest.Set("Service", "StartLimitBurst", "3")

	// Note that StartLimitInterval was renamed to StartLimitIntervalSec in systemd 230.
	// Both the old, and new name are accepted by systemd 230 and up, so using the old name to make
	// this option work for either version of systemd.
	manifest.Set("Service", "StartLimitInterval", "60s")

	// Having non-zero Limit*s causes performance problems due to accounting overhead
	// in the kernel. We recommend using cgroups to do container-local accounting.
	manifest.Set("Service", "LimitNOFILE", "infinity")
	manifest.Set("Service", "LimitNPROC", "infinity")
	manifest.Set("Service", "LimitCORE", "infinity")

	// Only systemd 226 and above support this option.
	manifest.Set("Service", "TasksMax", "infinity")

	// set delegate yes so that systemd does not reset the cgroups of docker containers
	manifest.Set("Service", "Delegate", "yes")

	// kill only the docker process, not all processes in the cgroup
	manifest.Set("Service", "KillMode", "process")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "docker", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdHealthCheckScript() *nodetasks.File {
	script := &nodetasks.File{
		Path:     "/opt/kops/bin/docker-healthcheck",
		Contents: fi.NewStringResource(resources.DockerHealthCheck),
		Type:     nodetasks.FileType_File,
		Mode:     s("0755"),
	}

	return script
}

func (b *DockerBuilder) buildSystemdHealthCheckService() *nodetasks.Service {
	manifest := &systemd.Manifest{}

	manifest.Set("Unit", "Description", "Run docker-healthcheck once")
	manifest.Set("Unit", "Documentation", "https://kops.sigs.k8s.io")
	manifest.Set("Service", "Type", "oneshot")
	manifest.Set("Service", "ExecStart", "/opt/kops/bin/docker-healthcheck")
	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "docker-healthcheck.service", manifestString)

	service := &nodetasks.Service{
		Name:       "docker-healthcheck.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildSystemdHealthCheckTimer() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Trigger docker-healthcheck periodically")
	manifest.Set("Unit", "Documentation", "https://kops.sigs.k8s.io")
	manifest.Set("Timer", "OnUnitInactiveSec", "10s")
	manifest.Set("Timer", "Unit", "docker-healthcheck.service")
	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built timer manifest %q\n%s", "docker-healthcheck.timer", manifestString)

	service := &nodetasks.Service{
		Name:       "docker-healthcheck.timer",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

// buildContainerOSConfigurationDropIn is responsible for configuring the docker daemon options
func (b *DockerBuilder) buildContainerOSConfigurationDropIn(c *fi.ModelBuilderContext) error {
	lines := []string{
		"[Service]",
		"EnvironmentFile=/etc/sysconfig/docker",
		"EnvironmentFile=/etc/environment",
	}

	// Equivalent of https://github.com/kubernetes/kubernetes/pull/51986
	lines = append(lines, "TasksMax=infinity")

	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		AfterFiles: []string{"/etc/sysconfig/docker"},
		Path:       "/etc/systemd/system/docker.service.d/10-kops.conf",
		Contents:   fi.NewStringResource(contents),
		Type:       nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "restart", "docker.service"},
			// We need to restart kops-configuration service since nodeup needs to load images
			// into docker with the new overlay storage. Restart is on the background because
			// kops-configuration is of type 'one-shot' so the restart command will wait for
			// nodeup to finish executing
			{"systemctl", "restart", "kops-configuration.service", "&"},
		},
	})

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

// buildSysconfig is responsible for extracting the docker configuration and writing the sysconfig file
func (b *DockerBuilder) buildSysconfig(c *fi.ModelBuilderContext) error {
	var docker kops.DockerConfig
	if b.Cluster.Spec.Docker != nil {
		docker = *b.Cluster.Spec.Docker
	}

	// ContainerOS now sets the storage flag in /etc/docker/daemon.json, and it is an error to set it twice
	if b.Distribution == distributions.DistributionContainerOS {
		// So that we can support older COS images though, we do check for /etc/docker/daemon.json
		if b, err := ioutil.ReadFile("/etc/docker/daemon.json"); err != nil {
			if os.IsNotExist(err) {
				klog.V(2).Infof("/etc/docker/daemon.json not found")
			} else {
				klog.Warningf("error reading /etc/docker/daemon.json: %v", err)
			}
		} else {
			// Maybe we get smarter here?
			data := make(map[string]interface{})
			if err := json.Unmarshal(b, &data); err != nil {
				klog.Warningf("error deserializing /etc/docker/daemon.json: %v", err)
			} else {
				storageDriver := data["storage-driver"]
				klog.Infof("/etc/docker/daemon.json has storage-driver: %q", storageDriver)
			}
			docker.Storage = nil
		}
	}

	// RHEL-family / docker has a bug with 17.x where it fails to use overlay2 because it does a broken kernel check
	if b.Distribution.IsRHELFamily() {
		dockerVersion, err := b.dockerVersion()
		if err != nil {
			return err
		}
		if strings.HasPrefix(dockerVersion, "17.") {
			storageOpts := strings.Join(docker.StorageOpts, ",")
			if strings.Contains(storageOpts, "overlay2.override_kernel_check=1") {
				// Already there
			} else if !strings.Contains(storageOpts, "overlay2.override_kernel_check") {
				docker.StorageOpts = append(docker.StorageOpts, "overlay2.override_kernel_check=1")
			} else {
				klog.Infof("detected image was RHEL and overlay2.override_kernel_check=1 was probably needed, but overlay2.override_kernel_check was already set (%q) so won't set", storageOpts)
			}
		}
	}

	flagsString, err := flagbuilder.BuildFlags(&docker)
	if err != nil {
		return fmt.Errorf("error building docker flags: %v", err)
	}

	lines := []string{
		"DOCKER_OPTS=" + flagsString,
		"DOCKER_NOFILE=1000000",
	}
	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		Path:     "/etc/sysconfig/docker",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})

	return nil
}

// skipInstall determines if kops should skip the installation and configuration of Docker
func (b *DockerBuilder) skipInstall() bool {
	d := b.Cluster.Spec.Docker

	// don't skip install if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.SkipInstall
}

// healthCheck determines if kops should enable the health-check for Docker
func (b *DockerBuilder) healthCheck() bool {
	d := b.Cluster.Spec.Docker

	// don't enable the health-check if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.HealthCheck
}

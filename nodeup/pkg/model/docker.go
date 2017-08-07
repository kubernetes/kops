/*
Copyright 2016 The Kubernetes Authors.

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
	"strings"

	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// DockerBuilder install docker (just the packages at the moment)
type DockerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &DockerBuilder{}

type dockerVersion struct {
	Name    string
	Version string
	Source  string
	Hash    string

	DockerVersion string
	Distros       []distros.Distribution
	Dependencies  []string
	Architectures []Architecture
}

const DefaultDockerVersion = "1.12.3"

var dockerVersions = []dockerVersion{
	// 1.11.2 - Jessie
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.11.2-0~jessie_amd64.deb",
		Hash:          "c312f1f6fa0b34df4589bb812e4f7af8e28fd51d",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.11.2 - Xenial
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2-0~xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.11.2-0~xenial_amd64.deb",
		Hash:          "194bfa864f0424d1bbdc7d499ccfa0445ce09b9f",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.11.2 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.11.2-1.el7.centos.x86_64.rpm",
		Hash:          "432e6d7948df9e05f4190fce2f423eedbfd673d5",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.11.2",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.11.2",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.11.2-1.el7.centos.noarch.rpm",
		Hash:          "f6da608fa8eeb2be8071489086ed9ff035f6daba",
	},

	// 1.12.1 - Jessie
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.1-0~jessie_amd64.deb",
		Hash:          "0401866231749abaabe8e09ee24432132839fe53",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.1 - Xenial
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1-0~xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.1-0~xenial_amd64.deb",
		Hash:          "30f7840704361673db2b62f25b6038628184b056",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.1 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.1-1.el7.centos.x86_64.rpm",
		Hash:          "636471665665546224444052c3b48001397036be",
		Dependencies:  []string{"libtool-ltdl"},
	},
	{
		DockerVersion: "1.12.1",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.1",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.1-1.el7.centos.noarch.rpm",
		Hash:          "52ec22128e70acc2f76b3a8e87ff96785995116a",
	},

	// 1.12.3 - k8s 1.5

	// 1.12.3 - Jessie
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.3-0~jessie_amd64.deb",
		Hash:          "7c7eb45542b67a9cfb33c292ba245710efb5d773",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.90), libltdl7 (>= 2.4.2), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.12.3 - Jessie on ARM
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "1.12.3-0~jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.3-0~jessie_armhf.deb",
		Hash:          "aa2f2f710360268dc5fd3eb066868c5883d95698",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.3 - Xenial
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3-0~xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.3-0~xenial_amd64.deb",
		Hash:          "b758fc88346a1e5eebf7408b0d0c99f4f134166c",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.3 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.3-1.el7.centos.x86_64.rpm",
		Hash:          "67fbb78cfb9526aaf8142c067c10384df199d8f9",
		Dependencies:  []string{"libtool-ltdl", "libseccomp"},
	},
	{
		DockerVersion: "1.12.3",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.3",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.3-1.el7.centos.noarch.rpm",
		Hash:          "a6b0243af348140236ed96f2e902b259c590eefa",
	},

	// 1.12.6 - k8s 1.6

	// 1.12.6 - Jessie
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6-0~debian-jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~debian-jessie_amd64.deb",
		Hash:          "1a8b0c4e3386e12964676a126d284cebf599cc8e",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
		//Depends: iptables, init-system-helpers (>= 1.18~), libapparmor1 (>= 2.6~devel), libc6 (>= 2.17), libdevmapper1.02.1 (>= 2:1.02.90), libltdl7 (>= 2.4.2), libsystemd0
		//Recommends: aufs-tools, ca-certificates, cgroupfs-mount | cgroup-lite, git, xz-utils
	},

	// 1.12.6 - Jessie on ARM
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionJessie},
		Architectures: []Architecture{ArchitectureArm},
		Version:       "1.12.6-0~debian-jessie",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~debian-jessie_armhf.deb",
		Hash:          "ac148e1f7381e4201e139584dd3c102372ad96fb",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.6 - Xenial
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionXenial},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6-0~ubuntu-xenial",
		Source:        "http://apt.dockerproject.org/repo/pool/main/d/docker-engine/docker-engine_1.12.6-0~ubuntu-xenial_amd64.deb",
		Hash:          "fffc22da4ad5b20715bbb6c485b2d2bb7e84fd33",
		Dependencies:  []string{"bridge-utils", "libapparmor1", "libltdl7", "perl"},
	},

	// 1.12.6 - Centos / Rhel7 (two packages)
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-1.12.6-1.el7.centos.x86_64.rpm",
		Hash:          "776dbefa9dc7733000e46049293555a9a422c50e",
		Dependencies:  []string{"libtool-ltdl", "libseccomp"},
	},
	{
		DockerVersion: "1.12.6",
		Name:          "docker-engine-selinux",
		Distros:       []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures: []Architecture{ArchitectureAmd64},
		Version:       "1.12.6",
		Source:        "https://yum.dockerproject.org/repo/main/centos/7/Packages/docker-engine-selinux-1.12.6-1.el7.centos.noarch.rpm",
		Hash:          "9a6ee0d631ca911b6927450a3c396e9a5be75047",
	},
}

func (d *dockerVersion) matches(arch Architecture, dockerVersion string, distro distros.Distribution) bool {
	if d.DockerVersion != dockerVersion {
		return false
	}
	foundDistro := false
	for _, d := range d.Distros {
		if d == distro {
			foundDistro = true
		}
	}
	if !foundDistro {
		return false
	}

	foundArch := false
	for _, a := range d.Architectures {
		if a == arch {
			foundArch = true
		}
	}
	if !foundArch {
		return false
	}

	return true
}

func (b *DockerBuilder) Build(c *fi.ModelBuilderContext) error {
	switch b.Distribution {
	case distros.DistributionCoreOS:
		glog.Infof("Detected CoreOS; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distros.DistributionContainerOS:
		glog.Infof("Detected ContainerOS; won't install Docker")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil
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

	dockerVersion := ""
	if b.Cluster.Spec.Docker != nil {
		dockerVersion = fi.StringValue(b.Cluster.Spec.Docker.Version)
	}
	if dockerVersion == "" {
		dockerVersion = DefaultDockerVersion
		glog.Warningf("DockerVersion not specified; using default %q", dockerVersion)
	}

	// Add packages
	{
		for i := range dockerVersions {
			dv := &dockerVersions[i]
			if !dv.matches(b.Architecture, dockerVersion, b.Distribution) {
				continue
			}

			c.AddTask(&nodetasks.Package{
				Name:    dv.Name,
				Version: s(dv.Version),
				Source:  s(dv.Source),
				Hash:    s(dv.Hash),

				// TODO: PreventStart is now unused?
				PreventStart: fi.Bool(true),
			})

			for _, dep := range dv.Dependencies {
				c.AddTask(&nodetasks.Package{Name: dep})
			}

			// Note we do _not_ stop looping... centos/rhel comprises multiple packages
		}
	}

	dockerSemver, err := semver.ParseTolerant(dockerVersion)
	if err != nil {
		return fmt.Errorf("error parsing docker version %q as semver: %v", dockerVersion, err)
	}

	c.AddTask(b.buildSystemdService(dockerSemver))

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

func (b *DockerBuilder) buildSystemdService(dockerVersion semver.Version) *nodetasks.Service {
	oldDocker := dockerVersion.Major <= 1 && dockerVersion.Minor <= 11
	usesDockerSocket := true
	hasDockerBabysitter := false

	var dockerdCommand string
	if oldDocker {
		dockerdCommand = "/usr/bin/docker daemon"
	} else {
		dockerdCommand = "/usr/bin/dockerd"
	}

	if b.Distribution.IsDebianFamily() {
		hasDockerBabysitter = true
	}

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Docker Application Container Engine")
	manifest.Set("Unit", "Documentation", "https://docs.docker.com")

	if b.Distribution.IsRHELFamily() && !oldDocker {
		// See https://github.com/docker/docker/pull/24804
		usesDockerSocket = false
	}

	if usesDockerSocket {
		manifest.Set("Unit", "After", "network.target docker.socket")
		manifest.Set("Unit", "Requires", "docker.socket")
	} else {
		manifest.Set("Unit", "After", "network.target")
	}

	manifest.Set("Service", "Type", "notify")
	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/docker")

	if usesDockerSocket {
		manifest.Set("Service", "ExecStart", dockerdCommand+" -H fd:// \"$DOCKER_OPTS\"")
	} else {
		manifest.Set("Service", "ExecStart", dockerdCommand+" \"$DOCKER_OPTS\"")
	}

	if !oldDocker {
		// This was added by docker 1.12
		// TODO: They seem sensible - should we backport them?

		manifest.Set("Service", "ExecReload", "/bin/kill -s HUP $MAINPID")
		// kill only the docker process, not all processes in the cgroup
		manifest.Set("Service", "KillMode", "process")

		manifest.Set("Service", "TimeoutStartSec", "0")
	}

	if oldDocker {
		// Only in older versions of docker (< 1.12)
		manifest.Set("Service", "MountFlags", "slave")
	}

	// Having non-zero Limit*s causes performance problems due to accounting overhead
	// in the kernel. We recommend using cgroups to do container-local accounting.
	// TODO: Should we set this? https://github.com/kubernetes/kubernetes/issues/39682
	//service.Set("Service", "LimitNOFILE", "infinity")
	//service.Set("Service", "LimitNPROC", "infinity")
	//service.Set("Service", "LimitCORE", "infinity")
	manifest.Set("Service", "LimitNOFILE", "1048576")
	manifest.Set("Service", "LimitNPROC", "1048576")
	manifest.Set("Service", "LimitCORE", "infinity")

	//# Uncomment TasksMax if your systemd version supports it.
	//# Only systemd 226 and above support this version.
	//#TasksMax=infinity

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")

	// set delegate yes so that systemd does not reset the cgroups of docker containers
	manifest.Set("Service", "Delegate", "yes")

	if hasDockerBabysitter {
		manifest.Set("Service", "ExecStartPre", "/opt/kubernetes/helpers/docker-prestart")
	}

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	glog.V(8).Infof("Built service manifest %q\n%s", "docker", manifestString)

	service := &nodetasks.Service{
		Name:       "docker.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

func (b *DockerBuilder) buildContainerOSConfigurationDropIn(c *fi.ModelBuilderContext) error {
	lines := []string{
		"[Service]",
		"EnvironmentFile=/etc/sysconfig/docker",
	}
	contents := strings.Join(lines, "\n")

	t := &nodetasks.File{
		Path:     "/etc/systemd/system/docker.service.d/10-kops.conf",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "restart", "docker.service"},
		},
	}
	c.AddTask(t)

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

func (b *DockerBuilder) buildSysconfig(c *fi.ModelBuilderContext) error {
	flagsString, err := flagbuilder.BuildFlags(b.Cluster.Spec.Docker)
	if err != nil {
		return fmt.Errorf("error building docker flags: %v", err)
	}

	lines := []string{
		"DOCKER_OPTS=" + flagsString,
		"DOCKER_NOFILE=1000000",
	}
	contents := strings.Join(lines, "\n")

	t := &nodetasks.File{
		Path:     "/etc/sysconfig/docker",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	}
	c.AddTask(t)

	return nil
}

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
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/nodeup/pkg/model/resources"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// ContainerdBuilder install containerd (just the packages at the moment)
type ContainerdBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ContainerdBuilder{}

// DefaultContainerdVersion is the (legacy) containerd version we use if one is not specified in the manifest.
// We don't change this with each version of kops, we expect newer versions of kops to populate the field.
const DefaultContainerdVersion = "1.2.10"

var containerdVersions = []packageVersion{
	// 1.2.10 - Debian Stretch
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "186f2f2c570f37b363102e6b879073db6dec671d",
		Dependencies:   []string{"libseccomp2", "pigz"},
	},

	// 1.2.10 - Debian Buster
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian10},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "365e4a7541ce2cf3c3036ea2a9bf6b40a50893a8",
		Dependencies:   []string{"libseccomp2", "pigz"},
	},

	// 1.2.10 - Ubuntu Xenial
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "b64e7170d9176bc38967b2e12147c69b65bdd0fc",
		Dependencies:   []string{"libseccomp2", "pigz"},
	},

	// 1.2.10 - Ubuntu Bionic
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "f4c941807310e3fa470dddfb068d599174a3daec",
		Dependencies:   []string{"libseccomp2", "pigz"},
	},

	// 1.2.10 - CentOS / Rhel 7
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.x86_64.rpm",
		Hash:           "f6447e84479df3a58ce04a3da87ccc384663493b",
		ExtraPackages: map[string]packageInfo{
			"container-selinux": {
				Version: "2.107",
				Source:  "http://vault.centos.org/7.6.1810/extras/x86_64/Packages/container-selinux-2.107-1.el7_6.noarch.rpm",
				Hash:    "7de4211fa0dfd240d8827b93763e1eb5f0d56411",
			},
		},
		Dependencies: []string{"libseccomp", "pigz", "policycoreutils-python"},
	},

	// 1.2.10 - CentOS / Rhel 8
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionRhel8, distros.DistributionCentos8},
		Architectures:  []Architecture{ArchitectureAmd64},
		Version:        "1.2.10",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.x86_64.rpm",
		Hash:           "f6447e84479df3a58ce04a3da87ccc384663493b",
		Dependencies:   []string{"container-selinux", "libseccomp", "pigz"},
	},

	// TIP: When adding the next version, copy the previous
	// version, string replace the version, run `VERIFY_HASHES=1
	// go test ./nodeup/pkg/model` (you might want to temporarily
	// comment out older versions on a slower connection), and
	// then validate the dependencies etc
}

func (b *ContainerdBuilder) containerdVersion() string {
	containerdVersion := ""
	if b.Cluster.Spec.Containerd != nil {
		containerdVersion = fi.StringValue(b.Cluster.Spec.Containerd.Version)
	}
	if containerdVersion == "" {
		containerdVersion = DefaultContainerdVersion
		klog.Warningf("Containerd version not specified; using default %q", containerdVersion)
	}
	return containerdVersion
}

// Build is responsible for configuring the containerd daemon
func (b *ContainerdBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.skipInstall() {
		klog.Infof("SkipInstall is set to true; won't install containerd")
		return nil
	}

	// @check: neither coreos or containeros need provision containerd.service, just the containerd daemon options
	switch b.Distribution {
	case distros.DistributionCoreOS:
		klog.Infof("Detected CoreOS; won't install containerd")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distros.DistributionFlatcar:
		klog.Infof("Detected Flatcar; won't install containerd")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil

	case distros.DistributionContainerOS:
		klog.Infof("Detected ContainerOS; won't install containerd")
		if err := b.buildContainerOSConfigurationDropIn(c); err != nil {
			return err
		}
		return nil
	}

	// Add Apache2 license
	{
		t := &nodetasks.File{
			Path:     "/usr/share/doc/containerd/apache.txt",
			Contents: fi.NewStringResource(resources.ContainerdApache2License),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add config file
	{
		containerdConfigFile := ""
		if b.Cluster.Spec.Containerd != nil {
			containerdConfigFile = fi.StringValue(b.Cluster.Spec.Containerd.ConfigFile)
		}

		t := &nodetasks.File{
			Path:     "/etc/containerd/config-kops.toml",
			Contents: fi.NewStringResource(containerdConfigFile),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	containerdVersion := b.containerdVersion()

	// Add packages
	{
		count := 0
		for i := range containerdVersions {
			dv := &containerdVersions[i]
			if !dv.matches(b.Architecture, containerdVersion, b.Distribution) {
				continue
			}

			count++

			var packageTask fi.Task
			if dv.PlainBinary {
				packageTask = &nodetasks.Archive{
					Name:            "containerd",
					Source:          dv.Source,
					Hash:            dv.Hash,
					TargetDir:       "/usr/bin/",
					StripComponents: 1,
				}
				c.AddTask(packageTask)
			} else {
				var extraPkgs []*nodetasks.Package
				for name, pkg := range dv.ExtraPackages {
					dep := &nodetasks.Package{
						Name:         name,
						Version:      s(pkg.Version),
						Source:       s(pkg.Source),
						Hash:         s(pkg.Hash),
						PreventStart: fi.Bool(true),
					}
					extraPkgs = append(extraPkgs, dep)
				}
				packageTask = &nodetasks.Package{
					Name:    dv.Name,
					Version: s(dv.Version),
					Source:  s(dv.Source),
					Hash:    s(dv.Hash),
					Deps:    extraPkgs,

					// TODO: PreventStart is now unused?
					PreventStart: fi.Bool(true),
				}
				c.AddTask(packageTask)
			}

			// As a mitigation for CVE-2019-5736 (possibly a fix, definitely defense-in-depth) we chattr docker-runc to be immutable
			for _, f := range dv.MarkImmutable {
				c.AddTask(&nodetasks.Chattr{
					File: f,
					Mode: "+i",
					Deps: []fi.Task{packageTask},
				})
			}

			for _, dep := range dv.Dependencies {
				c.AddTask(&nodetasks.Package{Name: dep})
			}

			// Note we do _not_ stop looping... centos/rhel comprises multiple packages
		}

		if count == 0 {
			klog.Warningf("Did not find containerd package for %s %s %s", b.Distribution, b.Architecture, containerdVersion)
		}
	}

	c.AddTask(b.buildSystemdService())

	if err := b.buildSysconfig(c); err != nil {
		return err
	}

	return nil
}

func (b *ContainerdBuilder) buildSystemdService() *nodetasks.Service {
	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "containerd container runtime")
	manifest.Set("Unit", "Documentation", "https://containerd.io")
	manifest.Set("Unit", "After", "network.target local-fs.target")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/containerd")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")
	manifest.Set("Service", "ExecStartPre", "-/sbin/modprobe overlay")
	manifest.Set("Service", "ExecStart", "/usr/bin/containerd -c /etc/containerd/config-kops.toml \"$CONTAINERD_OPTS\"")

	// kill only the containerd process, not all processes in the cgroup
	manifest.Set("Service", "KillMode", "process")
	// set delegate yes so that systemd does not reset the cgroups of containerd containers
	manifest.Set("Service", "Delegate", "yes")

	manifest.Set("Service", "LimitNOFILE", "1048576")
	manifest.Set("Service", "LimitNPROC", "infinity")
	manifest.Set("Service", "LimitCORE", "infinity")
	manifest.Set("Service", "TasksMax", "infinity")

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")
	manifest.Set("Service", "TimeoutStartSec", "0")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	manifestString := manifest.Render()
	klog.V(8).Infof("Built service manifest %q\n%s", "containerd", manifestString)

	service := &nodetasks.Service{
		Name:       "containerd.service",
		Definition: s(manifestString),
	}

	service.InitDefaults()

	return service
}

// buildContainerOSConfigurationDropIn is responsible for configuring the containerd daemon options
func (b *ContainerdBuilder) buildContainerOSConfigurationDropIn(c *fi.ModelBuilderContext) error {
	lines := []string{
		"[Service]",
		"EnvironmentFile=/etc/sysconfig/containerd",
		"EnvironmentFile=/etc/environment",
		"TasksMax=infinity",
	}

	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		AfterFiles: []string{"/etc/sysconfig/containerd"},
		Path:       "/etc/systemd/system/containerd.service.d/10-kops.conf",
		Contents:   fi.NewStringResource(contents),
		Type:       nodetasks.FileType_File,
		OnChangeExecute: [][]string{
			{"systemctl", "daemon-reload"},
			{"systemctl", "restart", "containerd.service"},
			// We need to restart kops-configuration service since nodeup needs to load images
			// into containerd with the new config. Restart is on the background because
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

// buildSysconfig is responsible for extracting the containerd configuration and writing the sysconfig file
func (b *ContainerdBuilder) buildSysconfig(c *fi.ModelBuilderContext) error {
	var containerd kops.ContainerdConfig
	if b.Cluster.Spec.Containerd != nil {
		containerd = *b.Cluster.Spec.Containerd
	}

	flagsString, err := flagbuilder.BuildFlags(&containerd)
	if err != nil {
		return fmt.Errorf("error building containerd flags: %v", err)
	}

	lines := []string{
		"CONTAINERD_OPTS=" + flagsString,
	}
	contents := strings.Join(lines, "\n")

	c.AddTask(&nodetasks.File{
		Path:     "/etc/sysconfig/containerd",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})

	return nil
}

// skipInstall determines if kops should skip the installation and configuration of containerd
func (b *ContainerdBuilder) skipInstall() bool {
	d := b.Cluster.Spec.Containerd

	// don't skip install if the user hasn't specified anything
	if d == nil {
		return false
	}

	return d.SkipInstall
}

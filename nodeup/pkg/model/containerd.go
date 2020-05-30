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
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/architectures"
)

// ContainerdBuilder install containerd (just the packages at the moment)
type ContainerdBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &ContainerdBuilder{}

var containerdVersions = []packageVersion{
	// 1.2.4 - Debian Stretch
	{
		PackageVersion: "1.2.4",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.4-1",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/containerd.io_1.2.4-1_amd64.deb",
		Hash:           "5d4eeec093bc6f0b35921b88c3939b480acc619c790f4eab001a66efb957e6c1",
	},

	// 1.2.10 - Debian Stretch
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian9},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/debian/dists/stretch/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "786a625f773ec5ac5dc4ebd9463ba7deb6926da890fa57a9ac79be7a6839865c",
	},

	// 1.2.10 - Debian Buster
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionDebian10},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/debian/dists/buster/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "355973b23d3d172b5cfb05bc605f2b0cd7145f2fcc572264225d8910701c650d",
	},

	// 1.2.10 - Ubuntu Xenial
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionXenial},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/ubuntu/dists/xenial/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "ec7f4698df82c4dc683a8835ef5a3ecce4df1e41c5d65b4e17558cdccf44952b",
	},

	// 1.2.10 - Ubuntu Bionic
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionBionic},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.10-3",
		Source:         "https://download.docker.com/linux/ubuntu/dists/bionic/pool/stable/amd64/containerd.io_1.2.10-3_amd64.deb",
		Hash:           "ffa99f3f4b3a76c66f04ad327b8eb6b437e311887e3e0330447fbf5ea2ddd827",
	},

	// 1.2.10 - CentOS7 / Rhel 7
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionRhel7, distros.DistributionCentos7},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.10",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.x86_64.rpm",
		Hash:           "8f29304480a197343fb6fb0fd28e8753aed4e177545568cb4ea3b0f0c02ebf82",
	},

	// 1.2.10 - CentOS8 / Rhel 8
	{
		PackageVersion: "1.2.10",
		Name:           "containerd.io",
		Distros:        []distros.Distribution{distros.DistributionRhel8, distros.DistributionCentos8},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Version:        "1.2.10",
		Source:         "https://download.docker.com/linux/centos/7/x86_64/stable/Packages/containerd.io-1.2.10-3.2.el7.x86_64.rpm",
		Hash:           "8f29304480a197343fb6fb0fd28e8753aed4e177545568cb4ea3b0f0c02ebf82",
	},

	// 1.2.10 - Linux Generic
	//
	// * AmazonLinux2: the Centos7 package depends on container-selinux, but selinux isn't used on amazonlinux2
	// * UbuntuFocal: no focal version available at download.docker.com
	{
		PackageVersion: "1.2.10",
		PlainBinary:    true,
		Distros:        []distros.Distribution{distros.DistributionAmazonLinux2, distros.DistributionFocal},
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.10.linux-amd64.tar.gz",
		Hash:           "9125a6ae5a89dfe9403fea7d03a8d8ba9fa97b6863ee8698c4e6c258fb14f1fd",
	},

	// 1.2.11 - Linux Generic
	{
		PackageVersion: "1.2.11",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.11.linux-amd64.tar.gz",
		Hash:           "df89b00e927115f1f21eec139e55668a6284abb2f378512677c99a3751579e51",
	},

	// 1.2.12 - Linux Generic
	{
		PackageVersion: "1.2.12",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.12.linux-amd64.tar.gz",
		Hash:           "291c26f97546ad92ce51e1584c002b286a82a64f1138cf11f78ef7c91adf5c80",
	},

	// 1.2.13 - Linux Generic
	{
		PackageVersion: "1.2.13",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.2.13.linux-amd64.tar.gz",
		Hash:           "92d6ae6c60f6b068652b31811ce23d650ec0f6cc1e618ec9ae23db9321956258",
	},

	// 1.3.2 - Linux Generic
	{
		PackageVersion: "1.3.2",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.2.linux-amd64.tar.gz",
		Hash:           "95cf4d2cfa23c7a586980c51f8c283a9f0717e09d1a3cc54fc6ed7984923b7aa",
	},

	// 1.3.3 - Linux Generic
	{
		PackageVersion: "1.3.3",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.3.linux-amd64.tar.gz",
		Hash:           "24ce7ad6b489fb25d07d2a3bb50e443fcce1ac3318f8cc0831e00668c2c9fd86",
	},

	// 1.3.4 - Linux Generic
	{
		PackageVersion: "1.3.4",
		PlainBinary:    true,
		Architectures:  []architectures.Architecture{architectures.ArchitectureAmd64},
		Source:         "https://storage.googleapis.com/cri-containerd-release/cri-containerd-1.3.4.linux-amd64.tar.gz",
		Hash:           "4616971c3ad21c24f2f2320fa1c085577a91032a068dd56a41c7c4b71a458087",
	},

	// TIP: When adding the next version, copy the previous version, string replace the version and run:
	//   VERIFY_HASHES=1 go test -v ./nodeup/pkg/model -run TestContainerdPackageHashes
	// (you might want to temporarily comment out older versions on a slower connection and then validate)
}

func (b *ContainerdBuilder) containerdVersion() (string, error) {
	containerdVersion := ""
	if b.Cluster.Spec.Containerd != nil {
		containerdVersion = fi.StringValue(b.Cluster.Spec.Containerd.Version)
	}
	if containerdVersion == "" {
		return "", fmt.Errorf("error finding containerd version")
	}
	return containerdVersion, nil
}

// Build is responsible for configuring the containerd daemon
func (b *ContainerdBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.skipInstall() {
		klog.Infof("SkipInstall is set to true; won't install containerd")
		return nil
	}

	// @check: neither flatcar nor containeros need provision containerd.service, just the containerd daemon options
	switch b.Distribution {
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
		containerdConfigOverride := ""
		if b.Cluster.Spec.Containerd != nil {
			containerdConfigOverride = fi.StringValue(b.Cluster.Spec.Containerd.ConfigOverride)
		}

		t := &nodetasks.File{
			Path:     "/etc/containerd/config-kops.toml",
			Contents: fi.NewStringResource(containerdConfigOverride),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	containerdVersion, err := b.containerdVersion()
	if err != nil {
		return err
	}

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
					Name:      "containerd.io",
					Source:    dv.Source,
					Hash:      dv.Hash,
					TargetDir: "/",
					MapFiles: map[string]string{
						"./usr/local/bin":  "/usr",
						"./usr/local/sbin": "/usr",
					},
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

	// Using containerd with Kubenet requires special configuration. This is a temporary backwards-compatible solution
	// and will be deprecated when Kubenet is deprecated:
	// https://github.com/containerd/cri/blob/master/docs/config.md#cni-config-template
	usesKubenet := components.UsesKubenet(b.Cluster.Spec.Networking)
	if b.Cluster.Spec.ContainerRuntime == "containerd" && usesKubenet {
		b.buildKubenetCNIConfigTemplate(c)
	}

	return nil
}

func (b *ContainerdBuilder) buildSystemdService() *nodetasks.Service {
	// Based on https://github.com/containerd/cri/blob/master/contrib/systemd-units/containerd.service

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "containerd container runtime")
	manifest.Set("Unit", "Documentation", "https://containerd.io")
	manifest.Set("Unit", "After", "network.target local-fs.target")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/containerd")
	manifest.Set("Service", "EnvironmentFile", "/etc/environment")
	manifest.Set("Service", "ExecStartPre", "-/sbin/modprobe overlay")
	manifest.Set("Service", "ExecStart", "/usr/bin/containerd -c /etc/containerd/config-kops.toml \"$CONTAINERD_OPTS\"")

	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "5")

	// set delegate yes so that systemd does not reset the cgroups of containerd containers
	manifest.Set("Service", "Delegate", "yes")
	// kill only the containerd process, not all processes in the cgroup
	manifest.Set("Service", "KillMode", "process")
	// make killing of processes of this unit under memory pressure very unlikely
	manifest.Set("Service", "OOMScoreAdjust", "-999")

	manifest.Set("Service", "LimitNOFILE", "1048576")
	manifest.Set("Service", "LimitNPROC", "infinity")
	manifest.Set("Service", "LimitCORE", "infinity")
	manifest.Set("Service", "TasksMax", "infinity")

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

// buildKubenetCNIConfigTemplate is responsible for creating a special template for setups using Kubenet
func (b *ContainerdBuilder) buildKubenetCNIConfigTemplate(c *fi.ModelBuilderContext) {
	lines := []string{
		"{",
		"    \"cniVersion\": \"0.3.1\",",
		"    \"name\": \"kubenet\",",
		"    \"plugins\": [",
		"        {",
		"            \"type\": \"bridge\",",
		"            \"bridge\": \"cbr0\",",
		"            \"mtu\": 1460,",
		"            \"addIf\": \"eth0\",",
		"            \"isGateway\": true,",
		"            \"ipMasq\": true,",
		"            \"promiscMode\": true,",
		"            \"ipam\": {",
		"                \"type\": \"host-local\",",
		"                \"subnet\": \"{{.PodCIDR}}\",",
		"                \"routes\": [{ \"dst\": \"0.0.0.0/0\" }]",
		"            }",
		"        }",
		"    ]",
		"}",
	}
	contents := strings.Join(lines, "\n")
	klog.V(8).Infof("Built kubenet CNI config file\n%s", contents)

	c.AddTask(&nodetasks.File{
		Path:     "/etc/containerd/cni-config.template",
		Contents: fi.NewStringResource(contents),
		Type:     nodetasks.FileType_File,
	})
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

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
	"sort"
	"strconv"

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// s is a helper that builds a *string from a string value
func s(v string) *string {
	return fi.String(v)
}

// i64 is a helper that builds a *int64 from an int64 value
func i64(v int64) *int64 {
	return fi.Int64(v)
}

// b returns a pointer to a boolean
func b(v bool) *bool {
	return fi.Bool(v)
}

// containsRole checks if a collection roles contains role v
func containsRole(v kops.InstanceGroupRole, list []kops.InstanceGroupRole) bool {
	for _, x := range list {
		if v == x {
			return true
		}
	}

	return false
}

// buildDockerEnvironmentVars just converts a series of keypairs to docker environment variables switches
func buildDockerEnvironmentVars(env map[string]string) []string {
	var list []string
	for k, v := range env {
		list = append(list, []string{"-e", fmt.Sprintf("%s=%s", k, v)}...)
	}

	return list
}

// sortedStrings is just a one liner helper methods
func sortedStrings(list []string) []string {
	sort.Strings(list)

	return list
}

// addHostPathMapping is shorthand for mapping a host path into a container
func addHostPathMapping(pod *v1.Pod, container *v1.Container, name, path string) *v1.VolumeMount {
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: path,
			},
		},
	})

	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:      name,
		MountPath: path,
		ReadOnly:  true,
	})

	return &container.VolumeMounts[len(container.VolumeMounts)-1]
}

// addHostPathVolume is shorthand for mapping a host path into a container
func addHostPathVolume(pod *v1.Pod, container *v1.Container, hostPath v1.HostPathVolumeSource, volumeMount v1.VolumeMount) {
	vol := v1.Volume{
		Name: volumeMount.Name,
		VolumeSource: v1.VolumeSource{
			HostPath: &hostPath,
		},
	}

	if volumeMount.MountPath == "" {
		volumeMount.MountPath = hostPath.Path
	}

	pod.Spec.Volumes = append(pod.Spec.Volumes, vol)
	container.VolumeMounts = append(container.VolumeMounts, volumeMount)
}

// convEtcdSettingsToMs converts etcd settings to a string rep of int milliseconds
func convEtcdSettingsToMs(dur *metav1.Duration) string {
	return strconv.FormatInt(dur.Nanoseconds()/1000000, 10)
}

// packageInfo - fields required for extra packages setup
type packageInfo struct {
	Version string // Package version
	Source  string // URL to download the package from
	Hash    string // sha1sum of the package file
}

// packageVersion - fields required for downloaded packages setup
type packageVersion struct {
	Name string

	// Version is the version of docker, as specified in the kops
	Version string

	// Source is the url where the package/tarfile can be found
	Source string

	// Hash is the sha1 hash of the file
	Hash string

	// Extra packages to install during the same dpkg/yum transaction.
	// This is used for:
	//   - On RHEL/CentOS, the SELinux policy needs to be installed.
	//   - Starting from Docker 18.09, the Docker package has been split in 3
	//     separate packages: one for the daemon, one for the CLI, one for
	//     containerd.  All 3 must be installed at the same time when
	//     upgrading from an older version of Docker.
	ExtraPackages map[string]packageInfo

	PackageVersion string
	Distros        []distros.Distribution
	// List of dependencies that can be installed using the system's package
	// manager (e.g. apt-get install or yum install).
	Dependencies  []string
	Architectures []Architecture

	// PlainBinary indicates that the Source is not an OS, but a "bare" tar.gz
	PlainBinary bool

	// MarkImmutable is a list of files on which we should perform a `chattr +i <file>`
	MarkImmutable []string
}

// Match package version against configured values
func (d *packageVersion) matches(arch Architecture, packageVersion string, distro distros.Distribution) bool {
	if d.PackageVersion != packageVersion {
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
	return foundArch
}

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

package kubemanifest

import (
	v1 "k8s.io/api/core/v1"
)

// MapEtcHosts maps the /etc/hosts file into the pod (useful for gossip DNS)
func MapEtcHosts(pod *v1.Pod, container *v1.Container, readOnly bool) {
	mapping := AddHostPathMapping(pod, container, "hosts", "/etc/hosts").WithType(v1.HostPathFile)
	mapping.VolumeMount.ReadOnly = readOnly
}

// HostPathMapping allows fluent construction of a hostpath mount
type HostPathMapping struct {
	VolumeMount *v1.VolumeMount
	Volume      *v1.Volume
}

// AddHostPathMapping is a helper function for mapping a host path into a container
// It returns a HostPathMapping for tweaking the defaults (which are notably read-only)
func AddHostPathMapping(pod *v1.Pod, container *v1.Container, name, path string) *HostPathMapping {
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

	return &HostPathMapping{
		Volume:      &pod.Spec.Volumes[len(pod.Spec.Volumes)-1],
		VolumeMount: &container.VolumeMounts[len(container.VolumeMounts)-1],
	}
}

// WithReadWrite changes the hostpath mapping to be read-write (the default is read-only)
func (m *HostPathMapping) WithReadWrite() *HostPathMapping {
	m.VolumeMount.ReadOnly = false
	return m
}

// WithType changes the hostpath mount type
func (m *HostPathMapping) WithType(t v1.HostPathType) *HostPathMapping {
	m.Volume.VolumeSource.HostPath.Type = &t
	return m
}

// WithHostPath changes the hostpath path
func (m *HostPathMapping) WithHostPath(p string) *HostPathMapping {
	m.Volume.VolumeSource.HostPath.Path = p
	return m
}

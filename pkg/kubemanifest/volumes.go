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

// HostPathMappingOption implements the "functional options pattern" for named variable parameters.
type HostPathMappingOption func(volumeMount *v1.VolumeMount, volume *v1.Volume)

// AddHostPathMapping is a helper function for mapping a host path into a container
// It returns a HostPathMapping for tweaking the defaults (which are notably read-only)
func AddHostPathMapping(pod *v1.Pod, container *v1.Container, name, path string, options ...HostPathMappingOption) {
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

	volume := &pod.Spec.Volumes[len(pod.Spec.Volumes)-1]
	volumeMount := &container.VolumeMounts[len(container.VolumeMounts)-1]

	for _, option := range options {
		option(volumeMount, volume)
	}
}

// WithReadWrite changes the hostpath mapping to be read-write (the default is read-only)
func WithReadWrite() HostPathMappingOption {
	return func(volumeMount *v1.VolumeMount, volume *v1.Volume) {
		volumeMount.ReadOnly = false
	}
}

// WithType changes the hostpath mount type
func WithType(t v1.HostPathType) HostPathMappingOption {
	return func(volumeMount *v1.VolumeMount, volume *v1.Volume) {
		volume.VolumeSource.HostPath.Type = &t
	}
}

// WithHostPath changes the host path (the path in the host)
func WithHostPath(p string) HostPathMappingOption {
	return func(volumeMount *v1.VolumeMount, volume *v1.Volume) {
		volume.VolumeSource.HostPath.Path = p
	}
}

// WithMountPath changes the mount path (the path in the container)
func WithMountPath(p string) HostPathMappingOption {
	return func(volumeMount *v1.VolumeMount, volume *v1.Volume) {
		volumeMount.MountPath = p
	}
}

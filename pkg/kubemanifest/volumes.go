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
	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:      "hosts",
		MountPath: "/etc/hosts",
		ReadOnly:  readOnly,
	})
	hostPathFile := v1.HostPathFile
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: "hosts",
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: "/etc/hosts",
				Type: &hostPathFile,
			},
		},
	})
}

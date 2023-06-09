/*
Copyright 2023 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/apis/nodeup"
)

// AddHostPathSELinuxContext allows a non-privileged pod to access any file on
// the Host via HostPath volumes when SELinux is enabled.
func AddHostPathSELinuxContext(pod *v1.Pod, cfg *nodeup.Config) {
	if cfg.ContainerdConfig == nil || !cfg.ContainerdConfig.SeLinuxEnabled {
		return
	}

	if pod.Spec.SecurityContext == nil {
		pod.Spec.SecurityContext = &v1.PodSecurityContext{}
	}

	// This context basically disables all SELinux checks for the pod.
	// Among other things, it allows the pod to access HostPath volumes.
	// The option is ignored on non-selinux systems.
	pod.Spec.SecurityContext.SELinuxOptions = &v1.SELinuxOptions{
		Type:  "spc_t",
		Level: "s0",
	}
}

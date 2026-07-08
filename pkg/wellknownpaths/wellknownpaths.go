/*
Copyright 2026 The Kubernetes Authors.

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

// Package wellknownpaths holds file paths that are shared between the
// cloudup model (which generates the files) and nodeup (which writes them
// to the nodes).
package wellknownpaths

const (
	// KubeSchedulerConfig is the path where we write the kube-scheduler config file (on the control-plane nodes).
	KubeSchedulerConfig = "/var/lib/kube-scheduler/config.yaml"

	// KubeSchedulerKubeConfig is the path where we write the kube-scheduler kubeconfig file (on the control-plane nodes).
	KubeSchedulerKubeConfig = "/var/lib/kube-scheduler/kubeconfig"
)

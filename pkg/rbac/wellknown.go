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

package rbac

// well-known user and group names
// Copied from k8s.io/apiserver; we don't want to depend on that any more
const (
	SystemPrivilegedGroup = "system:masters"
	NodesGroup            = "system:nodes"
	AllUnauthenticated    = "system:unauthenticated"
	AllAuthenticated      = "system:authenticated"

	Anonymous     = "system:anonymous"
	APIServerUser = "system:apiserver"

	// core kubernetes process identities
	KubeProxy             = "system:kube-proxy"
	KubeControllerManager = "system:kube-controller-manager"
	KubeScheduler         = "system:kube-scheduler"
)

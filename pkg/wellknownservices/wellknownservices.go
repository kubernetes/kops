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

package wellknownservices

type WellKnownService string

const (
	// KubeAPIServerExternal is the service where kube-apiserver listens for user traffic.
	KubeAPIServerExternal WellKnownService = "kube-apiserver-external"

	// KubeAPIServerInternal is the service where kube-apiserver listens for internal (in-cluster) traffic.
	// Note that this might still be exposed publicly, "internal" refers to whether the source of the traffic
	// is from inside or outside the cluster.
	KubeAPIServerInternal WellKnownService = "kube-apiserver-internal"

	// KopsControllerInternal is the service where kops-controller listens for internal (in-cluster) traffic.
	// As with KubeAPIServerInternal, this might still be exposed publicly,
	// "internal" refers to whether whether the source of the traffic is from inside or outside the cluster.
	// There is no "KopsControllerExternal" because the only client of kops-controller should be the Nodes;
	// and generally we do not need or want to expose kops-controller to the internet.
	// However, on some clouds it's not easy to restrict access, and we don't rely on kops-controller being
	// unreachable publicly.
	KopsControllerInternal WellKnownService = "kops-controller"
)

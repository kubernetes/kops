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

package declarative

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

// This will likely become a topological sort in the future, for owner refs.
// For now, we implement a simple kind-based heuristic for the sort.

func DefaultObjectOrder(ctx context.Context) func(o *manifest.Object) int {
	log := log.Log

	return func(o *manifest.Object) int {
		gk := o.Group + "/" + o.Kind
		switch gk {
		// Create CRDs asap - both because they are slow and because we will likely create instances of them soon
		case "apiextensions.k8s.io/CustomResourceDefinition":
			return -1000

		// We need to create ServiceAccounts, Roles before we bind them with a RoleBinding
		case "/ServiceAccount", "rbac.authorization.k8s.io/ClusterRole":
			return 1
		case "rbac.authorization.k8s.io/ClusterRoleBinding":
			return 2

		// Pods might need configmap or secrets - avoid backoff by creating them first
		case "/ConfigMap", "/Secrets":
			return 100

		// Create the pods after we've created other things they might be waiting for
		case "extensions/Deployment", "apps/Deployment":
			return 1000

		// Autoscalers typically act on a deployment
		case "autoscaling/HorizontalPodAutoscaler":
			return 1001

		// Create services late - after pods have been started
		case "/Service":
			return 10000

		default:
			log.WithValues("group", o.Group).WithValues("kind", o.Kind).V(2).Info("unknown group / kind")
			return 1000
		}
	}
}

/*
Copyright 2025 The Kubernetes Authors.

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

package clusterapi

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kops/clusterapi/controlplane/kops/api/v1beta1"
	clusterv1 "k8s.io/kops/clusterapi/snapshot/cluster-api/api/v1beta1"
	kopsapi "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func makeRef(u *unstructured.Unstructured) map[string]any {
	apiVersion, kind := u.GroupVersionKind().ToAPIVersionAndKind()
	ref := map[string]any{
		"name":       u.GetName(),
		"apiVersion": apiVersion,
		"kind":       kind,
	}
	return ref
}

func setOwnerRef(u *unstructured.Unstructured, owner client.Object) {
	apiVersion, kind := owner.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

	u.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: apiVersion,
			Kind:       kind,
			Name:       owner.GetName(),
			UID:        owner.GetUID(),
			Controller: PtrTo(true),
		},
	})
}

func PtrTo[T any](t T) *T {
	return &t
}

func getCAPIClusterFromCAPIObject(ctx context.Context, kube client.Client, obj client.Object) (*unstructured.Unstructured, error) {
	id := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}

	capiClusterName := obj.GetLabels()[clusterv1.ClusterNameLabel]
	if capiClusterName == "" {
		return nil, fmt.Errorf("label %q not set on %v", clusterv1.ClusterNameLabel, id)
	}

	capiCluster := &unstructured.Unstructured{}
	capiCluster.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "cluster.x-k8s.io",
		Version: "v1beta1",
		Kind:    "Cluster",
	})
	clusterKey := types.NamespacedName{
		Namespace: obj.GetNamespace(),
		Name:      capiClusterName,
	}
	if err := kube.Get(ctx, clusterKey, capiCluster); err != nil {
		return nil, fmt.Errorf("error fetching cluster %v: %w", clusterKey, err)
	}

	return capiCluster, nil
}

func getKopsClusterFromCAPICluster(ctx context.Context, kube client.Client, capiCluster *unstructured.Unstructured) (*kopsapi.Cluster, error) {
	var clusterKey types.NamespacedName

	for _, ownerRef := range capiCluster.GetOwnerReferences() {
		if ownerRef.Kind == "Cluster" && strings.HasPrefix(ownerRef.APIVersion, "kops.k8s.io/") {
			clusterKey = types.NamespacedName{
				Namespace: capiCluster.GetNamespace(),
				Name:      ownerRef.Name,
			}
		}
	}

	if clusterKey.Name == "" {
		return nil, fmt.Errorf("cluster ownerRef not set on CAPI cluster %v/%v", capiCluster.GetNamespace(), capiCluster.GetName())
	}

	cluster := &kopsapi.Cluster{}
	if err := kube.Get(ctx, clusterKey, cluster); err != nil {
		return nil, fmt.Errorf("error fetching cluster %v: %w", clusterKey, err)
	}

	return cluster, nil
}

func getKopsControlPlaneFromCAPICluster(ctx context.Context, kube client.Client, capiCluster *unstructured.Unstructured) (*v1beta1.KopsControlPlane, error) {
	capiClusterKey := types.NamespacedName{
		Namespace: capiCluster.GetNamespace(),
		Name:      capiCluster.GetName(),
	}
	name, _, _ := unstructured.NestedString(capiCluster.Object, "spec", "controlPlaneRef", "name")
	if name == "" {
		return nil, fmt.Errorf("controlPlaneRef.name not set for %v", capiClusterKey)
	}
	kind, _, _ := unstructured.NestedString(capiCluster.Object, "spec", "controlPlaneRef", "kind")
	if kind == "" {
		return nil, fmt.Errorf("controlPlaneRef.kind not set for %v", capiClusterKey)
	}
	if kind != "KopsControlPlane" {
		return nil, fmt.Errorf("controlPlaneRef.kind was %q for %v, expected KopsControlPlane", kind, capiClusterKey)
	}

	key := types.NamespacedName{
		Namespace: capiCluster.GetNamespace(),
		Name:      name,
	}

	// TODO: Add ip addresses to status

	kopsControlPlane := &v1beta1.KopsControlPlane{}
	if err := kube.Get(ctx, key, kopsControlPlane); err != nil {
		return nil, fmt.Errorf("error fetching KopsControlPlane %v: %w", key, err)
	}

	return kopsControlPlane, nil
}

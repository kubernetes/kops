/*
Copyright 2024 The Kubernetes Authors.

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

package controllers

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/clusterapi/pkg/builders"
	kopsinternal "k8s.io/kops/pkg/apis/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops/v1alpha2"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewInstanceGroupReconciler is the constructor for an InstanceGroupReconciler
func NewInstanceGroupReconciler(mgr manager.Manager) error {
	r := &InstanceGroupReconciler{
		client: mgr.GetClient(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kopsapi.InstanceGroup{}).
		Complete(r)
}

// InstanceGroupReconciler observes Node objects, and labels them with the correct labels for the instancegroup
// This used to be done by the kubelet, but is moving to a central controller for greater security in 1.16
type InstanceGroupReconciler struct {
	// client is the controller-runtime client
	client client.Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *InstanceGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instanceGroup := &kopsapi.InstanceGroup{}
	if err := r.client.Get(ctx, req.NamespacedName, instanceGroup); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cluster, err := r.getCluster(ctx, instanceGroup)
	if err != nil {
		return ctrl.Result{}, err
	}

	replicas := 1
	if instanceGroup.Spec.MinSize != nil {
		replicas = int(*instanceGroup.Spec.MinSize)
	}

	subnets := sets.New[string]()
	for _, s := range instanceGroup.Spec.Subnets {
		subnets.Insert(s)
	}
	if subnets.Len() > 1 {
		return ctrl.Result{}, fmt.Errorf("found multiple subnets for instance group %q: %v; only a single subnet is supported", instanceGroup.Name, sets.List(subnets))
	}
	if subnets.Len() == 0 {
		return ctrl.Result{}, fmt.Errorf("cannot determine subnet for instance group %q: no subnets defined", instanceGroup.Name)
	}
	subnet := sets.List(subnets)[0]

	builder := &builders.MachineDeploymentBuilder{
		ClusterName:       cluster.Name,
		Name:              instanceGroup.Name,
		Namespace:         "kube-system",
		Replicas:          replicas,
		Zones:             instanceGroup.Spec.Zones,
		MachineType:       instanceGroup.Spec.MachineType,
		Subnet:            subnet,
		Image:             instanceGroup.Spec.Image,
		Role:              kopsinternal.InstanceGroupRole(instanceGroup.Spec.Role),
		KubernetesVersion: cluster.Spec.KubernetesVersion,
	}

	builder.AdditionalMetadata = map[string]string{
		nodeidentitygce.MetadataKeyInstanceGroupName: instanceGroup.Name,
	}

	objects, err := builder.BuildObjects(ctx)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error building machine deployments for instance group %q: %w", instanceGroup.Name, err)
	}

	for _, obj := range objects {
		if err := r.ssa(ctx, r.client, obj); err != nil {
			return ctrl.Result{}, fmt.Errorf("error applying %s/%s: %w", obj.GetKind(), obj.GetName(), err)
		}
	}

	return ctrl.Result{}, nil
}

func (r *InstanceGroupReconciler) getCluster(ctx context.Context, ig *kopsapi.InstanceGroup) (*kopsapi.Cluster, error) {
	clusters := &kopsapi.ClusterList{}
	if err := r.client.List(ctx, clusters, client.InNamespace(ig.Namespace)); err != nil {
		return nil, fmt.Errorf("listing clusters in namespace %q: %w", ig.Namespace, err)
	}
	if len(clusters.Items) == 0 {
		return nil, fmt.Errorf("no cluster found in namespace %q", ig.Namespace)
	}
	if len(clusters.Items) > 1 {
		return nil, fmt.Errorf("multiple clusters found in namespace %q", ig.Namespace)
	}
	return &clusters.Items[0], nil
}

func (s *InstanceGroupReconciler) ssa(ctx context.Context, kube client.Client, u *unstructured.Unstructured) error {
	return kube.Patch(ctx, u, client.Apply, client.FieldOwner("kops-instancegroup-controller"))
}

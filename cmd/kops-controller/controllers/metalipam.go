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

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	kopsapi "k8s.io/kops/pkg/apis/kops/v1alpha2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewMetalIPAMReconciler is the constructor for a MetalIPAMReconciler
func NewMetalIPAMReconciler(ctx context.Context, mgr manager.Manager) (*MetalIPAMReconciler, error) {
	klog.Info("starting metal ipam controller")
	r := &MetalIPAMReconciler{
		client: mgr.GetClient(),
		log:    ctrl.Log.WithName("controllers").WithName("metal_ipam"),
	}

	coreClient, err := corev1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("building corev1 client: %w", err)
	}
	r.coreV1Client = coreClient

	return r, nil
}

// MetalIPAMReconciler observes Node objects, assigning their `PodCIDRs` from the instance's `ExternalIpv6`.
type MetalIPAMReconciler struct {
	// client is the controller-runtime client
	client client.Client

	// log is a logr
	log logr.Logger

	// coreV1Client is a client-go client for patching nodes
	coreV1Client *corev1client.CoreV1Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *MetalIPAMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := klog.FromContext(ctx)

	node := &corev1.Node{}
	if err := r.client.Get(ctx, req.NamespacedName, node); err != nil {
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch node", "node.name", node.Name)
		return ctrl.Result{}, err
	}

	host := &kopsapi.Host{}
	id := types.NamespacedName{
		Namespace: "kops-system",
		Name:      node.Name,
	}
	if err := r.client.Get(ctx, id, host); err != nil {
		log.Error(err, "unable to fetch host", "id", id)
		return ctrl.Result{}, err
	}

	if len(node.Spec.PodCIDRs) != 0 && len(host.Spec.PodCIDRs) > 0 {
		log.V(2).Info("node has pod cidrs, skipping", "node.name", node.Name)
		return ctrl.Result{}, nil
	}

	if len(host.Spec.PodCIDRs) == 0 {
		log.Info("host record has no pod cidrs, cannot assign pod cidrs to node", "node.name", node.Name)
		return ctrl.Result{}, nil
	}

	log.Info("assigning pod cidrs to node", "node.name", node.Name, "pod.cidrs", host.Spec.PodCIDRs)
	if err := patchNodePodCIDRs(r.coreV1Client, ctx, node, host.Spec.PodCIDRs); err != nil {
		log.Error(err, "unable to patch node", "node.name", node.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MetalIPAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("metal_ipam").
		For(&corev1.Node{}).
		Complete(r)
}

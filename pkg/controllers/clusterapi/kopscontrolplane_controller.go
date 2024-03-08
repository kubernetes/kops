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

package clusterapi

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"k8s.io/kops/clusterapi/controlplane/kops/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewKopsControlPlaneReconciler is the constructor for a KopsControlPlaneReconciler
func NewKopsControlPlaneReconciler(mgr manager.Manager) error {
	r := &KopsControlPlaneReconciler{
		client: mgr.GetClient(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1beta1.KopsControlPlane{}).
		Complete(r)
}

// KopsControlPlaneReconciler observes KopsControlPlane objects.
type KopsControlPlaneReconciler struct {
	// client is the controller-runtime client
	client client.Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch

// Reconcile is the main reconciler function that observes node changes.
func (r *KopsControlPlaneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &v1beta1.KopsControlPlane{}
	if err := r.client.Get(ctx, req.NamespacedName, obj); err != nil {
		klog.Warningf("unable to fetch object: %v", err)
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	capiCluster, err := getCAPIClusterFromCAPIObject(ctx, r.client, obj)
	if err != nil {
		return ctrl.Result{}, err
	}
	cluster, err := getKopsClusterFromCAPICluster(ctx, r.client, capiCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	log := klog.FromContext(ctx)
	log.Info("found cluster", "cluster", cluster)
	// if err := r.client.Status().Update(ctx, obj); err != nil {
	// 	return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	// }
	return ctrl.Result{}, nil
}

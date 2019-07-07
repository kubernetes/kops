/*
Copyright 2019 The Kubernetes authors.

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

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kopsv1alpha1 "k8s.io/kops/cmd/kops-controller/api/v1alpha1"
)

// NodeControllerReconciler reconciles a NodeController object
type NodeControllerReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=kops.kubernetes.io,resources=nodecontrollers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kops.kubernetes.io,resources=nodecontrollers/status,verbs=get;update;patch

func (r *NodeControllerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("nodecontroller", req.NamespacedName)

	// your logic here

	return ctrl.Result{}, nil
}

func (r *NodeControllerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kopsv1alpha1.NodeController{}).
		Complete(r)
}

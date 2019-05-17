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

package instancegroup

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/kops/cmd/kops-controller/pkg/clusterapi"
	api "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func Add(mgr manager.Manager) error {
	r, err := newReconciler(mgr)
	if err != nil {
		return err
	}
	return add(mgr, r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) (reconcile.Reconciler, error) {
	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("error building client: %v", err)
	}

	return &ReconcileInstanceGroup{
		Client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		dynamicClient: dynamicClient,
	}, nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("instancegroup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to InstanceGroup
	err = c.Watch(&source.Kind{Type: &api.InstanceGroup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	/*
		err = c.Watch(&source.Kind{Type: &appsv1.MachineDeployment{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &api.InstanceGroup{},
		})
		if err != nil {
			return err
		}
	*/

	return nil
}

var _ reconcile.Reconciler = &ReconcileInstanceGroup{}

type ReconcileInstanceGroup struct {
	client.Client
	scheme        *runtime.Scheme
	dynamicClient dynamic.Interface
}

// +kubebuilder:rbac:groups=kops.k8s.io,resources=instancegroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kops.k8s.io,resources=instancegroups/status,verbs=get;update;patch
func (r *ReconcileInstanceGroup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.TODO()

	// Fetch the InstanceGroup instance
	instance := &api.InstanceGroup{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	cluster, err := r.getCluster(ctx, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	b := &clusterapi.Builder{
		//	ClientSet: todo,
	}
	md, err := b.BuildMachineDeployment(cluster, instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	if err := r.updateMachineDeployment(md); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileInstanceGroup) updateMachineDeployment(md *unstructured.Unstructured) error {
	gvr := schema.GroupVersionResource{
		Group:    "cluster.k8s.io",
		Version:  "v1alpha1",
		Resource: "machinedeployments",
	}
	namespace := md.GetNamespace()
	name := md.GetName()
	res := r.dynamicClient.Resource(gvr).Namespace(namespace)
	existing, err := res.Get(name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			existing = nil
		} else {
			return fmt.Errorf("error getting machindeployment %s/%s: %v", namespace, name, err)
		}
	}
	if existing == nil {
		var opts metav1.CreateOptions
		if _, err := res.Create(md, opts); err != nil {
			return fmt.Errorf("error creating machindeployment %s/%s: %v", namespace, name, err)
		}
		return nil
	}

	existing.Object["spec"] = md.Object["spec"]

	if _, err := res.Update(existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("error updating machindeployment %s/%s: %v", namespace, name, err)
	}

	return nil
}

func (r *ReconcileInstanceGroup) getCluster(ctx context.Context, ig *api.InstanceGroup) (*api.Cluster, error) {
	clusters := &api.ClusterList{}
	var opts client.ListOptions
	opts.Namespace = ig.Namespace
	err := r.List(ctx, &opts, clusters)
	if err != nil {
		return nil, fmt.Errorf("error fetching clusters: %v", err)
	}

	if len(clusters.Items) == 0 {
		return nil, fmt.Errorf("cluster not found in namespace %q", ig.Namespace)
	}

	if len(clusters.Items) > 1 {
		return nil, fmt.Errorf("multiple clusters found in namespace %q", ig.Namespace)
	}

	return &clusters.Items[0], nil

}

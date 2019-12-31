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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	clusterapi "k8s.io/kops/cmd/kops-controller/pkg/clusterapi"
	api "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// InstanceGroupReconciler reconciles a kops InstanceGroup object, creating a cluster-api MachineDeployment
type InstanceGroupReconciler struct {
	client.Client
	Log           logr.Logger
	DynamicClient dynamic.Interface
	ConfigServer  *cloudup.ConfigServer
}

// +kubebuilder:rbac:groups=kops.k8s.io,resources=instancegroups,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kops.k8s.io,resources=instancegroups/status,verbs=get;update;patch

func (r *InstanceGroupReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.TODO()

	klog.Infof("reconcile %v", req)

	// Fetch the InstanceGroup instance
	instance := &api.InstanceGroup{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	cluster, err := r.getCluster(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	b := &clusterapi.Builder{
		ConfigServer: r.ConfigServer,
	}
	objects, err := b.BuildMachineDeployment(cluster, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, obj := range objects {
		if err := r.updateObject(obj); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *InstanceGroupReconciler) updateObject(obj *unstructured.Unstructured) error {
	gvk := obj.GetObjectKind().GroupVersionKind()

	var resource string

	switch gvk.Group + ":" + gvk.Kind {
	case "cluster.x-k8s.io:MachineDeployment":
		resource = "machinedeployments"
	case "infrastructure.cluster.x-k8s.io:GCPMachineTemplate":
		resource = "gcpmachinetemplates"
	case ":Secret":
		resource = "secrets"
	default:
		return fmt.Errorf("unsupported gvk: %v", gvk)
	}

	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: resource,
	}
	namespace := obj.GetNamespace()
	name := obj.GetName()
	res := r.DynamicClient.Resource(gvr).Namespace(namespace)
	existing, err := res.Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			existing = nil
		} else {
			return fmt.Errorf("error getting %s %s/%s: %v", gvk.Kind, namespace, name, err)
		}
	}
	if existing == nil {
		var opts metav1.CreateOptions
		if _, err := res.Create(obj, opts); err != nil {
			return fmt.Errorf("error creating %s %s/%s: %v", gvk.Kind, namespace, name, err)
		}
		return nil
	}

	// We could do some better logic in here based on the Kind
	existing.Object["spec"] = obj.Object["spec"]
	existing.SetOwnerReferences(obj.GetOwnerReferences())

	// TODO: Skip update if no changes

	if _, err := res.Update(existing, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("error updating %s %s/%s: %v", gvk.Kind, namespace, name, err)
	}

	return nil
}

func (r *InstanceGroupReconciler) getCluster(ctx context.Context, ig *api.InstanceGroup) (*api.Cluster, error) {
	clusters := &api.ClusterList{}
	var opts client.ListOptions
	opts.Namespace = ig.Namespace
	err := r.List(ctx, clusters, &opts)
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

func (r *InstanceGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.InstanceGroup{}).
		Complete(r)
}

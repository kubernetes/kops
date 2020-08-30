/*
Copyright 2020 The Kubernetes Authors.

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
	"encoding/json"
	"fmt"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/nodeidentity"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewNodeReconciler is the constructor for a NodeReconciler
func NewNodeReconciler(mgr manager.Manager, identifier nodeidentity.Identifier) (*NodeReconciler, error) {
	r := &NodeReconciler{
		client:     mgr.GetClient(),
		log:        ctrl.Log.WithName("controllers").WithName("Node"),
		identifier: identifier,
	}

	coreClient, err := corev1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("error building corev1 client: %v", err)
	}
	r.coreV1Client = coreClient

	return r, nil
}

// NodeReconciler observes Node objects, and labels them with the correct labels for the instancegroup
// This used to be done by the kubelet, but is moving to a central controller for greater security in 1.16
type NodeReconciler struct {
	// client is the controller-runtime client
	client client.Client

	// log is a logr
	log logr.Logger

	// coreV1Client is a client-go client for patching nodes
	coreV1Client *corev1client.CoreV1Client

	// identifier is a provider that can securely map node ProviderIDs to labels
	identifier nodeidentity.Identifier
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *NodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.log.WithValues("nodecontroller", req.NamespacedName)

	node := &corev1.Node{}
	if err := r.client.Get(ctx, req.NamespacedName, node); err != nil {
		klog.Warningf("unable to fetch node %s: %v", node.Name, err)
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	info, err := r.identifier.IdentifyNode(ctx, node)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error identifying node %q: %v", node.Name, err)
	}

	labels := info.Labels

	updateLabels := make(map[string]string)
	for k, v := range labels {
		actual, found := node.Labels[k]
		if !found || actual != v {
			updateLabels[k] = v
		}
	}

	if len(updateLabels) == 0 {
		klog.V(4).Infof("no label changes needed for %s", node.Name)
		return ctrl.Result{}, nil
	}

	if err := patchNodeLabels(r.coreV1Client, ctx, node, updateLabels); err != nil {
		klog.Warningf("failed to patch node labels on %s: %v", node.Name, err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

type nodePatch struct {
	Metadata *nodePatchMetadata `json:"metadata,omitempty"`
}

type nodePatchMetadata struct {
	Labels map[string]string `json:"labels,omitempty"`
}

// patchNodeLabels patches the node labels to set the specified labels
func patchNodeLabels(client *corev1client.CoreV1Client, ctx context.Context, node *corev1.Node, setLabels map[string]string) error {
	nodePatchMetadata := &nodePatchMetadata{
		Labels: setLabels,
	}
	nodePatch := &nodePatch{
		Metadata: nodePatchMetadata,
	}
	nodePatchJson, err := json.Marshal(nodePatch)
	if err != nil {
		return fmt.Errorf("error building node patch: %v", err)
	}

	klog.V(2).Infof("sending patch for node %q: %q", node.Name, string(nodePatchJson))

	_, err = client.Nodes().Patch(ctx, node.Name, types.StrategicMergePatchType, nodePatchJson, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch to node: %v", err)
	}

	return nil
}

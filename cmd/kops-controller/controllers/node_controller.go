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
	"reflect"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/nodeidentity"
	"k8s.io/kops/pkg/nodelabels"
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

const externalCloudProviderTaint = "node.cloudprovider.kubernetes.io/uninitialized"

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// +kubebuilder:rbac:groups=,resources=nodes/status,verbs=get;patch;update
// Reconcile is the main reconciler function that observes node changes.
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
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

	deleteLabels := make(map[string]struct{})
	for k := range node.Labels {
		// If it is one of our managed labels, "prune" values we don't want to be there
		switch k {
		case nodelabels.RoleLabelAPIServer16, nodelabels.RoleLabelNode16, nodelabels.RoleLabelControlPlane20:
			if _, found := labels[k]; !found {
				deleteLabels[k] = struct{}{}
			}
		}
	}

	providerID := ""
	if info.ProviderID != "" && node.Spec.ProviderID != info.ProviderID {
		providerID = info.ProviderID
	}

	var taints *[]corev1.Taint
	if info.Initialized {
		if updatedTaints, changed := removeTaint(node.Spec.Taints, externalCloudProviderTaint); changed {
			taints = &updatedTaints
		}
	}

	if len(updateLabels) == 0 && len(deleteLabels) == 0 && providerID == "" && taints == nil {
		klog.V(4).Infof("no spec or label changes needed for %s", node.Name)
	} else if err := patchNode(r.coreV1Client, ctx, node, updateLabels, deleteLabels, providerID, taints); err != nil {
		klog.Warningf("failed to patch node on %s: %v", node.Name, err)
		return ctrl.Result{}, err
	}

	if len(info.Addresses) != 0 && !reflect.DeepEqual(node.Status.Addresses, info.Addresses) {
		if err := patchNodeStatusAddresses(r.coreV1Client, ctx, node, info.Addresses); err != nil {
			klog.Warningf("failed to patch node status addresses on %s: %v", node.Name, err)
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("node").
		For(&corev1.Node{}).
		Complete(r)
}

type nodePatch struct {
	Spec     *nodePatchSpec     `json:"spec,omitempty"`
	Metadata *nodePatchMetadata `json:"metadata,omitempty"`
}

type nodePatchMetadata struct {
	Labels map[string]*string `json:"labels,omitempty"`
}

// patchNodeLabels patches the node labels to set the specified labels
func patchNodeLabels(client *corev1client.CoreV1Client, ctx context.Context, node *corev1.Node, setLabels map[string]string, deleteLabels map[string]struct{}) error {
	return patchNode(client, ctx, node, setLabels, deleteLabels, "", nil)
}

// patchNode patches node metadata and spec fields managed by the node controller.
func patchNode(client *corev1client.CoreV1Client, ctx context.Context, node *corev1.Node, setLabels map[string]string, deleteLabels map[string]struct{}, providerID string, taints *[]corev1.Taint) error {
	nodePatchMetadata := &nodePatchMetadata{
		Labels: make(map[string]*string),
	}
	for k, v := range setLabels {
		v := v
		nodePatchMetadata.Labels[k] = &v
	}
	for k := range deleteLabels {
		nodePatchMetadata.Labels[k] = nil
	}

	nodePatch := &nodePatch{}
	if len(nodePatchMetadata.Labels) != 0 {
		nodePatch.Metadata = nodePatchMetadata
	}
	if providerID != "" {
		nodePatch.Spec = &nodePatchSpec{ProviderID: &providerID}
	}

	if nodePatch.Metadata != nil || nodePatch.Spec != nil {
		nodePatchJson, err := json.Marshal(nodePatch)
		if err != nil {
			return fmt.Errorf("error building node patch: %v", err)
		}

		klog.V(2).Infof("sending patch for node %q: %q", node.Name, string(nodePatchJson))

		_, err = client.Nodes().Patch(ctx, node.Name, types.StrategicMergePatchType, nodePatchJson, metav1.PatchOptions{})
		if err != nil {
			return fmt.Errorf("error applying patch to node: %v", err)
		}
	}

	if taints != nil {
		if err := patchNodeTaints(client, ctx, node, *taints); err != nil {
			return err
		}
	}

	return nil
}

func patchNodeTaints(client *corev1client.CoreV1Client, ctx context.Context, node *corev1.Node, taints []corev1.Taint) error {
	nodePatchJson, err := json.Marshal(struct {
		Spec struct {
			Taints []corev1.Taint `json:"taints"`
		} `json:"spec"`
	}{Spec: struct {
		Taints []corev1.Taint `json:"taints"`
	}{Taints: taints}})
	if err != nil {
		return fmt.Errorf("error building node taints patch: %v", err)
	}

	klog.V(2).Infof("sending taints patch for node %q: %q", node.Name, string(nodePatchJson))

	_, err = client.Nodes().Patch(ctx, node.Name, types.MergePatchType, nodePatchJson, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying taints patch to node: %v", err)
	}

	return nil
}

func patchNodeStatusAddresses(client *corev1client.CoreV1Client, ctx context.Context, node *corev1.Node, addresses []corev1.NodeAddress) error {
	nodePatchJson, err := json.Marshal(struct {
		Status struct {
			Addresses []corev1.NodeAddress `json:"addresses"`
		} `json:"status"`
	}{Status: struct {
		Addresses []corev1.NodeAddress `json:"addresses"`
	}{Addresses: addresses}})
	if err != nil {
		return fmt.Errorf("error building node status patch: %v", err)
	}

	klog.V(2).Infof("sending status patch for node %q: %q", node.Name, string(nodePatchJson))

	_, err = client.Nodes().Patch(ctx, node.Name, types.MergePatchType, nodePatchJson, metav1.PatchOptions{}, "status")
	if err != nil {
		return fmt.Errorf("error applying status patch to node: %v", err)
	}

	return nil
}

func removeTaint(taints []corev1.Taint, key string) ([]corev1.Taint, bool) {
	updated := make([]corev1.Taint, 0, len(taints))
	changed := false
	for _, taint := range taints {
		if taint.Key == key {
			changed = true
			continue
		}
		updated = append(updated, taint)
	}
	return updated, changed
}

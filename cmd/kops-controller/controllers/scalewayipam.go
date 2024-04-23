/*
Copyright 2021 The Kubernetes Authors.

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
	"strings"

	"github.com/go-logr/logr"
	"github.com/scaleway/scaleway-sdk-go/api/ipam/v1"
	"github.com/scaleway/scaleway-sdk-go/scw"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	kopsv "k8s.io/kops"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewScalewayIPAMReconciler is the constructor for a IPAMReconciler
func NewScalewayIPAMReconciler(mgr manager.Manager) (*ScalewayIPAMReconciler, error) {
	klog.Info("Starting scaleway ipam controller")
	r := &ScalewayIPAMReconciler{
		client: mgr.GetClient(),
		log:    ctrl.Log.WithName("controllers").WithName("IPAM"),
	}

	coreClient, err := corev1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("error building corev1 client: %v", err)
	}
	r.coreV1Client = coreClient

	profile, err := scaleway.CreateValidScalewayProfile()
	if err != nil {
		return nil, err
	}
	scwClient, err := scw.NewClient(
		scw.WithProfile(profile),
		scw.WithUserAgent(scaleway.KopsUserAgentPrefix+kopsv.Version),
	)
	if err != nil {
		return nil, fmt.Errorf("creating client for Scaleway IPAM controller: %w", err)
	}
	r.scwClient = scwClient

	return r, nil
}

// ScalewayIPAMReconciler observes Node objects, and labels them with the correct labels for the instancegroup
// This used to be done by the kubelet, but is moving to a central controller for greater security in 1.16
type ScalewayIPAMReconciler struct {
	// client is the controller-runtime client
	client client.Client

	// log is a logr
	log logr.Logger

	// coreV1Client is a client-go client for patching nodes
	coreV1Client *corev1client.CoreV1Client

	scwClient *scw.Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *ScalewayIPAMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.log.WithValues("ipam-controller", req.NamespacedName)

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

	if len(node.Spec.PodCIDRs) == 0 {
		// CCM Node Controller has not done its thing yet
		if node.Spec.ProviderID == "" {
			klog.Infof("Node %q has empty provider ID", node.Name)
			return ctrl.Result{}, nil
		}

		// providerID scaleway://instance/fr-par-1/instance-id
		uuid := strings.Split(node.Spec.ProviderID, "/")
		if len(uuid) != 3 {
			return ctrl.Result{}, fmt.Errorf("unexpected format for server id %s", node.Spec.ProviderID)
		}
		serverID := uuid[2]
		zone := scw.Zone(uuid[1])

		ip, err := scaleway.GetIPAMPublicIP(ipam.NewAPI(r.scwClient), serverID, zone)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("could not get IPAM public IP for server %s: %w", serverID, err)
		}
		if err := patchNodePodCIDRs(r.coreV1Client, ctx, node, ip); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ScalewayIPAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

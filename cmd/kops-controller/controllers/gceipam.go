/*
Copyright 2023 The Kubernetes Authors.

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
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"google.golang.org/api/compute/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewGCEIPAMReconciler is the constructor for a GCEIPAMReconciler
func NewGCEIPAMReconciler(mgr manager.Manager) (*GCEIPAMReconciler, error) {
	klog.Info("starting gce ipam controller")
	r := &GCEIPAMReconciler{
		client: mgr.GetClient(),
		log:    ctrl.Log.WithName("controllers").WithName("gce-ipam"),
	}

	coreClient, err := corev1client.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("building corev1 client: %w", err)
	}
	r.coreV1Client = coreClient

	gceClient, err := compute.NewService(context.Background())
	if err != nil {
		return nil, fmt.Errorf("building compute API client: %w", err)
	}
	r.gceClient = gceClient

	return r, nil
}

// GCEIPAMReconciler observes Node objects, assigning their`PodCIDRs` from the instance's `ExternalIpv6`.
type GCEIPAMReconciler struct {
	// client is the controller-runtime client
	client client.Client

	// log is a logr
	log logr.Logger

	// coreV1Client is a client-go client for patching nodes
	coreV1Client *corev1client.CoreV1Client

	// gceClient is a client for GCE
	gceClient *compute.Service
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *GCEIPAMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.log.WithValues("node", req.NamespacedName)

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
			klog.Infof("node %q has empty provider ID", node.Name)
			return ctrl.Result{}, nil
		}

		// e.g. providerID: gce://example-project-id/us-west2-a/instance-id
		providerURL, err := url.Parse(node.Spec.ProviderID)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("parsing providerID %q: %w", node.Spec.ProviderID, err)
		}
		tokens := strings.Split(strings.Trim(providerURL.Path, "/"), "/")
		if len(tokens) != 2 {
			return ctrl.Result{}, fmt.Errorf("unexpected format for providerID %q", node.Spec.ProviderID)
		}
		project := providerURL.Host
		zone := tokens[0]
		instanceID := tokens[1]
		if project == "" || zone == "" || instanceID == "" {
			return ctrl.Result{}, fmt.Errorf("unexpected format for providerID %q", node.Spec.ProviderID)
		}

		instance, err := r.gceClient.Instances.Get(project, zone, instanceID).Context(ctx).Do()
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("getting instance %s/%s/%s: %w", project, zone, instanceID, err)
		}

		var ipv6Addresses []string
		for _, nic := range instance.NetworkInterfaces {
			for _, ipv6AccessConfig := range nic.Ipv6AccessConfigs {
				if ipv6AccessConfig.ExternalIpv6 != "" {
					ipv6Address := fmt.Sprintf("%s/%d", ipv6AccessConfig.ExternalIpv6, ipv6AccessConfig.ExternalIpv6PrefixLength)
					ipv6Addresses = append(ipv6Addresses, ipv6Address)
				}
			}
		}

		if len(ipv6Addresses) == 0 {
			return ctrl.Result{}, fmt.Errorf("no ipv6 address found on interface %q", instance.NetworkInterfaces[0].Name)
		}
		if len(ipv6Addresses) != 1 {
			return ctrl.Result{}, fmt.Errorf("multiple ipv6 addresses found on interface %q: %v", instance.NetworkInterfaces[0].Name, ipv6Addresses)
		}

		ipv6Address := ipv6Addresses[0]
		podCIDRs := []string{ipv6Address}
		if err := patchNodePodCIDRs(r.coreV1Client, ctx, node, podCIDRs); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *GCEIPAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("gce_ipam").
		For(&corev1.Node{}).
		Complete(r)
}

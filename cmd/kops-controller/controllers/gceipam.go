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
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/go-logr/logr"
	"google.golang.org/api/compute/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
		client:     mgr.GetClient(),
		fieldOwner: "kops-controller",
		log:        ctrl.Log.WithName("controllers").WithName("gce-ipam"),
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

	// fieldOwner is the field-manager owner for fields that we apply
	fieldOwner string

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
		if err := patchNodePodCIDRs(r.coreV1Client, ctx, node, ipv6Address); err != nil {
			return ctrl.Result{}, err
		}
	}

	if len(node.Spec.PodCIDRs) != 0 {
		allIPv6 := true

		for _, podCIDR := range node.Spec.PodCIDRs {
			_, cidr, err := net.ParseCIDR(podCIDR)
			if err != nil {
				klog.Warning("failed to parse podCIDR %q", podCIDR)
				allIPv6 = false
				continue
			}

			// Split into ipv4s and ipv6s, but treat IPv4-mapped IPv6 addresses as IPv6
			if cidr.IP.To4() != nil && !strings.Contains(podCIDR, ":") {
				// ipv4
				allIPv6 = false
			} else {
				// ipv6
			}
		}

		if allIPv6 {
			// IPv6 does not require a route to be set up, so mark the node as NetworkReady
			for _, condition := range node.Status.Conditions {
				if condition.Type == "NetworkUnavailable" {
					if condition.Status == corev1.ConditionTrue {
						newCondition := metav1.Condition{
							Message: "Node has IPv6",
							Status:  metav1.ConditionFalse,
							Reason:  "RouteCreated",
							Type:    "NetworkUnavailable",
						}
						if err := patchStatusCondition(ctx, r.client, node, r.fieldOwner, newCondition); err != nil {
							return ctrl.Result{}, fmt.Errorf("updating NetworkUnavailable condition: %w", err)
						}
					}
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *GCEIPAMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}

type statusConditions struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

type objectStatusPatch struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Metadata   metav1.ObjectMeta `json:"metadata"`
	Status     statusConditions  `json:"status,omitempty"`
}

// patchStatusCondition server-side-applies the node status to set the specified status condition.
func patchStatusCondition(ctx context.Context, kube client.Client, obj client.Object, fieldOwner string, condition metav1.Condition) error {
	apiVersion, kind := obj.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()

	klog.Infof("setting condition %v on %v %q", condition, kind, obj.GetName())
	patch := &objectStatusPatch{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: metav1.ObjectMeta{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		},
		Status: statusConditions{
			Conditions: []metav1.Condition{condition},
		},
	}

	patchJSON, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error building patch: %w", err)
	}

	klog.V(2).Infof("sending patch for %v %q: %q", kind, obj.GetName(), string(patchJSON))

	if err := kube.Status().Patch(ctx, obj, client.RawPatch(types.ApplyPatchType, patchJSON), client.ForceOwnership, client.FieldOwner(fieldOwner)); err != nil {
		return fmt.Errorf("applying patch to %v %v: %w", kind, obj.GetName(), err)
	}

	return nil
}

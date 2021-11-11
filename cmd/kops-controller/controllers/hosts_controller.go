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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/ipaddr"
	"k8s.io/kops/pkg/nodelabels"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// HostsReconciler populates an /etc/hosts style file in the CoreDNS config map,
// supporting in-pod resolution of our k8s.local entries.
// Currently we only populate the apiserver internal record.
type HostsReconciler struct {
	// configMapID identifies the configmap we should update
	configMapID types.NamespacedName

	// roleControlPlaneHostnames lists the DNS hostnames we should populate for the (full) control-plane nodes.
	roleControlPlaneHostnames []string

	// roleAPIServerHostnames lists the DNS hostnames we should populate for apiserver-only nodes.
	roleAPIServerHostnames []string

	// client is the controller-runtime client
	client client.Client

	// log is a logr
	log logr.Logger

	// dynamicClient is a client-go client for patching ConfigMaps
	dynamicClient dynamic.Interface

	// lastUpdate holds the last value we updated, to reduce spurious updates.
	lastUpdate *managedConfigMap

	// addressFamilies holds the list of address families we should populate for internal IPs
	addressFamilies map[ipaddr.Family]bool

	// matchLabels is the set of labels we filter on; only nodes matching one of these labels will be considered
	matchLabels []string
}

// NewHostsReconciler is the constructor for a HostsReconciler
func NewHostsReconciler(mgr manager.Manager, configMapID types.NamespacedName, roleControlPlaneHostnames []string, roleAPIServerHostnames []string, addressFamilies []ipaddr.Family) (*HostsReconciler, error) {
	r := &HostsReconciler{
		client:                    mgr.GetClient(),
		log:                       ctrl.Log.WithName("controllers").WithName("Hosts"),
		configMapID:               configMapID,
		roleControlPlaneHostnames: roleControlPlaneHostnames,
		roleAPIServerHostnames:    roleAPIServerHostnames,
		matchLabels:               []string{nodelabels.RoleLabelAPIServer16, nodelabels.RoleLabelControlPlane20},
	}

	r.addressFamilies = make(map[ipaddr.Family]bool)
	for _, addressFamily := range addressFamilies {
		r.addressFamilies[addressFamily] = true
	}

	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("error building dynamic client: %v", err)
	}
	r.dynamicClient = dynamicClient

	return r, nil
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch

// +kubebuilder:rbac:groups=,resources=configmaps,namespace=kube-system,resourceNames=coredns,verbs=get;patch

// Reconcile is the main reconciler function that observes node changes.
func (r *HostsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	addrToHosts := make(map[string][]string)

	for _, label := range r.matchLabels {
		nodes := &corev1.NodeList{}
		if err := r.client.List(ctx, nodes, client.HasLabels([]string{label})); err != nil {
			klog.Warningf("unable to list nodes with label %q: %v", label, err)
			return ctrl.Result{}, err
		}
		if err := r.buildAddrToHosts(ctx, nodes, addrToHosts); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, r.updateConfigMap(ctx, addrToHosts)
}

func (r *HostsReconciler) buildAddrToHosts(ctx context.Context, nodes *corev1.NodeList, addrToHosts map[string][]string) error {
	for i := range nodes.Items {
		node := &nodes.Items[i]

		role := util.GetNodeRole(node)

		isRoleAPIServer := false
		isRoleControlPlane := false

		switch role {
		case "apiserver":
			isRoleAPIServer = true

		case "control-plane", "master":
			isRoleControlPlane = true

		default:
			klog.Warningf("ignoring node that should have been filtered by label selector: %v", node.Name)
			continue
		}

		for j := range node.Status.Addresses {
			address := &node.Status.Addresses[j]

			if address.Type == corev1.NodeInternalIP {
				if address.Address == "" {
					continue
				}

				family, err := ipaddr.GetFamily(address.Address)
				if err != nil {
					klog.Warningf("cannot get family for address %q: %w", address.Address, err)
					continue
				}

				if !r.addressFamilies[family] {
					continue
				}

				if isRoleAPIServer {
					addrToHosts[address.Address] = append(addrToHosts[address.Address], r.roleAPIServerHostnames...)
				}
				if isRoleControlPlane {
					addrToHosts[address.Address] = append(addrToHosts[address.Address], r.roleControlPlaneHostnames...)
				}
			}
		}
	}

	return nil
}

// managedConfigMap holds the fields we manage
type managedConfigMap struct {
	APIVersion      string `json:"apiVersion"`
	Kind            string `json:"kind"`
	managedMetadata `json:"metadata"`

	Data map[string]string `json:"data"`
}

type managedMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (r *HostsReconciler) updateConfigMap(ctx context.Context, addrToHosts map[string][]string) error {
	var block []string
	for addr, hosts := range addrToHosts {
		hosts = normalizeStringSlice(hosts)
		block = append(block, addr+"\t"+strings.Join(hosts, " "))
	}
	// Sort into a consistent order to minimize updates
	sort.Strings(block)

	hosts := strings.Join(block, "\n")

	data := &managedConfigMap{}
	// These fields are needed, even though they can be determined by looking at the path
	data.APIVersion = "v1"
	data.Kind = "ConfigMap"
	data.Name = r.configMapID.Name
	data.Namespace = r.configMapID.Namespace

	data.Data = map[string]string{"hosts": hosts}
	if r.lastUpdate != nil && reflect.DeepEqual(r.lastUpdate, data) {
		klog.Infof("skipping hosts configmap update (unchanged): %#v", data)
		return nil
	}

	klog.Infof("patching hosts configmap: %#v", data)

	configmapGVR := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmaps"}

	patch, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	// It is strongly recommended for controllers to always "force" conflicts, since they might not be able to resolve or act on these conflicts.
	force := true
	patchOpts := metav1.PatchOptions{
		FieldManager: "kops.k8s.io/hosts",
		Force:        &force,
	}
	if _, err := r.dynamicClient.Resource(configmapGVR).Namespace(r.configMapID.Namespace).Patch(ctx, r.configMapID.Name, types.ApplyPatchType, patch, patchOpts); err != nil {
		return fmt.Errorf("failed to patch configmap: %w", err)
	}

	r.lastUpdate = data

	return nil
}

// normalizeStringSlice returns a de-duplicated and sorted list of strings
func normalizeStringSlice(in []string) []string {
	if len(in) <= 1 {
		return in
	}

	sort.Strings(in)
	// Remove duplicates from sorted slice
	out := make([]string, 0, len(in))
	for i, s := range in {
		if i == 0 || in[i-1] != s {
			out = append(out, s)
		}
	}
	return out
}

func (r *HostsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// We use a predicate to filter, ideally we would push-down the watch to filter by labels,
	// but as we're watching all nodes anyway for the labelling controller, this isn't too bad.

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}, builder.WithPredicates(
			&hostsPredicate{matchLabels: r.matchLabels},
		)).
		Complete(r)
}

// hostsPredicate filters watch events only to nodes of interest.
type hostsPredicate struct {
	matchLabels []string
}

func (p *hostsPredicate) isRelevant(obj *corev1.Node) bool {
	for _, label := range p.matchLabels {
		if _, found := obj.Labels[label]; found {
			return true
		}
	}
	return false
}

func (p *hostsPredicate) Create(ev event.CreateEvent) bool {
	return p.isRelevant(ev.Object.(*corev1.Node))
}

// Delete returns true if the Delete event should be processed
func (p *hostsPredicate) Delete(ev event.DeleteEvent) bool {
	return p.isRelevant(ev.Object.(*corev1.Node))
}

// Update returns true if the Update event should be processed
func (p *hostsPredicate) Update(ev event.UpdateEvent) bool {
	return p.isRelevant(ev.ObjectOld.(*corev1.Node)) || p.isRelevant(ev.ObjectNew.(*corev1.Node))
}

// Generic returns true if the Generic event should be processed
func (p *hostsPredicate) Generic(ev event.GenericEvent) bool {
	return p.isRelevant(ev.Object.(*corev1.Node))
}

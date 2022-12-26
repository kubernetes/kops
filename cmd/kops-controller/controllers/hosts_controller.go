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
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/pkg/apis/kops"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// HostsReconciler populates an /etc/hosts style file in the CoreDNS config map,
// supporting in-pod resolution of our k8s.local entries.
type HostsReconciler struct {
	// clusterName identifies the kOps cluster
	clusterName string

	// configMapID identifies the configmap we should update
	configMapID types.NamespacedName

	// client is the controller-runtime client
	client client.Client

	// log is a logr
	log logr.Logger

	// dynamicClient is a client-go client for patching ConfigMaps
	dynamicClient dynamic.Interface

	// lastUpdate holds the last value we updated, to reduce spurious updates.
	lastUpdate *managedConfigMap
}

// NewHostsReconciler is the constructor for a HostsReconciler
func NewHostsReconciler(mgr manager.Manager, opt *config.Options, configMapID types.NamespacedName) (*HostsReconciler, error) {
	r := &HostsReconciler{
		clusterName: opt.ClusterName,
		client:      mgr.GetClient(),
		log:         ctrl.Log.WithName("controllers").WithName("Hosts"),
		configMapID: configMapID,
	}

	dynamicClient, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return nil, fmt.Errorf("error building dynamic client: %v", err)
	}
	r.dynamicClient = dynamicClient

	return r, nil
}

// +kubebuilder:rbac:groups=,resources=endpoints,verbs=get;list;watch

// +kubebuilder:rbac:groups=,resources=configmaps,namespace=kube-system,resourceNames=coredns,verbs=get;patch

// Reconcile is the main reconciler function that observes node changes.
func (r *HostsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.log.WithValues("endpoints", req.NamespacedName)

	// Although we label the service, the labels get copied to the endpoints by the kube-controller-manager.
	endpointsLabels := client.HasLabels([]string{kops.DiscoveryLabelKey})

	endpointsList := &corev1.EndpointsList{}

	// For security, we only process endpoints in kube-system
	if err := r.client.List(ctx, endpointsList, endpointsLabels, client.InNamespace("kube-system")); err != nil {
		klog.Warningf("unable to list endpoints: %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.updateHosts(ctx, endpointsList)
}

func (r *HostsReconciler) updateHosts(ctx context.Context, endpointsList *corev1.EndpointsList) error {
	addrToHosts := make(map[string][]string)

	for i := range endpointsList.Items {
		endpoints := &endpointsList.Items[i]

		hostname := endpoints.Labels[kops.DiscoveryLabelKey]
		hostname = strings.TrimSuffix(hostname, ".")
		if hostname == "" {
			klog.Warningf("endpoints %s/%s found without discovery label %q; filtering is not working correctly", endpoints.Name, endpoints.Namespace, kops.DiscoveryLabelKey)
			continue
		}
		suffix := ".internal." + r.clusterName
		if !strings.HasSuffix(hostname, suffix) {
			hostname = hostname + suffix
		} else {
			klog.Warningf("endpoints %s/%s found with full internal name for discovery label %q; use short name %q instead", endpoints.Name, endpoints.Namespace, kops.DiscoveryLabelKey, strings.TrimSuffix(hostname, suffix))
		}

		for j := range endpoints.Subsets {
			subset := &endpoints.Subsets[j]

			for k := range subset.Addresses {
				address := &subset.Addresses[k]

				ip := address.IP
				if ip != "" {
					addrToHosts[ip] = append(addrToHosts[ip], hostname)
				}
			}
		}
	}

	return r.updateConfigMap(ctx, addrToHosts)
}

// managedConfigMap holds the fields we manage
type managedConfigMap struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Data       map[string]string `json:"data"`
}

func (r *HostsReconciler) updateConfigMap(ctx context.Context, addrToHosts map[string][]string) error {
	var block []string
	for addr, hosts := range addrToHosts {
		sort.Strings(hosts)
		block = append(block, addr+"\t"+strings.Join(hosts, " "))
	}
	// Sort into a consistent order to minimize updates
	sort.Strings(block)

	hosts := strings.Join(block, "\n")

	data := &managedConfigMap{}
	data.APIVersion = "v1"
	data.Kind = "ConfigMap"
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
		FieldManager: "kops-controller.kops.k8s.io/hosts",
		Force:        &force,
	}
	if _, err := r.dynamicClient.Resource(configmapGVR).Namespace(r.configMapID.Namespace).Patch(ctx, r.configMapID.Name, types.ApplyPatchType, patch, patchOpts); err != nil {
		return fmt.Errorf("failed to patch configmap: %w", err)
	}

	r.lastUpdate = data

	return nil
}

func (r *HostsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Endpoints{}).
		Complete(r)
}

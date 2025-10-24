/*
Copyright 2024 The Kubernetes Authors.

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

package clusterapi

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/pkg/apis/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/yaml"

	capikops "k8s.io/kops/clusterapi/controlplane/kops/api/v1beta1"
)

// NewClusterReconciler is the constructor for an ClusterReconciler
func NewClusterReconciler(mgr manager.Manager) error {
	r := &ClusterReconciler{
		client: mgr.GetClient(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kopsapi.Cluster{}).
		Complete(r)
}

// ClusterReconciler observes Node objects, and labels them with the correct labels for the instancegroup
// This used to be done by the kubelet, but is moving to a central controller for greater security in 1.16
type ClusterReconciler struct {
	// client is the controller-runtime client
	client client.Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cluster := &kopsapi.Cluster{}
	if err := r.client.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	clusterScope := &clusterScope{
		Cluster: cluster,
	}
	if err := clusterScope.Apply(ctx, r.client); err != nil {
		return ctrl.Result{}, fmt.Errorf("error configuring cluster CAPI objects: %w", err)
	}

	return ctrl.Result{}, nil
}

type clusterScope struct {
	Cluster *kopsapi.Cluster

	capiCluster      *unstructured.Unstructured
	capiInfra        *unstructured.Unstructured
	capiControlPlane *unstructured.Unstructured
}

func (s *clusterScope) namespace() string {
	return "kube-system"
}

func (s *clusterScope) ssa(ctx context.Context, kube client.Client, u *unstructured.Unstructured) error {
	return kube.Patch(ctx, u, client.Apply, client.FieldOwner("cluster-controller"))
}

func (s *clusterScope) ssaStatus(ctx context.Context, kube client.Client, u *unstructured.Unstructured) error {
	return kube.Status().Patch(ctx, u, client.Apply, client.FieldOwner("cluster-controller"))
}

func (s *clusterScope) Apply(ctx context.Context, kube client.Client) error {
	if err := s.applyGCPCluster(ctx, kube); err != nil {
		return err
	}
	if err := s.createKopsControlPlane(ctx, kube); err != nil {
		return err
	}
	if err := s.createClusterObject(ctx, kube); err != nil {
		return err
	}
	return nil
}

func (s *clusterScope) applyGCPCluster(ctx context.Context, kube client.Client) error {
	// This is because of network tags in cloud-provider-gcp
	// TODO: cloud-provider-gcp should not assume cluster name is a valid prefix
	name := gce.SafeClusterName(s.Cluster.GetName())

	gcpProject := s.Cluster.Spec.Project
	if gcpProject == "" {
		return fmt.Errorf("unable to determine gcp project for cluster")
	}
	gcpRegion := ""
	for _, subnet := range s.Cluster.Spec.Subnets {
		if gcpRegion == "" {
			gcpRegion = subnet.Region
		} else if gcpRegion != subnet.Region {
			return fmt.Errorf("found multiple gcp regions for cluster")
		}
	}
	if gcpRegion == "" {
		return fmt.Errorf("unable to determine gcp region for cluster")
	}

	// TODO: Sync with LinkToNetwork
	gcpNetworkName := s.Cluster.Spec.NetworkID
	if gcpNetworkName == "" {
		gcpNetworkName = gce.SafeTruncatedClusterName(s.Cluster.ObjectMeta.Name, 63)
	}
	if gcpNetworkName == "" {
		return fmt.Errorf("unable to determine gcp network for cluster")
	}

	obj := map[string]any{
		"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
		"kind":       "GCPCluster",
		"metadata": map[string]any{
			"name":      name,
			"namespace": s.namespace(),
		},
		"spec": map[string]any{
			"project": gcpProject,
			"region":  gcpRegion,
			"network": map[string]any{
				"name": gcpNetworkName,
			},
		},
	}

	u := &unstructured.Unstructured{Object: obj}

	// setOwnerRef(u, s.Cluster)

	if err := s.ssa(ctx, kube, u); err != nil {
		return fmt.Errorf("applying GCPCluster object to cluster: %w", err)
	}

	s.capiInfra = u
	return nil
}

func (s *clusterScope) findSystemEndpoints(ctx context.Context) ([]capikops.SystemEndpoint, error) {
	cluster := s.Cluster

	clusterInternal := &kops.Cluster{}
	if err := kopscodecs.Scheme.Convert(cluster, clusterInternal, nil); err != nil {
		return nil, fmt.Errorf("converting cluster object: %w", err)
	}

	cloud, err := cloudup.BuildCloud(clusterInternal)
	if err != nil {
		return nil, err
	}

	// TODO: Sync with BuildKubecfg

	ingresses, err := cloud.GetApiIngressStatus(clusterInternal)
	if err != nil {
		return nil, fmt.Errorf("error getting ingress status: %v", err)
	}

	var targets []capikops.SystemEndpoint

	for _, ingress := range ingresses {
		var target capikops.SystemEndpoint
		if ingress.Hostname != "" {
			target.Endpoint = ingress.Hostname
		}
		if ingress.IP != "" {
			target.Endpoint = ingress.IP
		}
		target.Type = capikops.SystemEndpointTypeKopsController
		if ingress.InternalEndpoint {
			target.Scope = capikops.SystemEndpointScopeInternal
		} else {
			target.Scope = capikops.SystemEndpointScopeExternal
		}
		targets = append(targets, target)
	}

	for _, ingress := range ingresses {
		var target capikops.SystemEndpoint
		if ingress.Hostname != "" {
			target.Endpoint = ingress.Hostname
		}
		if ingress.IP != "" {
			target.Endpoint = ingress.IP
		}
		target.Type = capikops.SystemEndpointTypeKubeAPIServer
		if ingress.InternalEndpoint {
			target.Scope = capikops.SystemEndpointScopeInternal
		} else {
			target.Scope = capikops.SystemEndpointScopeExternal
		}
		targets = append(targets, target)
	}

	// TODO: Sort targets
	// TODO: Mark targets as atomic list

	if len(targets) == 0 {
		return nil, fmt.Errorf("did not find API endpoint")
	}

	return targets, nil
}

func (s *clusterScope) createKopsControlPlane(ctx context.Context, kube client.Client) error {
	// This is because of network tags in cloud-provider-gcp
	// TODO: cloud-provider-gcp should not assume cluster name is a valid prefix
	name := gce.SafeClusterName(s.Cluster.GetName())

	status := capikops.KopsControlPlaneStatus{}

	systemEndpoints, err := s.findSystemEndpoints(ctx)
	if err != nil {
		return err
	}
	status.SystemEndpoints = systemEndpoints

	status.Initialization = capikops.KopsControlPlaneInitializationStatus{}
	controlPlaneInitialized := true
	status.Initialization.ControlPlaneInitialized = &controlPlaneInitialized

	// Create secret
	{
		kubeconfig := map[string]any{
			"apiVersion": "v1",
			"clusters": []map[string]any{
				{
					"cluster": map[string]any{
						"server":                "https://kubernetes.default.svc.cluster.local",
						"certificate-authority": "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
					},
					"name": "in-cluster",
				},
			},
			"contexts": []map[string]any{
				{
					"context": map[string]any{
						"cluster": "in-cluster",
						"user":    "in-cluster",
					},
					"name": "in-cluster",
				},
			},
			"current-context": "in-cluster",
			"kind":            "Config",
			"preferences":     map[string]any{},
			"users": []map[string]any{
				{
					"name": "in-cluster",
					"user": map[string]any{
						"tokenFile": "/var/run/secrets/kubernetes.io/serviceaccount/token",
					},
				},
			},
		}

		kubeconfigBytes, err := yaml.Marshal(kubeconfig)
		if err != nil {
			return fmt.Errorf("converting kubeconfig to yaml: %w", err)
		}

		obj := map[string]any{
			"apiVersion": "v1",
			"kind":       "Secret",
			"metadata": map[string]any{
				"name":      name + "-kubeconfig",
				"namespace": s.namespace(),
			},
			"data": map[string]any{
				"value": kubeconfigBytes,
			},
			"type": "Opaque",
		}

		u := &unstructured.Unstructured{Object: obj}

		// Needed so that capi manager has "permission" to read the secret
		labels := map[string]string{
			"cluster.x-k8s.io/cluster-name": name,
		}

		u.SetLabels(labels)

		setOwnerRef(u, s.Cluster)
		if err := s.ssa(ctx, kube, u); err != nil {
			return fmt.Errorf("applying kubeconfig secret to cluster: %w", err)
		}
	}

	// TODO: Sync with LinkToNetwork
	obj := map[string]any{
		"apiVersion": "controlplane.cluster.x-k8s.io/v1beta1",
		"kind":       "KopsControlPlane",
		"metadata": map[string]any{
			"name":      name,
			"namespace": s.namespace(),
		},
		"spec": map[string]any{},
	}

	u := &unstructured.Unstructured{Object: obj}

	// setOwnerRef(u, s.Cluster)

	if err := s.ssa(ctx, kube, u); err != nil {
		return fmt.Errorf("applying object to cluster: %w", err)
	}

	// TODO: Sync with LinkToNetwork
	statusObj := map[string]any{
		"apiVersion": "controlplane.cluster.x-k8s.io/v1beta1",
		"kind":       "KopsControlPlane",
		"metadata": map[string]any{
			"name":      name,
			"namespace": s.namespace(),
		},
		"status": status,
	}

	if err := s.ssaStatus(ctx, kube, &unstructured.Unstructured{Object: statusObj}); err != nil {
		return fmt.Errorf("applying object to cluster: %w", err)
	}

	s.capiControlPlane = u
	return nil
}

func (s *clusterScope) createClusterObject(ctx context.Context, kube client.Client) error {
	// This is because of network tags in cloud-provider-gcp
	// TODO: cloud-provider-gcp should not assume cluster name is a valid prefix
	name := gce.SafeClusterName(s.Cluster.GetName())

	obj := map[string]any{
		"apiVersion": "cluster.x-k8s.io/v1beta1",
		"kind":       "Cluster",
		"metadata": map[string]any{
			"name":      name,
			"namespace": s.namespace(),
		},
		"spec": map[string]any{
			"infrastructureRef": makeRef(s.capiInfra),
			"controlPlaneRef":   makeRef(s.capiControlPlane),
		},
	}

	u := &unstructured.Unstructured{Object: obj}

	setOwnerRef(u, s.Cluster)

	if err := s.ssa(ctx, kube, u); err != nil {
		return fmt.Errorf("applying object to cluster: %w", err)
	}

	s.capiCluster = u
	return nil
}

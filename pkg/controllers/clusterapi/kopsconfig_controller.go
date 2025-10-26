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

package clusterapi

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	api "k8s.io/kops/clusterapi/bootstrap/kops/api/v1beta1"
	capikops "k8s.io/kops/clusterapi/controlplane/kops/api/v1beta1"
	clusterv1 "k8s.io/kops/clusterapi/snapshot/cluster-api/api/v1beta1"
	"k8s.io/kops/pkg/apis/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/client/simple"
	"k8s.io/kops/pkg/commands"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi/cloudup"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewKopsConfigReconciler is the constructor for a KopsConfigReconciler
func NewKopsConfigReconciler(mgr manager.Manager, clientset simple.Clientset) error {
	r := &KopsConfigReconciler{
		client:    mgr.GetClient(),
		clientset: clientset,
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&api.KopsConfig{}).
		Complete(r)
}

// KopsConfigReconciler observes KopsConfig objects.
type KopsConfigReconciler struct {
	// client is the controller-runtime client
	client client.Client

	// clientset is a kops clientset
	clientset simple.Clientset
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch

// Reconcile is the main reconciler function that observes node changes.
func (r *KopsConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	obj := &api.KopsConfig{}
	if err := r.client.Get(ctx, req.NamespacedName, obj); err != nil {
		klog.Warningf("unable to fetch object: %v", err)
		if apierrors.IsNotFound(err) {
			// we'll ignore not-found errors, since they can't be fixed by an immediate
			// requeue (we'll need to wait for a new notification), and we can get them
			// on deleted requests.
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	capiCluster, err := getCAPIClusterFromCAPIObject(ctx, r.client, obj)
	if err != nil {
		return ctrl.Result{}, err
	}

	cluster, err := getKopsClusterFromCAPICluster(ctx, r.client, capiCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	kopsControlPlane, err := getKopsControlPlaneFromCAPICluster(ctx, r.client, capiCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	data, err := r.buildBootstrapData(ctx, cluster, kopsControlPlane)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.storeBootstrapData(ctx, obj, data); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.client.Status().Update(ctx, obj); err != nil {
		return ctrl.Result{}, fmt.Errorf("error patching status: %w", err)
	}
	return ctrl.Result{}, nil
}

// storeBootstrapData creates a new secret with the data passed in as input,
// sets the reference in the configuration status and ready to true.
func (r *KopsConfigReconciler) storeBootstrapData(ctx context.Context, parent *api.KopsConfig, data []byte) error {
	// log := ctrl.LoggerFrom(ctx)

	clusterName := parent.Labels[clusterv1.ClusterNameLabel]

	if clusterName == "" {
		return fmt.Errorf("cluster name label %q not yet set", clusterv1.ClusterNameLabel)
	}

	secretName := types.NamespacedName{
		Namespace: parent.GetNamespace(),
		Name:      parent.GetName(),
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName.Name,
			Namespace: secretName.Namespace,
			Labels: map[string]string{
				clusterv1.ClusterNameLabel: clusterName,
			},
		},
		Data: map[string][]byte{
			"value": data,
			// "format": []byte(scope.Config.Spec.Format),
		},
		Type: clusterv1.ClusterSecretType,
	}

	parentAPIVersion, parentKind := parent.GetObjectKind().GroupVersionKind().ToAPIVersionAndKind()
	secret.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion: parentAPIVersion,
			Kind:       parentKind,
			Name:       parent.GetName(),
			UID:        parent.GetUID(),
			Controller: pointer.Bool(true),
		},
	}

	var existing corev1.Secret
	if err := r.client.Get(ctx, secretName, &existing); err != nil {
		if apierrors.IsNotFound(err) {
			if err := r.client.Create(ctx, secret); err != nil {
				return fmt.Errorf("failed to create bootstrap data secret for KopsConfig %s/%s: %w", parent.GetNamespace(), parent.GetName(), err)
			}
		} else {
			return fmt.Errorf("failed to get bootstrap data secret: %w", err)
		}
	} else {
		// TODO: Verify that the existing secret "matches"
		klog.Warningf("TODO: verify that the existing secret matches our expected value")
	}

	parent.Status.DataSecretName = pointer.String(secret.Name)
	parent.Status.Ready = true
	// conditions.MarkTrue(scope.Config, bootstrapv1.DataSecretAvailableCondition)
	return nil
}

func (r *KopsConfigReconciler) buildBootstrapData(ctx context.Context, cluster *kopsapi.Cluster, kopsControlPlane *capikops.KopsControlPlane) ([]byte, error) {
	wellKnownAddresses := model.WellKnownAddresses{}
	for _, systemEndpoint := range kopsControlPlane.Status.SystemEndpoints {
		switch systemEndpoint.Type {
		case capikops.SystemEndpointTypeKopsController:
			wellKnownAddresses[wellknownservices.KopsController] = append(wellKnownAddresses[wellknownservices.KopsController], systemEndpoint.Endpoint)
		case capikops.SystemEndpointTypeKubeAPIServer:
			wellKnownAddresses[wellknownservices.KubeAPIServer] = append(wellKnownAddresses[wellknownservices.KubeAPIServer], systemEndpoint.Endpoint)
		}
	}

	clusterInternal := &kops.Cluster{}

	configBuilder := &commands.ConfigBuilder{}
	configBuilder.Clientset = r.clientset

	{
		if err := kopscodecs.Scheme.Convert(cluster, clusterInternal, nil); err != nil {
			return nil, fmt.Errorf("converting cluster object: %w", err)
		}
		// TODO: Fix validation
		clusterInternal.Namespace = ""

		configBuilder.Cluster = clusterInternal
		configBuilder.ClusterName = clusterInternal.Name
	}

	ig := &kops.InstanceGroup{}
	{
		ig.SetName("placeholder-ig-name") // IG name is not used for nodeup config generation
		ig.Spec.Role = kops.InstanceGroupRoleNode

		configBuilder.InstanceGroup = ig
		configBuilder.InstanceGroupName = ig.Name
	}

	{
		cloud, err := cloudup.BuildCloud(clusterInternal)
		if err != nil {
			return nil, fmt.Errorf("building cloud: %w", err)
		}
		configBuilder.Cloud = cloud
	}

	bootstrapData, err := configBuilder.GetBootstrapData(ctx)
	if err != nil {
		return nil, fmt.Errorf("building bootstrap data: %w", err)
	}

	return bootstrapData.NodeupScript, nil
}

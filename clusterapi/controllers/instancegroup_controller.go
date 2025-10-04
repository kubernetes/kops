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

package controllers

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/kops/pkg/apis/kops"
	kopsapi "k8s.io/kops/pkg/apis/kops/v1alpha2"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcemetadata"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// NewInstanceGroupReconciler is the constructor for an InstanceGroupReconciler
func NewInstanceGroupReconciler(mgr manager.Manager) error {
	r := &InstanceGroupReconciler{
		client: mgr.GetClient(),
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&kopsapi.InstanceGroup{}).
		Complete(r)
}

// InstanceGroupReconciler observes Node objects, and labels them with the correct labels for the instancegroup
// This used to be done by the kubelet, but is moving to a central controller for greater security in 1.16
type InstanceGroupReconciler struct {
	// client is the controller-runtime client
	client client.Client
}

// +kubebuilder:rbac:groups=,resources=nodes,verbs=get;list;watch;patch
// Reconcile is the main reconciler function that observes node changes.
func (r *InstanceGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instanceGroup := &kopsapi.InstanceGroup{}
	if err := r.client.Get(ctx, req.NamespacedName, instanceGroup); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	cluster, err := r.getCluster(ctx, instanceGroup)
	if err != nil {
		return ctrl.Result{}, err
	}

	instanceGroupScope := &instanceGroupScope{
		Cluster:       cluster,
		InstanceGroup: instanceGroup,
	}
	if err := instanceGroupScope.Apply(ctx, r.client); err != nil {
		return ctrl.Result{}, fmt.Errorf("error configuring instance group CAPI objects: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *InstanceGroupReconciler) getCluster(ctx context.Context, ig *kopsapi.InstanceGroup) (*kopsapi.Cluster, error) {
	clusters := &kopsapi.ClusterList{}
	if err := r.client.List(ctx, clusters, client.InNamespace(ig.Namespace)); err != nil {
		return nil, fmt.Errorf("listing clusters in namespace %q: %w", ig.Namespace, err)
	}
	if len(clusters.Items) == 0 {
		return nil, fmt.Errorf("no cluster found in namespace %q", ig.Namespace)
	}
	if len(clusters.Items) > 1 {
		return nil, fmt.Errorf("multiple clusters found in namespace %q", ig.Namespace)
	}
	return &clusters.Items[0], nil
}

func (s *instanceGroupScope) Apply(ctx context.Context, kube client.Client) error {
	if err := s.createGCPMachineTemplate(ctx, kube); err != nil {
		return err
	}
	if err := s.createKopsConfigTemplate(ctx, kube); err != nil {
		return err
	}
	if err := s.createMachineDeployment(ctx, kube); err != nil {
		return err
	}
	return nil
}

func (s *instanceGroupScope) createMachineDeployment(ctx context.Context, kube client.Client) error {
	ig := s.InstanceGroup

	s.capiMachineDeployments = make(map[string]*unstructured.Unstructured)

	minSize := int32(1)
	if s.InstanceGroup.Spec.MinSize != nil {
		minSize = *s.InstanceGroup.Spec.MinSize
	}

	// TODO: Sync up with main replicas logic
	zoneCount := int32(len(ig.Spec.Zones))
	replicas := minSize / zoneCount
	for {
		if replicas*zoneCount >= minSize {
			break
		}
		replicas++
	}

	for _, zone := range ig.Spec.Zones {
		name := fmt.Sprintf("%s-%s", ig.GetName(), zone)

		// This is because of network tags in cloud-provider-gcp
		// TODO: cloud-provider-gcp should not assume cluster name is a valid prefix
		clusterName := gce.SafeClusterName(s.Cluster.GetName())
		kubernetesVersion := s.Cluster.Spec.KubernetesVersion

		configRef := makeRef(s.capiConfigTemplate)
		infrastructureRef := makeRef(s.capiInfra)

		spec := map[string]any{
			"clusterName":   clusterName,
			"version":       kubernetesVersion,
			"failureDomain": zone,
			"bootstrap": map[string]any{
				"configRef": configRef,
			},
			"infrastructureRef": infrastructureRef,
		}

		obj := map[string]any{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "MachineDeployment",
			"metadata": map[string]any{
				"name":      name,
				"namespace": s.namespace(),
			},
			"spec": map[string]any{
				"clusterName": clusterName,
				"replicas":    replicas,
				"template": map[string]any{
					"spec": spec,
				},
			},
		}

		u := &unstructured.Unstructured{Object: obj}

		if err := s.ssa(ctx, kube, u); err != nil {
			return fmt.Errorf("applying object to cluster: %w", err)
		}

		s.capiMachineDeployments[zone] = u
	}
	return nil

}

func (s *instanceGroupScope) namespace() string {
	return "kube-system"
}

func (s *instanceGroupScope) ssa(ctx context.Context, kube client.Client, u *unstructured.Unstructured) error {
	return kube.Patch(ctx, u, client.Apply, client.FieldOwner("kops-instancegroup-controller"))
}

func (s *instanceGroupScope) createKopsConfigTemplate(ctx context.Context, kube client.Client) error {
	name := s.InstanceGroup.GetName()

	spec := map[string]any{}

	obj := map[string]any{
		"apiVersion": "bootstrap.cluster.x-k8s.io/v1beta1",
		"kind":       "KopsConfigTemplate",
		"metadata": map[string]any{
			"name":      name,
			"namespace": s.namespace(),
		},
		"spec": map[string]any{
			"template": map[string]any{
				"spec": spec,
			},
		},
	}

	u := &unstructured.Unstructured{Object: obj}

	if err := s.ssa(ctx, kube, u); err != nil {
		return fmt.Errorf("applying object to cluster: %w", err)
	}

	s.capiConfigTemplate = u
	return nil
}

type instanceGroupScope struct {
	Cluster       *kopsapi.Cluster
	InstanceGroup *kopsapi.InstanceGroup

	// capiMachineDeployments holds the MachineDeployment objects,
	// we create one per zone
	capiMachineDeployments map[string]*unstructured.Unstructured

	capiInfra          *unstructured.Unstructured
	capiConfigTemplate *unstructured.Unstructured
}

func (s *instanceGroupScope) createGCPMachineTemplate(ctx context.Context, kube client.Client) error {
	ig := s.InstanceGroup

	name := ig.GetName()

	// name := s.name()

	// clusterLabel := gce.LabelForCluster(s.Cluster.GetName())
	// roleLabel := gce.GceLabelNameRolePrefix + ig.Spec.Role.ToLowerString()
	// labels := map[string]string{
	// 	clusterLabel.Key:              clusterLabel.Value,
	// 	roleLabel:                     "",
	// 	gce.GceLabelNameInstanceGroup: s.InstanceGroup.GetName(),
	// }

	metadata := map[string]string{
		gcemetadata.MetadataKeyClusterName:           s.Cluster.GetName(),
		nodeidentitygce.MetadataKeyInstanceGroupName: ig.GetName(),
	}

	additionalNetworkTags := []string{
		gce.TagForRole(s.Cluster.GetName(), kops.InstanceGroupRole(ig.Spec.Role)),
	}

	// TODO: Sync with LinkToSubnet and support ID
	subnet := ""
	for _, s := range ig.Spec.Subnets {
		if subnet == "" {
			subnet = s
		} else if subnet != s {
			return fmt.Errorf("found multiple subnets for instance group")
		}
	}
	if subnet == "" {
		return fmt.Errorf("cannot determine subnet for instance group")
	}
	subnet = gce.ClusterSuffixedName(subnet, s.Cluster.GetName(), 63)

	imageSpec := ig.Spec.Image
	tokens := strings.Split(imageSpec, "/")
	// TODO: Sync with GCE logic
	if len(tokens) == 2 {
		imageSpec = fmt.Sprintf("projects/%s/global/images/%s", tokens[0], tokens[1])
	}

	spec := map[string]any{}
	spec["instanceType"] = ig.Spec.MachineType
	spec["image"] = imageSpec
	spec["subnet"] = subnet
	spec["additionalNetworkTags"] = additionalNetworkTags
	spec["publicIP"] = true
	additionalMetadata := []map[string]string{}
	for k, v := range metadata {
		additionalMetadata = append(additionalMetadata, map[string]string{
			"key":   k,
			"value": v,
		})
	}
	// 	additionalMetadata = append(additionalMetadata, map[string]string{
	// 	"key":   "k8s-io-instance-group-name",
	// 	"value": ig.Name,
	// })
	// additionalMetadata = append(additionalMetadata, map[string]string{
	// 	"key":   "cluster-name",
	// 	"value": clusterName,
	// })
	spec["additionalMetadata"] = additionalMetadata

	obj := map[string]any{
		"apiVersion": "infrastructure.cluster.x-k8s.io/v1beta1",
		"kind":       "GCPMachineTemplate",
		"metadata": map[string]any{
			"name":      name,
			"namespace": s.namespace(),
		},
		"spec": map[string]any{
			"template": map[string]any{
				"spec": spec,
			},
		},
	}

	u := &unstructured.Unstructured{Object: obj}

	if err := s.ssa(ctx, kube, u); err != nil {
		return fmt.Errorf("applying object to cluster: %w", err)
	}

	s.capiInfra = u
	return nil
}

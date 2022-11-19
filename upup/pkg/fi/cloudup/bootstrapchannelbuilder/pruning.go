/*
Copyright 2019 The Kubernetes Authors.

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

package bootstrapchannelbuilder

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/kubemanifest"
)

func (b *BootstrapChannelBuilder) addPruneDirectives(addons *AddonList) error {
	for _, addon := range addons.Items {
		if !addon.BuildPrune {
			continue
		}

		id := *addon.Spec.Name

		if err := b.addPruneDirectivesForAddon(addon); err != nil {
			return fmt.Errorf("failed to configure pruning for %s: %w", id, err)
		}
	}
	return nil
}

func (b *BootstrapChannelBuilder) addPruneDirectivesForAddon(addon *Addon) error {
	addon.Spec.Prune = &channelsapi.PruneSpec{}

	// We add these labels to all objects we manage, so we reuse them for pruning.
	selectorMap := map[string]string{
		"app.kubernetes.io/managed-by": "kops",
		"addon.kops.k8s.io/name":       *addon.Spec.Name,
	}
	selector, err := labels.ValidatedSelectorFromSet(selectorMap)
	if err != nil {
		return fmt.Errorf("error parsing selector %v: %w", selectorMap, err)
	}

	// We always include a set of well-known group kinds,
	// so that we prune even if we end up removing something from the manifest.
	alwaysPruneGroupKinds := []schema.GroupKind{
		{Group: "", Kind: "ConfigMap"},
		{Group: "", Kind: "Service"},
		{Group: "", Kind: "ServiceAccount"},
		{Group: "apps", Kind: "Deployment"},
		{Group: "apps", Kind: "DaemonSet"},
		{Group: "apps", Kind: "StatefulSet"},
		{Group: "rbac.authorization.k8s.io", Kind: "ClusterRole"},
		{Group: "rbac.authorization.k8s.io", Kind: "ClusterRoleBinding"},
		{Group: "rbac.authorization.k8s.io", Kind: "Role"},
		{Group: "rbac.authorization.k8s.io", Kind: "RoleBinding"},
		{Group: "policy", Kind: "PodDisruptionBudget"},
		{Group: "admissionregistration.k8s.io/v1", Kind: "ValidatingWebhookConfiguration"},
		{Group: "admissionregistration.k8s.io/v1", Kind: "MutatingWebhookConfiguration"},
	}
	pruneGroupKind := make(map[schema.GroupKind]bool)
	for _, gk := range alwaysPruneGroupKinds {
		pruneGroupKind[gk] = true
	}

	// In addition, we deliberately exclude a few types that are riskier to delete:
	//
	//  * Namespace: because it deletes anything else that happens to be in the namespace
	//
	//  * CustomResourceDefinition: because it deletes all instances of the CRD
	neverPruneGroupKinds := map[schema.GroupKind]bool{
		{Group: "", Kind: "Namespace"}:                                    true,
		{Group: "apiextensions.k8s.io", Kind: "CustomResourceDefinition"}: true,
	}

	// Parse the manifest; we use this to scope pruning to namespaces
	objects, err := kubemanifest.LoadObjectsFrom(addon.ManifestData)
	if err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}
	objectsByGK := make(map[schema.GroupKind][]*kubemanifest.Object)
	for _, object := range objects {
		gv, err := schema.ParseGroupVersion(object.APIVersion())
		if err != nil || gv.Version == "" {
			return fmt.Errorf("failed to parse apiVersion %q", object.APIVersion())
		}
		gvk := gv.WithKind(object.Kind())
		if gvk.Kind == "" {
			return fmt.Errorf("failed to get kind for object")
		}

		gk := gvk.GroupKind()
		objectsByGK[gk] = append(objectsByGK[gk], object)

		// Warn if there are objects in the manifest that we haven't considered
		if !pruneGroupKind[gk] {
			if !neverPruneGroupKinds[gk] {
				klog.Warningf("manifest includes an object of GroupKind %v, which will not be pruned", gk)
			}
		}
	}

	var groupKinds []schema.GroupKind
	for gk := range pruneGroupKind {
		groupKinds = append(groupKinds, gk)
	}

	sort.Slice(groupKinds, func(i, j int) bool {
		if groupKinds[i].Group != groupKinds[j].Group {
			return groupKinds[i].Group < groupKinds[j].Group
		}
		return groupKinds[i].Kind < groupKinds[j].Kind
	})

	for _, gk := range groupKinds {
		pruneSpec := channelsapi.PruneKindSpec{}
		pruneSpec.Group = gk.Group
		pruneSpec.Kind = gk.Kind

		namespaces := sets.NewString()
		for _, object := range objectsByGK[gk] {
			namespace := object.GetNamespace()
			if namespace != "" {
				namespaces.Insert(namespace)
			}
		}
		if namespaces.Len() != 0 {
			pruneSpec.Namespaces = namespaces.List()
		}

		pruneSpec.LabelSelector = selector.String()

		addon.Spec.Prune.Kinds = append(addon.Spec.Prune.Kinds, pruneSpec)
	}

	return nil
}

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

package channels

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	"k8s.io/kops/channels/pkg/api"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Addon is a wrapper around a single version of an addon
type Addon struct {
	Name            string
	ChannelName     string
	ChannelLocation url.URL
	Spec            *api.AddonSpec
}

// AddonUpdate holds data about a proposed update to an addon
type AddonUpdate struct {
	Name            string
	ExistingVersion *ChannelVersion
	NewVersion      *ChannelVersion
}

// AddonMenu is a collection of addons, with helpers for computing the latest versions
type AddonMenu struct {
	Addons map[string]*Addon
}

func NewAddonMenu() *AddonMenu {
	return &AddonMenu{
		Addons: make(map[string]*Addon),
	}
}

func (m *AddonMenu) MergeAddons(o *AddonMenu) {
	for k, v := range o.Addons {
		existing := m.Addons[k]
		if existing == nil {
			m.Addons[k] = v
		} else {
			if v.ChannelVersion().replaces(existing.ChannelVersion()) {
				m.Addons[k] = v
			}
		}
	}
}

func (a *Addon) ChannelVersion() *ChannelVersion {
	return &ChannelVersion{
		Channel:      &a.ChannelName,
		Version:      a.Spec.Version,
		Id:           a.Spec.Id,
		ManifestHash: a.Spec.ManifestHash,
	}
}

func (a *Addon) buildChannel() *Channel {
	namespace := "kube-system"
	if a.Spec.Namespace != nil {
		namespace = *a.Spec.Namespace
	}

	channel := &Channel{
		Namespace: namespace,
		Name:      a.Name,
	}
	return channel
}

func (a *Addon) GetRequiredUpdates(ctx context.Context, k8sClient kubernetes.Interface) (*AddonUpdate, error) {
	newVersion := a.ChannelVersion()

	channel := a.buildChannel()

	existingVersion, err := channel.GetInstalledVersion(ctx, k8sClient)
	if err != nil {
		return nil, err
	}

	if existingVersion != nil && !newVersion.replaces(existingVersion) {
		return nil, nil
	}

	return &AddonUpdate{
		Name:            a.Name,
		ExistingVersion: existingVersion,
		NewVersion:      newVersion,
	}, nil
}

func (a *Addon) GetManifestFullUrl() (*url.URL, error) {
	if a.Spec.Manifest == nil || *a.Spec.Manifest == "" {
		return nil, field.Required(field.NewPath("spec", "manifest"), "")
	}

	manifest := *a.Spec.Manifest
	manifestURL, err := url.Parse(manifest)
	if err != nil {
		return nil, field.Invalid(field.NewPath("spec", "manifest"), manifest, "Not a valid URL")
	}
	if !manifestURL.IsAbs() {
		manifestURL = a.ChannelLocation.ResolveReference(manifestURL)
	}
	return manifestURL, nil
}

func (a *Addon) EnsureUpdated(ctx context.Context, k8sClient kubernetes.Interface) (*AddonUpdate, error) {
	required, err := a.GetRequiredUpdates(ctx, k8sClient)
	if err != nil {
		return nil, err
	}
	if required == nil {
		return nil, nil
	}
	manifestURL, err := a.GetManifestFullUrl()
	if err != nil {
		return nil, err
	}
	klog.Infof("Applying update from %q", manifestURL)

	err = Apply(manifestURL.String())
	if err != nil {
		return nil, fmt.Errorf("error applying update from %q: %v", manifestURL, err)
	}

	if a.Spec.NeedsRollingUpdate != "" {
		err = a.AddNeedsUpdateLabel(ctx, k8sClient)
		if err != nil {
			return nil, fmt.Errorf("error adding needs-update label: %v", err)
		}
	}

	channel := a.buildChannel()
	err = channel.SetInstalledVersion(ctx, k8sClient, a.ChannelVersion())
	if err != nil {
		return nil, fmt.Errorf("error applying annotation to record addon installation: %v", err)
	}

	return required, nil
}

func (a *Addon) AddNeedsUpdateLabel(ctx context.Context, k8sClient kubernetes.Interface) error {
	klog.Infof("addon %v wants to update %v nodes", a.Name, a.Spec.NeedsRollingUpdate)
	selector := ""
	switch a.Spec.NeedsRollingUpdate {
	case "control-plane":
		selector = "node-role.kubernetes.io/master="
	case "worker":
		selector = "node-role.kubernetes.io/node="
	}

	annotationPatch := &annotationPatch{Metadata: annotationPatchMetadata{Annotations: map[string]string{
		"kops.k8s.io/needs-update": "",
	}}}
	annotationPatchJSON, err := json.Marshal(annotationPatch)
	if err != nil {
		return err
	}

	nodeInterface := k8sClient.CoreV1().Nodes()
	nodes, err := nodeInterface.List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return err
	}
	for _, node := range nodes.Items {
		_, err = nodeInterface.Patch(ctx, node.Name, types.StrategicMergePatchType, annotationPatchJSON, metav1.PatchOptions{})

		if err != nil {
			return err
		}
	}
	return nil
}

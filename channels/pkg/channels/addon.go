/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kops/channels/pkg/api"
	"net/url"
)

type Addon struct {
	Name            string
	ChannelName     string
	ChannelLocation url.URL
	Spec            *api.AddonSpec
}

type AddonUpdate struct {
	Name            string
	ExistingVersion *ChannelVersion
	NewVersion      *ChannelVersion
}

func (a *Addon) ChannelVersion() *ChannelVersion {
	return &ChannelVersion{
		Channel: &a.ChannelName,
		Version: a.Spec.Version,
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
func (a *Addon) GetRequiredUpdates(k8sClient kubernetes.Interface) (*AddonUpdate, error) {
	newVersion := a.ChannelVersion()

	channel := a.buildChannel()

	existingVersion, err := channel.GetInstalledVersion(k8sClient)
	if err != nil {
		return nil, err
	}

	if existingVersion != nil && !newVersion.Replaces(existingVersion) {
		return nil, nil
	}

	return &AddonUpdate{
		Name:            a.Name,
		ExistingVersion: existingVersion,
		NewVersion:      newVersion,
	}, nil
}

func (a *Addon) EnsureUpdated(k8sClient kubernetes.Interface) (*AddonUpdate, error) {
	required, err := a.GetRequiredUpdates(k8sClient)
	if err != nil {
		return nil, err
	}
	if required == nil {
		return nil, nil
	}

	if a.Spec.Manifest == nil || *a.Spec.Manifest == "" {
		return nil, field.Required(field.NewPath("Spec", "Manifest"), "")
	}

	manifest := *a.Spec.Manifest
	manifestURL, err := url.Parse(manifest)
	if err != nil {
		return nil, field.Invalid(field.NewPath("Spec", "Manifest"), manifest, "Not a valid URL")
	}
	if !manifestURL.IsAbs() {
		manifestURL = a.ChannelLocation.ResolveReference(manifestURL)
	}
	glog.Infof("Applying update from %q", manifestURL)

	err = Apply(manifestURL.String())
	if err != nil {
		return nil, fmt.Errorf("error applying update from %q: %v", manifest, err)
	}

	channel := a.buildChannel()
	err = channel.SetInstalledVersion(k8sClient, a.ChannelVersion())
	if err != nil {
		return nil, fmt.Errorf("error applying annotation to to record addon installation: %v", err)
	}

	return required, nil
}

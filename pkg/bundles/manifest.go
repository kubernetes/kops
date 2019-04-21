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

package bundles

import (
	"fmt"
	"net/url"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

func resolveChannel(channel string) (string, error) {
	u, err := url.Parse(channel)
	if err != nil {
		return "", fmt.Errorf("invalid channel: %q", channel)
	}

	if !u.IsAbs() {
		base, err := url.Parse(kops.DefaultChannelBase)
		if err != nil {
			return "", fmt.Errorf("invalid base channel location: %q", kops.DefaultChannelBase)
		}
		klog.V(4).Infof("resolving %q against default channel location %q", channel, kops.DefaultChannelBase)
		u = base.ResolveReference(u)
	}

	return u.String(), nil
}

// loadBundleManifest loads a file object from the specified VFS location
func loadBundleManifest(clusterSpec *kops.ClusterSpec, bundleLocation string) ([]byte, string, error) {
	u, err := url.Parse(bundleLocation)
	if err != nil {
		return nil, "", fmt.Errorf("invalid bundle: %q", bundleLocation)
	}

	bundleBase := bundleLocation
	if !u.IsAbs() {
		channel, err := resolveChannel(clusterSpec.Channel)
		if err != nil {
			return nil, "", err
		}

		base, err := url.Parse(channel)
		if err != nil {
			return nil, "", fmt.Errorf("invalid channel location: %q", channel)
		}
		bundleBase = channel

		tokens := strings.Split(bundleLocation, ".")
		s := "bundles/" + tokens[0] + "/" + bundleLocation
		klog.V(4).Infof("resolving %q against channel location %q", s, channel)
		q, err := url.Parse(s)
		if err != nil {
			return nil, "", fmt.Errorf("invalid channel: %q", s)
		}
		u = base.ResolveReference(q)
	}

	resolved := u.String()
	klog.V(2).Infof("loading bundle from %q", resolved)
	contents, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, "", fmt.Errorf("error reading bundle %q: %v", resolved, err)
	}

	return contents, bundleBase, nil
}

// loadComponent loads a file object from the specified VFS location
func loadComponent(clusterSpec *kops.ClusterSpec, baseLocation string, componentRef *ComponentReference) ([]byte, error) {
	base, err := url.Parse(baseLocation)
	if err != nil {
		return nil, fmt.Errorf("invalid base location: %q", baseLocation)
	}

	tokens := strings.Split(componentRef.Version, ".")
	s := "components/" + componentRef.ComponentName + "/" + tokens[0] + "/" + componentRef.Version
	q, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("invalid channel: %q", s)
	}
	u := base.ResolveReference(q)

	resolved := u.String()
	klog.V(2).Infof("loading bundle from %q", resolved)
	contents, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading bundle %q: %v", resolved, err)
	}

	return contents, nil
}

var Scheme = runtime.NewScheme()
var Codecs = serializer.NewCodecFactory(Scheme)

func init() {
	addKnownTypes(Scheme)
}

// LoadComponent loads a file object from the specified VFS location
func LoadComponent(clusterSpec *kops.ClusterSpec, bundleLocation string, componentName string) (*Component, error) {
	var componentRef *ComponentReference
	var bundleBase string
	{
		manifest, base, err := loadBundleManifest(clusterSpec, bundleLocation)
		if err != nil {
			return nil, err
		}
		bundleBase = base

		codec := Codecs.UniversalDeserializer() //&serializer.DirectCodecFactory{CodecFactory: Codecs}
		componentSets, err := ParseBytes(manifest, codec)
		if err != nil {
			return nil, err
		}

		if len(componentSets) == 0 {
			return nil, fmt.Errorf("no ComponentSets found in %s", bundleLocation)
		}

		var componentRefs []*ComponentReference

		for _, componentSetObj := range componentSets {
			componentSet, ok := componentSetObj.(*ComponentSet)
			if !ok {
				return nil, fmt.Errorf("unexpected component type %T found, expected Component", componentSetObj)
			}

			for i := range componentSet.Spec.Components {
				component := &componentSet.Spec.Components[i]
				if component.ComponentName == componentName {
					componentRefs = append(componentRefs, component)
				}
			}
		}

		if len(componentRefs) == 0 {
			return nil, fmt.Errorf("no component with name %q found in ComponentSet %q", componentName, bundleLocation)
		}

		if len(componentRefs) > 1 {
			return nil, fmt.Errorf("multiple components with name %q found in ComponentSet %q", componentName, bundleLocation)
		}

		componentRef = componentRefs[0]
	}

	{
		manifest, err := loadComponent(clusterSpec, bundleBase, componentRef)
		if err != nil {
			return nil, err
		}

		codec := Codecs.UniversalDeserializer() //&serializer.DirectCodecFactory{CodecFactory: Codecs}
		components, err := ParseBytes(manifest, codec)
		if err != nil {
			return nil, err
		}

		if len(components) == 0 {
			return nil, fmt.Errorf("no Component found in %s", componentRef.Version)
		}

		if len(components) > 1 {
			return nil, fmt.Errorf("multiple Component found in %s", componentRef.Version)
		}

		component, ok := components[0].(*Component)
		if !ok {
			return nil, fmt.Errorf("unexpected component type %T found, expected Component", components[0])
		}

		if component.Spec.ComponentName != componentName {
			return nil, fmt.Errorf("component %q did not have expected name: %q", component.Spec.ComponentName, componentName)
		}

		return component, nil
	}
}

/*
Copyright 2020 The Kubernetes Authors.

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

package clusteraddons

import (
	"fmt"
	"net/url"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/util/pkg/vfs"
)

type ClusterAddon struct {
	Raw     string
	Objects kubemanifest.ObjectList
}

// LoadClusterAddon loads a set of objects from the specified VFS location
func LoadClusterAddon(location string) (*ClusterAddon, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("invalid addon location: %q", location)
	}

	// TODO: Should we support relative paths for "standard" addons?  See equivalent code in LoadChannel

	resolved := u.String()
	klog.V(2).Infof("Loading addon from %q", resolved)
	addonBytes, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading addon %q: %v", resolved, err)
	}
	addon, err := ParseClusterAddon(addonBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing addon %q: %v", resolved, err)
	}
	klog.V(4).Infof("Addon contents: %s", string(addonBytes))

	return addon, nil
}

// ParseClusterAddon parses a ClusterAddon object
func ParseClusterAddon(raw []byte) (*ClusterAddon, error) {
	objects, err := kubemanifest.LoadObjectsFrom(raw)
	if err != nil {
		return nil, fmt.Errorf("error parsing addon %v", err)
	}

	return &ClusterAddon{Raw: string(raw), Objects: objects}, nil
}

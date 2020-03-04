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
	"fmt"
	"net/url"
	"strings"

	"github.com/blang/semver"
	"k8s.io/klog"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

type Addons struct {
	ChannelName     string
	ChannelLocation url.URL
	APIObject       *api.Addons
}

func LoadAddons(name string, location *url.URL) (*Addons, error) {
	klog.V(2).Infof("Loading addons channel from %q", location)
	data, err := vfs.Context.ReadFile(location.String())
	if err != nil {
		return nil, fmt.Errorf("error reading addons from %q: %v", location, err)
	}

	return ParseAddons(name, location, data)
}

func ParseAddons(name string, location *url.URL, data []byte) (*Addons, error) {
	// Yaml can't parse empty strings
	configString := string(data)
	configString = strings.TrimSpace(configString)

	apiObject := &api.Addons{}
	if configString != "" {
		err := utils.YamlUnmarshal([]byte(configString), apiObject)
		if err != nil {
			return nil, fmt.Errorf("error parsing addons: %v", err)
		}
	}

	for _, addon := range apiObject.Spec.Addons {
		if addon != nil && addon.Version != nil && *addon.Version != "" {
			name := apiObject.ObjectMeta.Name
			if addon.Name != nil {
				name = *addon.Name
			}

			_, err := semver.ParseTolerant(*addon.Version)
			if err != nil {
				return nil, fmt.Errorf("addon %q has unparseable version %q: %v", name, *addon.Version, err)
			}
		}
	}

	return &Addons{ChannelName: name, ChannelLocation: *location, APIObject: apiObject}, nil
}

func (a *Addons) GetCurrent(kubernetesVersion semver.Version) (*AddonMenu, error) {
	all, err := a.wrapInAddons()
	if err != nil {
		return nil, err
	}

	menu := NewAddonMenu()
	for _, addon := range all {
		if !addon.matches(kubernetesVersion) {
			continue
		}
		name := addon.Name

		existing := menu.Addons[name]
		if existing == nil {
			menu.Addons[name] = addon
		} else if update, _ := addon.ChannelVersion().updates(existing.ChannelVersion()); update {
			menu.Addons[name] = addon
		}
	}

	return menu, nil
}

func (a *Addons) wrapInAddons() ([]*Addon, error) {
	var addons []*Addon
	for _, s := range a.APIObject.Spec.Addons {
		name := a.APIObject.ObjectMeta.Name
		if s.Name != nil {
			name = *s.Name
		}

		addon := &Addon{
			ChannelName:     a.ChannelName,
			ChannelLocation: a.ChannelLocation,
			Spec:            s,
			Name:            name,
		}

		addons = append(addons, addon)
	}
	return addons, nil
}

func (s *Addon) matches(kubernetesVersion semver.Version) bool {
	if s.Spec.KubernetesVersion != "" {
		versionRange, err := semver.ParseRange(s.Spec.KubernetesVersion)
		if err != nil {
			klog.Warningf("unable to parse KubernetesVersion %q; skipping", s.Spec.KubernetesVersion)
			return false
		}
		if !versionRange(kubernetesVersion) {
			klog.V(4).Infof("Skipping version range %q that does not match current version %s", s.Spec.KubernetesVersion, kubernetesVersion)
			return false
		}
	}

	return true
}

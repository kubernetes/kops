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
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
	"net/url"
	"strings"
)

type Addons struct {
	ChannelName     string
	ChannelLocation url.URL
	APIObject       *api.Addons
}

func LoadAddons(name string, location *url.URL) (*Addons, error) {
	glog.V(2).Infof("Loading addons channel from %q", location)
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

	return &Addons{ChannelName: name, ChannelLocation: *location, APIObject: apiObject}, nil
}

func (a *Addons) GetCurrent() ([]*Addon, error) {
	all, err := a.All()
	if err != nil {
		return nil, err
	}
	specs := make(map[string]*Addon)
	for _, addon := range all {
		name := addon.Name
		existing := specs[name]
		if existing == nil || addon.ChannelVersion().Replaces(existing.ChannelVersion()) {
			specs[name] = addon
		}
	}

	var addons []*Addon
	for _, addon := range specs {
		addons = append(addons, addon)
	}
	return addons, nil
}

func (a *Addons) All() ([]*Addon, error) {
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

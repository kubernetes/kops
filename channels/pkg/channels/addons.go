package channels

import (
	"fmt"
	"k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
	"strings"
)

type Addons struct {
	Channel   string
	APIObject *api.Addons
}

func LoadAddons(location string) (*Addons, error) {
	data, err := vfs.Context.ReadFile(location)
	if err != nil {
		return nil, fmt.Errorf("error reading addons from %q: %v", location, err)
	}

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

	return &Addons{Channel: location, APIObject: apiObject}, nil
}

func (a *Addons) GetCurrent() ([]*Addon, error) {
	specs := make(map[string]*Addon)
	for _, s := range a.APIObject.Spec.Addons {
		name := a.APIObject.Name
		if s.Name != nil {
			name = *s.Name
		}

		addon := &Addon{Channel: a.Channel, Spec: s, Name: name}
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

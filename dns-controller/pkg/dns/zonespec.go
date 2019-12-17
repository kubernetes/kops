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

package dns

import (
	"fmt"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

type ZoneSpec struct {
	Name string
	ID   string
}

func ParseZoneSpec(s string) (*ZoneSpec, error) {
	s = strings.TrimSpace(s)

	tokens := strings.SplitN(s, "/", 2)
	if len(tokens) == 2 && tokens[0] == "*" {
		// */1234: Match by ID
		return &ZoneSpec{ID: tokens[1]}, nil
	}
	name := EnsureDotSuffix(tokens[0])
	if len(tokens) == 1 {
		// example.com: Match by name
		return &ZoneSpec{Name: name}, nil
	}

	// example.com/1234: Match by name & id
	return &ZoneSpec{Name: name, ID: tokens[1]}, nil
}

type ZoneRules struct {
	// We don't use a map so we can support e.g. *.example.com later
	Zones    []*ZoneSpec
	Wildcard bool
}

func ParseZoneRules(zones []string) (*ZoneRules, error) {
	r := &ZoneRules{}

	for _, s := range zones {
		s = strings.TrimSpace(s)
		if s == "*" || s == "*/*" {
			r.Wildcard = true
			continue
		}

		zoneSpec, err := ParseZoneSpec(s)
		if err != nil {
			return nil, fmt.Errorf("error parsing %q: %v", s, err)
		}

		r.Zones = append(r.Zones, zoneSpec)
	}

	if len(zones) == 0 {
		klog.Infof("No rules specified, will permit management of all zones")
		r.Wildcard = true
	}

	return r, nil
}

// MatchesExplicitly returns true if this matches an explicit rule (not a wildcard)
func (r *ZoneRules) MatchesExplicitly(zone dnsprovider.Zone) bool {
	name := EnsureDotSuffix(zone.Name())
	id := zone.ID()

	for _, zoneSpec := range r.Zones {
		if zoneSpec.Name != "" && zoneSpec.Name != name {
			continue
		}

		if zoneSpec.ID != "" && zoneSpec.ID != id {
			return false
		}

		return true
	}

	return false
}

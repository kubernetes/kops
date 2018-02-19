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

package dns

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/golang/glog"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
)

type ZoneSpec struct {
	Name    string
	ID      string
	Pattern glob.Glob
}

type RecordWhiteList struct {
	Name    string
	Pattern glob.Glob
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
		return &ZoneSpec{Name: name, Pattern: glob.MustCompile(name, '.')}, nil
	}

	// example.com/1234: Match by name & id
	return &ZoneSpec{Name: name, Pattern: glob.MustCompile(name, '.'), ID: tokens[1]}, nil
}

func ParseRecordWhiteList(s string) (*RecordWhiteList, error) {
	s = strings.TrimSpace(s)
	g, err := glob.Compile(s, '.')
	if err != nil {
		return nil, fmt.Errorf("error parsing %q: %v", s, err)
	}
	return &RecordWhiteList{Name: s, Pattern: g}, nil
}

type ZoneRules struct {
	// We don't use a map so we can support e.g. *.example.com later
	Zones      []*ZoneSpec
	WhiteLists []*RecordWhiteList
	Catchall   bool
}

func ParseZoneRules(zones []string, whitelists []string) (*ZoneRules, error) {
	r := &ZoneRules{}

	for _, s := range zones {
		s = strings.TrimSpace(s)
		if s == "*" || s == "*/*" {
			r.Catchall = true
			continue
		}

		zoneSpec, err := ParseZoneSpec(s)
		if err != nil {
			return nil, fmt.Errorf("error parsing %q: %v", s, err)
		}

		r.Zones = append(r.Zones, zoneSpec)
	}

	if len(zones) == 0 {
		glog.Infof("No rules specified, will permit management of all zones")
		r.Catchall = true
	}

	for _, s := range whitelists {
		whiteList, err := ParseRecordWhiteList(s)
		if err != nil {
			return nil, fmt.Errorf("error parsing %q: %v", s, err)
		}

		glog.V(2).Infof("Adding whitelist %q", s)
		r.WhiteLists = append(r.WhiteLists, whiteList)
	}

	if len(whitelists) == 0 {
		glog.V(2).Infof("No whitelist rules specified, will permit management of any record")
	}

	return r, nil
}

// MatchesExplicitly returns true if this matches an explicit rule (not a catchall)
func (r *ZoneRules) MatchesExplicitly(zone dnsprovider.Zone) bool {
	name := EnsureDotSuffix(zone.Name())
	id := zone.ID()

	for _, zoneSpec := range r.Zones {
		if zoneSpec.Name != "" && zoneSpec.Pattern.Match(name) {
			continue
		}

		if zoneSpec.ID != "" && zoneSpec.ID != id {
			return false
		}

		return true
	}

	return false
}

// MatchesExplicitly returns true if this matches an explicit rule (not a catchall)
func (r *ZoneRules) PassesWhiteLists(record string) bool {
	if len(r.WhiteLists) == 0 {
		return true
	}

	name := EnsureDotSuffix(record)
	glog.V(4).Infof("checking %s against the white lists", name)

	for _, whiteList := range r.WhiteLists {

		if whiteList.Pattern.Match(name) {
			glog.V(4).Infof("%s matched white list pattern %s", name, whiteList.Name)
			return true
		} else {
			glog.V(4).Infof("%s did not match white list pattern %s", name, whiteList.Name)
		}
	}

	glog.V(4).Infof("%s did not match any white list patterns", name)
	return false
}

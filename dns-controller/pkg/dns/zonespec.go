package dns

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"strings"
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
		glog.Infof("No rules specified, will permit management of all zones")
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

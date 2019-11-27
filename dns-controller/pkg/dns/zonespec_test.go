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
	"reflect"
	"testing"
)

func TestEnsureDotSuffix(t *testing.T) {
	cases := []struct {
		s        string
		expected string
	}{
		{"example.com", "example.com."},
		{"example.com.", "example.com."},
	}

	for _, c := range cases {
		if actual := EnsureDotSuffix(c.s); actual != c.expected {
			t.Errorf("EnsureDotSuffix(%#v) expected %#v, but got %#v", c.s, c.expected, actual)
		}
	}
}

func TestParseZoneSpec(t *testing.T) {
	cases := []struct {
		s        string
		expected ZoneSpec
	}{
		{
			"example.com",
			ZoneSpec{Name: "example.com.", ID: ""},
		},
		{
			"example.com.",
			ZoneSpec{Name: "example.com.", ID: ""},
		},
		{
			"example.com/1234",
			ZoneSpec{Name: "example.com.", ID: "1234"},
		},
		{
			"example.com./1234",
			ZoneSpec{Name: "example.com.", ID: "1234"},
		},
		{
			"*/1234",
			ZoneSpec{Name: "", ID: "1234"},
		},
	}

	for _, c := range cases {
		if actual, _ := ParseZoneSpec(c.s); *actual != c.expected {
			t.Errorf("ParseZoneSpec(%#v) expected %#v, but got %#v", c.s, c.expected, *actual)
		}
	}
}

func TestParseZoneRules(t *testing.T) {
	cases := []struct {
		zones    []string
		expected ZoneRules
	}{
		{
			[]string{"*"},
			ZoneRules{
				Wildcard: true,
			},
		},
		{
			[]string{"*/*"},
			ZoneRules{
				Wildcard: true,
			},
		},
		{
			[]string{},
			ZoneRules{
				Wildcard: true,
			},
		},
		{
			[]string{"*/1234"},
			ZoneRules{
				Zones: []*ZoneSpec{
					{Name: "", ID: "1234"},
				},
				Wildcard: false,
			},
		},
		{
			[]string{"example.com"},
			ZoneRules{
				Zones: []*ZoneSpec{
					{Name: "example.com.", ID: ""},
				},
				Wildcard: false,
			},
		},
		{
			[]string{"example.com/1234"},
			ZoneRules{
				Zones: []*ZoneSpec{
					{Name: "example.com.", ID: "1234"},
				},
				Wildcard: false,
			},
		},
	}

	for _, c := range cases {
		if actual, _ := ParseZoneRules(c.zones); !reflect.DeepEqual(*actual, c.expected) {
			t.Errorf("ParseZoneRules(%#v) expected %#v, but got %#v", c.zones, c.expected, *actual)
		}
	}
}

// This is not correct

// func TestMatchesExplicitly(t *testing.T) {
// 	cases := []struct {
// 		r        ZoneRules
// 		zone     dnsprovider.Zone
// 		expected bool
// 	}{
// 		{
// 			ZoneRules{
// 				Zones: []*ZoneSpec{
// 					&ZoneSpec{Name: "example.com.", ID: "1234"},
// 				},
// 				Wildcard: false,
// 			},
// 			{},
// 			true,
// 		},
// 	}

// 	for _, c := range cases {
// 		if actual, _ := c.r.MatchesExplicitly(c.zone); !reflect.DeepEqual(*actual, c.expected) {
// 			t.Errorf("MatchesExplicitly(%#v) expected %#v, but got %#v", c.zones, c.expected, *actual)
// 		}
// 	}
// }

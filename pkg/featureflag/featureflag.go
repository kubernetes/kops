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

package featureflag

import (
	"os"
	"strings"
	"sync"
)

func Bool(b bool) *bool {
	return &b
}

// PreviewPrivateDNS turns on the preview of the private hosted zone support
var PreviewPrivateDNS = New("PreviewPrivateDNS", Bool(false))

// DNSPreCreate controls whether we pre-create DNS records
var DNSPreCreate = New("DNSPreCreate", Bool(true))

var flags = make(map[string]*FeatureFlag)
var flagsMutex sync.Mutex

var initFlags sync.Once

type FeatureFlag struct {
	Key          string
	enabled      *bool
	defaultValue *bool
}

func (f *FeatureFlag) Enabled() bool {
	if f.enabled != nil {
		return *f.enabled
	}
	if f.defaultValue != nil {
		return *f.defaultValue
	}
	return false
}

func readFlags() {
	f := os.Getenv("KOPS_FEATURE_FLAGS")
	f = strings.TrimSpace(f)
	for _, s := range strings.Split(f, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		enabled := true
		var ff *FeatureFlag
		if s[0] == '+' || s[0] == '-' {
			ff = New(s[1:], nil)
			if s[0] == '-' {
				enabled = false
			}
		} else {
			ff = New(s, nil)
		}
		ff.enabled = &enabled
	}

}

func New(key string, defaultValue *bool) *FeatureFlag {
	initFlags.Do(readFlags)

	flagsMutex.Lock()
	defer flagsMutex.Unlock()

	f := flags[key]
	if f == nil {
		f = &FeatureFlag{
			Key: key,
		}
		flags[key] = f
	}

	if f.defaultValue == nil {
		f.defaultValue = defaultValue
	}

	return f
}

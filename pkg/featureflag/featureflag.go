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

// Package featureflag implements simple feature-flagging.
// Feature flags can become an anti-pattern if abused.
// We should try to use them for two use-cases:
// * `Preview` feature flags enable a piece of functionality we haven't yet fully baked.  The user needs to 'opt-in'.
//   We expect these flags to be removed at some time.  Normally these will default to false.
// * Escape-hatch feature flags turn off a default that we consider risky (e.g. pre-creating DNS records).
//   This lets us ship a behaviour, and if we encounter unusual circumstances in the field, we can
//   allow the user to turn the behaviour off.  Normally these will default to true.
package featureflag

import (
	"os"
	"strings"
	"sync"

	"github.com/golang/glog"
)

func Bool(b bool) *bool {
	return &b
}

// DNSPreCreate controls whether we pre-create DNS records.
var DNSPreCreate = New("DNSPreCreate", Bool(true))

// DrainAndValidateRollingUpdate if set will use new rolling update code that will drain and validate.
var DrainAndValidateRollingUpdate = New("DrainAndValidateRollingUpdate", Bool(true))

// VPCSkipEnableDNSSupport if set will make that a VPC does not need DNSSupport enabled.
var VPCSkipEnableDNSSupport = New("VPCSkipEnableDNSSupport", Bool(false))

// SkipTerraformFormat if set will mean that we will not `tf fmt` the generated terraform.
var SkipTerraformFormat = New("SkipTerraformFormat", Bool(false))

var VSphereCloudProvider = New("VSphereCloudProvider", Bool(false))

var SpotinstCloudProvider = New("SpotinstCloudProvider", Bool(false))

var EnableExternalDNS = New("EnableExternalDNS", Bool(false))

//EnableExternalCloudController toggles the use of cloud-controller-manager introduced in v1.7
var EnableExternalCloudController = New("EnableExternalCloudController", Bool(false))

// EnableSeparateConfigBase allows a config-base that is different from the state store
var EnableSeparateConfigBase = New("EnableSeparateConfigBase", Bool(false))

// SpecOverrideFlag allows setting spec values on create
var SpecOverrideFlag = New("SpecOverrideFlag", Bool(false))

// GoogleCloudBucketAcl means the ACL will be set on a bucket when using GCS
// In particular, this is the only (?) way to grant the bucket.list permission
// However we should no longer need it, with the keyset.yaml fix
var GoogleCloudBucketAcl = New("GoogleCloudBucketAcl", Bool(false))

var flags = make(map[string]*FeatureFlag)
var flagsMutex sync.Mutex

func init() {
	ParseFlags(os.Getenv("KOPS_FEATURE_FLAGS"))
}

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

func ParseFlags(f string) {
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
		glog.Infof("FeatureFlag %q=%v", ff.Key, enabled)
		ff.enabled = &enabled
	}
}

func New(key string, defaultValue *bool) *FeatureFlag {
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

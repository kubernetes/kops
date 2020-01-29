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

	"k8s.io/klog"
)

const (
	// Name is the name of the environment variable which encapsulates feature flags
	Name = "KOPS_FEATURE_FLAGS"
)

func init() {
	ParseFlags(os.Getenv(Name))
}

var (
	flags      = make(map[string]*FeatureFlag)
	flagsMutex sync.Mutex
)

var (
	// DNSPreCreate controls whether we pre-create DNS records.
	DNSPreCreate = New("DNSPreCreate", Bool(true))
	// EnableLaunchTemplates indicates we wish to switch to using launch templates rather than launchconfigurations
	EnableLaunchTemplates = New("EnableLaunchTemplates", Bool(false))
	//EnableExternalCloudController toggles the use of cloud-controller-manager introduced in v1.7
	EnableExternalCloudController = New("EnableExternalCloudController", Bool(false))
	// EnableExternalDNS enables external DNS
	EnableExternalDNS = New("EnableExternalDNS", Bool(false))
	// EnableNodeAuthorization enables the node authorization features
	EnableNodeAuthorization = New("EnableNodeAuthorization", Bool(false))
	// EnableSeparateConfigBase allows a config-base that is different from the state store
	EnableSeparateConfigBase = New("EnableSeparateConfigBase", Bool(false))
	// ExperimentalClusterDNS allows for setting the kubelet dns flag to experimental values.
	ExperimentalClusterDNS = New("ExperimentalClusterDNS", Bool(false))
	// GoogleCloudBucketACL means the ACL will be set on a bucket when using GCS
	GoogleCloudBucketACL = New("GoogleCloudBucketAcl", Bool(false))
	// KeepLaunchConfigurations can be set to prevent garbage collection of old launch configurations
	KeepLaunchConfigurations = New("KeepLaunchConfigurations", Bool(false))
	// SkipTerraformFormat if set means we will not `tf fmt` the generated terraform.
	// However we should no longer need it, with the keyset.yaml fix
	// In particular, this is the only (?) way to grant the bucket.list permission
	// It allows for experiments with alternative DNS configurations - in particular local proxies.
	SkipTerraformFormat = New("SkipTerraformFormat", Bool(false))
	// SpecOverrideFlag allows setting spec values on create
	SpecOverrideFlag = New("SpecOverrideFlag", Bool(false))
	// Spotinst toggles the use of Spotinst integration.
	Spotinst = New("Spotinst", Bool(false))
	// SpotinstOcean toggles the use of Spotinst Ocean instance group implementation.
	SpotinstOcean = New("SpotinstOcean", Bool(false))
	// VPCSkipEnableDNSSupport if set will make that a VPC does not need DNSSupport enabled.
	VPCSkipEnableDNSSupport = New("VPCSkipEnableDNSSupport", Bool(false))
	// VSphereCloudProvider enables the vsphere cloud provider
	VSphereCloudProvider = New("VSphereCloudProvider", Bool(false))
	// SkipEtcdVersionCheck will bypass the check that etcd-manager is using a supported etcd version
	SkipEtcdVersionCheck = New("SkipEtcdVersionCheck", Bool(false))
	// Enable terraform JSON output instead of hcl output. JSON output can be also parsed by terraform 0.12
	TerraformJSON = New("TerraformJSON", Bool(false))
)

// FeatureFlag defines a feature flag
type FeatureFlag struct {
	Key          string
	enabled      *bool
	defaultValue *bool
}

// New creates a new feature flag
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

// Enabled checks if the flag is enabled
func (f *FeatureFlag) Enabled() bool {
	if f.enabled != nil {
		return *f.enabled
	}
	if f.defaultValue != nil {
		return *f.defaultValue
	}
	return false
}

// Bool returns a pointer to the boolean value
func Bool(b bool) *bool {
	return &b
}

// ParseFlags responsible for parse out the feature flag usage
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
		klog.Infof("FeatureFlag %q=%v", ff.Key, enabled)
		ff.enabled = &enabled
	}
}

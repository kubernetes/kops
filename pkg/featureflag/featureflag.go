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

	"k8s.io/klog/v2"
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
	// CacheNodeidentityInfo enables NodeidentityInfo caching
	// in order to reduce the number of EC2 DescribeInstance calls.
	CacheNodeidentityInfo = new("CacheNodeidentityInfo", Bool(false))
	// EnableSeparateConfigBase allows a config-base that is different from the state store
	EnableSeparateConfigBase = new("EnableSeparateConfigBase", Bool(false))
	// ExperimentalClusterDNS allows for setting the kubelet dns flag to experimental values.
	ExperimentalClusterDNS = new("ExperimentalClusterDNS", Bool(false))
	// GoogleCloudBucketACL means the ACL will be set on a bucket when using GCS
	GoogleCloudBucketACL = new("GoogleCloudBucketAcl", Bool(false))
	// SpecOverrideFlag allows setting spec values on create
	SpecOverrideFlag = new("SpecOverrideFlag", Bool(false))
	// Spotinst toggles the use of Spotinst integration.
	Spotinst = new("Spotinst", Bool(false))
	// SpotinstOcean toggles the use of Spotinst Ocean instance group implementation.
	SpotinstOcean = new("SpotinstOcean", Bool(false))
	// SpotinstHybrid toggles between hybrid and full instance group implementations.
	SpotinstHybrid = new("SpotinstHybrid", Bool(false))
	// SpotinstController toggles the installation of the Spotinst controller addon.
	SpotinstController = new("SpotinstController", Bool(true))
	// VFSVaultSupport enables setting Vault as secret/keystore
	VFSVaultSupport = new("VFSVaultSupport", Bool(false))
	// VPCSkipEnableDNSSupport if set will make that a VPC does not need DNSSupport enabled.
	VPCSkipEnableDNSSupport = new("VPCSkipEnableDNSSupport", Bool(false))
	// SkipEtcdVersionCheck will bypass the check that etcd-manager is using a supported etcd version
	SkipEtcdVersionCheck = new("SkipEtcdVersionCheck", Bool(false))
	// ClusterAddons activates experimental cluster-addons support
	ClusterAddons = new("ClusterAddons", Bool(false))
	// Azure toggles the Azure support.
	Azure = new("Azure", Bool(false))
	// KopsControllerStateStore enables fetching the kops state from kops-controller, instead of requiring access to S3/GCS/etc.
	KopsControllerStateStore = new("KopsControllerStateStore", Bool(false))
	// APIServerNodes enables ability to provision nodes that only run the kube-apiserver.
	APIServerNodes = new("APIServerNodes", Bool(false))
	// UseAddonOperators activates experimental addon operator support
	UseAddonOperators = new("UseAddonOperators", Bool(false))
	// TerraformManagedFiles enables rendering managed files into the Terraform configuration.
	TerraformManagedFiles = new("TerraformManagedFiles", Bool(true))
	// AlphaAllowGCE is a feature flag that gates GCE support while it is alpha.
	AlphaAllowGCE = new("AlphaAllowGCE", Bool(false))
	// Karpenter enables karpenter-managed Instance Groups
	Karpenter = new("Karpenter", Bool(false))
	// ContainerRegistryIsMirror specifies that the containerRegistry supports nested image names (e.g. registry-sandbox.gcr.io)
	ContainerRegistryIsMirror = new("ContainerRegistryIsMirror", Bool(false))
)

// FeatureFlag defines a feature flag
type FeatureFlag struct {
	Key          string
	enabled      *bool
	defaultValue *bool
}

// new creates a new feature flag
func new(key string, defaultValue *bool) *FeatureFlag {
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
	flagsMutex.Lock()
	defer flagsMutex.Unlock()

	f = strings.TrimSpace(f)
	for _, s := range strings.Split(f, ",") {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		enabled := true
		var ff *FeatureFlag
		if s[0] == '+' || s[0] == '-' {
			ff = flags[s[1:]]
			if s[0] == '-' {
				enabled = false
			}
		} else {
			ff = flags[s]
		}
		if ff != nil {
			klog.Infof("FeatureFlag %q=%v", ff.Key, enabled)
			ff.enabled = &enabled
		} else {
			klog.Infof("Unknown FeatureFlag %q", s)
		}
	}
}

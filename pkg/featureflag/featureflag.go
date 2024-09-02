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
//   - `Preview` feature flags enable a piece of functionality we haven't yet fully baked.  The user needs to 'opt-in'.
//     We expect these flags to be removed at some time.  Normally these will default to false.
//   - Escape-hatch feature flags turn off a default that we consider risky (e.g. pre-creating DNS records).
//     This lets us ship a behaviour, and if we encounter unusual circumstances in the field, we can
//     allow the user to turn the behaviour off.  Normally these will default to true.
package featureflag

import (
	"fmt"
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
	// Spotinst toggles the use of Spotinst integration.
	Spotinst = new("Spotinst", Bool(false))
	// SpotinstOcean toggles the use of Spotinst Ocean instance group implementation.
	SpotinstOcean = new("SpotinstOcean", Bool(false))
	// SpotinstOceanTemplate toggles the use of Spotinst Ocean object as a template for Virtual Node Groups.
	SpotinstOceanTemplate = new("SpotinstOceanTemplate", Bool(false))
	// SpotinstHybrid toggles between hybrid and full instance group implementations.
	SpotinstHybrid = new("SpotinstHybrid", Bool(false))
	// SpotinstController toggles the installation of the Spotinst controller addon.
	SpotinstController = new("SpotinstController", Bool(true))
	// VPCSkipEnableDNSSupport if set will make that a VPC does not need DNSSupport enabled.
	VPCSkipEnableDNSSupport = new("VPCSkipEnableDNSSupport", Bool(false))
	// SkipEtcdVersionCheck will bypass the check that etcd-manager is using a supported etcd version
	SkipEtcdVersionCheck = new("SkipEtcdVersionCheck", Bool(false))
	// ClusterAddons activates experimental cluster-addons support
	ClusterAddons = new("ClusterAddons", Bool(false))
	// Azure toggles the Azure support.
	Azure = new("Azure", Bool(false))
	// APIServerNodes enables ability to provision nodes that only run the kube-apiserver.
	APIServerNodes = new("APIServerNodes", Bool(false))
	// UseAddonOperators activates experimental addon operator support
	UseAddonOperators = new("UseAddonOperators", Bool(false))
	// TerraformManagedFiles enables rendering managed files into the Terraform configuration.
	TerraformManagedFiles = new("TerraformManagedFiles", Bool(true))
	// ImageDigest remaps all manifests with image digests
	ImageDigest = new("ImageDigest", Bool(true))
	// Scaleway toggles the Scaleway Cloud support.
	Scaleway = new("Scaleway", Bool(false))
	// SELinuxMount configures AWS EBS and GCE PD CSI drivers for SELinuxMount support.
	// It expects than Kubernetes feature gate SELinuxMountReadWriteOncePod is
	// enabled or GA in the API server, KCM and kubelet.
	// OS with SELinux support on all nodes is recommended, but not required
	// - the feature won't do anything when the node OS does not support SELinux.
	// TODO(jsafrane): add to all CSI drivers installed by kops.
	SELinuxMount = new("SELinuxMount", Bool(false))
	// DO Terraform toggles the DO terraform support.
	DOTerraform = new("DOTerraform", Bool(false))
	// Metal enables the experimental bare-metal support.
	Metal = new("Metal", Bool(false))
	// AWSSingleNodesInstanceGroup enables the creation of a single node instance group instead of one per availability zone.
	AWSSingleNodesInstanceGroup = new("AWSSingleNodesInstanceGroup", Bool(false))
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

// Get returns given FeatureFlag.
func Get(flagName string) (*FeatureFlag, error) {
	flagsMutex.Lock()
	defer flagsMutex.Unlock()

	flag, found := flags[flagName]
	if !found {
		return nil, fmt.Errorf("flag %s not found", flagName)
	}
	return flag, nil
}

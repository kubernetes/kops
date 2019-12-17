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

package kops

import (
	"fmt"
	"net/url"

	"github.com/blang/semver"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/util/pkg/vfs"
)

var DefaultChannelBase = "https://raw.githubusercontent.com/kubernetes/kops/master/channels/"

const (
	DefaultChannel = "stable"
	AlphaChannel   = "alpha"
)

type Channel struct {
	v1.TypeMeta `json:",inline"`
	ObjectMeta  metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ChannelSpec `json:"spec,omitempty"`
}

type ChannelSpec struct {
	Images []*ChannelImageSpec `json:"images,omitempty"`

	Cluster *ClusterSpec `json:"cluster,omitempty"`

	// KopsVersions allows us to recommend/require kops versions
	KopsVersions []KopsVersionSpec `json:"kopsVersions,omitempty"`

	// KubernetesVersions allows us to recommend/requires kubernetes versions
	KubernetesVersions []KubernetesVersionSpec `json:"kubernetesVersions,omitempty"`
}

type KopsVersionSpec struct {
	Range string `json:"range,omitempty"`

	// RecommendedVersion is the recommended version of kops to use for this Range of kops versions
	RecommendedVersion string `json:"recommendedVersion,omitempty"`

	// RequiredVersion is the required version of kops to use for this Range of kops versions, forcing an upgrade
	RequiredVersion string `json:"requiredVersion,omitempty"`

	// KubernetesVersion is the default version of kubernetes to use with this kops version e.g. for new clusters
	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
}

type KubernetesVersionSpec struct {
	Range string `json:"range,omitempty"`

	RecommendedVersion string `json:"recommendedVersion,omitempty"`
	RequiredVersion    string `json:"requiredVersion,omitempty"`
}

type ChannelImageSpec struct {
	Labels map[string]string `json:"labels,omitempty"`

	ProviderID string `json:"providerID,omitempty"`

	Name string `json:"name,omitempty"`

	KubernetesVersion string `json:"kubernetesVersion,omitempty"`
}

// LoadChannel loads a Channel object from the specified VFS location
func LoadChannel(location string) (*Channel, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("invalid channel: %q", location)
	}

	if !u.IsAbs() {
		base, err := url.Parse(DefaultChannelBase)
		if err != nil {
			return nil, fmt.Errorf("invalid base channel location: %q", DefaultChannelBase)
		}
		klog.V(4).Infof("resolving %q against default channel location %q", location, DefaultChannelBase)
		u = base.ResolveReference(u)
	}

	resolved := u.String()
	klog.V(2).Infof("Loading channel from %q", resolved)
	channelBytes, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading channel %q: %v", resolved, err)
	}
	channel, err := ParseChannel(channelBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %q: %v", resolved, err)
	}
	klog.V(4).Infof("Channel contents: %s", string(channelBytes))

	return channel, nil
}

// ParseChannel parses a Channel object
func ParseChannel(channelBytes []byte) (*Channel, error) {
	channel := &Channel{}
	err := ParseRawYaml(channelBytes, channel)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %v", err)
	}

	return channel, nil
}

// FindRecommendedUpgrade returns a string with a new version, if the current version is out of date
func (v *KubernetesVersionSpec) FindRecommendedUpgrade(version semver.Version) (*semver.Version, error) {
	if v.RecommendedVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RecommendedVersion")
		return nil, nil
	}

	recommendedVersion, err := util.ParseKubernetesVersion(v.RecommendedVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing RecommendedVersion %q from channel", v.RecommendedVersion)
	}
	if recommendedVersion.GT(version) {
		klog.V(2).Infof("RecommendedVersion=%q, Have=%q.  Recommending upgrade", recommendedVersion, version)
		return recommendedVersion, nil
	}
	klog.V(4).Infof("RecommendedVersion=%q, Have=%q.  No upgrade needed.", recommendedVersion, version)
	return nil, nil
}

// FindRecommendedUpgrade returns a string with a new version, if the current version is out of date
func (v *KopsVersionSpec) FindRecommendedUpgrade(version semver.Version) (*semver.Version, error) {
	if v.RecommendedVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RecommendedVersion")
		return nil, nil
	}

	recommendedVersion, err := semver.ParseTolerant(v.RecommendedVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing RecommendedVersion %q from channel", v.RecommendedVersion)
	}
	if recommendedVersion.GT(version) {
		klog.V(2).Infof("RecommendedVersion=%q, Have=%q.  Recommending upgrade", recommendedVersion, version)
		return &recommendedVersion, nil
	}
	klog.V(4).Infof("RecommendedVersion=%q, Have=%q.  No upgrade needed.", recommendedVersion, version)
	return nil, nil
}

// IsUpgradeRequired returns true if the current version is not acceptable
func (v *KubernetesVersionSpec) IsUpgradeRequired(version semver.Version) (bool, error) {
	if v.RequiredVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RequiredVersion")
		return false, nil
	}

	requiredVersion, err := util.ParseKubernetesVersion(v.RequiredVersion)
	if err != nil {
		return false, fmt.Errorf("error parsing RequiredVersion %q from channel", v.RequiredVersion)
	}
	if requiredVersion.GT(version) {
		klog.V(2).Infof("RequiredVersion=%q, Have=%q.  Requiring upgrade", requiredVersion, version)
		return true, nil
	}
	klog.V(4).Infof("RequiredVersion=%q, Have=%q.  No upgrade needed.", requiredVersion, version)
	return false, nil
}

// IsUpgradeRequired returns true if the current version is not acceptable
func (v *KopsVersionSpec) IsUpgradeRequired(version semver.Version) (bool, error) {
	if v.RequiredVersion == "" {
		klog.V(2).Infof("VersionRecommendationSpec does not specify RequiredVersion")
		return false, nil
	}

	requiredVersion, err := semver.ParseTolerant(v.RequiredVersion)
	if err != nil {
		return false, fmt.Errorf("error parsing RequiredVersion %q from channel", v.RequiredVersion)
	}
	if requiredVersion.GT(version) {
		klog.V(2).Infof("RequiredVersion=%q, Have=%q.  Requiring upgrade", requiredVersion, version)
		return true, nil
	}
	klog.V(4).Infof("RequiredVersion=%q, Have=%q.  No upgrade needed.", requiredVersion, version)
	return false, nil
}

// FindKubernetesVersionSpec returns a KubernetesVersionSpec for the current version
func FindKubernetesVersionSpec(versions []KubernetesVersionSpec, version semver.Version) *KubernetesVersionSpec {
	for i := range versions {
		v := &versions[i]
		if v.Range != "" {
			versionRange, err := semver.ParseRange(v.Range)
			if err != nil {
				klog.Warningf("unable to parse range in channel version spec: %q", v.Range)
				continue
			}
			if !versionRange(version) {
				klog.V(8).Infof("version range %q does not apply to version %q; skipping", v.Range, version)
				continue
			}
		}
		return v
	}

	return nil
}

// FindKopsVersionSpec returns a KopsVersionSpec for the current version
func FindKopsVersionSpec(versions []KopsVersionSpec, version semver.Version) *KopsVersionSpec {
	for i := range versions {
		v := &versions[i]
		if v.Range != "" {
			versionRange, err := semver.ParseRange(v.Range)
			if err != nil {
				klog.Warningf("unable to parse range in channel version spec: %q", v.Range)
				continue
			}
			if !versionRange(version) {
				klog.V(8).Infof("version range %q does not apply to version %q; skipping", v.Range, version)
				continue
			}
		}
		return v
	}

	return nil
}

type CloudProviderID string

const (
	CloudProviderALI       CloudProviderID = "alicloud"
	CloudProviderAWS       CloudProviderID = "aws"
	CloudProviderBareMetal CloudProviderID = "baremetal"
	CloudProviderDO        CloudProviderID = "digitalocean"
	CloudProviderGCE       CloudProviderID = "gce"
	CloudProviderOpenstack CloudProviderID = "openstack"
	CloudProviderVSphere   CloudProviderID = "vsphere"
)

// FindImage returns the image for the cloudprovider, or nil if none found
func (c *Channel) FindImage(provider CloudProviderID, kubernetesVersion semver.Version) *ChannelImageSpec {
	var matches []*ChannelImageSpec

	for _, image := range c.Spec.Images {
		if image.ProviderID != string(provider) {
			continue
		}
		if image.KubernetesVersion != "" {
			versionRange, err := semver.ParseRange(image.KubernetesVersion)
			if err != nil {
				klog.Warningf("cannot parse KubernetesVersion=%q", image.KubernetesVersion)
				continue
			}

			if !versionRange(kubernetesVersion) {
				klog.V(2).Infof("Kubernetes version %q does not match range: %s", kubernetesVersion, image.KubernetesVersion)
				continue
			}
		}
		matches = append(matches, image)
	}

	if len(matches) == 0 {
		klog.V(2).Infof("No matching images in channel for cloudprovider %q", provider)
		return nil
	}

	if len(matches) != 1 {
		klog.Warningf("Multiple matching images in channel for cloudprovider %q", provider)
	}
	return matches[0]
}

// RecommendedKubernetesVersion returns the recommended kubernetes version for a version of kops
// It is used by default when creating a new cluster, for example
func RecommendedKubernetesVersion(c *Channel, kopsVersionString string) *semver.Version {
	kopsVersion, err := semver.ParseTolerant(kopsVersionString)
	if err != nil {
		klog.Warningf("unable to parse kops version %q", kopsVersionString)
	} else {
		kopsVersionSpec := FindKopsVersionSpec(c.Spec.KopsVersions, kopsVersion)
		if kopsVersionSpec != nil {
			if kopsVersionSpec.KubernetesVersion != "" {
				sv, err := util.ParseKubernetesVersion(kopsVersionSpec.KubernetesVersion)
				if err != nil {
					klog.Warningf("unable to parse kubernetes version %q", kopsVersionSpec.KubernetesVersion)
				} else {
					return sv
				}
			}
		}
	}

	if c.Spec.Cluster != nil {
		sv, err := util.ParseKubernetesVersion(c.Spec.Cluster.KubernetesVersion)
		if err != nil {
			klog.Warningf("unable to parse kubernetes version %q", c.Spec.Cluster.KubernetesVersion)
		} else {
			return sv
		}
	}

	return nil
}

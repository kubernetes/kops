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

package kops

import (
	"fmt"
	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/meta/v1"
	"net/url"
)

const DefaultChannelBase = "https://raw.githubusercontent.com/kubernetes/kops/master/channels/"
const DefaultChannel = "stable"
const AlphaChannel = "alpha"

type Channel struct {
	v1.TypeMeta `json:",inline"`
	ObjectMeta  api.ObjectMeta `json:"metadata,omitempty"`

	Spec ChannelSpec `json:"spec,omitempty"`
}

type ChannelSpec struct {
	Images []*ChannelImageSpec `json:"images,omitempty"`

	Cluster *ClusterSpec `json:"cluster,omitempty"`

	// KopsVersions allows us to recommend/require kops versions
	KopsVersions []VersionSpec `json:"kopsVersions,omitempty"`

	// KubernetesVersions allows us to recommend/requires kubernetes versions
	KubernetesVersions []VersionSpec `json:"kubernetesVersions,omitempty"`
}

type VersionSpec struct {
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
		glog.V(4).Infof("resolving %q against default channel location %q", location, DefaultChannelBase)
		u = base.ResolveReference(u)
	}

	resolved := u.String()
	glog.V(2).Infof("Loading channel from %q", resolved)
	channelBytes, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading channel %q: %v", resolved, err)
	}
	channel, err := ParseChannel(channelBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %q: %v", resolved, err)
	}
	glog.V(4).Infof("Channel contents: %s", string(channelBytes))

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
func FindRecommendedUpgrade(v *VersionSpec, version semver.Version) (string, error) {
	if v.RecommendedVersion == "" {
		glog.V(2).Infof("VersionRecommendationSpec does not specify RecommendedVersion")
		return "", nil
	}

	recommendedVersion, err := semver.Parse(v.RecommendedVersion)
	if err != nil {
		return "", fmt.Errorf("error parsing RecommendedVersion %q from channel", v.RecommendedVersion)
	}
	if recommendedVersion.GT(version) {
		glog.V(2).Infof("RecommendedVersion=%q, Have=%q.  Recommending upgrade", recommendedVersion, version)
		return v.RecommendedVersion, nil
	} else {
		glog.V(4).Infof("RecommendedVersion=%q, Have=%q.  No upgrade needed.", recommendedVersion, version)
	}
	return "", nil
}

// IsUpgradeRequired returns true if the current version is not acceptable
func IsUpgradeRequired(v *VersionSpec, version semver.Version) (bool, error) {
	if v.RequiredVersion == "" {
		glog.V(2).Infof("VersionRecommendationSpec does not specify RequiredVersion")
		return false, nil
	}

	requiredVersion, err := semver.Parse(v.RequiredVersion)
	if err != nil {
		return false, fmt.Errorf("error parsing RequiredVersion %q from channel", v.RequiredVersion)
	}
	if requiredVersion.GT(version) {
		glog.V(2).Infof("RequiredVersion=%q, Have=%q.  Requiring upgrade", requiredVersion, version)
		return true, nil
	} else {
		glog.V(4).Infof("RequiredVersion=%q, Have=%q.  No upgrade needed.", requiredVersion, version)
	}
	return false, nil
}

// FindVersionInfo returns a VersionSpec for the current version
func FindVersionInfo(versions []VersionSpec, version semver.Version) *VersionSpec {
	for i := range versions {
		v := &versions[i]
		if v.Range != "" {
			versionRange, err := semver.ParseRange(v.Range)
			if err != nil {
				glog.Warningf("unable to parse range in channel version spec: %q", v.Range)
				continue
			}
			if !versionRange(version) {
				glog.V(8).Infof("version range %q does not apply to version %q; skipping", v.Range, version)
				continue
			}
		}
		return v
	}

	return nil
}

// FindImage returns the image for the cloudprovider, or nil if none found
func (c *Channel) FindImage(provider fi.CloudProviderID, kubernetesVersion semver.Version) *ChannelImageSpec {
	var matches []*ChannelImageSpec

	for _, image := range c.Spec.Images {
		if image.ProviderID != string(provider) {
			continue
		}
		if image.KubernetesVersion != "" {
			versionRange, err := semver.ParseRange(image.KubernetesVersion)
			if err != nil {
				glog.Warningf("cannot parse KubernetesVersion=%q", image.KubernetesVersion)
				continue
			}

			if !versionRange(kubernetesVersion) {
				glog.V(2).Infof("Kubernetes version %q does not match range: %s", kubernetesVersion, image.KubernetesVersion)
				continue
			}
		}
		matches = append(matches, image)
	}

	if len(matches) == 0 {
		glog.V(2).Infof("No matching images in channel for cloudprovider %q", provider)
		return nil
	}

	if len(matches) != 1 {
		glog.Warningf("Multiple matching images in channel for cloudprovider %q", provider)
	}
	return matches[0]
}

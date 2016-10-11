package kops

import (
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"net/url"
)

const DefaultChannelBase = "https://raw.githubusercontent.com/kubernetes/kops/master/channels/"
const DefaultChannel = "stable"

type Channel struct {
	unversioned.TypeMeta `json:",inline"`
	ObjectMeta    `json:"metadata,omitempty"`

	Spec ChannelSpec `json:"spec,omitempty"`
}

type ChannelSpec struct {
	Images []*ChannelImageSpec `json:"images,omitempty"`

	Cluster *ClusterSpec `json:"cluster,omitempty"`
}

type ChannelImageSpec struct {
	Labels map[string]string `json:"labels,omitempty"`

	ProviderID string `json:"providerID,omitempty"`

	Name string `json:"name,omitempty"`
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
		u = base.ResolveReference(u)
	}

	resolved := u.String()
	glog.V(2).Infof("Loading channel from %q", resolved)
	channel := &Channel{}
	channelBytes, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading channel %q: %v", resolved, err)
	}
	err = ParseYaml(channelBytes, channel)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %q: %v", resolved, err)
	}
	glog.V(4).Info("Channel contents: %s", string(channelBytes))
	return channel, nil
}

// FindImage returns the image for the cloudprovider, or nil if none found
func (c *Channel) FindImage(provider fi.CloudProviderID) *ChannelImageSpec {
	var matches []*ChannelImageSpec

	for _, image := range c.Spec.Images {
		if image.ProviderID != string(provider) {
			continue
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

package bundles

import (
	"fmt"
	"net/url"

	"github.com/golang/glog"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

/*
// LoadChannelManifest loads a file object from the specified VFS location
func LoadChannelManifest(channel string) ([]byte, error) {
	resolved, err := resolveChannel(channel)
	if err != nil {
		return nil, err
	}

	glog.V(2).Infof("loading channel from %q", resolved)
	contents, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading channel %q: %v", resolved, err)
	}

	return contents, nil
}
*/

func resolveChannel(channel string) (string, error) {
	u, err := url.Parse(channel)
	if err != nil {
		return "", fmt.Errorf("invalid channel: %q", channel)
	}

	if !u.IsAbs() {
		base, err := url.Parse(kops.DefaultChannelBase)
		if err != nil {
			return "", fmt.Errorf("invalid base channel location: %q", kops.DefaultChannelBase)
		}
		glog.V(4).Infof("resolving %q against default channel location %q", channel, kops.DefaultChannelBase)
		u = base.ResolveReference(u)
	}

	return u.String(), nil
}

// LoadBundleManifest loads a file object from the specified VFS location
func LoadBundleManifest(clusterSpec *kops.ClusterSpec, location string) ([]byte, error) {
	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("invalid bundle: %q", location)
	}

	if !u.IsAbs() {
		channel, err := resolveChannel(clusterSpec.Channel)
		if err != nil {
			return nil, err
		}

		base, err := url.Parse(channel)
		if err != nil {
			return nil, fmt.Errorf("invalid channel location: %q", channel)
		}
		glog.V(4).Infof("resolving %q against channel location %q", location, channel)
		u = base.ResolveReference(u)
	}

	resolved := u.String()
	glog.V(2).Infof("loading bundle from %q", resolved)
	contents, err := vfs.Context.ReadFile(resolved)
	if err != nil {
		return nil, fmt.Errorf("error reading bundle %q: %v", resolved, err)
	}

	return contents, nil
}

/*
Copyright 2022 The Kubernetes Authors.

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

package channel

import (
	"fmt"
	"net/url"

	"k8s.io/klog/v2"
)

const GithubChannelBase = "https://raw.githubusercontent.com/kubernetes/kops/master/channels/"

// TestChannelBase can be set by tests to override the channel base.
var TestChannelBase = ""

func ChannelBase() string {
	if TestChannelBase != "" {
		return TestChannelBase
	}
	return GithubChannelBase
}

// ResolveChannel maps a channel to an absolute URL (possibly a VFS URL)
// If the channel is the well-known "none" value, we return (nil, nil)
func ResolveChannel(location string) (*url.URL, error) {
	if location == "none" {
		return nil, nil
	}

	u, err := url.Parse(location)
	if err != nil {
		return nil, fmt.Errorf("invalid channel location: %q", location)
	}

	if !u.IsAbs() {
		channelBase := ChannelBase()
		base, err := url.Parse(channelBase)
		if err != nil {
			return nil, fmt.Errorf("invalid base channel location: %q", channelBase)
		}
		klog.V(4).Infof("resolving %q against default channel location %q", location, channelBase)
		u = base.ResolveReference(u)
	}

	if isOCI(u) {
		return resolveOCIChannel(u)
	}

	return u, nil
}

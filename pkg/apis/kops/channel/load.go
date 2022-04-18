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

	"k8s.io/klog/v2"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/vfs"
)

// ParseChannel parses a Channel object
func ParseChannel(channelBytes []byte) (*api.Channel, error) {
	channel := &api.Channel{}
	err := api.ParseRawYaml(channelBytes, channel)
	if err != nil {
		return nil, fmt.Errorf("error parsing channel %v", err)
	}

	return channel, nil
}

// LoadChannel loads a Channel object from the specified VFS location
func LoadChannel(location string) (*api.Channel, error) {
	resolvedURL, err := ResolveChannel(location)
	if err != nil {
		return nil, err
	}

	if resolvedURL == nil {
		return &api.Channel{}, nil
	}

	if isOCI(resolvedURL) {
		return LoadChannelFromOCI(resolvedURL)
	}

	resolved := resolvedURL.String()

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

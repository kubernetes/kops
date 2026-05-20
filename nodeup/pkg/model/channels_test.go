/*
Copyright 2026 The Kubernetes Authors.

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

package model

import (
	"reflect"
	"testing"

	"k8s.io/kops/upup/pkg/fi"
)

func TestChannelsBuilder(t *testing.T) {
	RunGoldenTest(t, "tests/channels/", "channels", func(nodeupModelContext *NodeupModelContext, target *fi.NodeupModelBuilderContext) error {
		nodeupModelContext.NodeupConfig.Channels = []string{
			"memfs://clusters.example.com/minimal.example.com/addons/bootstrap-channel.yaml",
		}
		builder := ChannelsBuilder{NodeupModelContext: nodeupModelContext}
		return builder.Build(target)
	})
}

func TestFileChannelDirs(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{name: "none", in: []string{"s3://bucket/x/bootstrap-channel.yaml"}, want: nil},
		{name: "one", in: []string{"file:///etc/kubernetes/kops/config/addons/bootstrap-channel.yaml"}, want: []string{"/etc/kubernetes/kops/config/addons"}},
		{name: "mixed and dedup", in: []string{
			"file:///etc/kubernetes/kops/config/addons/bootstrap-channel.yaml",
			"s3://bucket/x/bootstrap-channel.yaml",
			"file:///etc/kubernetes/kops/config/addons/extra-channel.yaml",
		}, want: []string{"/etc/kubernetes/kops/config/addons"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := fileChannelDirs(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("fileChannelDirs(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

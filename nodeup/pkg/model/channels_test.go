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

	v1 "k8s.io/api/core/v1"
	"k8s.io/kops/pkg/apis/nodeup"
)

func channelsPod(args []string, containerName string) *v1.Pod {
	return &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: containerName, Args: args},
			},
		},
	}
}

// TestSetChannelsNodeLabels checks that only the --node-labels argument is rewritten. The
// manifest in the state store is shared by every instance group and carries the cluster-wide
// default; nodeup narrows it to what this instance group claims.
func TestSetChannelsNodeLabels(t *testing.T) {
	baseArgs := []string{
		"apply", "channel",
		"--v=4",
		"--yes",
		"--interval=5m0s",
		"--node-labels=node-role.kops.k8s.io/kops-channel,node-role.kops.k8s.io/cert-manager",
		"--node-name=$(NODE_NAME)",
		"memfs://channel.yaml",
	}

	pod := channelsPod(append([]string{}, baseArgs...), "kops-channels")
	if err := setChannelsNodeLabels(pod, []string{"node-role.kops.k8s.io/kops-controller"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := append([]string{}, baseArgs...)
	want[5] = "--node-labels=node-role.kops.k8s.io/kops-controller"
	if got := pod.Spec.Containers[0].Args; !reflect.DeepEqual(got, want) {
		t.Errorf("args = %v, want %v", got, want)
	}
}

func TestSetChannelsNodeLabelsMultiple(t *testing.T) {
	pod := channelsPod([]string{"apply", "channel", "--node-labels=a", "chan"}, "kops-channels")
	labels := []string{"node-role.kops.k8s.io/kops-channel", "node-role.kops.k8s.io/kops-controller"}
	if err := setChannelsNodeLabels(pod, labels); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "--node-labels=node-role.kops.k8s.io/kops-channel,node-role.kops.k8s.io/kops-controller"
	if got := pod.Spec.Containers[0].Args[2]; got != want {
		t.Errorf("arg = %q, want %q", got, want)
	}
}

// TestSetChannelsNodeLabelsErrors covers the cases where the manifest is not the shape we expect.
// Silently doing nothing would leave the node claiming the cluster-wide default set of
// components, which is what this rewrite exists to narrow.
func TestSetChannelsNodeLabelsErrors(t *testing.T) {
	grid := []struct {
		name string
		pod  *v1.Pod
	}{
		{
			name: "no node-labels argument",
			pod:  channelsPod([]string{"apply", "channel", "chan"}, "kops-channels"),
		},
		{
			name: "no kops-channels container",
			pod:  channelsPod([]string{"--node-labels=a"}, "something-else"),
		},
		{
			name: "no containers",
			pod:  &v1.Pod{},
		},
	}

	for _, g := range grid {
		t.Run(g.name, func(t *testing.T) {
			if err := setChannelsNodeLabels(g.pod, []string{"x"}); err == nil {
				t.Errorf("expected an error, got none")
			}
		})
	}
}

// TestRunsChannels checks that placement follows the manifest path cloudup sets, rather than a
// role check that could drift from it.
func TestRunsChannels(t *testing.T) {
	for _, g := range []struct {
		manifest string
		want     bool
	}{
		{manifest: "memfs://tests/manifests/channels/kops-channels.yaml", want: true},
		{manifest: "", want: false},
	} {
		b := &ChannelsBuilder{NodeupModelContext: &NodeupModelContext{
			NodeupConfig: &nodeup.Config{ChannelsManifest: g.manifest},
		}}
		if got := b.runsChannels(); got != g.want {
			t.Errorf("ChannelsManifest=%q: runsChannels() = %v, want %v", g.manifest, got, g.want)
		}
	}
}

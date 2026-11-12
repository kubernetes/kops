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
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/pflag"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/kopscodecs"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/upup/pkg/fi"
	kubelet "k8s.io/kubelet/config/v1beta1"
	"sigs.k8s.io/yaml"
)

func TestKubeletConfigurationClusterDNS(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      []string
		wantError bool
	}{
		{name: "ordered CSV", input: "192.168.0.1,10.4.240.10", want: []string{"192.168.0.1", "10.4.240.10"}},
		{name: "single IP", input: "10.4.240.10", want: []string{"10.4.240.10"}},
		{name: "token whitespace", input: " 192.168.0.1 , 10.4.240.10 ", want: []string{"192.168.0.1", "10.4.240.10"}},
		{name: "empty", want: []string{}},
		{name: "whitespace only", input: " 	 ", wantError: true},
		// CSV skips blank lines and strips quotes; only an exact empty API value
		// means unset, not a nonempty value containing no addresses.
		{name: "blank line only", input: "\n", wantError: true},
		{name: "quoted empty token", input: `""`, wantError: true},
		{name: "invalid IP", input: "192.168.0.1,not-an-ip", wantError: true},
		{name: "empty token", input: "192.168.0.1,,10.4.240.10", wantError: true},
		{name: "trailing comma", input: "192.168.0.1,", wantError: true},
		{name: "leading comma", input: ",10.4.240.10", wantError: true},
		{name: "whitespace token", input: "192.168.0.1, ,10.4.240.10", wantError: true},
		{name: "quoted CSV", input: `"192.168.0.1","10.4.240.10"`, want: []string{"192.168.0.1", "10.4.240.10"}},
		// The legacy parser removes all three quote styles, including whole-list quotes.
		{name: "legacy quotes", input: "'192.168.0.1,10.4.240.10'", want: []string{"192.168.0.1", "10.4.240.10"}},
		{name: "backtick quotes", input: "`192.168.0.1`,`10.4.240.10`", want: []string{"192.168.0.1", "10.4.240.10"}},
		// IPv6 normalization and duplicate retention match the old IP slice flag.
		{name: "IPv6", input: "2001:0db8::1,10.4.240.10", want: []string{"2001:db8::1", "10.4.240.10"}},
		{name: "duplicates", input: "10.4.240.10,10.4.240.10", want: []string{"10.4.240.10", "10.4.240.10"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Exercise the actual legacy kubelet flag parser as a compatibility oracle.
			flags := pflag.NewFlagSet("legacy-kubelet", pflag.ContinueOnError)
			legacy := flags.IPSlice("cluster-dns", nil, "")
			legacyErr := flags.Set("cluster-dns", tt.input)
			// Unlike a missing CLI flag, a nonempty API value must contain an IP.
			legacyEmpty := tt.input != "" && len(*legacy) == 0
			if (legacyErr != nil || legacyEmpty) != tt.wantError {
				t.Fatalf("legacy parser error = %v, empty = %v, wantError %v", legacyErr, legacyEmpty, tt.wantError)
			}
			if !tt.wantError {
				got := make([]string, len(*legacy))
				for i, ip := range *legacy {
					got[i] = ip.String()
				}
				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("legacy DNS = %v, want %v", got, tt.want)
				}
			}
			builder := &KubeletBuilder{NodeupModelContext: &NodeupModelContext{
				NodeupConfig: &nodeup.Config{ContainerdConfig: &kops.ContainerdConfig{}},
			}}
			config := &kops.KubeletConfigSpec{ClusterDNS: tt.input}
			cc, err := builder.kubeletConfiguration(config, "")
			file, fileErr := builder.buildKubeletComponentConfig(config, "")
			if tt.wantError {
				for _, e := range []error{err, fileErr} {
					if e == nil || !strings.Contains(e.Error(), "clusterDNS") {
						t.Errorf("error = %v, want clusterDNS field error", e)
					}
				}
				if file != nil {
					t.Error("invalid DNS produced a file task")
				}
				return
			}
			if err != nil || fileErr != nil {
				t.Fatalf("conversion error = %v, file error = %v", err, fileErr)
			}
			if !reflect.DeepEqual(cc.ClusterDNS, tt.want) {
				t.Errorf("converted DNS = %v, want ordered %v", cc.ClusterDNS, tt.want)
			}
			if file.Path != kubeletConfigFilePath {
				t.Errorf("file path = %q", file.Path)
			}
			data, err := fi.ResourceAsBytes(file.Contents)
			if err != nil {
				t.Fatal(err)
			}
			var decoded kubelet.KubeletConfiguration
			if err := yaml.UnmarshalStrict(data, &decoded); err != nil {
				t.Fatal(err)
			}
			// YAML may omit an empty list; no empty-string resolver may be emitted.
			if len(decoded.ClusterDNS) != len(tt.want) || (len(tt.want) > 0 && !reflect.DeepEqual(decoded.ClusterDNS, tt.want)) {
				t.Errorf("serialized DNS = %v, want ordered %v", decoded.ClusterDNS, tt.want)
			}
		})
	}
}

func TestKubeletClusterDNSVersionedAPI(t *testing.T) {
	for _, version := range []string{"v1alpha2", "v1alpha3"} {
		t.Run(version, func(t *testing.T) {
			const dns = "192.168.0.1,10.4.240.10"
			document := fmt.Sprintf("apiVersion: kops.k8s.io/%s\nkind: Cluster\nmetadata:\n  name: dns.example.com\nspec:\n  kubernetesVersion: 1.36.6\n  kubelet:\n    clusterDNS: %s\n", version, dns)
			obj, _, err := kopscodecs.Decode([]byte(document), nil)
			if err != nil {
				t.Fatal(err)
			}
			cluster := obj.(*kops.Cluster)
			if cluster.Spec.Kubelet.ClusterDNS != dns {
				t.Fatalf("API changed scalar to %q", cluster.Spec.Kubelet.ClusterDNS)
			}
			builder := &KubeletBuilder{NodeupModelContext: &NodeupModelContext{
				NodeupConfig: &nodeup.Config{ContainerdConfig: &kops.ContainerdConfig{}},
			}}
			assertKubeletClusterDNSFile(t, builder, cluster.Spec.Kubelet, []string{"192.168.0.1", "10.4.240.10"})
			listDocument := strings.Replace(document, dns, "[192.168.0.1, 10.4.240.10]", 1)
			if _, _, err := kopscodecs.Decode([]byte(listDocument), nil); err == nil {
				t.Fatal("list clusterDNS unexpectedly accepted; API must remain scalar")
			}
		})
	}
}

func TestKubeletClusterDNSDefaults(t *testing.T) {
	for _, tt := range []struct {
		name      string
		input     string
		absent    bool
		nodeLocal bool
		want      string
	}{
		{name: "absent kubelet service default", absent: true, want: "10.4.240.10"},
		{name: "service default", want: "10.4.240.10"},
		{name: "node local default", nodeLocal: true, want: "169.254.20.10"},
		{name: "explicit CSV", input: "192.168.0.1,10.4.240.10", want: "192.168.0.1,10.4.240.10"},
		{name: "whitespace is not defaulted", input: " ", want: " "},
	} {
		t.Run(tt.name, func(t *testing.T) {
			cluster := &kops.Cluster{Spec: kops.ClusterSpec{
				KubernetesVersion: "1.36.6",
				Kubelet:           &kops.KubeletConfigSpec{ClusterDNS: tt.input},
				Networking:        kops.NetworkingSpec{ServiceClusterIPRange: "10.4.240.0/20"},
			}}
			if tt.absent {
				cluster.Spec.Kubelet = nil
			}
			if tt.nodeLocal {
				cluster.Spec.KubeDNS = &kops.KubeDNSConfig{NodeLocalDNS: &kops.NodeLocalDNSConfig{Enabled: new(true), LocalIP: tt.want}}
			}
			options, err := components.NewOptionsContext(cluster, nil, "")
			if err != nil {
				t.Fatal(err)
			}
			if err := (&components.KubeletOptionsBuilder{OptionsContext: options}).BuildOptions(cluster); err != nil {
				t.Fatal(err)
			}
			if cluster.Spec.Kubelet.ClusterDNS != tt.want {
				t.Fatalf("defaulted DNS = %q, want %q", cluster.Spec.Kubelet.ClusterDNS, tt.want)
			}
			builder := &KubeletBuilder{NodeupModelContext: &NodeupModelContext{
				BootConfig:   &nodeup.BootConfig{InstanceGroupRole: kops.InstanceGroupRoleNode},
				NodeupConfig: &nodeup.Config{KubernetesVersion: "1.36.6", KubeletConfig: *cluster.Spec.Kubelet, ContainerdConfig: &kops.ContainerdConfig{}},
			}}
			if err := builder.Init(); err != nil {
				t.Fatal(err)
			}
			config, err := builder.buildKubeletConfigSpec(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if tt.input == " " {
				if _, err := builder.buildKubeletComponentConfig(config, ""); err == nil || !strings.Contains(err.Error(), "clusterDNS") {
					t.Fatalf("whitespace DNS error = %v", err)
				}
				return
			}
			want := []string{tt.want}
			if tt.input != "" {
				want = []string{"192.168.0.1", "10.4.240.10"}
			}
			assertKubeletClusterDNSFile(t, builder, config, want)
		})
	}
}

func assertKubeletClusterDNSFile(t *testing.T, builder *KubeletBuilder, config *kops.KubeletConfigSpec, want []string) {
	t.Helper()
	cc, err := builder.kubeletConfiguration(config, "")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cc.ClusterDNS, want) {
		t.Errorf("converted DNS = %v, want ordered %v", cc.ClusterDNS, want)
	}
	file, err := builder.buildKubeletComponentConfig(config, "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := fi.ResourceAsBytes(file.Contents)
	if err != nil {
		t.Fatal(err)
	}
	var decoded kubelet.KubeletConfiguration
	if err := yaml.UnmarshalStrict(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded.ClusterDNS, want) {
		t.Errorf("serialized DNS = %v, want ordered %v", decoded.ClusterDNS, want)
	}
}

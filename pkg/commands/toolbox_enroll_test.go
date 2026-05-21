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

package commands

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// TestRewriteChannelsManifestForEnroll pins both the URL rewrite and the addons hostPath mount:
// without either, an enrolled control plane would fail to reach the bootstrap channel.
func TestRewriteChannelsManifestForEnroll(t *testing.T) {
	in := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: kops-channels
spec:
  containers:
  - name: kops-channels
    args:
    - apply
    - channel
    - --yes
    - s3://example/clusters/my-cluster/addons/bootstrap-channel.yaml
    - s3://example/clusters/my-cluster/addons/custom.yaml
`)

	out, err := rewriteChannelsManifestForEnroll(
		in,
		"s3://example/clusters/my-cluster/addons/bootstrap-channel.yaml",
		"/etc/kubernetes/kops/config/addons",
	)
	if err != nil {
		t.Fatalf("rewriteChannelsManifestForEnroll: %v", err)
	}

	s := string(out)
	if strings.Contains(s, "s3://example/clusters/my-cluster/addons/bootstrap-channel.yaml") {
		t.Errorf("bootstrap URL not rewritten:\n%s", s)
	}
	if !strings.Contains(s, "file:///etc/kubernetes/kops/config/addons/bootstrap-channel.yaml") {
		t.Errorf("expected local file:// bootstrap URL in args:\n%s", s)
	}
	if !strings.Contains(s, "s3://example/clusters/my-cluster/addons/custom.yaml") {
		t.Errorf("non-bootstrap channel URL should not be rewritten:\n%s", s)
	}

	pod := &corev1.Pod{}
	if err := yaml.Unmarshal(out, pod); err != nil {
		t.Fatalf("parsing output: %v", err)
	}

	var hostVol *corev1.Volume
	for i := range pod.Spec.Volumes {
		v := &pod.Spec.Volumes[i]
		if v.HostPath != nil && v.HostPath.Path == "/etc/kubernetes/kops/config/addons" {
			hostVol = v
			break
		}
	}
	if hostVol == nil {
		t.Fatalf("expected hostPath volume for /etc/kubernetes/kops/config/addons, got volumes=%+v", pod.Spec.Volumes)
	}
	if hostVol.HostPath.Type == nil || *hostVol.HostPath.Type != corev1.HostPathDirectory {
		t.Errorf("expected HostPathDirectory type, got %+v", hostVol.HostPath.Type)
	}

	var mount *corev1.VolumeMount
	for i := range pod.Spec.Containers[0].VolumeMounts {
		m := &pod.Spec.Containers[0].VolumeMounts[i]
		if m.Name == hostVol.Name {
			mount = m
			break
		}
	}
	if mount == nil {
		t.Fatalf("expected volumeMount for %q in kops-channels container", hostVol.Name)
	}
	if mount.MountPath != "/etc/kubernetes/kops/config/addons" {
		t.Errorf("expected mountPath /etc/kubernetes/kops/config/addons, got %q", mount.MountPath)
	}
}

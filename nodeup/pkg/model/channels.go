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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"

	kopsroot "k8s.io/kops"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/env"
	"k8s.io/kops/util/pkg/vfs/openstackconfig"
)

const (
	channelsManifestPath = "/etc/kubernetes/manifests/kops-channels.manifest"
	channelsKubeconfig   = "/var/lib/kops/kubeconfig"
	channelsInterval     = 60 * time.Second
)

// ChannelsBuilder renders the kops-channels static pod that reconciles addon
// channels against the cluster.
type ChannelsBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &ChannelsBuilder{}

// Build emits the kops-channels static pod manifest on control-plane nodes.
func (b *ChannelsBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}
	if len(b.NodeupConfig.Channels) == 0 {
		klog.V(2).Infof("no channels configured; skipping kops-channels static pod")
		return nil
	}

	// Match runAsUser to the kubeconfig file's owner.
	c.AddTask(&nodetasks.UserTask{
		Name:  wellknownusers.KopsChannelsName,
		UID:   wellknownusers.KopsChannelsID,
		Shell: "/sbin/nologin",
		Home:  "/var/lib/kops",
	})

	kubeconfig := b.BuildIssuedKubeconfig("kops", nodetasks.PKIXName{
		CommonName:   "kops",
		Organization: []string{rbac.SystemPrivilegedGroup},
	}, c)
	c.AddTask(&nodetasks.File{
		Path:     channelsKubeconfig,
		Contents: kubeconfig,
		Type:     nodetasks.FileType_File,
		Mode:     fi.PtrTo("0400"),
		Owner:    fi.PtrTo(wellknownusers.KopsChannelsName),
	})

	pod, err := b.buildPod()
	if err != nil {
		return fmt.Errorf("building kops-channels pod: %w", err)
	}
	manifest, err := k8scodecs.ToVersionedYaml(pod)
	if err != nil {
		return fmt.Errorf("marshaling kops-channels pod: %w", err)
	}
	c.AddTask(&nodetasks.File{
		Path:     channelsManifestPath,
		Contents: fi.NewBytesResource(manifest),
		Type:     nodetasks.FileType_File,
	})
	return nil
}

func (b *ChannelsBuilder) buildPod() (*v1.Pod, error) {
	// TODO: route through AssetBuilder.RemapImage at cloudup so
	// containerRegistry/containerProxy/dev overrides apply.
	image := b.RemapImage("registry.k8s.io/kops/channels:" + kopsroot.KopsVersionImageTag())
	envVars := b.channelsEnvVars()

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kops-channels",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kops-channels",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
			// kops-channels installs CoreDNS via the bootstrap channel, so it
			// cannot itself depend on cluster DNS. Use the host resolver so
			// VFS can reach S3/GCS/HTTPS-backed channel stores at boot.
			DNSPolicy: v1.DNSDefault,
		},
	}

	args := []string{
		"apply", "channel",
		"--v=4",
		"--yes",
		"--interval=" + channelsInterval.String(),
		"--node-name=$(NODE_NAME)",
	}
	args = append(args, b.NodeupConfig.Channels...)

	container := v1.Container{
		Name:  "kops-channels",
		Image: image,
		Args:  args,
		Env: append([]v1.EnvVar{{
			Name: "NODE_NAME",
			ValueFrom: &v1.EnvVarSource{
				FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
			},
		}}, envVars...),
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("50m"),
				v1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
		// ko-distroless defaults to a nonroot uid that can't read /var/lib/kops/kubeconfig.
		SecurityContext: &v1.SecurityContext{
			RunAsUser:    fi.PtrTo(int64(wellknownusers.KopsChannelsID)),
			RunAsNonRoot: fi.PtrTo(true),
		},
	}
	kubemanifest.AddHostPathMapping(pod, &container, "kubeconfig", channelsKubeconfig,
		kubemanifest.WithType(v1.HostPathFile))
	// `kops toolbox enroll` rewrites channel URLs to file:// pointing at
	// /etc/kubernetes/kops/config/addons/ on the host; the container needs
	// that tree mounted to read the channel and its referenced manifests.
	for i, dir := range fileChannelDirs(b.NodeupConfig.Channels) {
		name := fmt.Sprintf("channels-%d", i)
		kubemanifest.AddHostPathMapping(pod, &container, name, dir,
			kubemanifest.WithType(v1.HostPathDirectory))
	}
	pod.Spec.Containers = append(pod.Spec.Containers, container)

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsNodeCritical(pod)
	kubemanifest.AddHostPathSELinuxContext(pod, b.NodeupConfig)

	return pod, nil
}

// fileChannelDirs returns the unique parent directories of any file:// channel
// URLs in channels, preserving input order.
func fileChannelDirs(channels []string) []string {
	seen := map[string]bool{}
	var dirs []string
	for _, c := range channels {
		p, ok := strings.CutPrefix(c, "file://")
		if !ok {
			continue
		}
		dir := filepath.Dir(p)
		if seen[dir] {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}
	return dirs
}

// channelsEnvVars returns the credentials and proxy env vars VFS reads to
// fetch channel manifests. Only S3 (non-AWS backends) and OpenStack
// credentials are wired through; DO/Hetzner/Scaleway use S3-compatible
// config stores so their cloud-API tokens aren't needed. AWS_REGION is a
// non-secret hint kept to skip an IMDS round-trip.
func (b *ChannelsBuilder) channelsEnvVars() []v1.EnvVar {
	var out []v1.EnvVar
	add := func(name, value string) {
		if value != "" {
			out = append(out, v1.EnvVar{Name: name, Value: value})
		}
	}

	add("KUBECONFIG", channelsKubeconfig)
	add("AWS_REGION", os.Getenv("AWS_REGION"))

	if os.Getenv("S3_ENDPOINT") != "" {
		for _, name := range []string{"S3_ENDPOINT", "S3_REGION", "S3_ACCESS_KEY_ID", "S3_SECRET_ACCESS_KEY"} {
			add(name, os.Getenv(name))
		}
	}

	if os.Getenv("OS_AUTH_URL") != "" {
		for _, name := range []string{
			"OS_TENANT_ID", "OS_TENANT_NAME", "OS_PROJECT_ID", "OS_PROJECT_NAME",
			"OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID",
			"OS_DOMAIN_NAME", "OS_DOMAIN_ID",
			"OS_USERNAME", "OS_PASSWORD", "OS_AUTH_URL", "OS_REGION_NAME",
			"OS_APPLICATION_CREDENTIAL_ID", "OS_APPLICATION_CREDENTIAL_SECRET",
			openstackconfig.EnvKeyOpenstackTLSInsecureSkipVerify,
		} {
			add(name, os.Getenv(name))
		}
	}

	for _, ev := range env.GetProxyEnvVars(b.NodeupConfig.Networking.EgressProxy) {
		out = append(out, v1.EnvVar{Name: ev.Name, Value: ev.Value})
	}

	return out
}

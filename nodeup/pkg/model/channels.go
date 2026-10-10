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
	"strings"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	channelsManifestPath   = "/etc/kubernetes/manifests/kops-channels.manifest"
	channelsKubeconfigPath = "/var/lib/kops/kubeconfig"
)

// ChannelsBuilder writes the host-side artifacts the kops-channels static pod needs: the
// kops-channels user, the kubeconfig owned by that user, and the pod manifest copied from
// the state store (built by the cloudup channels builder).
type ChannelsBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &ChannelsBuilder{}

// runsChannels reports whether kops-channels runs on this instance. cloudup decides placement
// and says so by setting the manifest path, so there is one answer rather than a role check here
// that has to be kept in step with it.
func (b *ChannelsBuilder) runsChannels() bool {
	return b.NodeupConfig.ChannelsManifest != ""
}

func (b *ChannelsBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	if !b.runsChannels() {
		return nil
	}

	// Host user that owns the kubeconfig the kops-channels container reads via hostPath.
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
		Path:     channelsKubeconfigPath,
		Contents: kubeconfig,
		Type:     nodetasks.FileType_File,
		Mode:     new("0400"),
		Owner:    new(wellknownusers.KopsChannelsName),
	})

	manifest, err := b.readChannelsManifest(c)
	if err != nil {
		return err
	}
	c.AddTask(&nodetasks.File{
		Path:     channelsManifestPath,
		Contents: fi.NewBytesResource(manifest),
		Type:     nodetasks.FileType_File,
	})
	return nil
}

// readChannelsManifest fetches the cloudup-built manifest and applies node-local SELinux
// decoration when needed. Otherwise it passes the bytes through unchanged.
func (b *ChannelsBuilder) readChannelsManifest(c *fi.NodeupModelBuilderContext) ([]byte, error) {
	ctx := c.Context()
	p, err := vfs.Context.BuildVfsPath(b.NodeupConfig.ChannelsManifest)
	if err != nil {
		return nil, fmt.Errorf("parsing path for kops-channels manifest %s: %w", b.NodeupConfig.ChannelsManifest, err)
	}
	data, err := p.ReadFile(ctx)
	if err != nil {
		return nil, fmt.Errorf("reading kops-channels manifest %s: %w", b.NodeupConfig.ChannelsManifest, err)
	}

	// SELinux is per-IG via containerdConfig, and so are the node labels when the cluster places
	// the scheduled components explicitly; both can only be applied at nodeup. Skip the
	// parse/reserialize round-trip when there is nothing to change.
	selinux := b.NodeupConfig.ContainerdConfig != nil && b.NodeupConfig.ContainerdConfig.SeLinuxEnabled
	if !selinux && len(b.NodeupConfig.ChannelsNodeLabels) == 0 {
		return data, nil
	}
	pod := &v1.Pod{}
	if err := yaml.Unmarshal(data, pod); err != nil {
		return nil, fmt.Errorf("parsing kops-channels manifest: %w", err)
	}
	if selinux {
		kubemanifest.AddHostPathSELinuxContext(pod, b.NodeupConfig)
	}
	if len(b.NodeupConfig.ChannelsNodeLabels) > 0 {
		if err := setChannelsNodeLabels(pod, b.NodeupConfig.ChannelsNodeLabels); err != nil {
			return nil, err
		}
	}
	out, err := k8scodecs.ToVersionedYaml(pod)
	if err != nil {
		return nil, fmt.Errorf("re-marshaling kops-channels manifest: %w", err)
	}
	return out, nil
}

// setChannelsNodeLabels replaces the --node-labels argument in the kops-channels manifest. The
// manifest in the state store is shared by every instance group, carrying the cluster-wide
// default; this narrows it to what this instance group claims.
func setChannelsNodeLabels(pod *v1.Pod, labels []string) error {
	const flag = "--node-labels="

	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		if container.Name != "kops-channels" {
			continue
		}
		for j, arg := range container.Args {
			if strings.HasPrefix(arg, flag) {
				container.Args[j] = flag + strings.Join(labels, ",")
				return nil
			}
		}
		return fmt.Errorf("no %s argument in kops-channels manifest", flag)
	}

	return fmt.Errorf("no kops-channels container in kops-channels manifest")
}

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

package channels

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kopsroot "k8s.io/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/env"
	"k8s.io/kops/util/pkg/vfs"
)

const (
	channelsManifestPath = "manifests/channels/kops-channels.yaml"

	channelsInterval = 60 * time.Second

	// channelsKubeconfigPath is the on-host kubeconfig the kops-channels container reads via hostPath.
	channelsKubeconfigPath = "/var/lib/kops/kubeconfig"
)

// ChannelsBuilder builds the kops-channels static pod manifest at cloudup and writes it to
// the state store. Nodeup copies the manifest onto each control-plane node, mirroring how
// etcd-manager manifests are delivered.
type ChannelsBuilder struct {
	*model.KopsModelContext
	Lifecycle    fi.Lifecycle
	AssetBuilder *assets.AssetBuilder
}

var _ fi.CloudupModelBuilder = &ChannelsBuilder{}

func (b *ChannelsBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	channels, err := b.channelList()
	if err != nil {
		return err
	}

	pod, err := b.buildPod(channels)
	if err != nil {
		return fmt.Errorf("building kops-channels pod: %w", err)
	}
	manifest, err := k8scodecs.ToVersionedYaml(pod)
	if err != nil {
		return fmt.Errorf("marshaling kops-channels pod: %w", err)
	}

	c.AddTask(&fitasks.ManagedFile{
		Contents:  fi.NewBytesResource(manifest),
		Lifecycle: b.Lifecycle,
		Location:  fi.PtrTo(channelsManifestPath),
		Name:      fi.PtrTo("manifests-channels-kops-channels"),
	})
	return nil
}

// channelList returns the bootstrap channel plus cluster.Spec.Addons. The bootstrap URL is
// built via vfs path joining so toolbox_enroll's identity comparison stays byte-identical.
func (b *ChannelsBuilder) channelList() ([]string, error) {
	configBase, err := vfs.Context.BuildVfsPath(b.Cluster.Spec.ConfigStore.Base)
	if err != nil {
		return nil, fmt.Errorf("parsing configStore.base %q: %w", b.Cluster.Spec.ConfigStore.Base, err)
	}
	channels := []string{
		configBase.Join("addons", "bootstrap-channel.yaml").Path(),
	}
	for i := range b.Cluster.Spec.Addons {
		channels = append(channels, b.Cluster.Spec.Addons[i].Manifest)
	}
	return channels, nil
}

func (b *ChannelsBuilder) buildPod(channels []string) (*v1.Pod, error) {
	image := b.AssetBuilder.RemapImage("registry.k8s.io/kops/channels:" + kopsroot.KopsVersionImageTag())

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
			// kops-channels installs CoreDNS via the bootstrap channel, so it can't depend on
			// cluster DNS. Use the host resolver so VFS can reach S3/GCS/HTTPS channel stores at boot.
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
	args = append(args, channels...)

	container := v1.Container{
		Name:  "kops-channels",
		Image: image,
		Args:  args,
		Env: append([]v1.EnvVar{
			{
				Name: "NODE_NAME",
				ValueFrom: &v1.EnvVarSource{
					FieldRef: &v1.ObjectFieldSelector{FieldPath: "spec.nodeName"},
				},
			},
			{
				Name:  "KUBECONFIG",
				Value: channelsKubeconfigPath,
			},
			{
				// client-go's discovery cache writes to $HOME/.kube; /tmp is the writable dir
				// for the non-root uid.
				Name:  "HOME",
				Value: "/tmp",
			},
		}, env.BuildSystemComponentEnvVars(&b.Cluster.Spec).ToEnvVars()...),
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("50m"),
				v1.ResourceMemory: resource.MustParse("50Mi"),
			},
		},
		// ko-distroless's default nonroot uid can't read /var/lib/kops/kubeconfig.
		SecurityContext: &v1.SecurityContext{
			RunAsUser:    fi.PtrTo(int64(wellknownusers.KopsChannelsID)),
			RunAsNonRoot: fi.PtrTo(true),
		},
	}
	kubemanifest.AddHostPathMapping(pod, &container, "kubeconfig", channelsKubeconfigPath,
		kubemanifest.WithType(v1.HostPathFile))
	pod.Spec.Containers = append(pod.Spec.Containers, container)

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsNodeCritical(pod)

	return pod, nil
}

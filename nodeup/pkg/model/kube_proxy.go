/*
Copyright 2017 The Kubernetes Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/architectures"
)

// KubeProxyBuilder installs kube-proxy
type KubeProxyBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeProxyBuilder{}

// Build is responsible for building the kube-proxy manifest
// @TODO we should probably change this to a daemonset in the future and follow the kubeadm path
func (b *KubeProxyBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.Cluster.Spec.KubeProxy.Enabled != nil && !*b.Cluster.Spec.KubeProxy.Enabled {
		klog.V(2).Infof("Kube-proxy is disabled, will not create configuration for it.")
		return nil
	}

	if b.IsMaster {
		// If this is a master that is not isolated, run it as a normal node also (start kube-proxy etc)
		// This lets e.g. daemonset pods communicate with other pods in the system
		if fi.BoolValue(b.Cluster.Spec.IsolateMasters) {
			klog.V(2).Infof("Running on Master with IsolateMaster=true; skipping kube-proxy installation")
			return nil
		}
	}

	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-proxy manifest: %v", err)
		}

		pod.ObjectMeta.Labels["kubernetes.io/managed-by"] = "nodeup"

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshaling manifest to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-proxy.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	{
		var kubeconfig fi.Resource
		var err error

		if b.HasAPIServer {
			kubeconfig = b.BuildIssuedKubeconfig("kube-proxy", nodetasks.PKIXName{CommonName: rbac.KubeProxy}, c)
		} else {
			kubeconfig, err = b.BuildBootstrapKubeconfig("kube-proxy", c)
			if err != nil {
				return err
			}
		}

		c.AddTask(&nodetasks.File{
			Path:           "/var/lib/kube-proxy/kubeconfig",
			Contents:       kubeconfig,
			Type:           nodetasks.FileType_File,
			Mode:           s("0400"),
			BeforeServices: []string{kubeletService},
		})
	}

	{
		c.AddTask(&nodetasks.File{
			Path:        "/var/log/kube-proxy.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		})
	}

	return nil
}

// buildPod is responsible constructing the pod spec
func (b *KubeProxyBuilder) buildPod() (*v1.Pod, error) {
	c := b.Cluster.Spec.KubeProxy
	if c == nil {
		return nil, fmt.Errorf("KubeProxy not configured")
	}

	if c.Master == "" {
		if b.IsMaster {
			// As a special case, if this is the master, we point kube-proxy to the local IP
			// This prevents a circular dependency where kube-proxy can't come up until DNS comes up,
			// which would mean that DNS can't rely on API to come up
			c.Master = "https://127.0.0.1"
		} else {
			c.Master = "https://" + b.Cluster.Spec.MasterInternalName
		}
	}

	resourceRequests := v1.ResourceList{}
	resourceLimits := v1.ResourceList{}

	resourceRequests["cpu"] = *c.CPURequest

	if c.CPULimit != nil {
		resourceLimits["cpu"] = *c.CPULimit
	}

	if c.MemoryRequest != nil {
		resourceRequests["memory"] = *c.MemoryRequest
	}

	if c.MemoryLimit != nil {
		resourceLimits["memory"] = *c.MemoryLimit
	}

	if c.ConntrackMaxPerCore == nil {
		defaultConntrackMaxPerCore := int32(131072)
		c.ConntrackMaxPerCore = &defaultConntrackMaxPerCore
	}

	flags, err := flagbuilder.BuildFlagsList(c)
	if err != nil {
		return nil, fmt.Errorf("error building kubeproxy flags: %v", err)
	}

	flags = append(flags, []string{
		"--kubeconfig=/var/lib/kube-proxy/kubeconfig",
		"--oom-score-adj=-998",
	}...)

	image := c.Image
	if b.Architecture != architectures.ArchitectureAmd64 {
		image = strings.Replace(image, "-amd64", "-"+string(b.Architecture), 1)
	}
	if components.IsBaseURL(b.Cluster.Spec.KubernetesVersion) && b.IsKubernetesLT("1.25") {
		image = strings.Replace(image, "registry.k8s.io", "k8s.gcr.io", 1)
	}

	container := &v1.Container{
		Name:  "kube-proxy",
		Image: image,
		Resources: v1.ResourceRequirements{
			Requests: resourceRequests,
			Limits:   resourceLimits,
		},
		SecurityContext: &v1.SecurityContext{
			Privileged: fi.Bool(true),
		},
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kube-proxy",
				"tier":    "node",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
			Tolerations: tolerateMasterTaints(),
		},
	}

	// Log both to docker and to the logfile
	kubemanifest.AddHostPathMapping(pod, container, "logfile", "/var/log/kube-proxy.log").WithReadWrite()
	// We use lighter containers that don't include shells
	// But they have richer logging support via klog
	if b.IsKubernetesGTE("1.23") {
		container.Command = []string{"/go-runner"}
		container.Args = []string{
			"--log-file=/var/log/kube-proxy.log",
			"--also-stdout",
			"/usr/local/bin/kube-proxy",
		}
		container.Args = append(container.Args, sortedStrings(flags)...)
	} else {
		container.Command = []string{"/usr/local/bin/kube-proxy"}
		container.Args = append(
			sortedStrings(flags),
			"--logtostderr=false", // https://github.com/kubernetes/klog/issues/60
			"--alsologtostderr",
			"--log-file=/var/log/kube-proxy.log")
	}
	{
		kubemanifest.AddHostPathMapping(pod, container, "kubeconfig", "/var/lib/kube-proxy/kubeconfig")
		// @note: mapping the host modules directory to fix the missing ipvs kernel module
		kubemanifest.AddHostPathMapping(pod, container, "modules", "/lib/modules")

		// Map SSL certs from host: /usr/share/ca-certificates -> /etc/ssl/certs
		sslCertsHost := kubemanifest.AddHostPathMapping(pod, container, "ssl-certs-hosts", "/usr/share/ca-certificates")
		sslCertsHost.VolumeMount.MountPath = "/etc/ssl/certs"
	}

	if dns.IsGossipHostname(b.Cluster.Name) {
		// Map /etc/hosts from host, so that we see the updates that are made by protokube
		kubemanifest.AddHostPathMapping(pod, container, "etchosts", "/etc/hosts")
	}

	// Mount the iptables lock file
	{
		kubemanifest.AddHostPathMapping(pod, container, "iptableslock", "/run/xtables.lock").WithReadWrite()

		vol := pod.Spec.Volumes[len(pod.Spec.Volumes)-1]
		if vol.Name != "iptableslock" {
			// Sanity check
			klog.Fatalf("expected volume to be last volume added")
		}
		hostPathType := v1.HostPathFileOrCreate
		vol.HostPath.Type = &hostPathType
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	// Note that e.g. kubeadm has this as a daemonset, but this doesn't have a lot of test coverage AFAICT
	// ServiceAccountName: "kube-proxy",

	//d := &v1beta1.DaemonSet{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Labels: map[string]string{
	//			"k8s-app": "kube-proxy",
	//		},
	//		Name: "kube-proxy",
	//		Namespace: "kube-proxy",
	//	},
	//	Spec: v1beta1.DeploymentSpec{
	//		Selector: &metav1.LabelSelector{
	//			MatchLabels: map[string]string{
	//				"k8s-app": "kube-proxy",
	//			},
	//		},
	//		Template: template,
	//	},
	//}

	// This annotation ensures that kube-proxy does not get evicted if the node
	// supports critical pod annotation based priority scheme.
	// Note that kube-proxy runs as a static pod so this annotation does NOT have
	// any effect on rescheduler (default scheduler and rescheduler are not
	// involved in scheduling kube-proxy).
	kubemanifest.MarkPodAsCritical(pod)

	// Also set priority so that kube-proxy does not get evicted in clusters where
	// PodPriority is enabled.
	kubemanifest.MarkPodAsNodeCritical(pod)

	return pod, nil
}

func tolerateMasterTaints() []v1.Toleration {
	tolerations := []v1.Toleration{}

	// As long as we are a static pod, we don't need any special tolerations
	//	{
	//		Key:    MasterTaintKey,
	//		Effect: NoSchedule,
	//	},
	//}

	return tolerations
}

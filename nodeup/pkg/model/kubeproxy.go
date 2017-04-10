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
	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KubeProxyBuilder installs kube-proxy
type KubeProxyBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeAPIServerBuilder{}

func (b *KubeProxyBuilder) Build(c *fi.ModelBuilderContext) error {
	if b.IsMaster {
		// If this is a master that is not isolated, run it as a normal node also (start kube-proxy etc)
		// This lets e.g. daemonset pods communicate with other pods in the system
		if fi.BoolValue(b.Cluster.Spec.IsolateMasters) {
			glog.V(2).Infof("Running on Master with IsolateMaster=true; skipping kube-proxy installation")
			return nil
		}
	}

	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-proxy manifest: %v", err)
		}

		manifest, err := ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshalling manifest to yaml: %v", err)
		}

		t := &nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-proxy.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add kubeconfig
	{
		kubeconfig, err := b.buildPKIKubeconfig("kube-proxy")
		if err != nil {
			return err
		}
		t := &nodetasks.File{
			Path:     "/var/lib/kube-proxy/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		}
		c.AddTask(t)
	}

	// Touch log file, so that docker doesn't create a directory instead
	{
		t := &nodetasks.File{
			Path:        "/var/log/kube-proxy.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		}
		c.AddTask(t)
	}

	return nil
}

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
			if b.IsKubernetesGTE("1.6") {
				c.Master = "https://127.0.0.1"
			} else {
				c.Master = "http://127.0.0.1:8080"
			}
		} else {
			c.Master = "https://" + b.Cluster.Spec.MasterInternalName
		}
	}

	cpuRequest, err := resource.ParseQuantity(c.CPURequest)
	if err != nil {
		return nil, fmt.Errorf("Error parsing CPURequest=%q", c.CPURequest)
	}

	flags, err := flagbuilder.BuildFlags(c)
	if err != nil {
		return nil, fmt.Errorf("error building kubeproxy flags: %v", err)
	}

	image := c.Image
	cmd := "echo -998 > /proc/$$$/oom_score_adj && kube-proxy --kubeconfig=/var/lib/kube-proxy/kubeconfig --conntrack-max-per-core=131072 --resource-container=\"\" "
	cmd += flags

	// cmd += " 1>>/var/log/kube-proxy.log 2>&1"
	cmd += " 2>&1 | /usr/bin/tee /var/log/kube-proxy.log"

	glog.V(2).Infof("built kube-proxy command: %s", cmd)

	command := []string{
		"/bin/sh", "-c",
		cmd,
	}

	container := &v1.Container{
		Name:            "kube-proxy",
		Image:           image,
		ImagePullPolicy: v1.PullIfNotPresent,
		Command:         command,
		SecurityContext: &v1.SecurityContext{
			Privileged: fi.Bool(true),
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				"cpu": cpuRequest,
			},
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
			Annotations: map[string]string{
				// This annotation ensures that kube-proxy does not get evicted if the node
				// supports critical pod annotation based priority scheme.
				// Note that kube-proxy runs as a static pod so this annotation does NOT have
				// any effect on rescheduler (default scheduler and rescheduler are not
				// involved in scheduling kube-proxy).
				"scheduler.alpha.kubernetes.io/critical-pod": "",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
			Tolerations: tolerateMasterTaints(),
		},
	}

	{
		// Map /var/log to host
		varlog := addHostPathMapping(pod, container, "varlog", "/var/log")
		varlog.ReadOnly = false

		// Map kubeconfig from host
		addHostPathMapping(pod, container, "kubeconfig", "/var/lib/kube-proxy/kubeconfig")

		// Map SSL certs from host: /usr/share/ca-certificates -> /etc/ssl/certs
		sslCertsHost := addHostPathMapping(pod, container, "ssl-certs-hosts", "/usr/share/ca-certificates")
		sslCertsHost.MountPath = "/etc/ssl/certs"
	}

	if dns.IsGossipHostname(b.Cluster.Name) {
		// Map /etc/hosts from host, so that we see the updates that are made by protokube
		addHostPathMapping(pod, container, "etchosts", "/etc/hosts")
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	// Note that e.g. kubeadm has this as a daemonset, but this doesn't have a lot of test coverage AFAICT
	//ServiceAccountName: "kube-proxy",

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

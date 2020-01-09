/*
Copyright 2019 The Kubernetes Authors.

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
	"bytes"
	"fmt"
	"strconv"

	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/exec"
	"k8s.io/kops/util/pkg/proxy"

	"k8s.io/apimachinery/pkg/runtime"
	config "k8s.io/kube-scheduler/config/v1alpha1"
	"k8s.io/kubernetes/pkg/scheduler/apis/config/scheme"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// KubeSchedulerBuilder install kube-scheduler
type KubeSchedulerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeSchedulerBuilder{}

// Build is responsible for building the manifest for the kube-scheduler
func (b *KubeSchedulerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-scheduler pod: %v", err)
		}

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshaling pod to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-scheduler.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	{
		kubeconfig, err := b.BuildPKIKubeconfig("kube-scheduler")
		if err != nil {
			return err
		}

		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kube-scheduler/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})
	}

	conf := b.Cluster.Spec.KubeScheduler.Config
	if conf != nil {
		if conf.ClientConnection.Kubeconfig == "" {
			conf.ClientConnection.Kubeconfig = "/var/lib/kube-scheduler/kubeconfig"
		}

		confB, err := b.encodeConfig(conf)
		if err != nil {
			return fmt.Errorf("error marshaling configuration to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/var/lib/kube-scheduler/schedulerconfig",
			Contents: fi.NewBytesResource(confB),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		})
	}

	{
		c.AddTask(&nodetasks.File{
			Path:        "/var/log/kube-scheduler.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		})
	}

	return nil
}

func (b *KubeSchedulerBuilder) encodeConfig(cfg *config.KubeSchedulerConfiguration) ([]byte, error) {
	const mediaType = runtime.ContentTypeYAML
	info, ok := runtime.SerializerInfoForMediaType(scheme.Codecs.SupportedMediaTypes(), mediaType)
	if !ok {
		return nil, fmt.Errorf("unable to locate encoder -- %q is not a supported media type", mediaType)
	}

	encoder := scheme.Codecs.EncoderForVersion(info.Serializer, config.SchemeGroupVersion)

	var w bytes.Buffer
	if err := encoder.Encode(cfg, &w); err != nil {
		return nil, err
	}

	return w.Bytes(), nil
}

// buildPod is responsible for constructing the pod specification
func (b *KubeSchedulerBuilder) buildPod() (*v1.Pod, error) {
	c := b.Cluster.Spec.KubeScheduler

	flags, err := flagbuilder.BuildFlagsList(c)
	if err != nil {
		return nil, fmt.Errorf("error building kube-scheduler flags: %v", err)
	}
	// Add kubeconfig flag
	flags = append(flags, "--kubeconfig="+"/var/lib/kube-scheduler/kubeconfig")

	if c.UsePolicyConfigMap != nil {
		flags = append(flags, "--policy-configmap=scheduler-policy --policy-configmap-namespace=kube-system")
	}

	if c.Config != nil {
		flags = append(flags, "--config=/var/lib/kube-scheduler/schedulerconfig")
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-scheduler",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kube-scheduler",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
		},
	}

	container := &v1.Container{
		Name:  "kube-scheduler",
		Image: c.Image,
		Env:   proxy.GetProxyEnvVars(b.Cluster.Spec.EgressProxy),
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Host: "127.0.0.1",
					Path: "/healthz",
					Port: intstr.FromInt(10251),
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		},
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("100m"),
			},
		},
	}
	addHostPathMapping(pod, container, "varlibkubescheduler", "/var/lib/kube-scheduler")

	// Log both to docker and to the logfile
	addHostPathMapping(pod, container, "logfile", "/var/log/kube-scheduler.log").ReadOnly = false
	if b.IsKubernetesGTE("1.15") {
		// From k8s 1.15, we use lighter containers that don't include shells
		// But they have richer logging support via klog
		container.Command = []string{"/usr/local/bin/kube-scheduler"}
		container.Args = append(
			sortedStrings(flags),
			"--logtostderr=false", //https://github.com/kubernetes/klog/issues/60
			"--alsologtostderr",
			"--log-file=/var/log/kube-scheduler.log")
	} else {
		container.Command = exec.WithTee(
			"/usr/local/bin/kube-scheduler",
			sortedStrings(flags),
			"/var/log/kube-scheduler.log")
	}

	if c.MaxPersistentVolumes != nil {
		maxPDV := v1.EnvVar{
			Name:  "KUBE_MAX_PD_VOLS", // https://kubernetes.io/docs/concepts/storage/storage-limits/
			Value: strconv.Itoa(int(*c.MaxPersistentVolumes)),
		}
		container.Env = append(container.Env, maxPDV)
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsClusterCritical(pod)

	return pod, nil
}

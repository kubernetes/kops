/*
Copyright 2016 The Kubernetes Authors.

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
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/v1"
	metav1 "k8s.io/kubernetes/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/intstr"
	"strings"
)

// KubeControllerManagerBuilder install kube-controller-manager (just the manifest at the moment)
type KubeControllerManagerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeControllerManagerBuilder{}

func (b *KubeControllerManagerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-controller-manager pod: %v", err)
		}

		manifest, err := ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshalling pod to yaml: %v", err)
		}

		t := &nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-controller-manager.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	return nil
}

func (b *KubeControllerManagerBuilder) buildPod() (*v1.Pod, error) {
	flags, err := flagbuilder.BuildFlags(b.Cluster.Spec.KubeControllerManager)
	if err != nil {
		return nil, fmt.Errorf("error building kube-controller-manager flags: %v", err)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		flags += " --cloud-config=" + CloudConfigFilePath
	}

	redirectCommand := []string{
		"/bin/sh", "-c", "/usr/local/bin/kube-controller-manager " + flags + " 1>>/var/log/kube-controller-manager.log 2>&1",
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      "kube-controller-manager",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kube-controller-manager",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
		},
	}

	container := &v1.Container{
		Name:  "kube-controller-manager",
		Image: b.Cluster.Spec.KubeControllerManager.Image,
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("100m"),
			},
		},
		Command: redirectCommand,
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: &v1.HTTPGetAction{
					Host: "127.0.0.1",
					Path: "/healthz",
					Port: intstr.FromInt(10252),
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		},
	}

	for _, path := range b.SSLHostPaths() {
		name := strings.Replace(path, "/", "", -1)

		addHostPathMapping(pod, container, name, path, true)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		addHostPathMapping(pod, container, "cloudconfig", CloudConfigFilePath, true)
	}

	if b.Cluster.Spec.KubeControllerManager.PathSrvKubernetes != "" {
		addHostPathMapping(pod, container, "srvkube", b.Cluster.Spec.KubeControllerManager.PathSrvKubernetes, true)
	}

	addHostPathMapping(pod, container, "logfile", "/var/log/kube-controller-manager.log", false)

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	return pod, nil
}

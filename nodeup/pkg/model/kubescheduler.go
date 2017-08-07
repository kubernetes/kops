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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// KubeSchedulerBuilder install kube-scheduler
type KubeSchedulerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeSchedulerBuilder{}

func (b *KubeSchedulerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-scheduler pod: %v", err)
		}

		manifest, err := ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshalling pod to yaml: %v", err)
		}

		t := &nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-scheduler.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Add kubeconfig
	{
		kubeconfig, err := b.buildPKIKubeconfig("kube-scheduler")
		if err != nil {
			return err
		}
		t := &nodetasks.File{
			Path:     "/var/lib/kube-scheduler/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		}
		c.AddTask(t)
	}

	// Touch log file, so that docker doesn't create a directory instead
	{
		t := &nodetasks.File{
			Path:        "/var/log/kube-scheduler.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		}
		c.AddTask(t)
	}

	return nil
}

func (b *KubeSchedulerBuilder) buildPod() (*v1.Pod, error) {
	c := b.Cluster.Spec.KubeScheduler

	flags, err := flagbuilder.BuildFlags(c)
	if err != nil {
		return nil, fmt.Errorf("error building kube-scheduler flags: %v", err)
	}

	// Add kubeconfig flag
	flags += " --kubeconfig=" + "/var/lib/kube-scheduler/kubeconfig"

	redirectCommand := []string{
		"/bin/sh", "-c", "/usr/local/bin/kube-scheduler " + flags + " 1>>/var/log/kube-scheduler.log 2>&1",
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
					Port: intstr.FromInt(10251),
				},
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		},
		Env: getProxyEnvVars(b.Cluster.Spec.EgressProxy),
	}

	addHostPathMapping(pod, container, "varlibkubescheduler", "/var/lib/kube-scheduler")

	addHostPathMapping(pod, container, "logfile", "/var/log/kube-scheduler.log").ReadOnly = false

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	return pod, nil
}

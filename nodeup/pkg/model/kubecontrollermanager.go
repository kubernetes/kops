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
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
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

	// If we're using the CertificateSigner, include the CA Key
	// TODO: use a per-machine key?  use KMS?
	if b.useCertificateSigner() {
		ca, err := b.KeyStore.PrivateKey(fi.CertificateId_CA)
		if err != nil {
			return err
		}

		serialized, err := ca.AsString()
		if err != nil {
			return err
		}

		t := &nodetasks.File{
			Path:     filepath.Join(b.PathSrvKubernetes(), "ca.key"),
			Contents: fi.NewStringResource(serialized),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
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

	// Add kubeconfig
	{
		// TODO: Change kubeconfig to be https

		kubeconfig, err := b.buildPKIKubeconfig("kube-controller-manager")
		if err != nil {
			return err
		}
		t := &nodetasks.File{
			Path:     "/var/lib/kube-controller-manager/kubeconfig",
			Contents: fi.NewStringResource(kubeconfig),
			Type:     nodetasks.FileType_File,
			Mode:     s("0400"),
		}
		c.AddTask(t)
	}

	// Touch log file, so that docker doesn't create a directory instead
	{
		t := &nodetasks.File{
			Path:        "/var/log/kube-controller-manager.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		}
		c.AddTask(t)
	}

	return nil
}

func (b *KubeControllerManagerBuilder) useCertificateSigner() bool {
	// For now, we enable this on 1.6 and later
	return b.IsKubernetesGTE("1.6")
}

func (b *KubeControllerManagerBuilder) buildPod() (*v1.Pod, error) {
	kcm := b.Cluster.Spec.KubeControllerManager

	kcm.RootCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")

	kcm.ServiceAccountPrivateKeyFile = filepath.Join(b.PathSrvKubernetes(), "server.key")

	flags, err := flagbuilder.BuildFlags(kcm)
	if err != nil {
		return nil, fmt.Errorf("error building kube-controller-manager flags: %v", err)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		flags += " --cloud-config=" + CloudConfigFilePath
	}

	// Add kubeconfig flag
	flags += " --kubeconfig=" + "/var/lib/kube-controller-manager/kubeconfig"

	// Configure CA certificate to be used to sign keys, if we are using CSRs
	if b.useCertificateSigner() {
		flags += " --cluster-signing-cert-file=" + filepath.Join(b.PathSrvKubernetes(), "ca.crt")
		flags += " --cluster-signing-key-file=" + filepath.Join(b.PathSrvKubernetes(), "ca.key")
	}

	redirectCommand := []string{
		"/bin/sh", "-c", "/usr/local/bin/kube-controller-manager " + flags + " 1>>/var/log/kube-controller-manager.log 2>&1",
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
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
		Env: getProxyEnvVars(b.Cluster.Spec.EgressProxy),
	}

	for _, path := range b.SSLHostPaths() {
		name := strings.Replace(path, "/", "", -1)

		addHostPathMapping(pod, container, name, path)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		addHostPathMapping(pod, container, "cloudconfig", CloudConfigFilePath)
	}

	pathSrvKubernetes := b.PathSrvKubernetes()
	if pathSrvKubernetes != "" {
		addHostPathMapping(pod, container, "srvkube", pathSrvKubernetes)
	}

	addHostPathMapping(pod, container, "logfile", "/var/log/kube-controller-manager.log").ReadOnly = false
	addHostPathMapping(pod, container, "varlibkcm", "/var/lib/kube-controller-manager")

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	return pod, nil
}

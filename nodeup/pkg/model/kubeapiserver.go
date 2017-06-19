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
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

const PathAuthnConfig = "/etc/kubernetes/authn.config"

// KubeAPIServerBuilder install kube-apiserver (just the manifest at the moment)
type KubeAPIServerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeAPIServerBuilder{}

// Build is responsible for generating the kubernetes api manifest
func (b *KubeAPIServerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	err := b.writeAuthenticationConfig(c)
	if err != nil {
		return err
	}

	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-apiserver manifest: %v", err)
		}

		manifest, err := ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshalling manifest to yaml: %v", err)
		}

		t := &nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-apiserver.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)
	}

	// Touch log file, so that docker doesn't create a directory instead
	{
		t := &nodetasks.File{
			Path:        "/var/log/kube-apiserver.log",
			Contents:    fi.NewStringResource(""),
			Type:        nodetasks.FileType_File,
			Mode:        s("0400"),
			IfNotExists: true,
		}
		c.AddTask(t)
	}

	return nil
}

func (b *KubeAPIServerBuilder) writeAuthenticationConfig(c *fi.ModelBuilderContext) error {
	if b.Cluster.Spec.Authentication == nil || b.Cluster.Spec.Authentication.IsEmpty() {
		return nil
	}

	if b.Cluster.Spec.Authentication.Kopeio != nil {
		cluster := kubeconfig.KubectlCluster{
			Server: "http://127.0.0.1:9001/hooks/authn",
		}
		context := kubeconfig.KubectlContext{
			Cluster: "webhook",
			User:    "kube-apiserver",
		}

		config := kubeconfig.KubectlConfig{
			Kind:       "Config",
			ApiVersion: "v1",
		}
		config.Clusters = append(config.Clusters, &kubeconfig.KubectlClusterWithName{
			Name:    "webhook",
			Cluster: cluster,
		})
		config.Users = append(config.Users, &kubeconfig.KubectlUserWithName{
			Name: "kube-apiserver",
		})
		config.CurrentContext = "webhook"
		config.Contexts = append(config.Contexts, &kubeconfig.KubectlContextWithName{
			Name:    "webhook",
			Context: context,
		})

		manifest, err := kops.ToRawYaml(config)
		if err != nil {
			return fmt.Errorf("error marshalling authentication config to yaml: %v", err)
		}

		t := &nodetasks.File{
			Path:     PathAuthnConfig,
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		}
		c.AddTask(t)

		return nil
	}

	return fmt.Errorf("Unrecognized authentication config %v", b.Cluster.Spec.Authentication)
}

func (b *KubeAPIServerBuilder) buildPod() (*v1.Pod, error) {
	kubeAPIServer := b.Cluster.Spec.KubeAPIServer
	kubeAPIServer.ClientCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")
	kubeAPIServer.TLSCertFile = filepath.Join(b.PathSrvKubernetes(), "server.cert")
	kubeAPIServer.TLSPrivateKeyFile = filepath.Join(b.PathSrvKubernetes(), "server.key")
	kubeAPIServer.BasicAuthFile = filepath.Join(b.PathSrvKubernetes(), "basic_auth.csv")
	kubeAPIServer.TokenAuthFile = filepath.Join(b.PathSrvKubernetes(), "known_tokens.csv")

	if b.UseEtcdTLS() {
		kubeAPIServer.EtcdCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")
		kubeAPIServer.EtcdCertFile = filepath.Join(b.PathSrvKubernetes(), "etcd-client.pem")
		kubeAPIServer.EtcdKeyFile = filepath.Join(b.PathSrvKubernetes(), "etcd-client-key.pem")
		kubeAPIServer.EtcdServers = []string{"https://127.0.0.1:4001"}
		kubeAPIServer.EtcdServersOverrides = []string{"/events#https://127.0.0.1:4002"}
	}

	flags, err := flagbuilder.BuildFlags(b.Cluster.Spec.KubeAPIServer)
	if err != nil {
		return nil, fmt.Errorf("error building kube-apiserver flags: %v", err)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		flags += " --cloud-config=" + CloudConfigFilePath
	}

	redirectCommand := []string{
		"/bin/sh", "-c", "/usr/local/bin/kube-apiserver " + flags + " 1>>/var/log/kube-apiserver.log 2>&1",
	}

	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "kube-apiserver",
			Namespace:   "kube-system",
			Annotations: b.buildAnnotations(),
			Labels: map[string]string{
				"k8s-app": "kube-apiserver",
			},
		},
		Spec: v1.PodSpec{
			HostNetwork: true,
		},
	}

	probeAction := &v1.HTTPGetAction{
		Host: "127.0.0.1",
		Path: "/healthz",
		Port: intstr.FromInt(8080),
	}
	if kubeAPIServer.InsecurePort != 0 {
		probeAction.Port = intstr.FromInt(int(kubeAPIServer.InsecurePort))
	} else if kubeAPIServer.SecurePort != 0 {
		probeAction.Port = intstr.FromInt(int(kubeAPIServer.SecurePort))
		probeAction.Scheme = v1.URISchemeHTTPS
	}

	container := &v1.Container{
		Name:  "kube-apiserver",
		Image: b.Cluster.Spec.KubeAPIServer.Image,
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("150m"),
			},
		},
		Command: redirectCommand,
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: probeAction,
			},
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "https",
				ContainerPort: b.Cluster.Spec.KubeAPIServer.SecurePort,
				HostPort:      b.Cluster.Spec.KubeAPIServer.SecurePort,
			},
			{
				Name:          "local",
				ContainerPort: 8080,
				HostPort:      8080,
			},
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

	pathSrvSshproxy := b.PathSrvSshproxy()
	if pathSrvSshproxy != "" {
		addHostPathMapping(pod, container, "srvsshproxy", pathSrvSshproxy)
	}

	addHostPathMapping(pod, container, "logfile", "/var/log/kube-apiserver.log").ReadOnly = false

	auditLogPath := b.Cluster.Spec.KubeAPIServer.AuditLogPath
	if auditLogPath != nil {
		// Mount the directory of the path instead, as kube-apiserver rotates the log by renaming the file.
		// Renaming is not possible when the file is mounted as the host path, and will return a
		// 'Device or resource busy' error
		auditLogPathDir := filepath.Dir(*auditLogPath)
		addHostPathMapping(pod, container, "auditlogpathdir", auditLogPathDir).ReadOnly = false
	}

	if b.Cluster.Spec.Authentication != nil {
		if b.Cluster.Spec.Authentication.Kopeio != nil {
			addHostPathMapping(pod, container, "authn-config", PathAuthnConfig)
		}
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	return pod, nil
}

func addHostPathMapping(pod *v1.Pod, container *v1.Container, name string, path string) *v1.VolumeMount {
	pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
		Name: name,
		VolumeSource: v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: path,
			},
		},
	})

	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:      name,
		MountPath: path,
		ReadOnly:  true,
	})

	return &container.VolumeMounts[len(container.VolumeMounts)-1]
}

func (b *KubeAPIServerBuilder) buildAnnotations() map[string]string {
	annotations := make(map[string]string)
	annotations["dns.alpha.kubernetes.io/internal"] = b.Cluster.Spec.MasterInternalName
	if b.Cluster.Spec.API != nil && b.Cluster.Spec.API.DNS != nil {
		annotations["dns.alpha.kubernetes.io/external"] = b.Cluster.Spec.MasterPublicName
	}
	return annotations
}

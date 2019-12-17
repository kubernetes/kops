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

package protokube

import (
	"fmt"
	"path/filepath"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/util/pkg/exec"
)

// BuildEtcdManifest creates the pod spec, based on the etcd cluster
func BuildEtcdManifest(c *EtcdCluster) *v1.Pod {

	pod := &v1.Pod{}
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	pod.Name = c.PodName
	pod.Namespace = "kube-system"
	pod.Labels = map[string]string{"k8s-app": c.PodName}
	pod.Spec.HostNetwork = true

	// dereference our resource requests if they exist
	// cpu
	var cpuRequest resource.Quantity
	if c.CPURequest != nil {
		cpuRequest = *c.CPURequest
	}

	// memory
	var memoryRequest resource.Quantity
	if c.MemoryRequest != nil {
		memoryRequest = *c.MemoryRequest
	}

	{
		container := v1.Container{
			Name:  "etcd-container",
			Image: c.ImageSource,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    cpuRequest,
					v1.ResourceMemory: memoryRequest,
				},
			},
			Command: exec.WithTee("/usr/local/bin/etcd", []string{}, "/var/log/etcd.log"),
		}
		// build the environment variables for etcd service
		container.Env = buildEtcdEnvironmentOptions(c)

		container.LivenessProbe = &v1.Probe{
			InitialDelaySeconds: 15,
			TimeoutSeconds:      15,
		}
		// ensure we have the correct probe schema
		if c.isTLS() {
			container.LivenessProbe.TCPSocket = &v1.TCPSocketAction{
				Host: "127.0.0.1",
				Port: intstr.FromInt(c.ClientPort),
			}
		} else {
			container.LivenessProbe.HTTPGet = &v1.HTTPGetAction{
				Host:   "127.0.0.1",
				Port:   intstr.FromInt(c.ClientPort),
				Path:   "/health",
				Scheme: v1.URISchemeHTTP,
			}
		}
		container.Ports = append(container.Ports, v1.ContainerPort{
			Name:          "serverport",
			ContainerPort: int32(c.PeerPort),
			HostPort:      int32(c.PeerPort),
		})
		container.Ports = append(container.Ports, v1.ContainerPort{
			Name:          "clientport",
			ContainerPort: int32(c.ClientPort),
			HostPort:      int32(c.ClientPort),
		})
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "varetcdata",
			MountPath: "/var/etcd/" + c.DataDirName,
			ReadOnly:  false,
		})
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "varlogetcd",
			MountPath: "/var/log/etcd.log",
			ReadOnly:  false,
		})
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "hosts",
			MountPath: "/etc/hosts",
			ReadOnly:  true,
		})
		// add the host path mount to the pod spec
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "varetcdata",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: c.VolumeMountPath + "/var/etcd/" + c.DataDirName,
				},
			},
		})
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "varlogetcd",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: c.LogFile,
				},
			},
		})
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "hosts",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: "/etc/hosts",
				},
			},
		})
		// @check if tls is enabled and mount the directory. It might be worth considering
		// if we you use our own directory in /srv i.e /srv/etcd rather than the default /src/kubernetes
		if c.isTLS() {
			for _, dirname := range buildCertificateDirectories(c) {
				normalized := strings.Replace(dirname, "/", "", -1)
				pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
					Name: normalized,
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: dirname,
						},
					},
				})
				container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
					Name:      normalized,
					MountPath: dirname,
					ReadOnly:  true,
				})
			}
		}

		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}

	if c.BackupStore != "" && c.BackupImage != "" {
		backupContainer := buildEtcdBackupManagerContainer(c)
		pod.Spec.Containers = append(pod.Spec.Containers, *backupContainer)
	}

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsClusterCritical(pod)

	return pod
}

// buildEtcdEnvironmentOptions is responsible for building the environment variables for etcd
// @question should we perhaps make this version specific in prep for v3 support?
func buildEtcdEnvironmentOptions(c *EtcdCluster) []v1.EnvVar {
	var options []v1.EnvVar

	// @check if we are using TLS
	scheme := "http"
	if c.isTLS() {
		scheme = "https"
	}

	// add the default setting for masters - http or https
	options = append(options, []v1.EnvVar{
		{Name: "ETCD_NAME", Value: c.Me.Name},
		{Name: "ETCD_DATA_DIR", Value: "/var/etcd/" + c.DataDirName},
		{Name: "ETCD_LISTEN_PEER_URLS", Value: fmt.Sprintf("%s://0.0.0.0:%d", scheme, c.PeerPort)},
		{Name: "ETCD_LISTEN_CLIENT_URLS", Value: fmt.Sprintf("%s://0.0.0.0:%d", scheme, c.ClientPort)},
		{Name: "ETCD_ADVERTISE_CLIENT_URLS", Value: fmt.Sprintf("%s://%s:%d", scheme, c.Me.InternalName, c.ClientPort)},
		{Name: "ETCD_INITIAL_ADVERTISE_PEER_URLS", Value: fmt.Sprintf("%s://%s:%d", scheme, c.Me.InternalName, c.PeerPort)},
		{Name: "ETCD_INITIAL_CLUSTER_STATE", Value: "new"},
		{Name: "ETCD_INITIAL_CLUSTER_TOKEN", Value: c.ClusterToken}}...)

	// add timeout/hearbeat settings
	if notEmpty(c.ElectionTimeout) {
		options = append(options, v1.EnvVar{Name: "ETCD_ELECTION_TIMEOUT", Value: c.ElectionTimeout})
	}
	if notEmpty(c.HeartbeatInterval) {
		options = append(options, v1.EnvVar{Name: "ETCD_HEARTBEAT_INTERVAL", Value: c.HeartbeatInterval})
	}

	// @check if we are using peer certificates
	if notEmpty(c.PeerCA) {
		options = append(options, []v1.EnvVar{
			{Name: "ETCD_PEER_TRUSTED_CA_FILE", Value: c.PeerCA}}...)
	}
	if notEmpty(c.PeerCert) {
		options = append(options, v1.EnvVar{Name: "ETCD_PEER_CERT_FILE", Value: c.PeerCert})
	}
	if notEmpty(c.PeerKey) {
		options = append(options, v1.EnvVar{Name: "ETCD_PEER_KEY_FILE", Value: c.PeerKey})
	}
	if notEmpty(c.TLSCA) {
		options = append(options, v1.EnvVar{Name: "ETCD_TRUSTED_CA_FILE", Value: c.TLSCA})
	}
	if notEmpty(c.TLSCert) {
		options = append(options, v1.EnvVar{Name: "ETCD_CERT_FILE", Value: c.TLSCert})
	}
	if notEmpty(c.TLSKey) {
		options = append(options, v1.EnvVar{Name: "ETCD_KEY_FILE", Value: c.TLSKey})
	}
	if c.isTLS() {
		if c.TLSAuth {
			options = append(options, v1.EnvVar{Name: "ETCD_CLIENT_CERT_AUTH", Value: "true"})
			options = append(options, v1.EnvVar{Name: "ETCD_PEER_CLIENT_CERT_AUTH", Value: "true"})
		}
	}

	// @step: generate the initial cluster
	var hosts []string
	for _, node := range c.Nodes {
		hosts = append(hosts, node.Name+"="+fmt.Sprintf("%s://%s:%d", scheme, node.InternalName, c.PeerPort))
	}
	options = append(options, v1.EnvVar{Name: "ETCD_INITIAL_CLUSTER", Value: strings.Join(hosts, ",")})

	return options
}

// buildCertificateDirectories generates a list of the base directories which the certificates are located
// so we can map in as volumes. They will probably all be placed into /src/kubernetes, but just to make it
// generic.
func buildCertificateDirectories(c *EtcdCluster) []string {
	tracked := make(map[string]bool)

	for _, x := range []string{c.TLSCA, c.TLSCert, c.TLSKey, c.PeerCA, c.PeerKey, c.PeerKey} {
		if x == "" || tracked[filepath.Dir(x)] {
			continue
		}
		tracked[filepath.Dir(x)] = true
	}

	var list []string
	for k := range tracked {
		list = append(list, k)
	}

	return list
}

// notEmpty is just a code pretty version if string != ""
func notEmpty(v string) bool {
	return v != ""
}

// buildEtcdBackupManagerContainer builds a container for the standalone etcd backup manager
func buildEtcdBackupManagerContainer(c *EtcdCluster) *v1.Container {
	command := []string{"/etcd-backup"}
	command = append(command, "--backup-store", c.BackupStore)
	command = append(command, "--cluster-name", c.ClusterName)
	command = append(command, "--data-dir", "/var/etcd/"+c.DataDirName)

	if c.isTLS() {
		command = append(command, "--client-url", "https://127.0.0.1:4001")
		command = append(command, "--client-ca-file", c.TLSCA)
		command = append(command, "--client-cert-file", c.TLSCert)
		command = append(command, "--client-key-file", c.TLSKey)
	}

	container := v1.Container{
		Name:    "etcd-backup",
		Image:   c.BackupImage,
		Command: command,
	}

	// TODO: TLS options
	// TODO: Liveness probe?

	// volume should already have been registered
	container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
		Name:      "varetcdata",
		MountPath: "/var/etcd/" + c.DataDirName,
		ReadOnly:  false,
	})

	if c.isTLS() {
		for _, dirname := range buildCertificateDirectories(c) {
			normalized := strings.Replace(dirname, "/", "", -1)

			// pod volume already registered for etcd container above
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				Name:      normalized,
				MountPath: dirname,
				ReadOnly:  true,
			})
		}
	}

	return &container
}

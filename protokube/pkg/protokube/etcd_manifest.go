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

package protokube

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/pkg/api/v1"
	"sort"
	"strings"
)

const DefaultEtcdImageV3 = "3.0.17"

// BuildEtcdManifest creates the pod spec, based on the etcd cluster
func BuildEtcdManifest(c *EtcdCluster) *v1.Pod {
	migrate := "if [ -e /usr/local/bin/migrate-if-needed.sh ]; then /usr/local/bin/migrate-if-needed.sh 1>>/var/log/etcd.log 2>&1; fi; "

	etcdVersion := c.Spec.EtcdVersion
	if etcdVersion == "" {
		// For backwards compatibility
		etcdVersion = "2.2.1"
		migrate = ""
	}

	image := "gcr.io/google_containers/etcd:" + etcdVersion
	if etcdVersion == "2.2.1" {
		// Even with V2, we still use the V3 image, because the V3 image embeds etcd2 and can downgrade to it
		image = "gcr.io/google_containers/etcd:" + DefaultEtcdImageV3
	}

	etcdCommand := "/usr/local/bin/etcd"

	peerProtocol := "http"

	keystorePath := "/etc/kubernetes/ssl/etcd"
	if c.Spec.UseSSL {
		peerProtocol = "https"

		etcdCommand += " --peer-trusted-ca-file /etc/kubernetes/ssl/etcd/etcd-ca.crt"
		etcdCommand += " --peer-cert-file /etc/kubernetes/ssl/etcd/etcd-peer.crt"
		etcdCommand += " --peer-key-file /etc/kubernetes/ssl/etcd/etcd-peer.key"
		etcdCommand += " --peer-client-cert-auth"
	}

	env := make(map[string]string)

	env["ETCD_NAME"] = c.Me.Name
	env["ETCD_DATA_DIR"] = "/var/etcd/" + c.DataDirName

	// Note that we listen on 0.0.0.0, not 127.0.0.1, so we can support etcd clusters
	env["ETCD_LISTEN_PEER_URLS"] = fmt.Sprintf("%s://0.0.0.0:%d", peerProtocol, c.PeerPort)
	env["ETCD_INITIAL_ADVERTISE_PEER_URLS"] = fmt.Sprintf("%s://%s:%d", peerProtocol, c.Me.InternalName, c.PeerPort)
	env["ETCD_INITIAL_CLUSTER_TOKEN"] = c.ClusterToken

	if c.Spec.LockdownClient {
		env["ETCD_LISTEN_CLIENT_URLS"] = fmt.Sprintf("http://127.0.0.1:%d", c.ClientPort)
		env["ETCD_ADVERTISE_CLIENT_URLS"] = fmt.Sprintf("http://127.0.0.1:%d", c.ClientPort)
	} else {
		env["ETCD_LISTEN_CLIENT_URLS"] = fmt.Sprintf("http://0.0.0.0:%d", c.ClientPort)
		env["ETCD_ADVERTISE_CLIENT_URLS"] = fmt.Sprintf("http://%s:%d", c.Me.InternalName, c.ClientPort)
	}

	storageBackend := c.Spec.StorageBackend
	if storageBackend == "" {
		// For backwards compatibility
		storageBackend = "etcd2"
	}
	env["TARGET_STORAGE"] = storageBackend
	if storageBackend == "etcd3" {
		etcdCommand += " --quota-backend-bytes=4294967296"
	}

	// For upgrade
	env["TARGET_VERSION"] = etcdVersion
	env["DATA_DIRECTORY"] = "/var/etcd/" + c.DataDirName

	// TODO: tee or similar, so we can see it in kubectl logs
	etcdCommand += " 1>>/var/log/etcd.log 2>&1"

	var initialCluster []string
	for _, node := range c.Nodes {
		// TODO: Use localhost for ourselves?  Does the cluster view have to be symmetric?
		initialCluster = append(initialCluster, node.Name+"="+fmt.Sprintf("%s://%s:%d", peerProtocol, node.InternalName, c.PeerPort))
	}
	env["ETCD_INITIAL_CLUSTER"] = strings.Join(initialCluster, ",")
	if c.Spec.JoinExistingCluster {
		env["ETCD_INITIAL_CLUSTER_STATE"] = "existing"
	} else {
		env["ETCD_INITIAL_CLUSTER_STATE"] = "new"
	}

	pod := &v1.Pod{}
	pod.APIVersion = "v1"
	pod.Kind = "Pod"
	pod.Name = c.PodName
	pod.Namespace = "kube-system"

	pod.Labels = map[string]string{
		"k8s-app": c.PodName,
	}

	pod.Spec.HostNetwork = true

	{
		container := v1.Container{
			Name:  "etcd-container",
			Image: image,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: c.CPURequest,
				},
			},
			Command: []string{
				"/bin/sh",
				"-c",
				migrate + etcdCommand,
			},
		}

		{
			var keys []string
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				container.Env = append(container.Env, v1.EnvVar{Name: k, Value: env[k]})
			}
		}

		container.LivenessProbe = &v1.Probe{
			// kube-up has 15 seconds, but I still think this is risky
			// Tracking as https://github.com/kubernetes/kubernetes/issues/43362
			InitialDelaySeconds: 300,
			TimeoutSeconds:      15,
		}
		container.LivenessProbe.HTTPGet = &v1.HTTPGetAction{
			Host: "127.0.0.1",
			Port: intstr.FromInt(c.ClientPort),
			Path: "/health",
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

		// We need all of /var/etcd, because the downgrade script renames /var/etcd/data -> /var/etcd/data.bak
		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "varetcdata",
			MountPath: "/var/etcd/",
			ReadOnly:  false,
		})
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "varetcdata",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: c.VolumeMountPath + "/var/etcd/",
				},
			},
		})

		container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
			Name:      "varlogetcd",
			MountPath: "/var/log/etcd.log",
			ReadOnly:  false,
		})
		pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
			Name: "varlogetcd",
			VolumeSource: v1.VolumeSource{
				HostPath: &v1.HostPathVolumeSource{
					Path: c.LogFile,
				},
			},
		})

		if c.Spec.UseSSL {
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{
				Name:      "keystore",
				MountPath: keystorePath,
				ReadOnly:  true,
			})
			pod.Spec.Volumes = append(pod.Spec.Volumes, v1.Volume{
				Name: "keystore",
				VolumeSource: v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: keystorePath,
					},
				},
			})
		}

		pod.Spec.Containers = append(pod.Spec.Containers, container)
	}

	return pod
}

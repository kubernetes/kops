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

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/exec"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PathAuthnConfig is the path to the custom webhook authentication config
const PathAuthnConfig = "/etc/kubernetes/authn.config"

// KubeAPIServerBuilder install kube-apiserver (just the manifest at the moment)
type KubeAPIServerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeAPIServerBuilder{}

// Build is responsible for generating the configuration for the kube-apiserver
func (b *KubeAPIServerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.IsMaster {
		return nil
	}

	if err := b.writeAuthenticationConfig(c); err != nil {
		return err
	}

	if b.Cluster.Spec.EncryptionConfig != nil {
		if *b.Cluster.Spec.EncryptionConfig && b.IsKubernetesGTE("1.7") {
			b.Cluster.Spec.KubeAPIServer.ExperimentalEncryptionProviderConfig = fi.String(filepath.Join(b.PathSrvKubernetes(), "encryptionconfig.yaml"))
			key := "encryptionconfig"
			encryptioncfg, _ := b.SecretStore.Secret(key)
			if encryptioncfg != nil {
				contents := string(encryptioncfg.Data)
				t := &nodetasks.File{
					Path:     *b.Cluster.Spec.KubeAPIServer.ExperimentalEncryptionProviderConfig,
					Contents: fi.NewStringResource(contents),
					Mode:     fi.String("600"),
					Type:     nodetasks.FileType_File,
				}
				c.AddTask(t)
			}
		}
	}
	{
		pod, err := b.buildPod()
		if err != nil {
			return fmt.Errorf("error building kube-apiserver manifest: %v", err)
		}

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshalling manifest to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-apiserver.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	// @check if we are using secure client certificates for kubelet and grab the certificates
	if b.UseSecureKubelet() {
		name := "kubelet-api"
		if err := b.BuildCertificateTask(c, name, name+".pem"); err != nil {
			return err
		}
		if err := b.BuildPrivateKeyTask(c, name, name+"-key.pem"); err != nil {
			return err
		}
	}

	c.AddTask(&nodetasks.File{
		Path:        "/var/log/kube-apiserver.log",
		Contents:    fi.NewStringResource(""),
		Type:        nodetasks.FileType_File,
		Mode:        s("0400"),
		IfNotExists: true,
	})

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

		c.AddTask(&nodetasks.File{
			Path:     PathAuthnConfig,
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})

		return nil
	}

	if b.Cluster.Spec.Authentication.Heptio != nil {
		id := "heptio-authenticator-aws"
		b.Cluster.Spec.KubeAPIServer.AuthenticationTokenWebhookConfigFile = fi.String(PathAuthnConfig)

		{
			caCertificate, err := b.NodeupModelContext.KeyStore.FindCert(fi.CertificateId_CA)
			if err != nil {
				return fmt.Errorf("error fetching Heptio Authentication CA certificate from keystore: %v", err)
			}
			if caCertificate == nil {
				return fmt.Errorf("Heptio Authentication CA certificate %q not found", fi.CertificateId_CA)
			}

			cluster := kubeconfig.KubectlCluster{
				Server: "https://127.0.0.1:21362/authenticate",
			}
			context := kubeconfig.KubectlContext{
				Cluster: "heptio-authenticator-aws",
				User:    "kube-apiserver",
			}

			cluster.CertificateAuthorityData, err = caCertificate.AsBytes()
			if err != nil {
				return fmt.Errorf("error encoding Heptio Authentication CA certificate: %v", err)
			}

			config := kubeconfig.KubectlConfig{}
			config.Clusters = append(config.Clusters, &kubeconfig.KubectlClusterWithName{
				Name:    "heptio-authenticator-aws",
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

			c.AddTask(&nodetasks.File{
				Path:     PathAuthnConfig,
				Contents: fi.NewBytesResource(manifest),
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("600"),
			})
		}

		{
			c.AddTask(&nodetasks.UserTask{
				Name:  "aws-iam-authenticator",
				UID:   10000,
				Shell: "/sbin/nologin",
				Home:  "/srv/kubernetes/aws-iam-authenticator",
			})
		}

		{
			certificate, err := b.NodeupModelContext.KeyStore.FindCert(id)
			if err != nil {
				return fmt.Errorf("error fetching %q certificate from keystore: %v", id, err)
			}
			if certificate == nil {
				return fmt.Errorf("certificate %q not found", id)
			}

			certificateData, err := certificate.AsBytes()
			if err != nil {
				return fmt.Errorf("error encoding %q certificate: %v", id, err)
			}

			c.AddTask(&nodetasks.File{
				Path:     "/srv/kubernetes/heptio-authenticator-aws/cert.pem",
				Contents: fi.NewBytesResource(certificateData),
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("600"),
				Owner:    fi.String("aws-iam-authenticator"),
				Group:    fi.String("aws-iam-authenticator"),
			})
		}

		{
			privateKey, err := b.NodeupModelContext.KeyStore.FindPrivateKey(id)
			if err != nil {
				return fmt.Errorf("error fetching %q private key from keystore: %v", id, err)
			}
			if privateKey == nil {
				return fmt.Errorf("private key %q not found", id)
			}

			keyData, err := privateKey.AsBytes()
			if err != nil {
				return fmt.Errorf("error encoding %q private key: %v", id, err)
			}

			c.AddTask(&nodetasks.File{
				Path:     "/srv/kubernetes/heptio-authenticator-aws/key.pem",
				Contents: fi.NewBytesResource(keyData),
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("600"),
				Owner:    fi.String("aws-iam-authenticator"),
				Group:    fi.String("aws-iam-authenticator"),
			})
		}

		return nil
	}

	return fmt.Errorf("Unrecognized authentication config %v", b.Cluster.Spec.Authentication)
}

// buildPod is responsible for generating the kube-apiserver pod and thus manifest file
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

	// @check if we are using secure kubelet client certificates
	if b.UseSecureKubelet() {
		// @note we are making assumption were using the ones created by the pki model, not custom defined ones
		kubeAPIServer.KubeletClientCertificate = filepath.Join(b.PathSrvKubernetes(), "kubelet-api.pem")
		kubeAPIServer.KubeletClientKey = filepath.Join(b.PathSrvKubernetes(), "kubelet-api-key.pem")
	}

	if b.IsKubernetesGTE("1.7") {
		certPath := filepath.Join(b.PathSrvKubernetes(), "apiserver-aggregator.cert")
		kubeAPIServer.ProxyClientCertFile = &certPath
		keyPath := filepath.Join(b.PathSrvKubernetes(), "apiserver-aggregator.key")
		kubeAPIServer.ProxyClientKeyFile = &keyPath
	}

	// APIServer aggregation options
	if b.IsKubernetesGTE("1.7") {
		cert, err := b.KeyStore.FindCert("apiserver-aggregator-ca")
		if err != nil {
			return nil, fmt.Errorf("apiserver aggregator CA cert lookup failed: %v", err.Error())
		}

		if cert != nil {
			certPath := filepath.Join(b.PathSrvKubernetes(), "apiserver-aggregator-ca.cert")
			kubeAPIServer.RequestheaderClientCAFile = certPath
		}
	}

	// @fixup: the admission controller migrated from --admission-control to --enable-admission-plugins, but
	// most people will still have c.Spec.KubeAPIServer.AdmissionControl references into their configuration we need
	// to fix up. A PR https://github.com/kubernetes/kops/pull/5221/ introduced the issue and since the command line
	// flags are mutually exclusive the API refuses to come up.
	if b.IsKubernetesGTE("1.10") {
		// @note: note sure if this is the best place to put it, I could place into the validation.go which has the benefit of
		// fixing up the manifests itself, but that feels VERY hacky
		// @note: it's fine to use AdmissionControl here and it's not populated by the model, thus the only data could have come from the cluster spec
		c := b.Cluster.Spec.KubeAPIServer
		if len(c.AdmissionControl) > 0 {
			copy(c.EnableAdmissionPlugins, c.AdmissionControl)
			c.AdmissionControl = []string{}
		}
	}

	// build the kube-apiserver flags for the service
	flags, err := flagbuilder.BuildFlagsList(b.Cluster.Spec.KubeAPIServer)
	if err != nil {
		return nil, fmt.Errorf("error building kube-apiserver flags: %v", err)
	}

	// add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		flags = append(flags, fmt.Sprintf("--cloud-config=%s", CloudConfigFilePath))
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
		Command: exec.WithTee(
			"/usr/local/bin/kube-apiserver",
			sortedStrings(flags),
			"/var/log/kube-apiserver.log"),
		Env: getProxyEnvVars(b.Cluster.Spec.EgressProxy),
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
		Resources: v1.ResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("150m"),
			},
		},
	}

	for _, path := range b.SSLHostPaths() {
		name := strings.Replace(path, "/", "", -1)
		addHostPathMapping(pod, container, name, path)
	}

	addHostPathMapping(pod, container, "logfile", "/var/log/kube-apiserver.log").ReadOnly = false

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

	auditLogPath := b.Cluster.Spec.KubeAPIServer.AuditLogPath
	if auditLogPath != nil {
		// Mount the directory of the path instead, as kube-apiserver rotates the log by renaming the file.
		// Renaming is not possible when the file is mounted as the host path, and will return a
		// 'Device or resource busy' error
		auditLogPathDir := filepath.Dir(*auditLogPath)
		addHostPathMapping(pod, container, "auditlogpathdir", auditLogPathDir).ReadOnly = false
	}

	if b.Cluster.Spec.Authentication != nil {
		if b.Cluster.Spec.Authentication.Kopeio != nil || b.Cluster.Spec.Authentication.Heptio != nil {
			addHostPathMapping(pod, container, "authn-config", PathAuthnConfig)
		}
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	kubemanifest.MarkPodAsCritical(pod)

	return pod, nil
}

func (b *KubeAPIServerBuilder) buildAnnotations() map[string]string {
	annotations := make(map[string]string)
	annotations["dns.alpha.kubernetes.io/internal"] = b.Cluster.Spec.MasterInternalName
	if b.Cluster.Spec.API != nil && b.Cluster.Spec.API.DNS != nil {
		annotations["dns.alpha.kubernetes.io/external"] = b.Cluster.Spec.MasterPublicName
	}

	return annotations
}

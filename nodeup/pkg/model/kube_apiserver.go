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
	"fmt"
	"path/filepath"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/k8scodecs"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/tokens"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/pkg/wellknownusers"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/proxy"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PathAuthnConfig is the path to the custom webhook authentication config.
const PathAuthnConfig = "/etc/kubernetes/authn.config"

// KubeAPIServerBuilder installs kube-apiserver.
type KubeAPIServerBuilder struct {
	*NodeupModelContext
}

var _ fi.ModelBuilder = &KubeAPIServerBuilder{}

// Build is responsible for generating the configuration for the kube-apiserver.
func (b *KubeAPIServerBuilder) Build(c *fi.ModelBuilderContext) error {
	if !b.HasAPIServer {
		return nil
	}

	pathSrvKAPI := filepath.Join(b.PathSrvKubernetes(), "kube-apiserver")

	var kubeAPIServer kops.KubeAPIServerConfig
	if b.NodeupConfig.APIServerConfig.KubeAPIServer != nil {
		kubeAPIServer = *b.NodeupConfig.APIServerConfig.KubeAPIServer
	}

	if err := b.writeAuthenticationConfig(c, &kubeAPIServer); err != nil {
		return err
	}

	if b.NodeupConfig.APIServerConfig.EncryptionConfigSecretHash != "" {
		encryptionConfigPath := fi.String(filepath.Join(pathSrvKAPI, "encryptionconfig.yaml"))

		kubeAPIServer.EncryptionProviderConfig = encryptionConfigPath

		key := "encryptionconfig"
		encryptioncfg, err := b.SecretStore.Secret(key)
		if err == nil {
			contents := string(encryptioncfg.Data)
			t := &nodetasks.File{
				Path:     *encryptionConfigPath,
				Contents: fi.NewStringResource(contents),
				Mode:     fi.String("600"),
				Type:     nodetasks.FileType_File,
			}
			c.AddTask(t)
		} else {
			return fmt.Errorf("encryptionConfig enabled, but could not load encryptionconfig secret: %v", err)
		}
	}

	kubeAPIServer.ServiceAccountKeyFile = append(kubeAPIServer.ServiceAccountKeyFile, filepath.Join(pathSrvKAPI, "service-account.pub"))
	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(pathSrvKAPI, "service-account.pub"),
		Contents: fi.NewStringResource(b.NodeupConfig.APIServerConfig.ServiceAccountPublicKeys),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	})

	// Set the signing key if we're using Service Account Token VolumeProjection
	if kubeAPIServer.ServiceAccountSigningKeyFile == nil {
		if fi.StringValue(kubeAPIServer.ServiceAccountIssuer) != "" {
			s := filepath.Join(pathSrvKAPI, "service-account.key")
			kubeAPIServer.ServiceAccountSigningKeyFile = &s
			if err := b.BuildPrivateKeyTask(c, "service-account", pathSrvKAPI, "service-account", nil, nil); err != nil {
				return err
			}
		}
	}

	if b.UseEtcdManager() {
		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(pathSrvKAPI, "etcd-ca.crt"),
			Contents: fi.NewStringResource(b.NodeupConfig.CAs["etcd-clients-ca"]),
			Type:     nodetasks.FileType_File,
			Mode:     fi.String("0644"),
		})
		kubeAPIServer.EtcdCAFile = filepath.Join(pathSrvKAPI, "etcd-ca.crt")

		issueCert := &nodetasks.IssueCert{
			Name:      "etcd-client",
			Signer:    "etcd-clients-ca",
			KeypairID: b.NodeupConfig.KeypairIDs["etcd-clients-ca"],
			Type:      "client",
			Subject: nodetasks.PKIXName{
				CommonName: "kube-apiserver",
			},
		}
		c.AddTask(issueCert)
		if err := issueCert.AddFileTasks(c, pathSrvKAPI, issueCert.Name, "", nil); err != nil {
			return err
		}
	} else if b.UseEtcdTLS() {
		kubeAPIServer.EtcdCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")
	}
	kubeAPIServer.EtcdCertFile = filepath.Join(pathSrvKAPI, "etcd-client.crt")
	kubeAPIServer.EtcdKeyFile = filepath.Join(pathSrvKAPI, "etcd-client.key")

	{
		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(pathSrvKAPI, "apiserver-aggregator-ca.crt"),
			Contents: fi.NewStringResource(b.NodeupConfig.CAs["apiserver-aggregator-ca"]),
			Type:     nodetasks.FileType_File,
			Mode:     fi.String("0644"),
		})
		kubeAPIServer.RequestheaderClientCAFile = filepath.Join(pathSrvKAPI, "apiserver-aggregator-ca.crt")

		issueCert := &nodetasks.IssueCert{
			Name:      "apiserver-aggregator",
			Signer:    "apiserver-aggregator-ca",
			KeypairID: b.NodeupConfig.KeypairIDs["apiserver-aggregator-ca"],
			Type:      "client",
			// Must match RequestheaderAllowedNames
			Subject: nodetasks.PKIXName{CommonName: "aggregator"},
		}
		c.AddTask(issueCert)
		err := issueCert.AddFileTasks(c, pathSrvKAPI, "apiserver-aggregator", "", nil)
		if err != nil {
			return err
		}
		kubeAPIServer.ProxyClientCertFile = fi.String(filepath.Join(pathSrvKAPI, "apiserver-aggregator.crt"))
		kubeAPIServer.ProxyClientKeyFile = fi.String(filepath.Join(pathSrvKAPI, "apiserver-aggregator.key"))
	}

	if err := b.writeServerCertificate(c, &kubeAPIServer); err != nil {
		return err
	}

	if err := b.writeKubeletAPICertificate(c, &kubeAPIServer); err != nil {
		return err
	}

	if err := b.writeStaticCredentials(c, &kubeAPIServer); err != nil {
		return err
	}

	{
		pod, err := b.buildPod(&kubeAPIServer)
		if err != nil {
			return fmt.Errorf("error building kube-apiserver manifest: %v", err)
		}

		manifest, err := k8scodecs.ToVersionedYaml(pod)
		if err != nil {
			return fmt.Errorf("error marshaling manifest to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     "/etc/kubernetes/manifests/kube-apiserver.manifest",
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})
	}

	// If we're using kube-apiserver-healthcheck, we need to set up the client cert etc
	if b.findHealthcheckManifest() != nil {
		if err := b.addHealthcheckSidecarTasks(c); err != nil {
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

func (b *KubeAPIServerBuilder) writeAuthenticationConfig(c *fi.ModelBuilderContext, kubeAPIServer *kops.KubeAPIServerConfig) error {
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
			return fmt.Errorf("error marshaling authentication config to yaml: %v", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     PathAuthnConfig,
			Contents: fi.NewBytesResource(manifest),
			Type:     nodetasks.FileType_File,
		})

		return nil
	}

	if b.Cluster.Spec.Authentication.Aws != nil {
		id := "aws-iam-authenticator"
		kubeAPIServer.AuthenticationTokenWebhookConfigFile = fi.String(PathAuthnConfig)

		{
			cluster := kubeconfig.KubectlCluster{
				Server:                   "https://127.0.0.1:21362/authenticate",
				CertificateAuthorityData: []byte(b.NodeupConfig.CAs[fi.CertificateIDCA]),
			}
			context := kubeconfig.KubectlContext{
				Cluster: "aws-iam-authenticator",
				User:    "kube-apiserver",
			}

			config := kubeconfig.KubectlConfig{}
			config.Clusters = append(config.Clusters, &kubeconfig.KubectlClusterWithName{
				Name:    "aws-iam-authenticator",
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
				return fmt.Errorf("error marshaling authentication config to yaml: %v", err)
			}

			c.AddTask(&nodetasks.File{
				Path:     PathAuthnConfig,
				Contents: fi.NewBytesResource(manifest),
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("600"),
			})
		}

		// We create user aws-iam-authenticator and hardcode its UID to 10000 as
		// that is the ID used inside the aws-iam-authenticator container.
		// The owner/group for the keypair to aws-iam-authenticator
		{
			c.AddTask(&nodetasks.UserTask{
				Name:  "aws-iam-authenticator",
				UID:   wellknownusers.AWSAuthenticator,
				Shell: "/sbin/nologin",
				Home:  "/srv/kubernetes/aws-iam-authenticator",
			})
		}

		{
			issueCert := &nodetasks.IssueCert{
				Name:      id,
				Signer:    fi.CertificateIDCA,
				KeypairID: b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
				Type:      "server",
				Subject:   nodetasks.PKIXName{CommonName: id},
				AlternateNames: []string{
					"localhost",
					"127.0.0.1",
				},
			}
			c.AddTask(issueCert)
			certificate, privateKey, _ := issueCert.GetResources()

			c.AddTask(&nodetasks.File{
				Path:     "/srv/kubernetes/aws-iam-authenticator/cert.pem",
				Contents: certificate,
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("600"),
				Owner:    fi.String("aws-iam-authenticator"),
				Group:    fi.String("aws-iam-authenticator"),
			})

			c.AddTask(&nodetasks.File{
				Path:     "/srv/kubernetes/aws-iam-authenticator/key.pem",
				Contents: privateKey,
				Type:     nodetasks.FileType_File,
				Mode:     fi.String("600"),
				Owner:    fi.String("aws-iam-authenticator"),
				Group:    fi.String("aws-iam-authenticator"),
			})
		}

		return nil
	}

	return fmt.Errorf("unrecognized authentication config %v", b.Cluster.Spec.Authentication)
}

func (b *KubeAPIServerBuilder) writeServerCertificate(c *fi.ModelBuilderContext, kubeAPIServer *kops.KubeAPIServerConfig) error {
	pathSrvKAPI := filepath.Join(b.PathSrvKubernetes(), "kube-apiserver")

	{
		// A few names used from inside the cluster, which all resolve the same based on our default suffixes
		alternateNames := []string{
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc." + b.Cluster.Spec.ClusterDNSDomain,
		}

		// Names specified in the cluster spec
		alternateNames = append(alternateNames, b.Cluster.Spec.MasterPublicName)
		alternateNames = append(alternateNames, b.Cluster.Spec.MasterInternalName)
		alternateNames = append(alternateNames, b.Cluster.Spec.AdditionalSANs...)

		// Load balancer IPs passed in through NodeupConfig
		alternateNames = append(alternateNames, b.NodeupConfig.ApiserverAdditionalIPs...)

		// Referencing it by internal IP should work also
		{
			ip, err := components.WellKnownServiceIP(&b.Cluster.Spec, 1)
			if err != nil {
				return err
			}
			alternateNames = append(alternateNames, ip.String())
		}

		// We also want to be able to reference it locally via https://127.0.0.1
		alternateNames = append(alternateNames, "127.0.0.1")

		if b.Cluster.Spec.CloudProvider == "openstack" {
			if b.Cluster.Spec.Topology != nil && b.Cluster.Spec.Topology.Masters == kops.TopologyPrivate {
				instanceAddress, err := getInstanceAddress()
				if err != nil {
					return err
				}
				alternateNames = append(alternateNames, instanceAddress)
			}
		}

		issueCert := &nodetasks.IssueCert{
			Name:           "master",
			Signer:         fi.CertificateIDCA,
			KeypairID:      b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
			Type:           "server",
			Subject:        nodetasks.PKIXName{CommonName: "kubernetes-master"},
			AlternateNames: alternateNames,
		}

		// Including the CA certificate is more correct, and is needed for e.g. AWS WebIdentity federation
		issueCert.IncludeRootCertificate = true

		c.AddTask(issueCert)
		err := issueCert.AddFileTasks(c, pathSrvKAPI, "server", "", nil)
		if err != nil {
			return err
		}
	}

	// If clientCAFile is not specified, set it to the default value ${PathSrvKubernetes}/ca.crt
	if kubeAPIServer.ClientCAFile == "" {
		kubeAPIServer.ClientCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")
	}
	kubeAPIServer.TLSCertFile = filepath.Join(pathSrvKAPI, "server.crt")
	kubeAPIServer.TLSPrivateKeyFile = filepath.Join(pathSrvKAPI, "server.key")

	return nil
}

func (b *KubeAPIServerBuilder) writeKubeletAPICertificate(c *fi.ModelBuilderContext, kubeAPIServer *kops.KubeAPIServerConfig) error {
	pathSrvKAPI := filepath.Join(b.PathSrvKubernetes(), "kube-apiserver")

	issueCert := &nodetasks.IssueCert{
		Name:      "kubelet-api",
		Signer:    fi.CertificateIDCA,
		KeypairID: b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
		Type:      "client",
		Subject:   nodetasks.PKIXName{CommonName: "kubelet-api"},
	}
	c.AddTask(issueCert)
	err := issueCert.AddFileTasks(c, pathSrvKAPI, "kubelet-api", "", nil)
	if err != nil {
		return err
	}

	// @note we are making assumption were using the ones created by the pki model, not custom defined ones
	kubeAPIServer.KubeletClientCertificate = filepath.Join(pathSrvKAPI, "kubelet-api.crt")
	kubeAPIServer.KubeletClientKey = filepath.Join(pathSrvKAPI, "kubelet-api.key")

	return nil
}

func (b *KubeAPIServerBuilder) writeStaticCredentials(c *fi.ModelBuilderContext, kubeAPIServer *kops.KubeAPIServerConfig) error {
	pathSrvKAPI := filepath.Join(b.PathSrvKubernetes(), "kube-apiserver")

	// Support for basic auth was deprecated 1.16 and removed in 1.19
	// https://github.com/kubernetes/kubernetes/pull/89069
	if b.IsKubernetesLT("1.19") && b.SecretStore != nil {
		key := "kube"
		token, err := b.SecretStore.FindSecret(key)
		if err != nil {
			return err
		}
		if token == nil {
			return fmt.Errorf("token not found: %q", key)
		}
		csv := string(token.Data) + "," + adminUser + "," + adminUser + "," + adminGroup

		t := &nodetasks.File{
			Path:     filepath.Join(pathSrvKAPI, "basic_auth.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		}
		c.AddTask(t)
	}

	if b.SecretStore != nil {
		allTokens, err := b.allAuthTokens()
		if err != nil {
			return err
		}

		var lines []string
		for id, token := range allTokens {
			if id == adminUser {
				lines = append(lines, token+","+id+","+id+","+adminGroup)
			} else {
				lines = append(lines, token+","+id+","+id)
			}
		}
		csv := strings.Join(lines, "\n")

		c.AddTask(&nodetasks.File{
			Path:     filepath.Join(pathSrvKAPI, "known_tokens.csv"),
			Contents: fi.NewStringResource(csv),
			Type:     nodetasks.FileType_File,
			Mode:     s("0600"),
		})
	}

	// Support for basic auth was deprecated 1.16 and removed in 1.19
	// https://github.com/kubernetes/kubernetes/pull/89069
	if b.IsKubernetesLT("1.18") {
		kubeAPIServer.TokenAuthFile = filepath.Join(pathSrvKAPI, "known_tokens.csv")
		if kubeAPIServer.DisableBasicAuth == nil || !*kubeAPIServer.DisableBasicAuth {
			kubeAPIServer.BasicAuthFile = filepath.Join(pathSrvKAPI, "basic_auth.csv")
		}
	} else if b.IsKubernetesLT("1.19") {
		if kubeAPIServer.DisableBasicAuth != nil && !*kubeAPIServer.DisableBasicAuth {
			kubeAPIServer.BasicAuthFile = filepath.Join(pathSrvKAPI, "basic_auth.csv")
		}
	}

	return nil
}

// allTokens returns a map of all auth tokens that are present
func (b *KubeAPIServerBuilder) allAuthTokens() (map[string]string, error) {
	possibleTokens := tokens.GetKubernetesAuthTokens_Deprecated()

	tokens := make(map[string]string)
	for _, id := range possibleTokens {
		token, err := b.SecretStore.FindSecret(id)
		if err != nil {
			return nil, err
		}
		if token != nil {
			tokens[id] = string(token.Data)
		}
	}
	return tokens, nil
}

// buildPod is responsible for generating the kube-apiserver pod and thus manifest file
func (b *KubeAPIServerBuilder) buildPod(kubeAPIServer *kops.KubeAPIServerConfig) (*v1.Pod, error) {
	// we need to replace 127.0.0.1 for etcd urls with the dns names in case this apiserver is not
	// running on master nodes
	if !b.IsMaster {
		clusterName := b.Cluster.ObjectMeta.Name
		mainEtcdDNSName := "main.etcd." + clusterName
		eventsEtcdDNSName := "events.etcd." + clusterName
		for i := range kubeAPIServer.EtcdServers {
			kubeAPIServer.EtcdServers[i] = strings.ReplaceAll(kubeAPIServer.EtcdServers[i], "127.0.0.1", mainEtcdDNSName)
		}
		for i := range kubeAPIServer.EtcdServersOverrides {
			if strings.HasPrefix(kubeAPIServer.EtcdServersOverrides[i], "/events") {
				kubeAPIServer.EtcdServersOverrides[i] = strings.ReplaceAll(kubeAPIServer.EtcdServersOverrides[i], "127.0.0.1", eventsEtcdDNSName)
			}
		}
	}

	// @fixup: the admission controller migrated from --admission-control to --enable-admission-plugins, but
	// most people will still have c.Spec.KubeAPIServer.AdmissionControl references into their configuration we need
	// to fix up. A PR https://github.com/kubernetes/kops/pull/5221/ introduced the issue and since the command line
	// flags are mutually exclusive the API refuses to come up.
	{
		// @note: note sure if this is the best place to put it, I could place into the validation.go which has the benefit of
		// fixing up the manifests itself, but that feels VERY hacky
		// @note: it's fine to use AdmissionControl here and it's not populated by the model, thus the only data could have come from the cluster spec
		if len(kubeAPIServer.AdmissionControl) > 0 {
			kubeAPIServer.EnableAdmissionPlugins = append([]string(nil), kubeAPIServer.AdmissionControl...)
			kubeAPIServer.AdmissionControl = []string{}
		}
	}

	// build the kube-apiserver flags for the service
	flags, err := flagbuilder.BuildFlagsList(kubeAPIServer)
	if err != nil {
		return nil, fmt.Errorf("error building kube-apiserver flags: %v", err)
	}

	// add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		flags = append(flags, fmt.Sprintf("--cloud-config=%s", InTreeCloudConfigFilePath))
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

	useHealthcheckProxy := b.findHealthcheckManifest() != nil

	probeAction := &v1.HTTPGetAction{
		Host: "127.0.0.1",
		Path: "/healthz",
		Port: intstr.FromInt(wellknownports.KubeAPIServerHealthCheck),
	}
	if useHealthcheckProxy {
		// kube-apiserver-healthcheck sidecar container runs on port 3990
	} else if kubeAPIServer.InsecurePort != 0 {
		probeAction.Port = intstr.FromInt(int(kubeAPIServer.InsecurePort))
	} else if kubeAPIServer.SecurePort != 0 {
		probeAction.Port = intstr.FromInt(int(kubeAPIServer.SecurePort))
		probeAction.Scheme = v1.URISchemeHTTPS
	}

	resourceRequests := v1.ResourceList{}
	resourceLimits := v1.ResourceList{}

	cpuRequest := resource.MustParse("150m")
	if kubeAPIServer.CPURequest != "" {
		var err error
		cpuRequest, err = resource.ParseQuantity(kubeAPIServer.CPURequest)
		if err != nil {
			return nil, fmt.Errorf("error parsing CPURequest=%q", kubeAPIServer.CPURequest)
		}
	}
	resourceRequests["cpu"] = cpuRequest

	if kubeAPIServer.CPULimit != "" {
		cpuLimit, err := resource.ParseQuantity(kubeAPIServer.CPULimit)
		if err != nil {
			return nil, fmt.Errorf("error parsing CPULimit=%q", kubeAPIServer.CPULimit)
		}
		resourceLimits["cpu"] = cpuLimit
	}

	if kubeAPIServer.MemoryRequest != "" {
		memoryRequest, err := resource.ParseQuantity(kubeAPIServer.MemoryRequest)
		if err != nil {
			return nil, fmt.Errorf("error parsing MemoryRequest=%q", kubeAPIServer.MemoryRequest)
		}
		resourceRequests["memory"] = memoryRequest
	}

	if kubeAPIServer.MemoryLimit != "" {
		memoryLimit, err := resource.ParseQuantity(kubeAPIServer.MemoryLimit)
		if err != nil {
			return nil, fmt.Errorf("error parsing MemoryLimit=%q", kubeAPIServer.MemoryLimit)
		}
		resourceLimits["memory"] = memoryLimit
	}

	image := kubeAPIServer.Image
	if b.Architecture != architectures.ArchitectureAmd64 {
		image = strings.Replace(image, "-amd64", "-"+string(b.Architecture), 1)
	}

	container := &v1.Container{
		Name:  "kube-apiserver",
		Image: image,
		Env:   proxy.GetProxyEnvVars(b.Cluster.Spec.EgressProxy),
		LivenessProbe: &v1.Probe{
			Handler: v1.Handler{
				HTTPGet: probeAction,
			},
			InitialDelaySeconds: 45,
			TimeoutSeconds:      15,
		},
		Ports: []v1.ContainerPort{
			{
				Name:          "https",
				ContainerPort: kubeAPIServer.SecurePort,
				HostPort:      kubeAPIServer.SecurePort,
			},
		},
		Resources: v1.ResourceRequirements{
			Requests: resourceRequests,
			Limits:   resourceLimits,
		},
	}

	if kubeAPIServer.InsecurePort != 0 {
		container.Ports = append(container.Ports, v1.ContainerPort{
			Name:          "local",
			ContainerPort: kubeAPIServer.InsecurePort,
			HostPort:      kubeAPIServer.InsecurePort,
		})
	}

	// Log both to docker and to the logfile
	addHostPathMapping(pod, container, "logfile", "/var/log/kube-apiserver.log").ReadOnly = false
	// We use lighter containers that don't include shells
	// But they have richer logging support via klog
	container.Command = []string{"/usr/local/bin/kube-apiserver"}
	if kubeAPIServer.LogFormat != "" && kubeAPIServer.LogFormat != "text" {
		// When logging-format is not text, some flags are not accepted.
		// https://github.com/kubernetes/kops/issues/13245
		container.Args = sortedStrings(flags)
	} else {
		container.Args = append(
			sortedStrings(flags),
			"--logtostderr=false", // https://github.com/kubernetes/klog/issues/60
			"--alsologtostderr",
			"--log-file=/var/log/kube-apiserver.log")
	}

	for _, path := range b.SSLHostPaths() {
		name := strings.Replace(path, "/", "", -1)
		addHostPathMapping(pod, container, name, path)
	}

	// Add cloud config file if needed
	if b.Cluster.Spec.CloudConfig != nil {
		addHostPathMapping(pod, container, "cloudconfig", InTreeCloudConfigFilePath)
	}

	addHostPathMapping(pod, container, "kubernetesca", filepath.Join(b.PathSrvKubernetes(), "ca.crt"))

	pathSrvKAPI := filepath.Join(b.PathSrvKubernetes(), "kube-apiserver")
	addHostPathMapping(pod, container, "srvkapi", pathSrvKAPI)

	pathSrvSshproxy := b.PathSrvSshproxy()
	if pathSrvSshproxy != "" {
		addHostPathMapping(pod, container, "srvsshproxy", pathSrvSshproxy)
	}

	auditLogPath := kubeAPIServer.AuditLogPath
	// Don't mount a volume if the mount path is set to '-' for stdout logging
	// See https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-backends
	if auditLogPath != nil && *auditLogPath != "-" {
		// Mount the directory of the path instead, as kube-apiserver rotates the log by renaming the file.
		// Renaming is not possible when the file is mounted as the host path, and will return a
		// 'Device or resource busy' error
		auditLogPathDir := filepath.Dir(*auditLogPath)
		addHostPathMapping(pod, container, "auditlogpathdir", auditLogPathDir).ReadOnly = false
	}

	if b.Cluster.Spec.Authentication != nil {
		if b.Cluster.Spec.Authentication.Kopeio != nil || b.Cluster.Spec.Authentication.Aws != nil {
			addHostPathMapping(pod, container, "authn-config", PathAuthnConfig)
		}
	}

	pod.Spec.Containers = append(pod.Spec.Containers, *container)

	kubemanifest.MarkPodAsCritical(pod)
	kubemanifest.MarkPodAsClusterCritical(pod)

	if useHealthcheckProxy {
		if err := b.addHealthcheckSidecar(pod); err != nil {
			return nil, err
		}
	}

	return pod, nil
}

func (b *KubeAPIServerBuilder) buildAnnotations() map[string]string {
	annotations := make(map[string]string)
	annotations["kubectl.kubernetes.io/default-container"] = "kube-apiserver"

	if b.Cluster.Spec.API != nil {
		if b.Cluster.Spec.API.LoadBalancer == nil || !b.Cluster.Spec.API.LoadBalancer.UseForInternalApi {
			annotations["dns.alpha.kubernetes.io/internal"] = b.Cluster.Spec.MasterInternalName
		}

		if b.Cluster.Spec.API.DNS != nil {
			annotations["dns.alpha.kubernetes.io/external"] = b.Cluster.Spec.MasterPublicName
		}
	}

	return annotations
}

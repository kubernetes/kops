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

	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"

	"github.com/blang/semver"
)

// NodeupModelContext is the context supplied the nodeup tasks
type NodeupModelContext struct {
	Architecture      Architecture
	Assets            *fi.AssetStore
	Cluster           *kops.Cluster
	Distribution      distros.Distribution
	InstanceGroup     *kops.InstanceGroup
	IsMaster          bool
	KeyStore          fi.CAStore
	KubernetesVersion semver.Version
	NodeupConfig      *nodeup.Config
	SecretStore       fi.SecretStore
}

// SSLHostPaths returns the TLS paths for the distribution
func (c *NodeupModelContext) SSLHostPaths() []string {
	paths := []string{"/etc/ssl", "/etc/pki/tls", "/etc/pki/ca-trust"}

	switch c.Distribution {
	case distros.DistributionCoreOS:
		// Because /usr is read-only on CoreOS, we can't have any new directories; docker will try (and fail) to create them
		// TODO: Just check if the directories exist?
		paths = append(paths, "/usr/share/ca-certificates")
	case distros.DistributionContainerOS:
		paths = append(paths, "/usr/share/ca-certificates")
	default:
		paths = append(paths, "/usr/share/ssl", "/usr/ssl", "/usr/lib/ssl", "/usr/local/openssl", "/var/ssl", "/etc/openssl")
	}

	return paths
}

// PathSrvKubernetes returns the path for the kubernetes service files
func (c *NodeupModelContext) PathSrvKubernetes() string {
	switch c.Distribution {
	case distros.DistributionContainerOS:
		return "/etc/srv/kubernetes"
	default:
		return "/srv/kubernetes"
	}
}

// PathSrvSshproxy returns the path for the SSL proxy
func (c *NodeupModelContext) PathSrvSshproxy() string {
	switch c.Distribution {
	case distros.DistributionContainerOS:
		return "/etc/srv/sshproxy"
	default:
		return "/srv/sshproxy"
	}
}

// CNIBinDir returns the path for the CNI binaries
func (c *NodeupModelContext) CNIBinDir() string {
	switch c.Distribution {
	case distros.DistributionContainerOS:
		return "/home/kubernetes/bin/"
	default:
		return "/opt/cni/bin/"
	}
}

// CNIConfDir returns the CNI directory
func (c *NodeupModelContext) CNIConfDir() string {
	return "/etc/cni/net.d/"
}

// buildPKIKubeconfig generates a kubeconfig
func (c *NodeupModelContext) buildPKIKubeconfig(id string) (string, error) {
	caCertificate, err := c.KeyStore.Cert(fi.CertificateId_CA)
	if err != nil {
		return "", fmt.Errorf("error fetching CA certificate from keystore: %v", err)
	}

	certificate, err := c.KeyStore.Cert(id)
	if err != nil {
		return "", fmt.Errorf("error fetching %q certificate from keystore: %v", id, err)
	}
	privateKey, err := c.KeyStore.PrivateKey(id)
	if err != nil {
		return "", fmt.Errorf("error fetching %q private key from keystore: %v", id, err)
	}

	user := kubeconfig.KubectlUser{}
	user.ClientCertificateData, err = certificate.AsBytes()
	if err != nil {
		return "", fmt.Errorf("error encoding %q certificate: %v", id, err)
	}
	user.ClientKeyData, err = privateKey.AsBytes()
	if err != nil {
		return "", fmt.Errorf("error encoding %q private key: %v", id, err)
	}
	cluster := kubeconfig.KubectlCluster{}
	cluster.CertificateAuthorityData, err = caCertificate.AsBytes()
	if err != nil {
		return "", fmt.Errorf("error encoding CA certificate: %v", err)
	}

	if c.IsMaster {
		if c.IsKubernetesGTE("1.6") {
			// Use https in 1.6, even for local connections, so we can turn off the insecure port
			cluster.Server = "https://127.0.0.1"
		} else {
			cluster.Server = "http://127.0.0.1:8080"
		}
	} else {
		cluster.Server = "https://" + c.Cluster.Spec.MasterInternalName
	}

	config := &kubeconfig.KubectlConfig{
		ApiVersion: "v1",
		Kind:       "Config",
		Users: []*kubeconfig.KubectlUserWithName{
			{
				Name: id,
				User: user,
			},
		},
		Clusters: []*kubeconfig.KubectlClusterWithName{
			{
				Name:    "local",
				Cluster: cluster,
			},
		},
		Contexts: []*kubeconfig.KubectlContextWithName{
			{
				Name: "service-account-context",
				Context: kubeconfig.KubectlContext{
					Cluster: "local",
					User:    id,
				},
			},
		},
		CurrentContext: "service-account-context",
	}

	yaml, err := kops.ToRawYaml(config)
	if err != nil {
		return "", fmt.Errorf("error marshalling kubeconfig to yaml: %v", err)
	}

	return string(yaml), nil
}

// IsKubernetesGTE checks if the version is greater-than-or-equal
func (c *NodeupModelContext) IsKubernetesGTE(version string) bool {
	return util.IsKubernetesGTE(version, c.KubernetesVersion)
}

// UseEtcdTLS checks if the etcd cluster has TLS enabled bool
func (c *NodeupModelContext) UseEtcdTLS() bool {
	// @note: because we enforce that 'both' have to be enabled for TLS we only need to check one here.
	for _, x := range c.Cluster.Spec.EtcdClusters {
		if x.EnableEtcdTLS {
			return true
		}
	}

	return false
}

// UsesCNI checks if the cluster has CNI configured
func (c *NodeupModelContext) UsesCNI() bool {
	networking := c.Cluster.Spec.Networking
	if networking == nil || networking.Classic != nil {
		return false
	}
	return true
}

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

	"github.com/blang/semver"
	"github.com/golang/glog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
)

// NodeupModelContext is the context supplied the nodeup tasks
type NodeupModelContext struct {
	Architecture  Architecture
	Assets        *fi.AssetStore
	Cluster       *kops.Cluster
	Distribution  distros.Distribution
	InstanceGroup *kops.InstanceGroup
	IsMaster      bool
	KeyStore      fi.CAStore
	NodeupConfig  *nodeup.Config
	SecretStore   fi.SecretStore

	kubernetesVersion semver.Version
}

// Init completes initialization of the object, for example pre-parsing the kubernetes version
func (c *NodeupModelContext) Init() error {
	k8sVersion, err := util.ParseKubernetesVersion(c.Cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return fmt.Errorf("unable to parse KubernetesVersion %q", c.Cluster.Spec.KubernetesVersion)
	}
	c.kubernetesVersion = *k8sVersion

	return nil
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

// FileAssetsDefaultPath is the default location for assets which have no path
func (c *NodeupModelContext) FileAssetsDefaultPath() string {
	return filepath.Join(c.PathSrvKubernetes(), "assets")
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
	caCertificate, err := c.KeyStore.FindCert(fi.CertificateId_CA)
	if err != nil {
		return "", fmt.Errorf("error fetching CA certificate from keystore: %v", err)
	}
	if caCertificate == nil {
		return "", fmt.Errorf("CA certificate %q not found", fi.CertificateId_CA)
	}

	certificate, err := c.KeyStore.FindCert(id)
	if err != nil {
		return "", fmt.Errorf("error fetching %q certificate from keystore: %v", id, err)
	}
	if certificate == nil {
		return "", fmt.Errorf("certificate %q not found", id)
	}

	privateKey, err := c.KeyStore.FindPrivateKey(id)
	if err != nil {
		return "", fmt.Errorf("error fetching %q private key from keystore: %v", id, err)
	}
	if privateKey == nil {
		return "", fmt.Errorf("private key %q not found", id)
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
	if c.kubernetesVersion.Major == 0 {
		glog.Fatalf("kubernetesVersion not set (%s); Init not called", c.kubernetesVersion)
	}
	return util.IsKubernetesGTE(version, c.kubernetesVersion)
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

// UseTLSAuth checks the peer-auth is set in both cluster
// @NOTE: in retrospect i think we should have consolidated the common config in the wrapper struct; it
// feels weird we set things like version, tls etc per cluster since they both have to be the same.
func (c *NodeupModelContext) UseTLSAuth() bool {
	if !c.UseEtcdTLS() {
		return false
	}

	for _, x := range c.Cluster.Spec.EtcdClusters {
		if x.EnableTLSAuth {
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

// UseSecureKubelet checks if the kubelet api should be protected by a client certificate. Note: the settings are
// in one of three section, master specific kubelet, cluster wide kubelet or the InstanceGroup. Though arguably is
// doesn't make much sense to unset this on a per InstanceGroup level, but hey :)
func (c *NodeupModelContext) UseSecureKubelet() bool {
	cluster := &c.Cluster.Spec // just to shorten the typing
	group := &c.InstanceGroup.Spec

	// @check on the InstanceGroup itself
	if group.Kubelet != nil && group.Kubelet.AnonymousAuth != nil && *group.Kubelet.AnonymousAuth == false {
		return true
	}

	// @check if we have anything specific to master kubelet
	if c.IsMaster {
		if cluster.MasterKubelet != nil && cluster.MasterKubelet.AnonymousAuth != nil && *cluster.MasterKubelet.AnonymousAuth == false {
			return true
		}
	}

	// @check the default settings for master and kubelet
	if cluster.Kubelet != nil && cluster.Kubelet.AnonymousAuth != nil && *cluster.Kubelet.AnonymousAuth == false {
		return true
	}

	return false
}

// KubectlPath returns distro based path for kubectl
func (c *NodeupModelContext) KubectlPath() string {
	kubeletCommand := "/usr/local/bin"
	if c.Distribution == distros.DistributionCoreOS {
		kubeletCommand = "/opt/bin"
	}
	if c.Distribution == distros.DistributionContainerOS {
		kubeletCommand = "/home/kubernetes/bin"
	}
	return kubeletCommand
}

// BuildCertificateTask is responsible for build a certificate request task
func (c *NodeupModelContext) BuildCertificateTask(ctx *fi.ModelBuilderContext, name, filename string) error {
	cert, err := c.KeyStore.FindCert(name)
	if err != nil {
		return err
	}

	if cert == nil {
		return fmt.Errorf("certificate %q not found", name)
	}

	serialized, err := cert.AsString()
	if err != nil {
		return err
	}

	ctx.AddTask(&nodetasks.File{
		Path:     filepath.Join(c.PathSrvKubernetes(), filename),
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0400"),
	})

	return nil
}

// BuildPrivateKeyTask is responsible for build a certificate request task
func (c *NodeupModelContext) BuildPrivateTask(ctx *fi.ModelBuilderContext, name, filename string) error {
	cert, err := c.KeyStore.FindPrivateKey(name)
	if err != nil {
		return err
	}

	if cert == nil {
		return fmt.Errorf("private key %q not found", name)
	}

	serialized, err := cert.AsString()
	if err != nil {
		return err
	}

	ctx.AddTask(&nodetasks.File{
		Path:     filepath.Join(c.PathSrvKubernetes(), filename),
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0400"),
	})

	return nil
}

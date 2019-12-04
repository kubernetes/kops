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
	"os"
	"path/filepath"
	"strings"

	"k8s.io/klog"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/kubeconfig"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kubernetes/pkg/util/mount"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/blang/semver"
)

// NodeupModelContext is the context supplied the nodeup tasks
type NodeupModelContext struct {
	Architecture  Architecture
	Assets        *fi.AssetStore
	Cluster       *kops.Cluster
	Distribution  distros.Distribution
	InstanceGroup *kops.InstanceGroup
	KeyStore      fi.CAStore
	NodeupConfig  *nodeup.Config
	SecretStore   fi.SecretStore

	// IsMaster is true if the InstanceGroup has a role of master (populated by Init)
	IsMaster bool

	kubernetesVersion semver.Version
}

// Init completes initialization of the object, for example pre-parsing the kubernetes version
func (c *NodeupModelContext) Init() error {
	k8sVersion, err := util.ParseKubernetesVersion(c.Cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return fmt.Errorf("unable to parse KubernetesVersion %q", c.Cluster.Spec.KubernetesVersion)
	}
	c.kubernetesVersion = *k8sVersion

	if c.InstanceGroup == nil {
		klog.Warningf("cannot determine role, InstanceGroup not set")
	} else if c.InstanceGroup.Spec.Role == kops.InstanceGroupRoleMaster {
		c.IsMaster = true
	}

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
	case distros.DistributionFlatcar:
		paths = append(paths, "/usr/share/ca-certificates")
	case distros.DistributionContainerOS:
		paths = append(paths, "/usr/share/ca-certificates")
	default:
		paths = append(paths, "/usr/share/ssl", "/usr/ssl", "/usr/lib/ssl", "/usr/local/openssl", "/var/ssl", "/etc/openssl")
	}

	return paths
}

// VolumesServiceName is the name of the service which is downstream of any volume mounts
func (c *NodeupModelContext) VolumesServiceName() string {
	return c.EnsureSystemdSuffix("kops-volume-mounts")
}

// EnsureSystemdSuffix ensures that the hook name ends with a valid systemd unit file extension. If it
// doesn't, it adds ".service" for backwards-compatibility with older versions of Kops
func (c *NodeupModelContext) EnsureSystemdSuffix(name string) string {
	if !systemd.UnitFileExtensionValid(name) {
		name += ".service"
	}

	return name
}

// EnsureDirectory ensures the directory exists or creates it
func (c *NodeupModelContext) EnsureDirectory(path string) error {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0755)
		}

		return err
	}

	if !st.IsDir() {
		return fmt.Errorf("path: %s already exists but is not a directory", path)
	}

	return nil
}

// IsMounted checks if the device is mount
func (c *NodeupModelContext) IsMounted(m mount.Interface, device, path string) (bool, error) {
	list, err := m.List()
	if err != nil {
		return false, err
	}

	for _, x := range list {
		if x.Device == device {
			klog.V(3).Infof("Found mountpoint device: %s, path: %s, type: %s", x.Device, x.Path, x.Type)
			if strings.TrimSuffix(x.Path, "/") == strings.TrimSuffix(path, "/") {
				return true, nil
			}
		}
	}

	return false, nil
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

// PathSrvSshproxy returns the path for the SSH proxy
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

// KubeletBootstrapKubeconfig is the path the bootstrap config file
func (c *NodeupModelContext) KubeletBootstrapKubeconfig() string {
	path := c.Cluster.Spec.Kubelet.BootstrapKubeconfig

	if c.IsMaster {
		if c.Cluster.Spec.MasterKubelet != nil && c.Cluster.Spec.MasterKubelet.BootstrapKubeconfig != "" {
			path = c.Cluster.Spec.MasterKubelet.BootstrapKubeconfig
		}
	}

	if path != "" {
		return path
	}

	return "/var/lib/kubelet/bootstrap-kubeconfig"
}

// KubeletKubeConfig is the path of the kubelet kubeconfig file
func (c *NodeupModelContext) KubeletKubeConfig() string {
	return "/var/lib/kubelet/kubeconfig"
}

// CNIConfDir returns the CNI directory
func (c *NodeupModelContext) CNIConfDir() string {
	return "/etc/cni/net.d/"
}

// BuildPKIKubeconfig generates a kubeconfig
func (c *NodeupModelContext) BuildPKIKubeconfig(name string) (string, error) {
	ca, err := c.FindCert(fi.CertificateId_CA)
	if err != nil {
		return "", err
	}

	cert, err := c.FindCert(name)
	if err != nil {
		return "", err
	}

	key, err := c.FindPrivateKey(name)
	if err != nil {
		return "", err
	}

	return c.BuildKubeConfig(name, ca, cert, key)
}

// BuildKubeConfig is responsible for building a kubeconfig
func (c *NodeupModelContext) BuildKubeConfig(username string, ca, certificate, privateKey []byte) (string, error) {
	user := kubeconfig.KubectlUser{
		ClientCertificateData: certificate,
		ClientKeyData:         privateKey,
	}
	cluster := kubeconfig.KubectlCluster{
		CertificateAuthorityData: ca,
	}

	if c.IsMaster {
		if c.IsKubernetesGTE("1.6") {
			// @note: use https >= 1.6m even for local connections, so we can turn off the insecure port
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
				Name: username,
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
					User:    username,
				},
			},
		},
		CurrentContext: "service-account-context",
	}

	yaml, err := kops.ToRawYaml(config)
	if err != nil {
		return "", fmt.Errorf("error marshaling kubeconfig to yaml: %v", err)
	}

	return string(yaml), nil
}

// IsKubernetesGTE checks if the version is greater-than-or-equal
func (c *NodeupModelContext) IsKubernetesGTE(version string) bool {
	if c.kubernetesVersion.Major == 0 {
		klog.Fatalf("kubernetesVersion not set (%s); Init not called", c.kubernetesVersion)
	}
	return util.IsKubernetesGTE(version, c.kubernetesVersion)
}

// UseEtcdManager checks if the etcd cluster has etcd-manager enabled
func (c *NodeupModelContext) UseEtcdManager() bool {
	for _, x := range c.Cluster.Spec.EtcdClusters {
		if x.Provider == kops.EtcdProviderTypeManager {
			return true
		}
	}

	return false
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

// UseVolumeMounts is used to check if we have volume mounts enabled as we need to
// insert requires and afters in various places
func (c *NodeupModelContext) UseVolumeMounts() bool {
	if c.InstanceGroup != nil {
		return len(c.InstanceGroup.Spec.VolumeMounts) > 0
	}

	return false
}

// UseEtcdTLSAuth checks the peer-auth is set in both cluster
// @NOTE: in retrospect i think we should have consolidated the common config in the wrapper struct; it
// feels weird we set things like version, tls etc per cluster since they both have to be the same.
func (c *NodeupModelContext) UseEtcdTLSAuth() bool {
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

// UseNodeAuthorization checks if have a node authorization policy
func (c *NodeupModelContext) UseNodeAuthorization() bool {
	return c.Cluster.Spec.NodeAuthorization != nil
}

// UseNodeAuthorizer checks if node authorization is enabled
func (c *NodeupModelContext) UseNodeAuthorizer() bool {
	if !c.UseNodeAuthorization() || !c.UseBootstrapTokens() {
		return false
	}

	return c.Cluster.Spec.NodeAuthorization.NodeAuthorizer != nil
}

// UsesSecondaryIP checks if the CNI in use attaches secondary interfaces to the host.
func (c *NodeupModelContext) UsesSecondaryIP() bool {
	if (c.Cluster.Spec.Networking.CNI != nil && c.Cluster.Spec.Networking.CNI.UsesSecondaryIP) || c.Cluster.Spec.Networking.AmazonVPC != nil || c.Cluster.Spec.Networking.LyftVPC != nil {
		return true
	}

	return false
}

// UseBootstrapTokens checks if we are using bootstrap tokens
func (c *NodeupModelContext) UseBootstrapTokens() bool {
	if c.IsMaster {
		return fi.BoolValue(c.Cluster.Spec.KubeAPIServer.EnableBootstrapAuthToken)
	}

	return c.Cluster.Spec.Kubelet != nil && c.Cluster.Spec.Kubelet.BootstrapKubeconfig != ""
}

// UseSecureKubelet checks if the kubelet api should be protected by a client certificate. Note: the settings are
// in one of three section, master specific kubelet, cluster wide kubelet or the InstanceGroup. Though arguably is
// doesn't make much sense to unset this on a per InstanceGroup level, but hey :)
func (c *NodeupModelContext) UseSecureKubelet() bool {
	cluster := &c.Cluster.Spec // just to shorten the typing
	group := &c.InstanceGroup.Spec

	// @check on the InstanceGroup itself
	if group.Kubelet != nil && group.Kubelet.AnonymousAuth != nil && !*group.Kubelet.AnonymousAuth {
		return true
	}

	// @check if we have anything specific to master kubelet
	if c.IsMaster {
		if cluster.MasterKubelet != nil && cluster.MasterKubelet.AnonymousAuth != nil && !*cluster.MasterKubelet.AnonymousAuth {
			return true
		}
	}

	// @check the default settings for master and kubelet
	if cluster.Kubelet != nil && cluster.Kubelet.AnonymousAuth != nil && !*cluster.Kubelet.AnonymousAuth {
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
	if c.Distribution == distros.DistributionFlatcar {
		kubeletCommand = "/opt/bin"
	}
	if c.Distribution == distros.DistributionContainerOS {
		kubeletCommand = "/home/kubernetes/bin"
	}
	return kubeletCommand
}

// BuildCertificatePairTask creates the tasks to pull down the certificate and private key
func (c *NodeupModelContext) BuildCertificatePairTask(ctx *fi.ModelBuilderContext, key, path, filename string) error {
	certificateName := filepath.Join(path, filename+".pem")
	keyName := filepath.Join(path, filename+"-key.pem")

	if err := c.BuildCertificateTask(ctx, key, certificateName); err != nil {
		return err
	}

	return c.BuildPrivateKeyTask(ctx, key, keyName)
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

	p := filename
	if !filepath.IsAbs(p) {
		p = filepath.Join(c.PathSrvKubernetes(), filename)
	}

	ctx.AddTask(&nodetasks.File{
		Path:     p,
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	})

	return nil
}

// BuildPrivateKeyTask is responsible for build a certificate request task
func (c *NodeupModelContext) BuildPrivateKeyTask(ctx *fi.ModelBuilderContext, name, filename string) error {
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

	p := filename
	if !filepath.IsAbs(p) {
		p = filepath.Join(c.PathSrvKubernetes(), filename)
	}

	ctx.AddTask(&nodetasks.File{
		Path:     p,
		Contents: fi.NewStringResource(serialized),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
	})

	return nil
}

// NodeName returns the name of the local Node, as it will be created in k8s
func (c *NodeupModelContext) NodeName() (string, error) {
	// This mirrors nodeutil.GetHostName
	hostnameOverride := c.Cluster.Spec.Kubelet.HostnameOverride

	if c.IsMaster && c.Cluster.Spec.MasterKubelet.HostnameOverride != "" {
		hostnameOverride = c.Cluster.Spec.MasterKubelet.HostnameOverride
	}

	nodeName, err := EvaluateHostnameOverride(hostnameOverride)
	if err != nil {
		return "", fmt.Errorf("error evaluating hostname: %v", err)
	}

	if nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Couldn't determine hostname: %v", err)
		}
		nodeName = hostname
	}

	return strings.ToLower(strings.TrimSpace(nodeName)), nil
}

// EvaluateHostnameOverride returns the hostname after replacing some well-known placeholders
func EvaluateHostnameOverride(hostnameOverride string) (string, error) {
	if hostnameOverride == "" || hostnameOverride == "@hostname" {
		return "", nil
	}
	k := strings.TrimSpace(hostnameOverride)
	k = strings.ToLower(k)

	if k != "@aws" {
		return hostnameOverride, nil
	}

	// We recognize @aws as meaning "the private DNS name from AWS", to generate this we need to get a few pieces of information
	azBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/placement/availability-zone")
	if err != nil {
		return "", fmt.Errorf("error reading availability zone from AWS metadata: %v", err)
	}

	instanceIDBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/instance-id")
	if err != nil {
		return "", fmt.Errorf("error reading instance-id from AWS metadata: %v", err)
	}
	instanceID := string(instanceIDBytes)

	config := aws.NewConfig()
	config = config.WithCredentialsChainVerboseErrors(true)

	s, err := session.NewSession(config)
	if err != nil {
		return "", fmt.Errorf("error starting new AWS session: %v", err)
	}

	svc := ec2.New(s, config.WithRegion(string(azBytes[:len(azBytes)-1])))

	result, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: []*string{&instanceID},
	})
	if err != nil {
		return "", fmt.Errorf("error describing instances: %v", err)
	}

	if len(result.Reservations) != 1 {
		return "", fmt.Errorf("Too many reservations returned for the single instance-id")
	}

	if len(result.Reservations[0].Instances) != 1 {
		return "", fmt.Errorf("Too many instances returned for the single instance-id")
	}
	return *(result.Reservations[0].Instances[0].PrivateDnsName), nil
}

// FindCert is a helper method to retrieving a certificate from the store
func (c *NodeupModelContext) FindCert(name string) ([]byte, error) {
	cert, err := c.KeyStore.FindCert(name)
	if err != nil {
		return []byte{}, fmt.Errorf("error fetching certificate: %v from keystore: %v", name, err)
	}
	if cert == nil {
		return []byte{}, fmt.Errorf("unable to found certificate: %s", name)
	}

	return cert.AsBytes()
}

// FindPrivateKey is a helper method to retrieving a private key from the store
func (c *NodeupModelContext) FindPrivateKey(name string) ([]byte, error) {
	key, err := c.KeyStore.FindPrivateKey(name)
	if err != nil {
		return []byte{}, fmt.Errorf("error fetching private key: %v from keystore: %v", name, err)
	}
	if key == nil {
		return []byte{}, fmt.Errorf("unable to found private key: %s", name)
	}

	return key.AsBytes()
}

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
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/blang/semver/v4"
	hcloudmetadata "github.com/hetznercloud/hcloud-go/hcloud/metadata"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/kopscontrollerclient"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcediscovery"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm/gcetpmsigner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/mount-utils"
	"sigs.k8s.io/yaml"
)

const (
	ConfigurationModeWarming string = "Warming"
)

// NodeupModelContext is the context supplied the nodeup tasks
type NodeupModelContext struct {
	Cloud        fi.Cloud
	Architecture architectures.Architecture
	GPUVendor    architectures.GPUVendor
	Assets       *fi.AssetStore
	// Deprecated: Fields should be accessed from NodeupConfig or BootConfig.
	Cluster      *kops.Cluster
	ConfigBase   vfs.Path
	Distribution distributions.Distribution
	KeyStore     fi.KeystoreReader
	BootConfig   *nodeup.BootConfig
	NodeupConfig *nodeup.Config
	SecretStore  fi.SecretStoreReader

	// IsMaster is true if the InstanceGroup has a role of master (populated by Init)
	IsMaster bool

	// HasAPIServer is true if the InstanceGroup has a role of master or apiserver (pupulated by Init)
	HasAPIServer bool

	// IsGossip is true if the cluster runs Gossip DNS.
	IsGossip bool

	kubernetesVersion   semver.Version
	bootstrapCerts      map[string]*nodetasks.BootstrapCert
	bootstrapKeypairIDs map[string]string

	// ConfigurationMode determines if we are prewarming an instance or running it live
	ConfigurationMode string
	InstanceID        string
	MachineType       string

	CloudProvider kops.CloudProviderID

	KopsControllerClient *kopscontrollerclient.Client
}

// Init completes initialization of the object, for example pre-parsing the kubernetes version
func (c *NodeupModelContext) Init(ctx context.Context) error {
	k8sVersion, err := util.ParseKubernetesVersion(c.NodeupConfig.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return fmt.Errorf("unable to parse KubernetesVersion %q", c.NodeupConfig.KubernetesVersion)
	}
	c.kubernetesVersion = *k8sVersion
	c.bootstrapCerts = map[string]*nodetasks.BootstrapCert{}
	c.bootstrapKeypairIDs = map[string]string{}

	role := c.BootConfig.InstanceGroupRole

	if role == kops.InstanceGroupRoleControlPlane {
		c.IsMaster = true
	}

	if role == kops.InstanceGroupRoleControlPlane || role == kops.InstanceGroupRoleAPIServer {
		c.HasAPIServer = true
	}

	if !c.Cluster.UsesNoneDNS() && dns.IsGossipClusterName(c.NodeupConfig.ClusterName) {
		c.IsGossip = true
	}

	if !c.IsMaster && c.UseKopsControllerForNodeBootstrap() {
		if c.KopsControllerClient == nil {
			baseURL := url.URL{
				Scheme: "https",
				Host:   net.JoinHostPort("kops-controller.internal."+c.Cluster.ObjectMeta.Name, strconv.Itoa(wellknownports.KopsControllerPort)),
				Path:   "/",
			}
			caBundle := []byte(c.NodeupConfig.CAs[fi.CertificateIDCA])

			client, err := BuildKopsControllerClient(ctx, baseURL, caBundle, c.CloudProvider, c.Cloud.Region())
			if err != nil {
				return fmt.Errorf("failed to build kops-controller client: %w", err)
			}
			c.KopsControllerClient = client
		}
	}

	return nil
}

func (c *NodeupModelContext) APIInternalName() string {
	return "api.internal." + c.NodeupConfig.ClusterName
}

func (c *NodeupModelContext) IsIPv6Only() bool {
	return utils.IsIPv6CIDR(c.NodeupConfig.Networking.NonMasqueradeCIDR)
}

func (c *NodeupModelContext) IsKopsControllerIPAM() bool {
	return c.IsIPv6Only()
}

func BuildKopsControllerClient(ctx context.Context, baseURL url.URL, caBundle []byte, cloudProvider kops.CloudProviderID, region string) (*kopscontrollerclient.Client, error) {
	var authenticator bootstrap.Authenticator
	var resolver resolver.Resolver

	switch cloudProvider {
	case kops.CloudProviderAWS:
		a, err := awsup.NewAWSAuthenticator(region)
		if err != nil {
			return nil, err
		}
		authenticator = a
	case kops.CloudProviderGCE:
		a, err := gcetpmsigner.NewTPMAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a

		discovery, err := gcediscovery.New()
		if err != nil {
			return nil, err
		}
		resolver = discovery
	case kops.CloudProviderHetzner:
		a, err := hetzner.NewHetznerAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a

	case kops.CloudProviderOpenstack:
	a, err := openstack.NewOpenstackAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a
		
	default:
		return nil, fmt.Errorf("unsupported cloud provider for node configuration %s", cloudProvider)
	}



	default:
		return fmt.Errorf("unsupported cloud provider for authenticator %q", b.BootConfig.CloudProvider)
	}

	if err != nil {
		return err
	}

	baseURL := url.URL{
		Scheme: "https",
		Host:   net.JoinHostPort("kops-controller.internal."+b.NodeupConfig.ClusterName, strconv.Itoa(wellknownports.KopsControllerPort)),
		Path:   "/",
	}

	bootstrapClient := &kopscontrollerclient.Client{
		Authenticator: authenticator,
		CAs:           []byte(b.NodeupConfig.CAs[fi.CertificateIDCA]),
		BaseURL:       baseURL,
=======


	client := &kopscontrollerclient.Client{
		Authenticator: authenticator,
		CAs:           caBundle,
		BaseURL:       baseURL,
		Resolver:      resolver,
	}

	return client, nil
}

// SSLHostPaths returns the TLS paths for the distribution
func (c *NodeupModelContext) SSLHostPaths() []string {
	paths := []string{"/etc/ssl", "/etc/pki/tls", "/etc/pki/ca-trust"}

	switch c.Distribution {
	case distributions.DistributionFlatcar:
		// Because /usr is read-only on Flatcar, we can't have any new directories; docker will try (and fail) to create them
		// TODO: Just check if the directories exist?
		paths = append(paths, "/usr/share/ca-certificates")
	case distributions.DistributionContainerOS:
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
			return os.MkdirAll(path, 0o755)
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
	case distributions.DistributionContainerOS:
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
	case distributions.DistributionContainerOS:
		return "/etc/srv/sshproxy"
	default:
		return "/srv/sshproxy"
	}
}

// KubeletBootstrapKubeconfig is the path the bootstrap config file
func (c *NodeupModelContext) KubeletBootstrapKubeconfig() string {
	path := c.NodeupConfig.KubeletConfig.BootstrapKubeconfig

	if path != "" {
		return path
	}

	return "/var/lib/kubelet/bootstrap-kubeconfig"
}

// KubeletKubeConfig is the path of the kubelet kubeconfig file
func (c *NodeupModelContext) KubeletKubeConfig() string {
	return "/var/lib/kubelet/kubeconfig"
}

// BuildIssuedKubeconfig generates a kubeconfig with a locally issued client certificate.
func (c *NodeupModelContext) BuildIssuedKubeconfig(name string, subject nodetasks.PKIXName, ctx *fi.NodeupModelBuilderContext) *fi.NodeupTaskDependentResource {
	issueCert := &nodetasks.IssueCert{
		Name:      name,
		Signer:    fi.CertificateIDCA,
		KeypairID: c.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
		Type:      "client",
		Subject:   subject,
	}
	ctx.AddTask(issueCert)
	certResource, keyResource, caResource := issueCert.GetResources()

	kubeConfig := &nodetasks.KubeConfig{
		Name: name,
		Cert: certResource,
		Key:  keyResource,
		CA:   caResource,
	}
	if c.HasAPIServer {
		// @note: use https even for local connections, so we can turn off the insecure port
		kubeConfig.ServerURL = "https://127.0.0.1"
	} else {
		kubeConfig.ServerURL = "https://" + c.APIInternalName()
	}
	ctx.AddTask(kubeConfig)
	return kubeConfig.GetConfig()
}

// GetBootstrapCert requests a certificate keypair from kops-controller.
func (c *NodeupModelContext) GetBootstrapCert(name string, signer string) (cert, key fi.Resource, err error) {
	if c.IsMaster {
		panic("control plane nodes can't get certs from kops-controller")
	}
	b, ok := c.bootstrapCerts[name]
	if !ok {
		b = &nodetasks.BootstrapCert{
			Cert: &fi.NodeupTaskDependentResource{},
			Key:  &fi.NodeupTaskDependentResource{},
		}
		c.bootstrapCerts[name] = b
	}
	c.bootstrapKeypairIDs[signer] = c.NodeupConfig.KeypairIDs[signer]
	if c.bootstrapKeypairIDs[signer] == "" {
		return nil, nil, fmt.Errorf("no keypairID for %q", signer)
	}
	return b.Cert, b.Key, nil
}

// BuildBootstrapKubeconfig generates a kubeconfig with a client certificate from either kops-controller or the state store.
func (c *NodeupModelContext) BuildBootstrapKubeconfig(name string, ctx *fi.NodeupModelBuilderContext) (fi.Resource, error) {
	if c.UseKopsControllerForNodeBootstrap() {
		cert, key, err := c.GetBootstrapCert(name, fi.CertificateIDCA)
		if err != nil {
			return nil, err
		}

		kubeConfig := &nodetasks.KubeConfig{
			Name: name,
			Cert: cert,
			Key:  key,
			CA:   fi.NewStringResource(c.NodeupConfig.CAs[fi.CertificateIDCA]),
		}
		if c.HasAPIServer {
			// @note: use https even for local connections, so we can turn off the insecure port
			kubeConfig.ServerURL = "https://127.0.0.1"
		} else {
			kubeConfig.ServerURL = "https://" + c.APIInternalName()
		}

		ctx.EnsureTask(kubeConfig)

		return kubeConfig.GetConfig(), nil
	} else {
		keyset, err := c.KeyStore.FindKeyset(ctx.Context(), name)
		if err != nil {
			return nil, fmt.Errorf("error fetching keyset %q from keystore: %w", name, err)
		}
		if keyset == nil {
			return nil, fmt.Errorf("keyset %q not found", name)
		}

		keypairID := c.NodeupConfig.KeypairIDs[name]
		if keypairID == "" {
			return nil, fmt.Errorf("keypairID for %s missing from NodeupConfig", name)
		}
		item := keyset.Items[keypairID]
		if item == nil {
			return nil, fmt.Errorf("keypairID %s missing from %s keyset", keypairID, name)
		}

		cert, err := item.Certificate.AsBytes()
		if err != nil {
			return nil, err
		}

		key, err := item.PrivateKey.AsBytes()
		if err != nil {
			return nil, err
		}

		kubeConfig := &nodetasks.KubeConfig{
			Name: name,
			Cert: fi.NewBytesResource(cert),
			Key:  fi.NewBytesResource(key),
			CA:   fi.NewStringResource(c.NodeupConfig.CAs[fi.CertificateIDCA]),
		}
		if c.HasAPIServer {
			// @note: use https even for local connections, so we can turn off the insecure port
			// This code path is used for the kubelet cert in Kubernetes 1.18 and earlier.
			kubeConfig.ServerURL = "https://127.0.0.1"
		} else {
			kubeConfig.ServerURL = "https://" + c.APIInternalName()
		}

		err = kubeConfig.Run(nil)
		if err != nil {
			return nil, err
		}

		config, err := fi.ResourceAsBytes(kubeConfig.GetConfig())
		if err != nil {
			return nil, err
		}

		return fi.NewBytesResource(config), nil
	}
}

// RemapImage applies any needed remapping to an image reference.
func (c *NodeupModelContext) RemapImage(image string) string {
	if components.IsBaseURL(c.NodeupConfig.KubernetesVersion) && c.IsKubernetesLT("1.25") {
		image = strings.Replace(image, "registry.k8s.io", "k8s.gcr.io", 1)
	}
	if c.Architecture != architectures.ArchitectureAmd64 {
		image = strings.Replace(image, "-amd64", "-"+string(c.Architecture), 1)
	}
	return image
}

// IsKubernetesGTE checks if the version is greater-than-or-equal
func (c *NodeupModelContext) IsKubernetesGTE(version string) bool {
	if c.kubernetesVersion.Major == 0 {
		klog.Fatalf("kubernetesVersion not set (%s); Init not called", c.kubernetesVersion)
	}
	return util.IsKubernetesGTE(version, c.kubernetesVersion)
}

// IsKubernetesLT checks if the version is less-than
func (c *NodeupModelContext) IsKubernetesLT(version string) bool {
	if c.kubernetesVersion.Major == 0 {
		klog.Fatalf("kubernetesVersion not set (%s); Init not called", c.kubernetesVersion)
	}
	return !c.IsKubernetesGTE(version)
}

// UseVolumeMounts is used to check if we have volume mounts enabled as we need to
// insert requires and afters in various places
func (c *NodeupModelContext) UseVolumeMounts() bool {
	return len(c.NodeupConfig.VolumeMounts) > 0
}

// UseKopsControllerForNodeBootstrap checks if nodeup should use kops-controller to bootstrap.
func (c *NodeupModelContext) UseKopsControllerForNodeBootstrap() bool {
	return model.UseKopsControllerForNodeBootstrap(c.Cluster)
}

// UsesSecondaryIP checks if the CNI in use attaches secondary interfaces to the host.
func (c *NodeupModelContext) UsesSecondaryIP() bool {
	return (c.NodeupConfig.Networking.CNI != nil && c.NodeupConfig.Networking.CNI.UsesSecondaryIP) ||
		c.NodeupConfig.Networking.AmazonVPC != nil ||
		(c.NodeupConfig.Networking.Cilium != nil && c.NodeupConfig.Networking.Cilium.IPAM == kops.CiliumIpamEni) ||
		c.BootConfig.CloudProvider == kops.CloudProviderHetzner
}

// UseBootstrapTokens checks if we are using bootstrap tokens
func (c *NodeupModelContext) UseBootstrapTokens() bool {
	if c.HasAPIServer {
		return fi.ValueOf(c.NodeupConfig.APIServerConfig.KubeAPIServer.EnableBootstrapAuthToken)
	}

	return c.NodeupConfig.KubeletConfig.BootstrapKubeconfig != ""
}

// KubectlPath returns distro based path for kubectl
func (c *NodeupModelContext) KubectlPath() string {
	kubeletCommand := "/usr/local/bin"
	if c.Distribution == distributions.DistributionFlatcar {
		kubeletCommand = "/opt/kops/bin"
	}
	if c.Distribution == distributions.DistributionContainerOS {
		kubeletCommand = "/home/kubernetes/bin"
	}
	return kubeletCommand
}

// BuildCertificatePairTask creates the tasks to create the certificate and private key files.
func (c *NodeupModelContext) BuildCertificatePairTask(ctx *fi.NodeupModelBuilderContext, name, path, filename string, owner *string, beforeServices []string) error {
	return c.buildCertificatePairTask(ctx, name, path, filename, owner, beforeServices, true)
}

// BuildPrivateKeyTask builds a task to create the private key file.
func (c *NodeupModelContext) BuildPrivateKeyTask(ctx *fi.NodeupModelBuilderContext, name, path, filename string, owner *string, beforeServices []string) error {
	return c.buildCertificatePairTask(ctx, name, path, filename, owner, beforeServices, false)
}

func (c *NodeupModelContext) buildCertificatePairTask(ctx *fi.NodeupModelBuilderContext, name, path, filename string, owner *string, beforeServices []string, includeCert bool) error {
	p := filepath.Join(path, filename)
	if !filepath.IsAbs(p) {
		p = filepath.Join(c.PathSrvKubernetes(), p)
	}

	// We use the keypair ID passed in nodeup.Config instead of the primary
	// keypair so that the node will be updated when the primary keypair does
	// not match the one that we are using.
	keypairID := c.NodeupConfig.KeypairIDs[name]
	if keypairID == "" {
		// kOps bug where KeypairID was not populated for the node role.
		return fmt.Errorf("no keypair ID for %q", name)
	}

	keyset, err := c.KeyStore.FindKeyset(ctx.Context(), name)
	if err != nil {
		return err
	}
	if keyset == nil {
		return fmt.Errorf("keyset %q not found", name)
	}

	item := keyset.Items[keypairID]
	if item == nil {
		return fmt.Errorf("did not find keypair %s for %s", keypairID, name)
	}

	if includeCert {
		certificate := item.Certificate
		if certificate == nil {
			return fmt.Errorf("certificate %q not found", name)
		}

		cert, err := certificate.AsString()
		if err != nil {
			return err
		}

		ctx.AddTask(&nodetasks.File{
			Path:           p + ".crt",
			Contents:       fi.NewStringResource(cert),
			Type:           nodetasks.FileType_File,
			Mode:           s("0600"),
			Owner:          owner,
			BeforeServices: beforeServices,
		})
	}

	privateKey := item.PrivateKey
	if privateKey == nil {
		return fmt.Errorf("private key %q not found", name)
	}

	key, err := privateKey.AsString()
	if err != nil {
		return err
	}

	ctx.AddTask(&nodetasks.File{
		Path:     p + ".key",
		Contents: fi.NewStringResource(key),
		Type:     nodetasks.FileType_File,
		Mode:     s("0600"),
		Owner:    owner,
	})

	return nil
}

// BuildCertificateTask builds a task to create a certificate file.
func (c *NodeupModelContext) BuildCertificateTask(ctx *fi.NodeupModelBuilderContext, name, filename string, owner *string) error {
	keyset, err := c.KeyStore.FindKeyset(ctx.Context(), name)
	if err != nil {
		return err
	}

	if keyset == nil {
		return fmt.Errorf("keyset %q not found", name)
	}

	serialized, err := keyset.Primary.Certificate.AsString()
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
		Owner:    owner,
	})

	return nil
}

// BuildLegacyPrivateKeyTask builds a task to create a private key file.
func (c *NodeupModelContext) BuildLegacyPrivateKeyTask(ctx *fi.NodeupModelBuilderContext, name, filename string, owner *string) error {
	keyset, err := c.KeyStore.FindKeyset(ctx.Context(), name)
	if err != nil {
		return err
	}

	if keyset == nil {
		return fmt.Errorf("keyset %q not found", name)
	}

	serialized, err := keyset.Primary.PrivateKey.AsString()
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
		Owner:    owner,
	})

	return nil
}

// NodeName returns the name of the local Node, as it will be created in k8s
func (c *NodeupModelContext) NodeName() (string, error) {
	// This mirrors nodeutil.GetHostName
	nodeName := c.NodeupConfig.KubeletConfig.HostnameOverride

	if nodeName == "" {
		hostname, err := os.Hostname()
		if err != nil {
			klog.Fatalf("Couldn't determine hostname: %v", err)
		}
		nodeName = hostname
	}

	return strings.ToLower(strings.TrimSpace(nodeName)), nil
}

func (b *NodeupModelContext) AddCNIBinAssets(c *fi.NodeupModelBuilderContext, assetNames []string) error {
	for _, assetName := range assetNames {
		re, err := regexp.Compile(fmt.Sprintf("^%s$", regexp.QuoteMeta(assetName)))
		if err != nil {
			return err
		}
		if err := b.addCNIBinAsset(c, re); err != nil {
			return err
		}
	}
	return nil
}

func (b *NodeupModelContext) addCNIBinAsset(c *fi.NodeupModelBuilderContext, assetPath *regexp.Regexp) error {
	name, res, err := b.Assets.FindMatch(assetPath)
	if err != nil {
		return err
	}

	c.AddTask(&nodetasks.File{
		Path:     filepath.Join(b.CNIBinDir(), name),
		Contents: res,
		Type:     nodetasks.FileType_File,
		Mode:     fi.PtrTo("0755"),
	})

	return nil
}

// CNIBinDir returns the path for the CNI binaries
func (c *NodeupModelContext) CNIBinDir() string {
	// We used to map this on a per-distro basis, but this can require CNI manifests to be distro aware
	return "/opt/cni/bin/"
}

// CNIConfDir returns the CNI directory
func (c *NodeupModelContext) CNIConfDir() string {
	return "/etc/cni/net.d/"
}

func (c *NodeupModelContext) InstallNvidiaRuntime() bool {
	return c.NodeupConfig.NvidiaGPU != nil &&
		fi.ValueOf(c.NodeupConfig.NvidiaGPU.Enabled) &&
		c.GPUVendor == architectures.GPUVendorNvidia
}

// RunningOnGCE returns true if we are running on GCE
func (c *NodeupModelContext) RunningOnGCE() bool {
	return c.BootConfig.CloudProvider == kops.CloudProviderGCE
}

// RunningOnAzure returns true if we are running on Azure
func (c *NodeupModelContext) RunningOnAzure() bool {
	return c.BootConfig.CloudProvider == kops.CloudProviderAzure
}

// GetMetadataLocalIP returns the local IP address read from metadata
func (c *NodeupModelContext) GetMetadataLocalIP() (string, error) {
	var internalIP string

	switch c.BootConfig.CloudProvider {
	case kops.CloudProviderAWS:
		sess := session.Must(session.NewSession())
		metadata := ec2metadata.New(sess)
		localIPv4, err := metadata.GetMetadata("local-ipv4")
		if err != nil {
			return "", fmt.Errorf("failed to get local-ipv4 address from ec2 metadata: %w", err)
		}
		internalIP = localIPv4

	case kops.CloudProviderHetzner:
		client := hcloudmetadata.NewClient()
		privateNetworksYaml, err := client.PrivateNetworks()
		if err != nil {
			return "", fmt.Errorf("failed to get private networks from hetzner cloud metadata: %w", err)
		}
		var privateNetworks []struct {
			IP           net.IP   `json:"ip"`
			AliasIPs     []net.IP `json:"alias_ips"`
			InterfaceNum int      `json:"interface_num"`
			MACAddress   string   `json:"mac_address"`
			NetworkID    int      `json:"network_id"`
			NetworkName  string   `json:"network_name"`
			Network      string   `json:"network"`
			Subnet       string   `json:"subnet"`
			Gateway      net.IP   `json:"gateway"`
		}
		err = yaml.Unmarshal([]byte(privateNetworksYaml), &privateNetworks)
		if err != nil {
			return "", fmt.Errorf("failed to convert private networks to object: %w", err)
		}
		for _, privateNetwork := range privateNetworks {
			if privateNetwork.InterfaceNum == 1 {
				internalIP = privateNetwork.IP.String()
			}
		}

	default:
		return "", fmt.Errorf("getting local IP from metadata is not supported for cloud provider: %q", c.BootConfig.CloudProvider)
	}

	return internalIP, nil
}

func (c *NodeupModelContext) findStaticManifest(key string) *nodeup.StaticManifest {
	if c == nil || c.NodeupConfig == nil {
		return nil
	}
	for _, manifest := range c.NodeupConfig.StaticManifests {
		if manifest.Key == key {
			return manifest
		}
	}
	return nil
}

func (c *NodeupModelContext) findFileAsset(path string) *kops.FileAssetSpec {
	if c == nil || c.NodeupConfig == nil {
		return nil
	}
	for i := range c.NodeupConfig.FileAssets {
		f := &c.NodeupConfig.FileAssets[i]
		if f.Path == path {
			return f
		}
	}
	return nil
}

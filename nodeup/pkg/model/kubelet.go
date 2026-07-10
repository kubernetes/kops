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
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	logsapi "k8s.io/component-base/logs/api/v1"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	kopsutil "k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/flagbuilder"
	"k8s.io/kops/pkg/rbac"
	"k8s.io/kops/pkg/systemd"
	"k8s.io/kops/upup/pkg/fi"
	azurecloud "k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/util/pkg/distributions"
	kubeletv1 "k8s.io/kubelet/config/v1"
	kubelet "k8s.io/kubelet/config/v1beta1"
)

const (
	// kubeletService is the name of the kubelet service
	kubeletService = "kubelet.service"

	kubeletConfigFilePath            = "/var/lib/kubelet/kubelet.conf"
	credentialProviderConfigFilePath = "/var/lib/kubelet/credential-provider.conf" //nolint:gosec // This is a config file path, not a credential.
)

// Scheme registration is shared across all calls in this file because the
// schemes are stateless after AddToScheme runs. Constructing them per-call
// allocates a populated scheme map and walks the scheme registry on every
// nodeup invocation.
var (
	kubeletV1Beta1Encoder runtime.Encoder
	kubeletV1Encoder      runtime.Encoder
)

func init() {
	v1beta1Scheme := runtime.NewScheme()
	utilruntime.Must(kubelet.AddToScheme(v1beta1Scheme))
	kubeletV1Beta1Encoder = mustYAMLEncoder(v1beta1Scheme, kubelet.SchemeGroupVersion)

	v1Scheme := runtime.NewScheme()
	utilruntime.Must(kubeletv1.AddToScheme(v1Scheme))
	kubeletV1Encoder = mustYAMLEncoder(v1Scheme, kubeletv1.SchemeGroupVersion)
}

// mustYAMLEncoder returns a YAML encoder for the given scheme + GV, panicking
// if the scheme has no YAML serializer registered (which would be a programmer
// error since the kubelet types we register always have one).
func mustYAMLEncoder(scheme *runtime.Scheme, gv runtime.GroupVersioner) runtime.Encoder {
	codecFactory := serializer.NewCodecFactory(scheme)
	info, ok := runtime.SerializerInfoForMediaType(codecFactory.SupportedMediaTypes(), "application/yaml")
	if !ok {
		panic(fmt.Sprintf("no YAML serializer registered for kubelet scheme %v", gv))
	}
	return codecFactory.EncoderForVersion(info.Serializer, gv)
}

// KubeletBuilder installs kubelet
type KubeletBuilder struct {
	*NodeupModelContext
}

var _ fi.NodeupModelBuilder = &KubeletBuilder{}

// Build is responsible for building the kubelet configuration
func (b *KubeletBuilder) Build(c *fi.NodeupModelBuilderContext) error {
	err := b.buildKubeletServingCertificate(c)
	if err != nil {
		return fmt.Errorf("error building kubelet server cert: %v", err)
	}

	ctx := c.Context()
	kubeletConfig, err := b.buildKubeletConfigSpec(ctx)
	if err != nil {
		return fmt.Errorf("error building kubelet config: %v", err)
	}

	{
		// Set the provider ID to help speed node registration on large clusters
		var providerID string
		if b.CloudProvider() == kops.CloudProviderAWS {
			config, err := awsconfig.LoadDefaultConfig(ctx)
			if err != nil {
				return fmt.Errorf("error loading AWS config: %v", err)
			}
			metadata := imds.NewFromConfig(config)
			instanceIdentity, err := metadata.GetInstanceIdentityDocument(ctx, &imds.GetInstanceIdentityDocumentInput{})
			if err != nil {
				return err
			}
			providerID = fmt.Sprintf("aws:///%s/%s", instanceIdentity.AvailabilityZone, instanceIdentity.InstanceID)
		} else if b.CloudProvider() == kops.CloudProviderAzure {
			metadata, err := azurecloud.QueryComputeInstanceMetadata(ctx)
			if err != nil {
				return fmt.Errorf("error querying Azure instance metadata: %v", err)
			}
			providerID = "azure://" + metadata.ResourceID
		}

		t, err := b.buildKubeletComponentConfig(kubeletConfig, providerID)
		if err != nil {
			return err
		}

		c.AddTask(t)
	}

	{
		t, err := b.buildSystemdEnvironmentFile(c.Context(), kubeletConfig)
		if err != nil {
			return err
		}
		c.AddTask(t)
	}

	{
		// @TODO Extract to common function?
		assetName := "kubelet"
		assetPath := ""
		// @TODO make Find call to an interface, we cannot mock out this function because it finds a file on disk
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		c.AddTask(&nodetasks.File{
			Path:     b.kubeletPath(),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		})
	}
	{
		if kubeletConfig.PodManifestPath != "" {
			t, err := b.buildManifestDirectory(kubeletConfig)
			if err != nil {
				return err
			}
			c.EnsureTask(t)
		}
	}
	{
		// We always create the directory, avoids circular dependency on a bind-mount
		c.EnsureTask(&nodetasks.File{
			Path: filepath.Dir(b.KubeletKubeConfig()), // e.g. "/var/lib/kubelet"
			Type: nodetasks.FileType_Directory,
			Mode: s("0755"),
		})

		{
			var kubeconfig fi.Resource
			if b.HasAPIServer {
				kubeconfig, err = b.buildControlPlaneKubeletKubeconfig(c)
			} else {
				kubeconfig, err = b.BuildBootstrapKubeconfig("kubelet", c)
			}
			if err != nil {
				return err
			}

			c.AddTask(&nodetasks.File{
				Path:           b.KubeletKubeConfig(),
				Contents:       kubeconfig,
				Type:           nodetasks.FileType_File,
				Mode:           s("0400"),
				BeforeServices: []string{kubeletService},
			})
		}
	}

	if !b.NodeupConfig.UsesKubenet {
		c.AddTask(&nodetasks.File{
			Path: b.CNIConfDir(),
			Type: nodetasks.FileType_Directory,
		})
	}

	if b.UseExternalKubeletCredentialProvider() {
		switch b.CloudProvider() {
		case kops.CloudProviderGCE:
			if err := b.addGCPCredentialProvider(c); err != nil {
				return fmt.Errorf("failed to add the %s kubelet credential provider: %w", b.CloudProvider(), err)
			}
		case kops.CloudProviderAWS:
			if err := b.addECRCredentialProvider(c); err != nil {
				return fmt.Errorf("failed to add the %s kubelet credential provider: %w", b.CloudProvider(), err)
			}
		}
	}

	{
		cgroup := kubeletConfig.KubeletCgroups
		if cgroup != "" {
			c.EnsureTask(b.buildCgroupService(cgroup))
		}

	}
	{
		cgroup := kubeletConfig.RuntimeCgroups
		if cgroup != "" {
			c.EnsureTask(b.buildCgroupService(cgroup))
		}

	}
	/* Kubelet incorrectly interprets this value when CgroupDriver is systemd
	See https://github.com/kubernetes/kubernetes/issues/101189
	{
		cgroup := kubeletConfig.KubeReservedCgroup
		if cgroup != "" {
			c.EnsureTask(b.buildCgroupService(cgroup))
		}
	}
	*/

	{
		cgroup := kubeletConfig.SystemCgroups
		if cgroup != "" {
			c.EnsureTask(b.buildCgroupService(cgroup))
		}
	}

	/* This suffers from the same issue as KubeReservedCgroup
	{
		cgroup := kubeletConfig.SystemReservedCgroup
		if cgroup != "" {
			c.EnsureTask(b.buildCgroupService(cgroup))
		}
	}
	*/

	c.AddTask(b.buildSystemdService())

	return nil
}

func (b *KubeletBuilder) buildKubeletComponentConfig(kubeletConfig *kops.KubeletConfigSpec, providerID string) (*nodetasks.File, error) {
	componentConfig, err := b.kubeletConfiguration(kubeletConfig, providerID)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := kubeletV1Beta1Encoder.Encode(componentConfig, &buf); err != nil {
		return nil, fmt.Errorf("encoding kubelet component config: %w", err)
	}

	return &nodetasks.File{
		Path:           kubeletConfigFilePath,
		Contents:       fi.NewBytesResource(buf.Bytes()),
		Type:           nodetasks.FileType_File,
		BeforeServices: []string{kubeletService},
	}, nil
}

// kubeletConfiguration translates a kops.KubeletConfigSpec into the upstream
// kubelet.KubeletConfiguration written to /var/lib/kubelet/kubelet.conf.
// Most kubelet CLI flags are deprecated upstream in favor of this config
// file; fields that have flag:"-" in the kops API land here.
func (b *KubeletBuilder) kubeletConfiguration(kubeletConfig *kops.KubeletConfigSpec, providerID string) (*kubelet.KubeletConfiguration, error) {
	cc := &kubelet.KubeletConfiguration{
		CgroupDriver:                     kubeletConfig.CgroupDriver,
		CgroupRoot:                       kubeletConfig.CgroupRoot,
		TLSCertFile:                      filepath.Join(b.PathSrvKubernetes(), "kubelet-server.crt"),
		TLSPrivateKeyFile:                filepath.Join(b.PathSrvKubernetes(), "kubelet-server.key"),
		TLSCipherSuites:                  kubeletConfig.TLSCipherSuites,
		TLSMinVersion:                    kubeletConfig.TLSMinVersion,
		ClusterDNS:                       []string{kubeletConfig.ClusterDNS},
		ClusterDomain:                    kubeletConfig.ClusterDomain,
		EnableDebuggingHandlers:          kubeletConfig.EnableDebuggingHandlers,
		HairpinMode:                      kubeletConfig.HairpinMode,
		StaticPodPath:                    kubeletConfig.PodManifestPath,
		VolumePluginDir:                  kubeletConfig.VolumePluginDirectory,
		ProviderID:                       providerID,
		KubeletCgroups:                   kubeletConfig.KubeletCgroups,
		SystemCgroups:                    kubeletConfig.SystemCgroups,
		PodCIDR:                          kubeletConfig.PodCIDR,
		ResolverConfig:                   kubeletConfig.ResolverConfig,
		SerializeImagePulls:              kubeletConfig.SerializeImagePulls,
		MaxParallelImagePulls:            kubeletConfig.MaxParallelImagePulls,
		AllowedUnsafeSysctls:             kubeletConfig.AllowedUnsafeSysctls,
		CPUCFSQuota:                      kubeletConfig.CPUCFSQuota,
		CPUCFSQuotaPeriod:                kubeletConfig.CPUCFSQuotaPeriod,
		CPUManagerPolicy:                 kubeletConfig.CpuManagerPolicy,
		RegistryPullQPS:                  kubeletConfig.RegistryPullQPS,
		TopologyManagerPolicy:            kubeletConfig.TopologyManagerPolicy,
		RotateCertificates:               fi.ValueOf(kubeletConfig.RotateCertificates),
		ContainerLogMaxSize:              kubeletConfig.ContainerLogMaxSize,
		ContainerLogMaxFiles:             kubeletConfig.ContainerLogMaxFiles,
		PodPidsLimit:                     kubeletConfig.PodPidsLimit,
		ImageGCHighThresholdPercent:      kubeletConfig.ImageGCHighThresholdPercent,
		ImageGCLowThresholdPercent:       kubeletConfig.ImageGCLowThresholdPercent,
		ImageMaximumGCAge:                fi.ValueOf(kubeletConfig.ImageMaximumGCAge),
		NodeStatusUpdateFrequency:        fi.ValueOf(kubeletConfig.NodeStatusUpdateFrequency),
		NodeLeaseDurationSeconds:         fi.ValueOf(kubeletConfig.NodeLeaseDurationSeconds),
		SeccompDefault:                   kubeletConfig.SeccompDefault,
		KubeReserved:                     kubeletConfig.KubeReserved,
		KubeReservedCgroup:               kubeletConfig.KubeReservedCgroup,
		SystemReserved:                   kubeletConfig.SystemReserved,
		SystemReservedCgroup:             kubeletConfig.SystemReservedCgroup,
		FailSwapOn:                       kubeletConfig.FailSwapOn,
		EvictionPressureTransitionPeriod: fi.ValueOf(kubeletConfig.EvictionPressureTransitionPeriod),
		EvictionMaxPodGracePeriod:        kubeletConfig.EvictionMaxPodGracePeriod,
		EventRecordQPS:                   kubeletConfig.EventRecordQPS,
		ProtectKernelDefaults:            fi.ValueOf(kubeletConfig.ProtectKernelDefaults),
		KernelMemcgNotification:          fi.ValueOf(kubeletConfig.KernelMemcgNotification),
		MaxPods:                          fi.ValueOf(kubeletConfig.MaxPods),
		ReadOnlyPort:                     fi.ValueOf(kubeletConfig.ReadOnlyPort),
		MemorySwap:                       kubelet.MemorySwapConfiguration{SwapBehavior: kubeletConfig.MemorySwapBehavior},
		Logging:                          logsapi.LoggingConfiguration{Format: kubeletConfig.LogFormat},
		CrashLoopBackOff:                 kubelet.CrashLoopBackOffConfig{MaxContainerRestartPeriod: kubeletConfig.CrashLoopBackOffMaxContainerRestartPeriod},
		ShutdownGracePeriod:              fi.ValueOf(kubeletConfig.ShutdownGracePeriod),
		ShutdownGracePeriodCriticalPods:  fi.ValueOf(kubeletConfig.ShutdownGracePeriodCriticalPods),
	}

	cc.Authentication.Anonymous.Enabled = kubeletConfig.AnonymousAuth
	cc.Authentication.Webhook.Enabled = kubeletConfig.AuthenticationTokenWebhook
	if kubeletConfig.ClientCAFile != "" {
		cc.Authentication.X509.ClientCAFile = kubeletConfig.ClientCAFile
	}
	if kubeletConfig.AuthorizationMode != "" {
		cc.Authorization.Mode = kubelet.KubeletAuthorizationMode(kubeletConfig.AuthorizationMode)
	}

	// EventQPS is the legacy kops field; EventRecordQPS is the newer alias
	// matching the kubelet config field name. Preserve historical precedence
	// where EventQPS overrides EventRecordQPS when both are set.
	if kubeletConfig.EventQPS != nil {
		cc.EventRecordQPS = kubeletConfig.EventQPS
	}

	if b.NodeupConfig.ContainerdConfig.Address == nil {
		cc.ContainerRuntimeEndpoint = "unix:///run/containerd/containerd.sock"
	} else {
		cc.ContainerRuntimeEndpoint = "unix://" + fi.ValueOf(b.NodeupConfig.ContainerdConfig.Address)
	}

	if kubeletConfig.EvictionHard != nil {
		evictionHard, err := parseKeyValueList(*kubeletConfig.EvictionHard, "<")
		if err != nil {
			return nil, fmt.Errorf("evictionHard: %w", err)
		}
		cc.EvictionHard = evictionHard
	}
	evictionSoft, err := parseKeyValueList(kubeletConfig.EvictionSoft, "<")
	if err != nil {
		return nil, fmt.Errorf("evictionSoft: %w", err)
	}
	cc.EvictionSoft = evictionSoft
	evictionSoftGracePeriod, err := parseKeyValueList(kubeletConfig.EvictionSoftGracePeriod, "=")
	if err != nil {
		return nil, fmt.Errorf("evictionSoftGracePeriod: %w", err)
	}
	cc.EvictionSoftGracePeriod = evictionSoftGracePeriod
	evictionMinimumReclaim, err := parseKeyValueList(kubeletConfig.EvictionMinimumReclaim, "=")
	if err != nil {
		return nil, fmt.Errorf("evictionMinimumReclaim: %w", err)
	}
	cc.EvictionMinimumReclaim = evictionMinimumReclaim

	featureGates, err := parseFeatureGates(kubeletConfig.FeatureGates)
	if err != nil {
		return nil, fmt.Errorf("featureGates: %w", err)
	}
	cc.FeatureGates = featureGates

	if kubeletConfig.EnforceNodeAllocatable != "" {
		cc.EnforceNodeAllocatable = strings.Split(kubeletConfig.EnforceNodeAllocatable, ",")
	}

	for _, t := range kubeletConfig.Taints {
		taint, err := parseTaint(t)
		if err != nil {
			return nil, fmt.Errorf("taints: %w", err)
		}
		cc.RegisterWithTaints = append(cc.RegisterWithTaints, taint)
	}

	return cc, nil
}

func (b *KubeletBuilder) binaryPath() string {
	path := "/usr/local/bin"
	if b.Distribution == distributions.DistributionFlatcar {
		path = "/opt/kubernetes/bin"
	}
	if b.Distribution == distributions.DistributionContainerOS {
		path = "/home/kubernetes/bin"
	}
	return path
}

// kubeletPath returns the path of the kubelet based on distro
func (b *KubeletBuilder) kubeletPath() string {
	return b.binaryPath() + "/kubelet"
}

// getECRCredentialProviderPath returns the path of the ECR Credentials Provider based on distro and archiecture
func (b *KubeletBuilder) getECRCredentialProviderPath() string {
	return b.binaryPath() + "/ecr-credential-provider"
}

// getGCPCredentialProviderPath returns the path of the GCP Credentials Provider based on distro and archiecture
func (b *KubeletBuilder) getGCPCredentialProviderPath() string {
	return b.binaryPath() + "/gcp-credential-provider"
}

// buildManifestDirectory creates the directory where kubelet expects static manifests to reside
func (b *KubeletBuilder) buildManifestDirectory(kubeletConfig *kops.KubeletConfigSpec) (*nodetasks.File, error) {
	if kubeletConfig.PodManifestPath == "" {
		return nil, fmt.Errorf("failed to build manifest path. Path was empty")
	}
	directory := &nodetasks.File{
		Path: kubeletConfig.PodManifestPath,
		Type: nodetasks.FileType_Directory,
		Mode: s("0755"),
	}
	return directory, nil
}

// buildSystemdEnvironmentFile renders the environment file for the kubelet
func (b *KubeletBuilder) buildSystemdEnvironmentFile(ctx context.Context, kubeletConfig *kops.KubeletConfigSpec) (*nodetasks.File, error) {
	// TODO: Dump the separate file for flags - just complexity!
	flags, err := flagbuilder.BuildFlags(kubeletConfig)
	if err != nil {
		return nil, fmt.Errorf("error building kubelet flags: %v", err)
	}

	if b.UsesSecondaryIP() {
		localIP, err := b.GetMetadataLocalIP(ctx)
		if err != nil {
			return nil, err
		}
		if localIP != "" {
			flags += " --node-ip=" + localIP
		}
	}

	if b.IsIPv6Only() {
		flags += " --node-ip=::"
	}

	flags += " --config=" + kubeletConfigFilePath

	if b.UseExternalKubeletCredentialProvider() {
		flags += " --image-credential-provider-config=" + credentialProviderConfigFilePath
		flags += " --image-credential-provider-bin-dir=" + b.binaryPath()
	}

	sysconfig := "DAEMON_ARGS=\"" + flags + "\"\n"
	// Makes kubelet read /root/.docker/config.json properly
	sysconfig = sysconfig + "HOME=\"/root" + "\"\n"

	t := &nodetasks.File{
		Path:     "/etc/sysconfig/kubelet",
		Contents: fi.NewStringResource(sysconfig),
		Type:     nodetasks.FileType_File,
	}

	return t, nil
}

// parseKeyValueList parses a comma-separated list of key/value pairs
// separated by sep (for example "memory.available<100Mi" with sep="<").
// Whitespace around keys and values is trimmed. An empty input returns
// (nil, nil) so callers can leave the corresponding kubelet config field
// unset. Returns an error if any entry is missing the separator.
//
// The kops API uses these CSV strings for fields that kubelet represents
// as map[string]string: eviction-hard / eviction-soft use "<", while
// eviction-soft-grace-period and eviction-minimum-reclaim use "=".
func parseKeyValueList(in string, sep string) (map[string]string, error) {
	if in == "" {
		return nil, nil
	}
	result := make(map[string]string, strings.Count(in, ",")+1)
	for kv := range strings.SplitSeq(in, ",") {
		k, v, ok := strings.Cut(kv, sep)
		if !ok {
			return nil, fmt.Errorf("invalid key/value pair %q (expected separator %q)", kv, sep)
		}
		result[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return result, nil
}

// parseTaint converts the kops "key=value:Effect" taint string (the form
// historically passed to --register-with-taints) into a v1.Taint, the type
// the kubelet config field RegisterWithTaints requires.
func parseTaint(s string) (v1.Taint, error) {
	parsed, err := kopsutil.ParseTaint(s)
	if err != nil {
		return v1.Taint{}, err
	}
	return v1.Taint{
		Key:    parsed["key"],
		Value:  parsed["value"],
		Effect: v1.TaintEffect(parsed["effect"]),
	}, nil
}

// parseFeatureGates converts the kops map[string]string feature-gate
// representation into the map[string]bool that the kubelet config schema
// requires. Values are parsed with strconv.ParseBool, so "true"/"false",
// "1"/"0", "t"/"f" etc. are all accepted. An empty or nil input returns
// (nil, nil); an unparseable value returns an error naming the offending
// gate.
func parseFeatureGates(gates map[string]string) (map[string]bool, error) {
	if len(gates) == 0 {
		return nil, nil
	}
	out := make(map[string]bool, len(gates))
	for name, raw := range gates {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid feature gate value %q=%q: %w", name, raw, err)
		}
		out[name] = parsed
	}
	return out, nil
}

// buildSystemdService is responsible for generating the kubelet systemd unit
func (b *KubeletBuilder) buildSystemdService() *nodetasks.Service {
	kubeletCommand := b.kubeletPath()

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Description", "Kubernetes Kubelet Server")
	manifest.Set("Unit", "Documentation", "https://github.com/kubernetes/kubernetes")
	manifest.Set("Unit", "After", "containerd.service")

	manifest.Set("Service", "EnvironmentFile", "/etc/sysconfig/kubelet")

	manifest.Set("Service", "ExecStart", kubeletCommand+" \"$DAEMON_ARGS\"")
	manifest.Set("Service", "Restart", "always")
	manifest.Set("Service", "RestartSec", "2s")
	manifest.Set("Service", "StartLimitInterval", "0")
	manifest.Set("Service", "KillMode", "process")
	manifest.Set("Service", "User", "root")
	manifest.Set("Service", "CPUAccounting", "true")
	manifest.Set("Service", "MemoryAccounting", "true")

	manifest.Set("Install", "WantedBy", "multi-user.target")

	cgroup := b.NodeupConfig.KubeletConfig.KubeletCgroups
	if cgroup != "" {
		manifest.Set("Service", "Slice", strings.Trim(cgroup, "/")+".slice")
	}

	manifestString := manifest.Render()

	klog.V(8).Infof("Built service manifest %q\n%s", "kubelet", manifestString)

	service := &nodetasks.Service{
		Name:       kubeletService,
		Definition: s(manifestString),
	}

	service.InitDefaults()

	if b.ConfigurationMode == "Warming" {
		service.Running = new(false)
		service.Enabled = new(false)
	}

	return service
}

// addECRCredentialProvider installs the ECR Kubelet Credential Provider
func (b *KubeletBuilder) addECRCredentialProvider(c *fi.NodeupModelBuilderContext) error {
	{
		assetName := "ecr-credential-provider-linux-" + string(b.Architecture)
		assetPath := ""
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		t := &nodetasks.File{
			Path:     b.getECRCredentialProviderPath(),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		}
		c.AddTask(t)
	}

	{

		providerConfig := &kubeletv1.CredentialProviderConfig{}

		// Build the list of container registry globs to match
		registryList := []string{
			"*.dkr.ecr.*.amazonaws.com",
			"*.dkr.ecr.*.amazonaws.com.cn",
			"*.dkr.ecr-fips.*.amazonaws.com",
			"*.dkr.ecr.us-iso-east-1.c2s.ic.gov",
		}

		containerd := b.NodeupConfig.ContainerdConfig
		if containerd.UseECRCredentialsForMirrors {
			for name := range containerd.RegistryMirrors {
				registryList = append(registryList, name)
			}
		}

		cacheDuration, err := time.ParseDuration("12h")
		if err != nil {
			return err
		}

		providerConfig.Providers = []kubeletv1.CredentialProvider{
			{
				APIVersion:           "credentialprovider.kubelet.k8s.io/v1",
				Name:                 "ecr-credential-provider",
				MatchImages:          registryList,
				DefaultCacheDuration: &metav1.Duration{Duration: cacheDuration},
				Args:                 []string{"get-credentials"},
				Env: []kubeletv1.ExecEnvVar{
					{
						Name:  "AWS_REGION",
						Value: b.Cloud.Region(),
					},
				},
			},
		}

		var buf bytes.Buffer
		if err := kubeletV1Encoder.Encode(providerConfig, &buf); err != nil {
			return fmt.Errorf("encoding ECR credential provider config: %w", err)
		}

		c.AddTask(&nodetasks.File{
			Path:     credentialProviderConfigFilePath,
			Contents: fi.NewBytesResource(buf.Bytes()),
			Type:     nodetasks.FileType_File,
			Mode:     s("0644"),
		})
	}
	return nil
}

// addGCPCredentialProvider installs the GCP Kubelet Credential Provider
func (b *KubeletBuilder) addGCPCredentialProvider(c *fi.NodeupModelBuilderContext) error {
	{
		assetName := "auth-provider-gcp"
		assetPath := ""
		asset, err := b.Assets.Find(assetName, assetPath)
		if err != nil {
			return fmt.Errorf("trying to locate asset %q: %v", assetName, err)
		}
		if asset == nil {
			return fmt.Errorf("unable to locate asset %q", assetName)
		}

		t := &nodetasks.File{
			Path:     b.getGCPCredentialProviderPath(),
			Contents: asset,
			Type:     nodetasks.FileType_File,
			Mode:     s("0755"),
		}
		c.AddTask(t)
	}

	{
		configContent := `apiVersion: kubelet.config.k8s.io/v1
kind: CredentialProviderConfig
providers:
  - apiVersion: credentialprovider.kubelet.k8s.io/v1
    name: gcp-credential-provider
    matchImages:
      - "gcr.io"
      - "*.gcr.io"
      - "container.cloud.google.com"
      - "*.pkg.dev"
    defaultCacheDuration: "1m"
    args:
      - get-credentials
      - --v=3
`

		t := &nodetasks.File{
			Path:     credentialProviderConfigFilePath,
			Contents: fi.NewStringResource(configContent),
			Type:     nodetasks.FileType_File,
			Mode:     s("0644"),
		}
		c.AddTask(t)
	}
	return nil
}

// NodeLabels are defined in the InstanceGroup, but set flags on the kubelet config.
// We have a conflict here: on the one hand we want an easy to use abstract specification
// for the cluster, on the other hand we don't want two fields that do the same thing.
// So we make the logic for combining a KubeletConfig part of our core logic.
// NodeLabels are set on the instanceGroup.  We might allow specification of them on the kubelet
// config as well, but for now the precedence is not fully specified.
// (Today, NodeLabels on the InstanceGroup are merged in to NodeLabels on the KubeletConfig in the Cluster).
// In future, we will likely deprecate KubeletConfig in the Cluster, and move it into componentconfig,
// once that is part of core k8s.

// buildKubeletConfigSpec returns the kubeletconfig for the specified instanceGroup
func (b *KubeletBuilder) buildKubeletConfigSpec(ctx context.Context) (*kops.KubeletConfigSpec, error) {
	// Merge KubeletConfig for NodeLabels
	c := b.NodeupConfig.KubeletConfig

	c.ClientCAFile = filepath.Join(b.PathSrvKubernetes(), "ca.crt")

	// Preserve the historical 15m default that used to be hard-coded into the
	// kubelet env file. Stays as a CLI flag because the upstream config field
	// is non-pointer with omitempty and would drop a zero value.
	if c.RuntimeRequestTimeout == nil {
		c.RuntimeRequestTimeout = &metav1.Duration{Duration: 15 * time.Minute}
	}

	// Wait less for pods to restart, especially during the bootstrap sequence
	if b.IsKubernetesGTE("1.35") && b.IsMaster {
		c.CrashLoopBackOffMaxContainerRestartPeriod = &metav1.Duration{Duration: time.Minute}
	}

	// Respect any MaxPods value the user sets explicitly.
	if (b.NodeupConfig.Networking.AmazonVPC != nil || (b.NodeupConfig.Networking.Cilium != nil && b.NodeupConfig.Networking.Cilium.IPAM == kops.CiliumIpamEni)) && c.MaxPods == nil {
		config, err := awsconfig.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("error loading AWS config: %v", err)
		}
		metadata := imds.NewFromConfig(config)

		var instanceTypeName ec2types.InstanceType
		// Get the actual instance type by querying the EC2 instance metadata service.
		resp, err := metadata.GetMetadata(ctx, &imds.GetMetadataInput{Path: "instance-type"})
		if err == nil {
			defer resp.Content.Close()
			itName, err := io.ReadAll(resp.Content)
			if err == nil {
				instanceTypeName = ec2types.InstanceType(string(itName))
			}
		}
		if instanceTypeName == "" {
			instanceTypeName = ec2types.InstanceType(*b.NodeupConfig.DefaultMachineType)
		}

		// Get the instance type's detailed information.
		instanceType, err := b.Cloud.GetMachineTypeInfo(ctx, instanceTypeName)
		if err != nil {
			return nil, err
		}

		// Default maximum pods per node defined by KubeletConfiguration
		maxPods := int32(110)

		// AWS VPC CNI plugin-specific maximum pod calculation based on:
		// https://github.com/aws/amazon-vpc-cni-k8s/blob/v1.9.3/README.md#setup
		enis := instanceType.InstanceENIs
		ips := instanceType.InstanceIPsPerENI
		if enis > 0 && ips > 0 {
			instanceMaxPods := enis*(ips-1) + 2
			if instanceMaxPods < maxPods {
				maxPods = instanceMaxPods
			}
		}

		// Write back values that could have changed
		c.MaxPods = new(int32(maxPods))
	}

	if c.VolumePluginDirectory == "" {
		switch b.Distribution {
		case distributions.DistributionContainerOS:
			// Default is different on ContainerOS, see https://github.com/kubernetes/kubernetes/pull/58171
			c.VolumePluginDirectory = "/home/kubernetes/flexvolume/"

		case distributions.DistributionFlatcar:
			// The /usr directory is read-only for Flatcar
			c.VolumePluginDirectory = "/var/lib/kubelet/volumeplugins/"

		default:
			c.VolumePluginDirectory = "/usr/libexec/kubernetes/kubelet-plugins/volume/exec/"
		}
	}

	// In certain configurations systemd-resolved will put the loopback address 127.0.0.53 as a nameserver into /etc/resolv.conf
	// https://github.com/coredns/coredns/blob/master/plugin/loop/README.md#troubleshooting-loops-in-kubernetes-clusters
	if c.ResolverConfig == nil {
		if b.Distribution.HasLoopbackEtcResolvConf() {
			c.ResolverConfig = s("/run/systemd/resolve/resolv.conf")
		}
	}

	// As of 1.16 we can no longer set critical labels.
	// kops-controller will set these labels.
	// For bootstrapping reasons, protokube sets the critical labels for kops-controller to run.
	c.NodeLabels = nil

	if c.AuthorizationMode == "" {
		c.AuthorizationMode = "Webhook"
	}

	if c.AuthenticationTokenWebhook == nil {
		c.AuthenticationTokenWebhook = new(true)
	}

	return &c, nil
}

// buildControlPlaneKubeletKubeconfig builds a kubeconfig for the master kubelet, self-signing the kubelet cert
func (b *KubeletBuilder) buildControlPlaneKubeletKubeconfig(c *fi.NodeupModelBuilderContext) (fi.Resource, error) {
	nodeName, err := b.NodeName()
	if err != nil {
		return nil, fmt.Errorf("error getting NodeName: %v", err)
	}
	certName := nodetasks.PKIXName{
		CommonName:   fmt.Sprintf("system:node:%s", nodeName),
		Organization: []string{rbac.NodesGroup},
	}

	return b.BuildIssuedKubeconfig("kubelet", certName, c), nil
}

func (b *KubeletBuilder) buildKubeletServingCertificate(c *fi.NodeupModelBuilderContext) error {
	name := "kubelet-server"
	dir := b.PathSrvKubernetes()

	names, err := b.kubeletNames(c.Context())
	if err != nil {
		return err
	}

	var cert, key fi.Resource
	if !b.HasAPIServer {
		cert, key, err = b.GetBootstrapCert(name, fi.CertificateIDCA)
		if err != nil {
			return err
		}
	} else {
		issueCert := &nodetasks.IssueCert{
			Name:      name,
			Signer:    fi.CertificateIDCA,
			KeypairID: b.NodeupConfig.KeypairIDs[fi.CertificateIDCA],
			Type:      "server",
			Subject: nodetasks.PKIXName{
				CommonName: names[0],
			},
			AlternateNames: names,
		}
		c.AddTask(issueCert)
		cert, key, _ = issueCert.GetResources()
		c.EnsureTask(&nodetasks.File{
			Path: dir,
			Type: nodetasks.FileType_Directory,
			Mode: new("0755"),
		})
	}

	c.AddTask(&nodetasks.File{
		Path:           filepath.Join(dir, name+".crt"),
		Contents:       cert,
		Type:           nodetasks.FileType_File,
		Mode:           new("0644"),
		BeforeServices: []string{kubeletService},
	})

	c.AddTask(&nodetasks.File{
		Path:           filepath.Join(dir, name+".key"),
		Contents:       key,
		Type:           nodetasks.FileType_File,
		Mode:           new("0400"),
		BeforeServices: []string{kubeletService},
	})

	return nil
}

func (b *KubeletBuilder) kubeletNames(ctx context.Context) ([]string, error) {
	if b.CloudProvider() != kops.CloudProviderAWS {
		name, err := os.Hostname()
		if err != nil {
			return nil, err
		}

		addrs, _ := net.LookupHost(name)

		return append(addrs, name), nil
	}

	addrs := []string{b.InstanceID}
	config, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("error loading AWS config: %v", err)
	}
	metadata := imds.NewFromConfig(config)

	if localHostname, err := getMetadata(ctx, metadata, "local-hostname"); err == nil {
		klog.V(2).Infof("Local Hostname: %s", localHostname)
		addrs = append(addrs, localHostname)
	}
	if localIPv4, err := getMetadata(ctx, metadata, "local-ipv4"); err == nil {
		klog.V(2).Infof("Local IPv4: %s", localIPv4)
		addrs = append(addrs, localIPv4)
	}
	if publicIPv4, err := getMetadata(ctx, metadata, "public-ipv4"); err == nil {
		klog.V(2).Infof("Public IPv4: %s", publicIPv4)
		addrs = append(addrs, publicIPv4)
	}
	if publicIPv6, err := getMetadata(ctx, metadata, "ipv6"); err == nil {
		klog.V(2).Infof("Public IPv6: %s", publicIPv6)
		addrs = append(addrs, publicIPv6)
	}

	return addrs, nil
}

func (b *KubeletBuilder) buildCgroupService(name string) *nodetasks.Service {
	name = strings.Trim(name, "/")

	manifest := &systemd.Manifest{}
	manifest.Set("Unit", "Documentation", "man:systemd.special(7)")
	manifest.Set("Unit", "Before", "slices.target")
	manifest.Set("Unit", "DefaultDependencies", "no")

	manifestString := manifest.Render()

	service := &nodetasks.Service{
		Name:       name + ".slice",
		Definition: s(manifestString),
	}

	return service
}

func getMetadata(ctx context.Context, client *imds.Client, path string) (string, error) {
	resp, err := client.GetMetadata(ctx, &imds.GetMetadataInput{Path: path})
	if err != nil {
		return "", err
	}
	defer resp.Content.Close()
	data, err := io.ReadAll(resp.Content)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

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

package nodeup

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/kms"
	"go.uber.org/multierr"
	"k8s.io/klog/v2"
	"k8s.io/kops/nodeup/pkg/model"
	"k8s.io/kops/nodeup/pkg/model/networking"
	api "k8s.io/kops/pkg/apis/kops"
	kopsmodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/bootstrap/pkibootstrap"
	"k8s.io/kops/pkg/configserver"
	"k8s.io/kops/pkg/kopscontrollerclient"
	"k8s.io/kops/pkg/resolver"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/gcediscovery"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm/gcetpmsigner"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/upup/pkg/fi/secrets"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/distributions"
	"k8s.io/kops/util/pkg/vfs"
)

// MaxTaskDuration is the amount of time to keep trying for; we retry for a long time - there is not really any great fallback
const MaxTaskDuration = 365 * 24 * time.Hour

// NodeUpCommand is the configuration for nodeup
type NodeUpCommand struct {
	CacheDir       string
	ConfigLocation string
	Target         string
}

// Run is responsible for perform the nodeup process
func (c *NodeUpCommand) Run(out io.Writer) error {
	ctx := context.Background()

	var bootConfig nodeup.BootConfig
	if c.ConfigLocation != "" {
		b, err := vfs.Context.ReadFile(c.ConfigLocation)
		if err != nil {
			return fmt.Errorf("error loading configuration %q: %v", c.ConfigLocation, err)
		}

		err = utils.YamlUnmarshal(b, &bootConfig)
		if err != nil {
			return fmt.Errorf("error parsing configuration %q: %v", c.ConfigLocation, err)
		}
	} else {
		return fmt.Errorf("ConfigLocation is required")
	}

	if c.CacheDir == "" {
		return fmt.Errorf("CacheDir is required")
	}

	region, err := getRegion(ctx, &bootConfig)
	if err != nil {
		return err
	}
	if err = seedRNG(ctx, &bootConfig, region); err != nil {
		return err
	}

	var configBase vfs.Path

	// If we're using a config server instead of vfs, nodeConfig will hold our configuration
	var nodeConfig *nodeup.NodeConfig

	if bootConfig.ConfigServer != nil && len(bootConfig.ConfigServer.Servers) > 0 {
		response, err := getNodeConfigFromServers(ctx, &bootConfig, region)
		if err != nil {
			return fmt.Errorf("failed to get node config from server: %w", err)
		}
		nodeConfig = response.NodeConfig
	} else if fi.ValueOf(bootConfig.ConfigBase) != "" {
		var err error
		configBase, err = vfs.Context.BuildVfsPath(*bootConfig.ConfigBase)
		if err != nil {
			return fmt.Errorf("cannot parse ConfigBase %q: %v", *bootConfig.ConfigBase, err)
		}
	} else {
		return fmt.Errorf("ConfigBase or ConfigServer is required")
	}

	var nodeupConfig nodeup.Config
	var nodeupConfigHash [32]byte
	if nodeConfig != nil {
		if err := utils.YamlUnmarshal([]byte(nodeConfig.NodeupConfig), &nodeupConfig); err != nil {
			return fmt.Errorf("error parsing BootConfig config response: %v", err)
		}
		nodeupConfigHash = sha256.Sum256([]byte(nodeConfig.NodeupConfig))
		nodeupConfig.CAs[fi.CertificateIDCA] = bootConfig.ConfigServer.CACertificates
	} else if bootConfig.InstanceGroupName != "" {
		nodeupConfigLocation := configBase.Join("igconfig", bootConfig.InstanceGroupRole.ToLowerString(), bootConfig.InstanceGroupName, "nodeupconfig.yaml")

		b, err := nodeupConfigLocation.ReadFile(ctx)
		if err != nil {
			return fmt.Errorf("error loading NodeupConfig %q: %v", nodeupConfigLocation, err)
		}

		if err = utils.YamlUnmarshal(b, &nodeupConfig); err != nil {
			return fmt.Errorf("error parsing NodeupConfig %q: %v", nodeupConfigLocation, err)
		}
		nodeupConfigHash = sha256.Sum256(b)
	} else {
		return fmt.Errorf("no instance group defined in nodeup config")
	}

	if want := bootConfig.NodeupConfigHash; want != "" {
		if got := base64.StdEncoding.EncodeToString(nodeupConfigHash[:]); got != want {
			return fmt.Errorf("nodeup config hash mismatch (was %q, expected %q)", got, want)
		}
	}

	err = evaluateSpec(&nodeupConfig, bootConfig.CloudProvider)
	if err != nil {
		return err
	}

	architecture, err := architectures.FindArchitecture()
	if err != nil {
		return fmt.Errorf("error determining OS architecture: %v", err)
	}

	distribution, err := distributions.FindDistribution("/")
	if err != nil {
		return fmt.Errorf("error determining OS distribution: %v", err)
	}

	configAssets := nodeupConfig.Assets[architecture]
	assetStore := fi.NewAssetStore(c.CacheDir)
	for _, asset := range configAssets {
		err := assetStore.Add(asset)
		if err != nil {
			return fmt.Errorf("error adding asset %q: %v", asset, err)
		}
	}

	var cloud fi.Cloud

	if bootConfig.CloudProvider == api.CloudProviderAWS {
		awsCloud, err := awsup.NewAWSCloud(region, nil)
		if err != nil {
			return err
		}
		cloud = awsCloud
	}

	modelContext := &model.NodeupModelContext{
		Cloud:        cloud,
		Architecture: architecture,
		Assets:       assetStore,
		ConfigBase:   configBase,
		Distribution: distribution,
		BootConfig:   &bootConfig,
		NodeupConfig: &nodeupConfig,
	}

	var secretStore fi.SecretStoreReader
	var keyStore fi.KeystoreReader
	if nodeConfig != nil {
		modelContext.SecretStore = configserver.NewSecretStore(nodeConfig.NodeSecrets)
	} else if nodeupConfig.ConfigStore.Secrets != "" {
		klog.Infof("Building SecretStore at %q", nodeupConfig.ConfigStore.Secrets)
		p, err := vfs.Context.BuildVfsPath(nodeupConfig.ConfigStore.Secrets)
		if err != nil {
			return fmt.Errorf("error building secret store path: %v", err)
		}

		secretStore = secrets.NewVFSSecretStoreReader(p)
		modelContext.SecretStore = secretStore
	} else {
		return fmt.Errorf("SecretStore not set")
	}

	if nodeConfig != nil {
		modelContext.KeyStore = configserver.NewKeyStore()
	} else if nodeupConfig.ConfigStore.Keypairs != "" {
		klog.Infof("Building KeyStore at %q", nodeupConfig.ConfigStore.Keypairs)
		p, err := vfs.Context.BuildVfsPath(nodeupConfig.ConfigStore.Keypairs)
		if err != nil {
			return fmt.Errorf("error building key store path: %v", err)
		}

		modelContext.KeyStore = fi.NewVFSKeystoreReader(p)
		keyStore = modelContext.KeyStore
	} else {
		return fmt.Errorf("KeyStore not set")
	}

	if err := modelContext.Init(); err != nil {
		return err
	}

	if bootConfig.CloudProvider == api.CloudProviderAWS {
		instanceIDBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/instance-id")
		if err != nil {
			return fmt.Errorf("error reading instance-id from AWS metadata: %v", err)
		}
		modelContext.InstanceID = string(instanceIDBytes)

		modelContext.ConfigurationMode, err = getAWSConfigurationMode(ctx, modelContext)
		if err != nil {
			return err
		}

		modelContext.MachineType, err = getMachineType()
		if err != nil {
			return fmt.Errorf("failed to get machine type: %w", err)
		}

		// If Nvidia is enabled in the cluster, check if this instance has support for it.
		nvidia := modelContext.NodeupConfig.ContainerdConfig.NvidiaGPU
		if nvidia != nil && fi.ValueOf(nvidia.Enabled) {
			awsCloud := cloud.(awsup.AWSCloud)
			// Get the instance type's detailed information.
			instanceType, err := awsup.GetMachineTypeInfo(awsCloud, modelContext.MachineType)
			if err != nil {
				return err
			}

			if instanceType.GPU {
				klog.Info("instance supports GPU acceleration")
				modelContext.GPUVendor = architectures.GPUVendorNvidia
			}
		}
	} else if bootConfig.CloudProvider == api.CloudProviderOpenstack {
		// NvidiaGPU possible to enable only in instance group level in OpenStack. When we assume that GPU is supported
		if nodeupConfig.NvidiaGPU != nil && fi.ValueOf(nodeupConfig.NvidiaGPU.Enabled) {
			klog.Info("instance supports GPU acceleration")
			modelContext.GPUVendor = architectures.GPUVendorNvidia
		}
	}

	if err := loadKernelModules(modelContext); err != nil {
		return err
	}

	loader := &Loader{}
	loader.Builders = append(loader.Builders, &model.EtcHostsBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.NTPBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.DirectoryBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.UpdateServiceBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.VolumesBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.ContainerdBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.ProtokubeBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.CloudConfigBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.FileAssetsBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.HookBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeletBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubectlBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.LogrotateBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.ManifestsBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.PackagesBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.NvidiaBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.SecretBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.FirewallBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.SysctlBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeAPIServerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeControllerManagerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeSchedulerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.EtcdManagerTLSBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeProxyBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KopsControllerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.WarmPoolBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.PrefixBuilder{NodeupModelContext: modelContext})

	loader.Builders = append(loader.Builders, &networking.CommonBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &networking.CalicoBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &networking.CiliumBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &networking.KuberouterBuilder{NodeupModelContext: modelContext})

	loader.Builders = append(loader.Builders, &model.BootstrapClientBuilder{NodeupModelContext: modelContext})
	taskMap, err := loader.Build()
	if err != nil {
		return fmt.Errorf("error building loader: %v", err)
	}

	for i, image := range nodeupConfig.Images[architecture] {
		taskMap["LoadImage."+strconv.Itoa(i)] = &nodetasks.LoadImageTask{
			Sources: image.Sources,
			Hash:    image.Hash,
		}
	}
	// Protokube load image task is in ProtokubeBuilder

	var target fi.NodeupTarget

	switch c.Target {
	case "direct":
		target = &local.LocalTarget{
			CacheDir: c.CacheDir,
			Cloud:    cloud,
		}
	case "dryrun":
		assetBuilder := assets.NewAssetBuilder(vfs.Context, nil, nodeupConfig.KubernetesVersion, false)
		target = fi.NewNodeupDryRunTarget(assetBuilder, out)
	default:
		return fmt.Errorf("unsupported target type %q", c.Target)
	}

	context, err := fi.NewNodeupContext(ctx, target, keyStore, &bootConfig, &nodeupConfig, taskMap)
	if err != nil {
		klog.Exitf("error building context: %v", err)
	}

	var options fi.RunTasksOptions
	options.InitDefaults()

	err = context.RunTasks(options)
	if err != nil {
		klog.Exitf("error running tasks: %v", err)
	}

	err = target.Finish(taskMap)
	if err != nil {
		klog.Exitf("error closing target: %v", err)
	}

	if nodeupConfig.EnableLifecycleHook {
		if bootConfig.CloudProvider == api.CloudProviderAWS {
			err := completeWarmingLifecycleAction(ctx, cloud.(awsup.AWSCloud), modelContext)
			if err != nil {
				return fmt.Errorf("failed to complete lifecylce action: %w", err)
			}
		}
	}
	return nil
}

func getMachineType() (string, error) {
	config := aws.NewConfig()
	config = config.WithCredentialsChainVerboseErrors(true)

	sess := session.Must(session.NewSession(config))
	metadata := ec2metadata.New(sess)

	// Get the actual instance type by querying the EC2 instance metadata service.
	instanceTypeName, err := metadata.GetMetadata("instance-type")
	if err != nil {
		return "", fmt.Errorf("failed to get instance metadata type: %w", err)
	}
	return instanceTypeName, err
}

func completeWarmingLifecycleAction(ctx context.Context, cloud awsup.AWSCloud, modelContext *model.NodeupModelContext) error {
	asgName := modelContext.BootConfig.InstanceGroupName + "." + modelContext.NodeupConfig.ClusterName
	hookName := "kops-warmpool"
	svc := cloud.Autoscaling()
	hooks, err := svc.DescribeLifecycleHooksWithContext(ctx, &autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: &asgName,
		LifecycleHookNames:   []*string{&hookName},
	})
	if err != nil {
		return fmt.Errorf("failed to find lifecycle hook %q: %w", hookName, err)
	}

	if len(hooks.LifecycleHooks) > 0 {
		klog.Info("Found ASG lifecycle hook")
		_, err := svc.CompleteLifecycleActionWithContext(ctx, &autoscaling.CompleteLifecycleActionInput{
			AutoScalingGroupName:  &asgName,
			InstanceId:            &modelContext.InstanceID,
			LifecycleHookName:     &hookName,
			LifecycleActionResult: fi.PtrTo("CONTINUE"),
		})
		if err != nil {
			return fmt.Errorf("failed to complete lifecycle hook %q for %q: %v", hookName, modelContext.InstanceID, err)
		}
		klog.Info("Lifecycle action completed")
	} else {
		klog.Info("No ASG lifecycle hook found")
	}
	return nil
}

func evaluateSpec(nodeupConfig *nodeup.Config, cloudProvider api.CloudProviderID) error {
	hostnameOverride, err := evaluateHostnameOverride(cloudProvider)
	if err != nil {
		return err
	}

	nodeupConfig.KubeletConfig.HostnameOverride = hostnameOverride

	if nodeupConfig.KubeProxy != nil {
		nodeupConfig.KubeProxy.HostnameOverride = hostnameOverride
		nodeupConfig.KubeProxy.BindAddress, err = evaluateBindAddress(nodeupConfig.KubeProxy.BindAddress)
		if err != nil {
			return err
		}
	}

	return nil
}

func evaluateHostnameOverride(cloudProvider api.CloudProviderID) (string, error) {
	switch cloudProvider {
	case api.CloudProviderAWS:
		instanceIDBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/instance-id")
		if err != nil {
			return "", fmt.Errorf("error reading instance-id from AWS metadata: %v", err)
		}

		return string(instanceIDBytes), nil

	case api.CloudProviderGCE:
		// This lets us tolerate broken hostnames (i.e. systemd)
		b, err := vfs.Context.ReadFile("metadata://gce/instance/hostname")
		if err != nil {
			return "", fmt.Errorf("error reading hostname from GCE metadata: %v", err)
		}

		// We only want to use the first portion of the fully-qualified name
		// e.g. foo.c.project.internal => foo
		fullyQualified := string(b)
		bareHostname := strings.Split(fullyQualified, ".")[0]
		return bareHostname, nil
	case api.CloudProviderDO:
		vBytes, err := vfs.Context.ReadFile("metadata://digitalocean/interfaces/private/0/ipv4/address")
		if err != nil {
			return "", fmt.Errorf("error reading droplet private IP from DigitalOcean metadata: %v", err)
		}

		hostname := string(vBytes)
		if hostname == "" {
			return "", errors.New("private IP for digitalocean droplet was empty")
		}

		return hostname, nil
	}

	return "", nil
}

func evaluateBindAddress(bindAddress string) (string, error) {
	if bindAddress == "" {
		return "", nil
	}
	if bindAddress == "@aws" {
		vBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/local-ipv4")
		if err != nil {
			return "", fmt.Errorf("error reading local IP from AWS metadata: %v", err)
		}

		// The local-ipv4 gets it's IP from the AWS.
		// For now just choose the first one.
		ips := strings.Fields(string(vBytes))
		if len(ips) == 0 {
			klog.Warningf("Local IP from AWS metadata service was empty")
			return "", nil
		}

		ip := ips[0]
		klog.Infof("Using IP from AWS metadata service: %s", ip)

		return ip, nil
	}

	if net.ParseIP(bindAddress) == nil {
		return "", fmt.Errorf("bindAddress is not valid IP address")
	}
	return bindAddress, nil
}

// kernelHasFilesystem checks if /proc/filesystems contains the specified filesystem
func kernelHasFilesystem(fs string) (bool, error) {
	contents, err := os.ReadFile("/proc/filesystems")
	if err != nil {
		return false, fmt.Errorf("error reading /proc/filesystems: %v", err)
	}

	for _, line := range strings.Split(string(contents), "\n") {
		tokens := strings.Fields(line)
		for _, token := range tokens {
			// Technically we should skip "nodev", but it doesn't matter
			if token == fs {
				return true, nil
			}
		}
	}

	return false, nil
}

// modprobe will exec `modprobe <module>`
func modprobe(module string) error {
	klog.Infof("Doing modprobe for module %v", module)
	out, err := exec.Command("/sbin/modprobe", module).CombinedOutput()
	outString := string(out)
	if err != nil {
		return fmt.Errorf("modprobe for module %q failed (%v): %s", module, err, outString)
	}
	if outString != "" {
		klog.Infof("Output from modprobe %s:\n%s", module, outString)
	}
	return nil
}

// loadKernelModules is a hack to force br_netfilter to be loaded
// TODO: Move to tasks architecture
func loadKernelModules(context *model.NodeupModelContext) error {
	err := modprobe("br_netfilter")
	if err != nil {
		// TODO: Return error in 1.11 (too risky for 1.10)
		klog.Warningf("error loading br_netfilter module: %v", err)
	}
	// TODO: Add to /etc/modules-load.d/ ?
	return nil
}

// getRegion queries the cloud provider for the region.
func getRegion(ctx context.Context, bootConfig *nodeup.BootConfig) (string, error) {
	switch bootConfig.CloudProvider {
	case api.CloudProviderAWS:
		region, err := awsup.RegionFromMetadata(ctx)
		if err != nil {
			return "", err
		}

		return region, nil
	}

	return "", nil
}

// seedRNG adds entropy to the random number generator.
func seedRNG(ctx context.Context, bootConfig *nodeup.BootConfig, region string) error {
	switch bootConfig.CloudProvider {
	case api.CloudProviderAWS:
		config := aws.NewConfig().WithCredentialsChainVerboseErrors(true).WithRegion(region)
		sess, err := session.NewSession(config)
		if err != nil {
			return err
		}

		random, err := kms.New(sess, config).GenerateRandom(&kms.GenerateRandomInput{
			NumberOfBytes: aws.Int64(64),
		})
		if err != nil {
			return fmt.Errorf("generating random seed: %v", err)
		}

		f, err := os.OpenFile("/dev/urandom", os.O_WRONLY, 0)
		if err != nil {
			return fmt.Errorf("opening /dev/urandom: %v", err)
		}
		_, err = f.Write(random.Plaintext)
		if err1 := f.Close(); err1 != nil && err == nil {
			err = err1
		}
		if err != nil {
			return fmt.Errorf("writing /dev/urandom: %v", err)
		}
	}

	return nil
}

// getNodeConfigFromServers queries kops-controllers for our node's configuration.
func getNodeConfigFromServers(ctx context.Context, bootConfig *nodeup.BootConfig, region string) (*nodeup.BootstrapResponse, error) {
	var authenticator bootstrap.Authenticator
	var resolver resolver.Resolver

	switch bootConfig.CloudProvider {
	case api.CloudProviderAWS:
		a, err := awsup.NewAWSAuthenticator(region)
		if err != nil {
			return nil, err
		}
		authenticator = a
	case api.CloudProviderGCE:
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
	case api.CloudProviderHetzner:
		a, err := hetzner.NewHetznerAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a
	case api.CloudProviderOpenstack:
		a, err := openstack.NewOpenstackAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a
	case api.CloudProviderDO:
		a, err := do.NewAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a
	case api.CloudProviderScaleway:
		a, err := scaleway.NewScalewayAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a
	case api.CloudProviderAzure:
		a, err := azure.NewAzureAuthenticator()
		if err != nil {
			return nil, err
		}
		authenticator = a

	case "metal":
		a, err := pkibootstrap.NewAuthenticatorFromFile("/etc/kubernetes/kops/pki/machine/private.pem")
		if err != nil {
			return nil, err
		}
		authenticator = a

	default:
		return nil, fmt.Errorf("unsupported cloud provider for node configuration %s", bootConfig.CloudProvider)
	}

	var challengeListener *bootstrap.ChallengeListener

	if kopsmodel.UseChallengeCallback(bootConfig.CloudProvider) {
		challengeServer, err := bootstrap.NewChallengeServer(bootConfig.ClusterName, []byte(bootConfig.ConfigServer.CACertificates))
		if err != nil {
			return nil, err
		}
		listen := ":" + strconv.Itoa(wellknownports.NodeupChallenge)

		l, err := challengeServer.NewListener(ctx, listen)
		if err != nil {
			return nil, fmt.Errorf("error starting challenge listener: %w", err)
		}
		challengeListener = l
		defer challengeListener.Stop()
	}

	client := &kopscontrollerclient.Client{
		Authenticator: authenticator,
		Resolver:      resolver,
		CAs:           []byte(bootConfig.ConfigServer.CACertificates),
	}

	var merr error
	for _, server := range bootConfig.ConfigServer.Servers {
		u, err := url.Parse(server)
		if err != nil {
			merr = multierr.Append(merr, fmt.Errorf("unable to parse configuration server url %q: %w", server, err))
			continue
		}
		client.BaseURL = *u

		request := nodeup.BootstrapRequest{
			APIVersion:        nodeup.BootstrapAPIVersion,
			IncludeNodeConfig: true,
		}

		if challengeListener != nil {
			request.Challenge = challengeListener.CreateChallenge()
		}

		var resp nodeup.BootstrapResponse
		err = client.Query(ctx, &request, &resp)
		if err != nil {
			merr = multierr.Append(merr, err)
			continue
		}
		return &resp, nil
	}
	return nil, merr
}

func getAWSConfigurationMode(ctx context.Context, c *model.NodeupModelContext) (string, error) {
	// Only worker nodes and apiservers can actually autoscale.
	// We are not adding describe permissions to the other roles
	role := c.BootConfig.InstanceGroupRole
	if role != api.InstanceGroupRoleNode && role != api.InstanceGroupRoleAPIServer {
		return "", nil
	}

	svc := c.Cloud.(awsup.AWSCloud).Autoscaling()

	result, err := svc.DescribeAutoScalingInstancesWithContext(ctx, &autoscaling.DescribeAutoScalingInstancesInput{
		InstanceIds: []*string{&c.InstanceID},
	})
	if err != nil {
		return "", fmt.Errorf("error describing instances: %v", err)
	}
	// If the instance is not a part of an ASG, it won't be in a warm pool either.
	if len(result.AutoScalingInstances) < 1 {
		return "", nil
	}
	lifecycle := fi.ValueOf(result.AutoScalingInstances[0].LifecycleState)
	if strings.HasPrefix(lifecycle, "Warmed:") {
		klog.Info("instance is entering warm pool")
		return model.ConfigurationModeWarming, nil
	} else {
		klog.Info("instance is entering the ASG")
		return "", nil
	}
}

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
package initscript

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"

	"k8s.io/kops/pkg/apis/kops"
	apimodel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
	"k8s.io/kops/util/pkg/vfs"
)

type nodeUpConfigBuilder struct {
	// Assets is a list of sources for files (primarily when not using everything containerized)
	// Formats:
	//  raw url: http://... or https://...
	//  url with hash: <hex>@http://... or <hex>@https://...
	assets map[architectures.Architecture][]*mirrors.MirroredAsset

	assetBuilder               *assets.AssetBuilder
	channels                   []string
	configBase                 vfs.Path
	cluster                    *kops.Cluster
	etcdManifests              map[string][]string
	images                     map[kops.InstanceGroupRole]map[architectures.Architecture][]*nodeup.Image
	protokubeAsset             map[architectures.Architecture][]*mirrors.MirroredAsset
	channelsAsset              map[architectures.Architecture][]*mirrors.MirroredAsset
	nodeUpAssets               map[architectures.Architecture]*mirrors.MirroredAsset
	encryptionConfigSecretHash string
}

func NewNodeUpConfigBuilder(cluster *kops.Cluster, assetBuilder *assets.AssetBuilder, assets map[architectures.Architecture][]*mirrors.MirroredAsset, encryptionConfigSecretHash string) (model.NodeUpConfigBuilder, error) {
	configBase, err := vfs.Context.BuildVfsPath(cluster.Spec.ConfigBase)
	if err != nil {
		return nil, fmt.Errorf("error parsing config base %q: %v", cluster.Spec.ConfigBase, err)
	}

	channels := []string{
		configBase.Join("addons", "bootstrap-channel.yaml").Path(),
	}

	for i := range cluster.Spec.Addons {
		channels = append(channels, cluster.Spec.Addons[i].Manifest)
	}

	protokubeAsset := map[architectures.Architecture][]*mirrors.MirroredAsset{}
	for _, arch := range architectures.GetSupported() {
		asset, err := ProtokubeAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		protokubeAsset[arch] = append(protokubeAsset[arch], asset)
	}

	channelsAsset := map[architectures.Architecture][]*mirrors.MirroredAsset{}
	for _, arch := range architectures.GetSupported() {
		asset, err := ChannelsAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		channelsAsset[arch] = append(channelsAsset[arch], asset)
	}

	nodeUpAssets := make(map[architectures.Architecture]*mirrors.MirroredAsset)
	for _, arch := range architectures.GetSupported() {
		asset, err := NodeUpAsset(assetBuilder, arch)
		if err != nil {
			return nil, err
		}
		nodeUpAssets[arch] = asset
	}

	images := map[kops.InstanceGroupRole]map[architectures.Architecture][]*nodeup.Image{}
	for _, role := range kops.AllInstanceGroupRoles {
		isMaster := role == kops.InstanceGroupRoleControlPlane
		isAPIServer := role == kops.InstanceGroupRoleAPIServer

		images[role] = make(map[architectures.Architecture][]*nodeup.Image)
		if components.IsBaseURL(cluster.Spec.KubernetesVersion) {
			// When using a custom version, we want to preload the images over http
			components := []string{"kube-proxy"}
			if isMaster {
				components = append(components, "kube-apiserver", "kube-controller-manager", "kube-scheduler")
			}
			if isAPIServer {
				components = append(components, "kube-apiserver")
			}

			for _, arch := range architectures.GetSupported() {
				for _, component := range components {
					baseURL, err := url.Parse(cluster.Spec.KubernetesVersion)
					if err != nil {
						return nil, err
					}

					baseURL.Path = path.Join(baseURL.Path, "/bin/linux", string(arch), component+".tar")

					u, hash, err := assetBuilder.RemapFileAndSHA(baseURL)
					if err != nil {
						return nil, err
					}

					image := &nodeup.Image{
						Sources: []string{u.String()},
						Hash:    hash.Hex(),
					}
					images[role][arch] = append(images[role][arch], image)
				}
			}
		}

		// `docker load` our images when using a KOPS_BASE_URL, so we
		// don't need to push/pull from a registry
		if os.Getenv("KOPS_BASE_URL") != "" && isMaster {
			for _, arch := range architectures.GetSupported() {
				for _, name := range []string{"kops-utils-cp", "kops-controller", "dns-controller", "kube-apiserver-healthcheck"} {
					baseURL, err := url.Parse(os.Getenv("KOPS_BASE_URL"))
					if err != nil {
						return nil, err
					}

					baseURL.Path = path.Join(baseURL.Path, "/images/"+name+"-"+string(arch)+".tar.gz")

					u, hash, err := assetBuilder.RemapFileAndSHA(baseURL)
					if err != nil {
						return nil, err
					}

					image := &nodeup.Image{
						Sources: []string{u.String()},
						Hash:    hash.Hex(),
					}
					images[role][arch] = append(images[role][arch], image)
				}
			}
		}
		if os.Getenv("KOPS_BASE_URL") != "" && isAPIServer {
			for _, arch := range architectures.GetSupported() {
				for _, name := range []string{"kube-apiserver-healthcheck"} {
					baseURL, err := url.Parse(os.Getenv("KOPS_BASE_URL"))
					if err != nil {
						return nil, err
					}

					baseURL.Path = path.Join(baseURL.Path, "/images/"+name+"-"+string(arch)+".tar.gz")

					u, hash, err := assetBuilder.RemapFileAndSHA(baseURL)
					if err != nil {
						return nil, err
					}

					image := &nodeup.Image{
						Sources: []string{u.String()},
						Hash:    hash.Hex(),
					}
					images[role][arch] = append(images[role][arch], image)
				}
			}
		}
	}

	etcdManifests := map[string][]string{}
	for _, role := range kops.AllInstanceGroupRoles {
		isMaster := role == kops.InstanceGroupRoleControlPlane
		if isMaster {
			for _, etcdCluster := range cluster.Spec.EtcdClusters {
				for _, member := range etcdCluster.Members {
					instanceGroup := fi.ValueOf(member.InstanceGroup)
					etcdManifest := fmt.Sprintf("manifests/etcd/%s-%s.yaml", etcdCluster.Name, instanceGroup)
					etcdManifests[instanceGroup] = append(etcdManifests[instanceGroup], configBase.Join(etcdManifest).Path())
				}
			}
		}
	}

	configBuilder := nodeUpConfigBuilder{
		assetBuilder:               assetBuilder,
		assets:                     assets,
		channels:                   channels,
		configBase:                 configBase,
		cluster:                    cluster,
		etcdManifests:              etcdManifests,
		images:                     images,
		protokubeAsset:             protokubeAsset,
		channelsAsset:              channelsAsset,
		nodeUpAssets:               nodeUpAssets,
		encryptionConfigSecretHash: encryptionConfigSecretHash,
	}

	return &configBuilder, nil
}

// NodeUpAssets implements model.NodeUpConfigBuilder.
func (n *nodeUpConfigBuilder) NodeUpAssets() map[architectures.Architecture]*mirrors.MirroredAsset {
	return n.nodeUpAssets
}

// BuildConfig returns the NodeUp config and auxiliary config.
func (n *nodeUpConfigBuilder) BuildConfig(ig *kops.InstanceGroup, apiserverAdditionalIPs []string, keysets map[string]*fi.Keyset) (*nodeup.Config, *nodeup.BootConfig, error) {
	cluster := n.cluster

	if ig == nil {
		return nil, nil, fmt.Errorf("instanceGroup cannot be nil")
	}

	role := ig.Spec.Role
	if role == "" {
		return nil, nil, fmt.Errorf("cannot determine role for instance group: %v", ig.ObjectMeta.Name)
	}

	usesLegacyGossip := cluster.UsesLegacyGossip()
	isMaster := role == kops.InstanceGroupRoleControlPlane
	hasAPIServer := isMaster || role == kops.InstanceGroupRoleAPIServer

	config, bootConfig := nodeup.NewConfig(cluster, ig)

	config.Assets = make(map[architectures.Architecture][]string)
	for _, arch := range architectures.GetSupported() {
		config.Assets[arch] = []string{}
		for _, a := range n.assets[arch] {
			config.Assets[arch] = append(config.Assets[arch], a.CompactString())
		}
	}

	if role != kops.InstanceGroupRoleBastion {
		if err := loadCertificates(keysets, fi.CertificateIDCA, config, true); err != nil {
			return nil, nil, err
		}
		if keysets["etcd-clients-ca-cilium"] != nil {
			if err := loadCertificates(keysets, "etcd-clients-ca-cilium", config, hasAPIServer || apimodel.UseKopsControllerForNodeBootstrap(n.cluster)); err != nil {
				return nil, nil, err
			}
		}

		if isMaster {
			if err := loadCertificates(keysets, "etcd-clients-ca", config, true); err != nil {
				return nil, nil, err
			}
			for _, etcdCluster := range cluster.Spec.EtcdClusters {
				k := etcdCluster.Name
				if err := loadCertificates(keysets, "etcd-manager-ca-"+k, config, true); err != nil {
					return nil, nil, err
				}
				if err := loadCertificates(keysets, "etcd-peers-ca-"+k, config, true); err != nil {
					return nil, nil, err
				}
				if k != "events" && k != "main" {
					if err := loadCertificates(keysets, "etcd-clients-ca-"+k, config, true); err != nil {
						return nil, nil, err
					}
				}
			}
			config.KeypairIDs["service-account"] = keysets["service-account"].Primary.Id
		} else {
			if keysets["etcd-client-cilium"] != nil {
				config.KeypairIDs["etcd-client-cilium"] = keysets["etcd-client-cilium"].Primary.Id
			}
		}

		if hasAPIServer {
			if err := loadCertificates(keysets, "apiserver-aggregator-ca", config, true); err != nil {
				return nil, nil, err
			}
			if keysets["etcd-clients-ca"] != nil {
				if err := loadCertificates(keysets, "etcd-clients-ca", config, true); err != nil {
					return nil, nil, err
				}
			}
			config.KeypairIDs["service-account"] = keysets["service-account"].Primary.Id

			config.APIServerConfig.EncryptionConfigSecretHash = n.encryptionConfigSecretHash
			serviceAccountPublicKeys, err := keysets["service-account"].ToPublicKeys()
			if err != nil {
				return nil, nil, fmt.Errorf("encoding service-account keys: %w", err)
			}
			config.APIServerConfig.ServiceAccountPublicKeys = serviceAccountPublicKeys
		} else {
			for _, key := range []string{"kubelet", "kube-proxy", "kube-router"} {
				if keysets[key] != nil {
					config.KeypairIDs[key] = keysets[key].Primary.Id
				}
			}
		}

		if isMaster || usesLegacyGossip {
			for _, arch := range architectures.GetSupported() {
				for _, a := range n.protokubeAsset[arch] {
					config.Assets[arch] = append(config.Assets[arch], a.CompactString())
				}
			}

			for _, arch := range architectures.GetSupported() {
				for _, a := range n.channelsAsset[arch] {
					config.Assets[arch] = append(config.Assets[arch], a.CompactString())
				}
			}
		}
	}

	if hasAPIServer {
		config.ApiserverAdditionalIPs = apiserverAdditionalIPs
	}

	// Set API server address to an IP from the cluster network CIDR
	var controlPlaneIPs []string
	switch cluster.Spec.GetCloudProvider() {
	case kops.CloudProviderAWS, kops.CloudProviderHetzner, kops.CloudProviderOpenstack:
		// Use a private IP address that belongs to the cluster network CIDR (some additional addresses may be FQDNs or public IPs)
		for _, additionalIP := range apiserverAdditionalIPs {
			for _, networkCIDR := range append(cluster.Spec.Networking.AdditionalNetworkCIDRs, cluster.Spec.Networking.NetworkCIDR) {
				_, cidr, err := net.ParseCIDR(networkCIDR)
				if err != nil {
					return nil, nil, fmt.Errorf("failed to parse network CIDR %q: %w", networkCIDR, err)
				}
				if cidr.Contains(net.ParseIP(additionalIP)) {
					controlPlaneIPs = append(controlPlaneIPs, additionalIP)
				}
			}
		}

	case kops.CloudProviderDO, kops.CloudProviderScaleway:
		// Use any IP address that is found (including public ones)
		for _, additionalIP := range apiserverAdditionalIPs {
			controlPlaneIPs = append(controlPlaneIPs, additionalIP)
		}

	case kops.CloudProviderGCE:
		// Use any IP address that is found (including public ones)
		for _, additionalIP := range apiserverAdditionalIPs {
			controlPlaneIPs = append(controlPlaneIPs, additionalIP)
		}
	}

	if cluster.UsesNoneDNS() {
		switch cluster.Spec.GetCloudProvider() {
		case kops.CloudProviderAWS, kops.CloudProviderHetzner, kops.CloudProviderOpenstack:
			bootConfig.APIServerIPs = controlPlaneIPs

		case kops.CloudProviderDO, kops.CloudProviderScaleway:
			bootConfig.APIServerIPs = controlPlaneIPs

		case kops.CloudProviderGCE:
			bootConfig.APIServerIPs = controlPlaneIPs

		default:
			return nil, nil, fmt.Errorf("'none' DNS topology is not supported for cloud %q", cluster.Spec.GetCloudProvider())
		}
	} else {
		// If we do have a fixed IP, we use it (on some clouds, initially)
		switch cluster.Spec.GetCloudProvider() {
		case kops.CloudProviderHetzner, kops.CloudProviderScaleway:
			bootConfig.APIServerIPs = controlPlaneIPs
		}
	}

	useConfigServer := apimodel.UseKopsControllerForNodeConfig(cluster) && !ig.HasAPIServer()
	if useConfigServer {
		hosts := []string{"kops-controller.internal." + cluster.ObjectMeta.Name}
		if len(bootConfig.APIServerIPs) > 0 {
			hosts = bootConfig.APIServerIPs
		}

		configServer := &nodeup.ConfigServerOptions{
			CACertificates: config.CAs[fi.CertificateIDCA],
		}
		for _, host := range hosts {
			baseURL := url.URL{
				Scheme: "https",
				Host:   net.JoinHostPort(host, strconv.Itoa(wellknownports.KopsControllerPort)),
				Path:   "/",
			}
			configServer.Servers = append(configServer.Servers, baseURL.String())
		}
		bootConfig.ConfigServer = configServer
		delete(config.CAs, fi.CertificateIDCA)
	} else {
		bootConfig.ConfigBase = fi.PtrTo(n.configBase.Path())
	}

	for _, manifest := range n.assetBuilder.StaticManifests {
		match := false
		for _, r := range manifest.Roles {
			if r == role {
				match = true
			}
		}

		if !match {
			continue
		}

		config.StaticManifests = append(config.StaticManifests, &nodeup.StaticManifest{
			Key:  manifest.Key,
			Path: manifest.Path,
		})
	}

	for _, staticFile := range n.assetBuilder.StaticFiles {
		match := false
		for _, r := range staticFile.Roles {
			if r == role {
				match = true
			}
		}

		if !match {
			continue
		}

		config.FileAssets = append(config.FileAssets, kops.FileAssetSpec{
			Content: staticFile.Content,
			Path:    staticFile.Path,
		})
	}

	config.Images = n.images[role]
	config.Channels = n.channels
	config.EtcdManifests = n.etcdManifests[ig.Name]

	if cluster.Spec.CloudProvider.AWS != nil {
		if ig.Spec.WarmPool != nil || cluster.Spec.CloudProvider.AWS.WarmPool != nil {
			config.WarmPoolImages = n.buildWarmPoolImages(ig)
		}
	}

	if ig.Spec.Packages != nil {
		config.Packages = ig.Spec.Packages
	}

	return config, bootConfig, nil
}

func loadCertificates(keysets map[string]*fi.Keyset, name string, config *nodeup.Config, includeKeypairID bool) error {
	keyset := keysets[name]
	if keyset == nil {
		return fmt.Errorf("key %q not found", name)
	}
	certificates, err := keyset.ToCertificateBytes()
	if err != nil {
		return fmt.Errorf("failed to read %q certificates: %w", name, err)
	}
	config.CAs[name] = string(certificates)
	if includeKeypairID {
		if keyset.Primary == nil {
			return fmt.Errorf("key %q did not have primary set", name)
		}
		config.KeypairIDs[name] = keyset.Primary.Id
	}
	return nil
}

// buildWarmPoolImages returns a list of container images that should be pre-pulled during instance pre-initialization
func (n *nodeUpConfigBuilder) buildWarmPoolImages(ig *kops.InstanceGroup) []string {
	if ig == nil || ig.Spec.Role == kops.InstanceGroupRoleControlPlane {
		return nil
	}

	images := map[string]bool{}

	// Add component and addon images that impact startup time
	// TODO: Exclude images that only run on control-plane nodes in a generic way
	desiredImagePrefixes := []string{
		// Ignore images hosted in private ECR repositories as containerd cannot actually pull these
		//"602401143452.dkr.ecr.us-west-2.amazonaws.com/", // Amazon VPC CNI
		// Ignore images hosted on docker.io until a solution for rate limiting is implemented
		//"docker.io/calico/",
		//"docker.io/cilium/",
		//"docker.io/cloudnativelabs/kube-router:",
		//"docker.io/weaveworks/",
		"registry.k8s.io/kube-proxy:",
		"registry.k8s.io/provider-aws/",
		"registry.k8s.io/sig-storage/csi-node-driver-registrar:",
		"registry.k8s.io/sig-storage/livenessprobe:",
		"quay.io/calico/",
		"quay.io/cilium/",
		"quay.io/coreos/flannel:",
		"quay.io/weaveworks/",
	}
	assetBuilder := n.assetBuilder
	if assetBuilder != nil {
		for _, image := range assetBuilder.ImageAssets {
			for _, prefix := range desiredImagePrefixes {
				if strings.HasPrefix(image.DownloadLocation, prefix) {
					images[image.DownloadLocation] = true
				}
			}
		}
	}

	var unique []string
	for image := range images {
		unique = append(unique, image)
	}
	sort.Strings(unique)

	return unique
}

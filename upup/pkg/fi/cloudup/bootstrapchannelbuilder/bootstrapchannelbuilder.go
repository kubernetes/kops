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

package bootstrapchannelbuilder

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"

	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/components/addonmanifests"
	"k8s.io/kops/pkg/model/components/addonmanifests/awscloudcontrollermanager"
	"k8s.io/kops/pkg/model/components/addonmanifests/awsebscsidriver"
	"k8s.io/kops/pkg/model/components/addonmanifests/awsloadbalancercontroller"
	"k8s.io/kops/pkg/model/components/addonmanifests/certmanager"
	"k8s.io/kops/pkg/model/components/addonmanifests/clusterautoscaler"
	"k8s.io/kops/pkg/model/components/addonmanifests/dnscontroller"
	"k8s.io/kops/pkg/model/components/addonmanifests/externaldns"
	"k8s.io/kops/pkg/model/components/addonmanifests/karpenter"
	"k8s.io/kops/pkg/model/components/addonmanifests/kuberouter"
	"k8s.io/kops/pkg/model/components/addonmanifests/nodeterminationhandler"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/templates"
	"k8s.io/kops/pkg/wellknownoperators"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
)

// BootstrapChannelBuilder is responsible for handling the addons in channels
type BootstrapChannelBuilder struct {
	*model.KopsModelContext
	ClusterAddons kubemanifest.ObjectList
	Lifecycle     fi.Lifecycle
	templates     *templates.Templates
	assetBuilder  *assets.AssetBuilder
}

var _ fi.ModelBuilder = &BootstrapChannelBuilder{}

// networkSelector is the labels set on networking addons
//
// The role.kubernetes.io/networking is used to label anything related to a networking addin,
// so that if we switch networking plugins (e.g. calico -> weave or vice-versa), we'll replace the
// old networking plugin, and there won't be old pods "floating around".
//
// This means whenever we create or update a networking plugin, we should be sure that:
// 1. the selector is role.kubernetes.io/networking=1
// 2. every object in the manifest is labeled with role.kubernetes.io/networking=1
//
// TODO: Some way to test/enforce this?
//
// TODO: Create "empty" configurations for others, so we can delete e.g. the kopeio configuration
// if we switch to kubenet?
//
// TODO: Create configuration object for cni providers (maybe create it but orphan it)?
//
// NOTE: we try to suffix with -kops.1, so that we can increment versions even if the upstream version
// hasn't changed.  The problem with semver is that there is nothing > 1.0.0 other than 1.0.1-pre.1
func networkingSelector() map[string]string {
	return map[string]string{"role.kubernetes.io/networking": "1"}
}

// NewBootstrapChannelBuilder creates a new BootstrapChannelBuilder
func NewBootstrapChannelBuilder(modelContext *model.KopsModelContext,
	clusterLifecycle fi.Lifecycle, assetBuilder *assets.AssetBuilder,
	templates *templates.Templates,
	addons kubemanifest.ObjectList,
) *BootstrapChannelBuilder {
	return &BootstrapChannelBuilder{
		KopsModelContext: modelContext,
		Lifecycle:        clusterLifecycle,
		assetBuilder:     assetBuilder,
		templates:        templates,
		ClusterAddons:    addons,
	}
}

// Build is responsible for adding the addons to the channel
func (b *BootstrapChannelBuilder) Build(c *fi.ModelBuilderContext) error {
	addons, serviceAccounts, err := b.buildAddons(c)
	if err != nil {
		return err
	}

	for _, a := range addons.Items {
		// Older versions of channels that may be running on the upgrading cluster requires Version to be set
		// We hardcode version to a high version to ensure an update is triggered on first run, and from then on
		// only a hash change will trigger an addon update.
		a.Spec.Version = "9.99.0"

		key := *a.Spec.Name
		if a.Spec.Id != "" {
			key = key + "-" + a.Spec.Id
		}
		name := b.Cluster.ObjectMeta.Name + "-addons-" + key
		manifestPath := "addons/" + *a.Spec.Manifest
		klog.V(4).Infof("Addon %q", name)

		manifestResource := b.templates.Find(manifestPath)
		if manifestResource == nil {
			return fmt.Errorf("unable to find manifest %s", manifestPath)
		}

		manifestBytes, err := fi.ResourceAsBytes(manifestResource)
		if err != nil {
			return fmt.Errorf("error reading manifest %s: %v", manifestPath, err)
		}

		// Go through any transforms that are best expressed as code
		remapped, err := addonmanifests.RemapAddonManifest(a.Spec, b.KopsModelContext, b.assetBuilder, manifestBytes, serviceAccounts)
		if err != nil {
			klog.Infof("invalid manifest: %s", string(manifestBytes))
			return fmt.Errorf("error remapping manifest %s: %v", manifestPath, err)
		}
		manifestBytes = remapped

		// Trim whitespace
		manifestBytes = []byte(strings.TrimSpace(string(manifestBytes)))

		a.ManifestData = manifestBytes

		rawManifest := string(manifestBytes)
		klog.V(4).Infof("Manifest %v", rawManifest)

		manifestHash, err := utils.HashString(rawManifest)
		klog.V(4).Infof("hash %s", manifestHash)
		if err != nil {
			return fmt.Errorf("error hashing manifest: %v", err)
		}
		a.Spec.ManifestHash = manifestHash

		c.AddTask(&fitasks.ManagedFile{
			Contents:  fi.NewBytesResource(manifestBytes),
			Lifecycle: b.Lifecycle,
			Location:  fi.String(manifestPath),
			Name:      fi.String(name),
		})
	}

	if featureflag.UseAddonOperators.Enabled() {
		ob := &wellknownoperators.Builder{
			Cluster: b.Cluster,
		}

		addonPackages, clusterAddons, err := ob.Build(b.ClusterAddons)
		if err != nil {
			return fmt.Errorf("error building well-known operators: %v", err)
		}
		b.ClusterAddons = clusterAddons

		for _, a := range addonPackages {
			key := *a.Spec.Name
			if a.Spec.Id != "" {
				key = key + "-" + a.Spec.Id
			}
			name := b.Cluster.ObjectMeta.Name + "-addons-" + key
			manifestPath := "addons/" + *a.Spec.Manifest

			// Go through any transforms that are best expressed as code
			manifestBytes, err := addonmanifests.RemapAddonManifest(&a.Spec, b.KopsModelContext, b.assetBuilder, a.Manifest, serviceAccounts)
			if err != nil {
				klog.Infof("invalid manifest: %s", string(a.Manifest))
				return fmt.Errorf("error remapping manifest %s: %v", manifestPath, err)
			}

			// Trim whitespace
			manifestBytes = []byte(strings.TrimSpace(string(manifestBytes)))

			rawManifest := string(manifestBytes)
			klog.V(4).Infof("Manifest %v", rawManifest)

			manifestHash, err := utils.HashString(rawManifest)
			klog.V(4).Infof("hash %s", manifestHash)
			if err != nil {
				return fmt.Errorf("error hashing manifest: %v", err)
			}
			a.Spec.ManifestHash = manifestHash

			c.AddTask(&fitasks.ManagedFile{
				Contents:  fi.NewBytesResource(manifestBytes),
				Lifecycle: b.Lifecycle,
				Location:  fi.String(manifestPath),
				Name:      fi.String(name),
			})

			addon := addons.Add(&a.Spec)
			addon.ManifestData = manifestBytes
		}
	}

	// Not all objects in ClusterAddons should be applied to the cluster - although most should.
	// However, there are a handful of well-known exceptions:
	// e.g. configuration objects which are instead configured via files on the nodes.
	var applyAdditionalObjectsToCluster kubemanifest.ObjectList
	if b.ClusterAddons != nil {
		for _, addon := range b.ClusterAddons {
			applyToCluster := true

			switch addon.GroupVersionKind().GroupKind() {
			case schema.GroupKind{Group: "kubescheduler.config.k8s.io", Kind: "KubeSchedulerConfiguration"}:
				applyToCluster = false
			}

			if applyToCluster {
				applyAdditionalObjectsToCluster = append(applyAdditionalObjectsToCluster, addon)
			}
		}
	}

	if len(applyAdditionalObjectsToCluster) != 0 {
		key := "cluster-addons.kops.k8s.io"
		location := key + "/default.yaml"

		a := &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		}

		name := b.Cluster.ObjectMeta.Name + "-addons-" + key
		manifestPath := "addons/" + *a.Manifest

		manifestBytes, err := applyAdditionalObjectsToCluster.ToYAML()
		if err != nil {
			return fmt.Errorf("error serializing addons: %v", err)
		}

		// Trim whitespace
		manifestBytes = []byte(strings.TrimSpace(string(manifestBytes)))

		rawManifest := string(manifestBytes)

		manifestHash, err := utils.HashString(rawManifest)
		if err != nil {
			return fmt.Errorf("error hashing manifest: %v", err)
		}
		a.ManifestHash = manifestHash

		c.AddTask(&fitasks.ManagedFile{
			Contents:  fi.NewBytesResource(manifestBytes),
			Lifecycle: b.Lifecycle,
			Location:  fi.String(manifestPath),
			Name:      fi.String(name),
		})

		addons.Add(a)
	}

	if err := b.addPruneDirectives(addons); err != nil {
		return err
	}

	addonsObject := &channelsapi.Addons{}
	addonsObject.Kind = "Addons"
	addonsObject.ObjectMeta.Name = "bootstrap"
	for _, addon := range addons.Items {
		addonsObject.Spec.Addons = append(addonsObject.Spec.Addons, addon.Spec)
	}

	if err := addonsObject.Verify(); err != nil {
		return err
	}

	addonsYAML, err := utils.YamlMarshal(addonsObject)
	if err != nil {
		return fmt.Errorf("error serializing addons yaml: %v", err)
	}

	name := b.Cluster.ObjectMeta.Name + "-addons-bootstrap"

	c.AddTask(&fitasks.ManagedFile{
		Contents:  fi.NewBytesResource(addonsYAML),
		Lifecycle: b.Lifecycle,
		Location:  fi.String("addons/bootstrap-channel.yaml"),
		Name:      fi.String(name),
	})

	return nil
}

type AddonList struct {
	Items []*Addon
}

func (a *AddonList) Add(spec *channelsapi.AddonSpec) *Addon {
	addon := &Addon{
		Spec: spec,
	}
	a.Items = append(a.Items, addon)
	return addon
}

type Addon struct {
	// Spec is the spec that will (eventually) be passed to the channels binary.
	Spec *channelsapi.AddonSpec

	// ManifestData is the object data loaded from the manifest.
	ManifestData []byte

	// BuildPrune is set if we should automatically build prune specifiers, based on the manifest.
	BuildPrune bool
}

func (b *BootstrapChannelBuilder) buildAddons(c *fi.ModelBuilderContext) (*AddonList, map[string]iam.Subject, error) {
	addons := &AddonList{}

	serviceAccountRoles := []iam.Subject{}

	{
		key := "kops-controller.addons.k8s.io"

		{
			location := key + "/k8s-1.16.yaml"
			id := "k8s-1.16"

			addons.Add(&channelsapi.AddonSpec{
				Name:               fi.String(key),
				Selector:           map[string]string{"k8s-addon": key},
				Manifest:           fi.String(location),
				NeedsRollingUpdate: "control-plane",
				Id:                 id,
			})
		}
	}

	// @check if podsecuritypolicies are enabled and if so, push the default kube-system policy
	if b.Cluster.Spec.KubeAPIServer != nil && b.Cluster.Spec.KubeAPIServer.HasAdmissionController("PodSecurityPolicy") {
		key := "podsecuritypolicy.addons.k8s.io"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	kubeDNS := b.Cluster.Spec.KubeDNS

	if kubeDNS.Provider == "" {
		kubeDNS.Provider = "CoreDNS"
	}

	if kubeDNS.Provider == "KubeDNS" {
		{
			key := "kube-dns.addons.k8s.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	if kubeDNS.Provider == "CoreDNS" && !featureflag.UseAddonOperators.Enabled() {
		{
			key := "coredns.addons.k8s.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	// @check if bootstrap tokens are enabled an if so we can forgo applying
	// this manifest. For clusters whom are upgrading from RBAC to Node,RBAC the clusterrolebinding
	// will remain and have to be deleted manually once all the nodes have been upgraded.
	enableRBACAddon := true
	if b.UseKopsControllerForNodeBootstrap() {
		enableRBACAddon = false
	}
	if b.Cluster.Spec.KubeAPIServer != nil {
		if b.Cluster.Spec.KubeAPIServer.EnableBootstrapAuthToken != nil && *b.Cluster.Spec.KubeAPIServer.EnableBootstrapAuthToken {
			enableRBACAddon = false
		}
	}

	if enableRBACAddon {
		{
			key := "rbac.addons.k8s.io"

			{
				location := key + "/k8s-1.8.yaml"
				id := "k8s-1.8"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	{
		// Adding the kubelet-api-admin binding: this is required when switching to webhook authorization on the kubelet
		// docs: https://kubernetes.io/docs/reference/access-authn-authz/rbac/#other-component-roles
		// issue: https://github.com/kubernetes/kops/issues/5176
		key := "kubelet-api.rbac.addons.k8s.io"

		{
			location := key + "/k8s-1.9.yaml"
			id := "k8s-1.9"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.IsKubernetesGTE("1.23") && b.IsKubernetesLT("1.26") &&
		(b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderAWS ||
			b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderGCE) {
		// AWS and GCE KCM-to-CCM leader migration
		key := "leader-migration.rbac.addons.k8s.io"

		{
			location := key + "/k8s-1.23.yaml"
			id := "k8s-1.23"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	{
		key := "limit-range.addons.k8s.io"
		version := "1.5.0"
		location := key + "/v" + version + ".yaml"

		addons.Add(&channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
	}

	if b.Cluster.Spec.ExternalDNS == nil || b.Cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderDNSController {
		{
			key := "dns-controller.addons.k8s.io"
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}

		// Generate dns-controller ServiceAccount IAM permissions.
		// Gossip clsuters do not require any cloud permissions.
		if b.UseServiceAccountExternalPermissions() && !b.IsGossip() {
			serviceAccountRoles = append(serviceAccountRoles, &dnscontroller.ServiceAccount{})
		}
	} else if b.Cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderExternalDNS {
		{
			key := "external-dns.addons.k8s.io"

			{
				location := key + "/k8s-1.19.yaml"
				id := "k8s-1.19"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}

			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &externaldns.ServiceAccount{})
			}
		}
	}

	// @check if the node-local-dns is enabled
	NodeLocalDNS := b.Cluster.Spec.KubeDNS.NodeLocalDNS
	if kubeDNS.Provider == "CoreDNS" && NodeLocalDNS != nil && fi.BoolValue(NodeLocalDNS.Enabled) {
		{
			key := "nodelocaldns.addons.k8s.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	if b.Cluster.Spec.ClusterAutoscaler != nil && fi.BoolValue(b.Cluster.Spec.ClusterAutoscaler.Enabled) {
		{
			key := "cluster-autoscaler.addons.k8s.io"

			{
				location := key + "/k8s-1.15.yaml"
				id := "k8s-1.15"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}

		if b.UseServiceAccountExternalPermissions() {
			serviceAccountRoles = append(serviceAccountRoles, &clusterautoscaler.ServiceAccount{})
		}

	}

	if b.Cluster.Spec.MetricsServer != nil && fi.BoolValue(b.Cluster.Spec.MetricsServer.Enabled) {
		{
			key := "metrics-server.addons.k8s.io"

			{
				location := key + "/k8s-1.11.yaml"
				id := "k8s-1.11"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-app": "metrics-server"},
					Manifest: fi.String(location),
					Id:       id,
					NeedsPKI: !fi.BoolValue(b.Cluster.Spec.MetricsServer.Insecure),
				})
			}
		}
	}

	if b.Cluster.Spec.CertManager != nil && fi.BoolValue(b.Cluster.Spec.CertManager.Enabled) && (b.Cluster.Spec.CertManager.Managed == nil || fi.BoolValue(b.Cluster.Spec.CertManager.Managed)) {
		{
			key := "certmanager.io"

			{
				location := key + "/k8s-1.16.yaml"
				id := "k8s-1.16"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}

		if len(b.Cluster.Spec.CertManager.HostedZoneIDs) > 0 {
			serviceAccountRoles = append(serviceAccountRoles, &certmanager.ServiceAccount{})
		}
	}

	nth := b.Cluster.Spec.NodeTerminationHandler

	if nth != nil && fi.BoolValue(nth.Enabled) {

		key := "node-termination-handler.aws"

		{
			location := key + "/k8s-1.11.yaml"
			id := "k8s-1.11"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}

		if b.UseServiceAccountExternalPermissions() {
			serviceAccountRoles = append(serviceAccountRoles, &nodeterminationhandler.ServiceAccount{})
		}
	}

	npd := b.Cluster.Spec.NodeProblemDetector

	if npd != nil && fi.BoolValue(npd.Enabled) {

		key := "node-problem-detector.addons.k8s.io"

		{
			location := key + "/k8s-1.17.yaml"
			id := "k8s-1.17"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	nvidia := b.Cluster.Spec.Containerd.NvidiaGPU
	igNvidia := false
	for _, ig := range b.KopsModelContext.InstanceGroups {
		if ig.Spec.Containerd != nil && ig.Spec.Containerd.NvidiaGPU != nil && fi.BoolValue(ig.Spec.Containerd.NvidiaGPU.Enabled) {
			igNvidia = true
			break
		}
	}

	if nvidia != nil && fi.BoolValue(nvidia.Enabled) || igNvidia {

		key := "nvidia.addons.k8s.io"

		{
			location := key + "/k8s-1.16.yaml"
			id := "k8s-1.16"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.AWSLoadBalancerController != nil && fi.BoolValue(b.Cluster.Spec.AWSLoadBalancerController.Enabled) {

		key := "aws-load-balancer-controller.addons.k8s.io"

		location := key + "/k8s-1.19.yaml"
		id := "k8s-1.19"

		addons.Add(&channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
			Id:       id,
			NeedsPKI: true,
		})

		// Generate aws-load-balancer-controller ServiceAccount IAM permissions
		if b.UseServiceAccountExternalPermissions() {
			serviceAccountRoles = append(serviceAccountRoles, &awsloadbalancercontroller.ServiceAccount{})
		}
	}

	if b.Cluster.Spec.PodIdentityWebhook != nil && fi.BoolValue(&b.Cluster.Spec.PodIdentityWebhook.Enabled) {

		key := "eks-pod-identity-webhook.addons.k8s.io"

		{
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name: fi.String(key),
				Selector: map[string]string{
					"k8s-addon": key,
				},
				Manifest: fi.String(location),
				Id:       id,
				NeedsPKI: true,
			})
		}
	}

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderAWS {
		key := "storage-aws.addons.k8s.io"

		{
			id := "v1.15.0"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderDO {
		key := "digitalocean-cloud-controller.addons.k8s.io"

		{
			id := "k8s-1.8"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}

		key = "digitalocean-csi-driver.addons.k8s.io"

		{
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderHetzner {
		{
			key := "hcloud-cloud-controller.addons.k8s.io"
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
		{
			key := "hcloud-csi-driver.addons.k8s.io"
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderGCE {
		key := "storage-gce.addons.k8s.io"

		{
			id := "v1.7.0"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}

		if b.Cluster.Spec.CloudConfig != nil && b.Cluster.Spec.CloudConfig.GCPPDCSIDriver != nil && fi.BoolValue(b.Cluster.Spec.CloudConfig.GCPPDCSIDriver.Enabled) {
			key := "gcp-pd-csi-driver.addons.k8s.io"
			{
				id := "k8s-1.23"
				location := key + "/" + id + ".yaml"
				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
		}
	}

	if featureflag.Spotinst.Enabled() && featureflag.SpotinstController.Enabled() {
		key := "spotinst-kubernetes-cluster-controller.addons.k8s.io"

		{
			id := "v1.14.0"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	// The metadata-proxy daemonset conceals node metadata endpoints in GCE.
	// It will land on nodes labeled cloud.google.com/metadata-proxy-ready=true
	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderGCE {
		key := "metadata-proxy.addons.k8s.io"

		{
			id := "v0.1.12"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}

		if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderGCE {
			if b.Cluster.Spec.ExternalCloudControllerManager != nil {
				key := "gcp-cloud-controller.addons.k8s.io"
				{
					id := "k8s-1.23"
					location := key + "/" + id + ".yaml"
					addon := addons.Add(&channelsapi.AddonSpec{
						Name:     fi.String(key),
						Manifest: fi.String(location),
						Selector: map[string]string{"k8s-addon": key},
						Id:       id,
					})
					addon.BuildPrune = true
				}
			}
		}
	}

	if b.Cluster.Spec.Networking.Kopeio != nil && !featureflag.UseAddonOperators.Enabled() {
		key := "networking.kope.io"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})

			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.Weave != nil {
		key := "networking.weave"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Flannel != nil {
		key := "networking.flannel"

		if b.IsKubernetesGTE("v1.25.0") {
			id := "k8s-1.25"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		} else {
			id := "k8s-1.12"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.Calico != nil {
		key := "networking.projectcalico.org"

		if b.IsKubernetesGTE("v1.25.0") {
			id := "k8s-1.25"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		} else if b.IsKubernetesGTE("v1.22.0") {
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		} else {
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.Canal != nil {
		key := "networking.projectcalico.org.canal"

		if b.IsKubernetesGTE("v1.25.0") {
			id := "k8s-1.25"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		} else if b.IsKubernetesGTE("v1.22.0") {
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		} else {
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.Kuberouter != nil {
		key := "networking.kuberouter"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}

		// Generate kube-router ServiceAccount IAM permissions
		if b.UseServiceAccountExternalPermissions() {
			serviceAccountRoles = append(serviceAccountRoles, &kuberouter.ServiceAccount{})
		}
	}

	if b.Cluster.Spec.Networking.AmazonVPC != nil {
		key := "networking.amazon-vpc-routed-eni"

		{
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:               fi.String(key),
				Selector:           networkingSelector(),
				Manifest:           fi.String(location),
				Id:                 id,
				NeedsRollingUpdate: "all",
			})
		}
	}

	err := addCiliumAddon(b, addons)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to add cilium addon: %w", err)
	}

	authenticationSelector := map[string]string{"role.kubernetes.io/authentication": "1"}

	if b.Cluster.Spec.Authentication != nil {
		if b.Cluster.Spec.Authentication.Kopeio != nil {
			key := "authentication.kope.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: authenticationSelector,
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
		if b.Cluster.Spec.Authentication.AWS != nil {
			key := "authentication.aws"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: authenticationSelector,
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderOpenstack {
		{
			key := "storage-openstack.addons.k8s.io"

			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Manifest: fi.String(location),
				Selector: map[string]string{"k8s-addon": key},
				Id:       id,
			})
			addon.BuildPrune = true
		}

		if b.Cluster.Spec.ExternalCloudControllerManager != nil {
			// cloudprovider specific out-of-tree controller
			{
				key := "openstack.addons.k8s.io"

				location := key + "/k8s-1.13.yaml"
				id := "k8s-1.13-ccm"

				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
		}
	}

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderAWS {

		if b.Cluster.Spec.ExternalCloudControllerManager != nil {
			key := "aws-cloud-controller.addons.k8s.io"

			{
				id := "k8s-1.18"
				location := key + "/" + id + ".yaml"
				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &awscloudcontrollermanager.ServiceAccount{})
			}
		}
		if b.Cluster.Spec.CloudConfig != nil && b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver != nil &&
			fi.BoolValue(b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver.Enabled) &&
			(b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver.Managed == nil || fi.BoolValue(b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver.Managed)) {
			key := "aws-ebs-csi-driver.addons.k8s.io"

			{
				id := "k8s-1.17"
				location := key + "/" + id + ".yaml"
				addons.Add(&channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &awsebscsidriver.ServiceAccount{})
			}
		}
	}

	if b.Cluster.Spec.SnapshotController != nil && fi.BoolValue(b.Cluster.Spec.SnapshotController.Enabled) {
		key := "snapshot-controller.addons.k8s.io"

		{
			id := "k8s-1.20"
			location := key + "/" + id + ".yaml"
			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Manifest: fi.String(location),
				Selector: map[string]string{"k8s-addon": key},
				NeedsPKI: true,
				Id:       id,
			})
		}
	}
	if b.Cluster.Spec.Karpenter != nil && fi.BoolValue(&b.Cluster.Spec.Karpenter.Enabled) {
		key := "karpenter.sh"

		{
			id := "k8s-1.19"
			location := key + "/" + id + ".yaml"
			addons.Add(&channelsapi.AddonSpec{
				Name:     fi.String(key),
				Manifest: fi.String(location),
				Selector: map[string]string{"k8s-addon": key},
				Id:       id,
			})
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &karpenter.ServiceAccount{})
			}
		}
	}

	if b.Cluster.Spec.KubeScheduler.UsePolicyConfigMap != nil {
		key := "scheduler.addons.k8s.io"
		version := "1.7.0"
		location := key + "/v" + version + ".yaml"

		addons.Add(&channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
	}

	serviceAccounts := make(map[string]iam.Subject)

	if b.Cluster.Spec.GetCloudProvider() == kops.CloudProviderAWS && b.Cluster.Spec.KubeAPIServer.ServiceAccountIssuer != nil {
		awsModelContext := &awsmodel.AWSModelContext{
			KopsModelContext: b.KopsModelContext,
		}

		for _, serviceAccountRole := range serviceAccountRoles {
			iamModelBuilder := &awsmodel.IAMModelBuilder{AWSModelContext: awsModelContext, Lifecycle: b.Lifecycle, Cluster: b.Cluster}

			_, err := iamModelBuilder.BuildServiceAccountRoleTasks(serviceAccountRole, c)
			if err != nil {
				return nil, nil, err
			}
			sa, _ := serviceAccountRole.ServiceAccount()
			serviceAccounts[sa.Name] = serviceAccountRole
		}
	}
	return addons, serviceAccounts, nil
}

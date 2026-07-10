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

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	"k8s.io/kops/util/pkg/vfs"
)

// BootstrapChannelBuilder is responsible for handling the addons in channels
type BootstrapChannelBuilder struct {
	*model.KopsModelContext
	ClusterAddons kubemanifest.ObjectList
	Lifecycle     fi.Lifecycle
	templates     *templates.Templates
	assetBuilder  *assets.AssetBuilder

	addonRenderer AddonTemplateRenderer
}

var _ fi.CloudupModelBuilder = (*BootstrapChannelBuilder)(nil)

// AddonTemplateRenderer renders addon manifest templates against a per-call func map
// bound to the given task graph, and exposes a few direct template-function calls
// (e.g. CloudControllerConfigArgv) used by the channel builder itself.
type AddonTemplateRenderer interface {
	RenderTemplate(name string, source []byte, tasks map[string]fi.CloudupTask) ([]byte, error)
	CloudControllerConfigArgv() ([]string, error)
}

// networkingSelector is the labels set on networking addons
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
	addonRenderer AddonTemplateRenderer,
) *BootstrapChannelBuilder {
	return &BootstrapChannelBuilder{
		KopsModelContext: modelContext,
		Lifecycle:        clusterLifecycle,
		assetBuilder:     assetBuilder,
		templates:        templates,
		ClusterAddons:    addons,
		addonRenderer:    addonRenderer,
	}
}

// Build is responsible for adding the addons to the channel.
func (b *BootstrapChannelBuilder) Build(c *fi.CloudupModelBuilderContext) error {
	addons, serviceAccounts, err := b.buildAddons(c)
	if err != nil {
		return err
	}

	for _, addon := range addons.Items {
		// Addons with a preset source do not have a template.
		if addon.Source != nil {
			continue
		}
		manifestPath := "addons/" + *addon.Spec.Manifest
		manifestResource := b.templates.Find(manifestPath)
		if manifestResource == nil {
			return fmt.Errorf("unable to find manifest %s", manifestPath)
		}
		addon.Source = manifestResource
		addon.SkipRender = !b.templates.IsTemplate(manifestPath)
	}

	if featureflag.UseAddonOperators.Enabled() {
		ob := &wellknownoperators.Builder{
			VFSContext: vfs.Context,
			Cluster:    b.Cluster,
		}

		addonPackages, clusterAddons, err := ob.Build(b.ClusterAddons)
		if err != nil {
			return fmt.Errorf("error building well-known operators: %v", err)
		}
		b.ClusterAddons = clusterAddons

		for _, pkg := range addonPackages {
			addons.AddWithSource(&pkg.Spec, fi.NewBytesResource(pkg.Manifest))
		}
	}

	// Not all objects in ClusterAddons should be applied to the cluster - although most should.
	// However, there are a handful of well-known exceptions:
	// e.g. configuration objects which are instead configured via files on the nodes.
	var applyAdditionalObjectsToCluster kubemanifest.ObjectList
	if b.ClusterAddons != nil {
		for _, addon := range b.ClusterAddons {
			applyToCluster := true

			if addon.GroupVersionKind().GroupKind() == (schema.GroupKind{Group: "kubescheduler.config.k8s.io", Kind: "KubeSchedulerConfiguration"}) {
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

		addon := addons.Add(&channelsapi.AddonSpec{
			Name:     new(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: new(location),
		})
		addon.SkipRemap = true

		manifestBytes, err := applyAdditionalObjectsToCluster.ToYAML()
		if err != nil {
			return fmt.Errorf("error serializing addons: %v", err)
		}
		addon.Source = fi.NewBytesResource(manifestBytes)
		addon.SkipRender = true
	}

	preRegisterAddonImages := b.shouldPreRegisterAddonImages()

	var addonTasks []*AddonManifest
	for _, addon := range addons.Items {
		if preRegisterAddonImages {
			if err := addon.CollectImages(b.assetBuilder, b.addonRenderer); err != nil {
				klog.Warningf("unable to pre-register addon images for %q: %v", addonKey(addon.Spec), err)
			}
		}

		task := &AddonManifest{
			Name:      new(b.Cluster.ObjectMeta.Name + "-addons-" + addonKey(addon.Spec)),
			Lifecycle: b.Lifecycle,
			Location:  new("addons/" + *addon.Spec.Manifest),

			addonRenderer:   b.addonRenderer,
			source:          addon.Source,
			addonSpec:       addon.Spec,
			buildPrune:      addon.BuildPrune,
			skipRemap:       addon.SkipRemap,
			skipRender:      addon.SkipRender,
			modelContext:    b.KopsModelContext,
			assetBuilder:    b.assetBuilder,
			serviceAccounts: serviceAccounts,
		}
		c.AddTask(task)
		addonTasks = append(addonTasks, task)
	}

	c.AddTask(&BootstrapChannel{
		Name:           new(b.Cluster.ObjectMeta.Name + "-addons-bootstrap"),
		Lifecycle:      b.Lifecycle,
		Location:       new("addons/bootstrap-channel.yaml"),
		addonManifests: addonTasks,
	})

	return nil
}

func (b *BootstrapChannelBuilder) shouldPreRegisterAddonImages() bool {
	if b.assetBuilder == nil || b.Cluster == nil || b.Cluster.Spec.CloudProvider.AWS == nil {
		return false
	}

	if b.Cluster.Spec.CloudProvider.AWS.WarmPool != nil {
		return true
	}

	for _, ig := range b.AllInstanceGroups {
		if ig != nil && ig.Spec.WarmPool != nil {
			return true
		}
	}

	return false
}

func addonKey(spec *channelsapi.AddonSpec) string {
	key := *spec.Name
	if spec.Id != "" {
		key = key + "-" + spec.Id
	}
	return key
}

type AddonList struct {
	Items []*Addon
}

func (a *AddonList) Add(spec *channelsapi.AddonSpec) *Addon {
	return a.AddWithSource(spec, nil)
}

func (a *AddonList) AddWithSource(spec *channelsapi.AddonSpec, source fi.Resource) *Addon {
	addon := &Addon{
		Spec:       spec,
		Source:     source,
		SkipRender: source != nil,
	}
	a.Items = append(a.Items, addon)
	return addon
}

type Addon struct {
	// Spec is the spec that will (eventually) be passed to the kops-channels static pod.
	Spec *channelsapi.AddonSpec

	// Source is the manifest template or static bytes used to build the addon file.
	Source fi.Resource

	// BuildPrune is set if we should automatically build prune specifiers, based on the manifest.
	BuildPrune bool

	// SkipRemap bypasses label stamping, service-account role injection, and asset image remapping.
	SkipRemap bool

	// SkipRender is true when Source is already a rendered manifest.
	SkipRender bool
}

// CollectImages renders template sources under stubbed task funcs and scans
// the resulting YAML for image references to pre-register with the builder.
// Best-effort only (used for AWS WarmPool image prewarm): a failure here
// warns rather than breaks the build.
func (a *Addon) CollectImages(assetBuilder *assets.AssetBuilder, renderer AddonTemplateRenderer) error {
	if a == nil || assetBuilder == nil || a.Source == nil {
		return nil
	}

	manifestBytes, err := fi.ResourceAsBytes(a.Source)
	if err != nil {
		return err
	}
	if !a.SkipRender && renderer != nil {
		manifestBytes, err = renderer.RenderTemplate(addonKey(a.Spec), manifestBytes, nil)
		if err != nil {
			return fmt.Errorf("rendering addon %q for image discovery: %w", addonKey(a.Spec), err)
		}
	}
	_, err = assetBuilder.RemapManifest(manifestBytes)
	return err
}

func (b *BootstrapChannelBuilder) buildAddons(c *fi.CloudupModelBuilderContext) (*AddonList, map[types.NamespacedName]iam.Subject, error) {
	addons := &AddonList{}

	serviceAccountRoles := []iam.Subject{}

	{
		key := "kops-controller.addons.k8s.io"

		{
			location := key + "/k8s-1.16.yaml"
			id := "k8s-1.16"

			addons.Add(&channelsapi.AddonSpec{
				Name:               new(key),
				Selector:           map[string]string{"k8s-addon": key},
				Manifest:           new(location),
				NeedsRollingUpdate: channelsapi.NeedsRollingUpdateControlPlane,
				Id:                 id,
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
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
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
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
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
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
	}

	{
		key := "limit-range.addons.k8s.io"
		version := "1.5.0"
		location := key + "/v" + version + ".yaml"

		addons.Add(&channelsapi.AddonSpec{
			Name:     new(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: new(location),
		})
	}

	{
		key := "dns-controller.addons.k8s.io"
		location := key + "/k8s-1.12.yaml"
		id := "k8s-1.12"

		spec := &channelsapi.AddonSpec{
			Name:     new(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: new(location),
			Id:       id,
		}

		if !b.Cluster.UsesNoneDNS() && (b.Cluster.Spec.ExternalDNS == nil || b.Cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderDNSController) {
			addons.Add(spec)

			// Generate dns-controller ServiceAccount IAM permissions. Gossip and dns=none clusters do
			// not require any cloud permissions.
			if b.UseServiceAccountExternalPermissions() && b.Cluster.PublishesDNSRecords() {
				serviceAccountRoles = append(serviceAccountRoles, &dnscontroller.ServiceAccount{})
			}
		} else {
			// dns-controller is not used, but the addon is still applied with an empty manifest, so
			// that pruning removes any resources deployed before migrating to dns=none.
			addon := addons.AddWithSource(spec, fi.NewBytesResource(nil))
			addon.BuildPrune = true
			addon.SkipRemap = true
		}
	}

	if !b.Cluster.UsesNoneDNS() && b.Cluster.Spec.ExternalDNS != nil && b.Cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderExternalDNS {
		key := "external-dns.addons.k8s.io"
		location := key + "/k8s-1.19.yaml"
		id := "k8s-1.19"

		addons.Add(&channelsapi.AddonSpec{
			Name:     new(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: new(location),
			Id:       id,
		})

		if b.UseServiceAccountExternalPermissions() {
			serviceAccountRoles = append(serviceAccountRoles, &externaldns.ServiceAccount{})
		}
	}

	// @check if the node-local-dns is enabled
	NodeLocalDNS := b.Cluster.Spec.KubeDNS.NodeLocalDNS
	if kubeDNS.Provider == "CoreDNS" && NodeLocalDNS != nil && fi.ValueOf(NodeLocalDNS.Enabled) {
		{
			key := "nodelocaldns.addons.k8s.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Add(&channelsapi.AddonSpec{
					Name:               new(key),
					Selector:           map[string]string{"k8s-addon": key},
					Manifest:           new(location),
					NeedsRollingUpdate: channelsapi.NeedsRollingUpdateAll,
					Id:                 id,
				})
			}
		}
	}

	if b.Cluster.Spec.ClusterAutoscaler != nil && fi.ValueOf(b.Cluster.Spec.ClusterAutoscaler.Enabled) {
		{
			key := "cluster-autoscaler.addons.k8s.io"

			{
				location := key + "/k8s-1.15.yaml"
				id := "k8s-1.15"

				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
					Id:       id,
				})
			}
		}

		if b.UseServiceAccountExternalPermissions() {
			serviceAccountRoles = append(serviceAccountRoles, &clusterautoscaler.ServiceAccount{})
		}

	}

	if b.Cluster.Spec.MetricsServer != nil && fi.ValueOf(b.Cluster.Spec.MetricsServer.Enabled) {
		{
			key := "metrics-server.addons.k8s.io"

			{
				location := key + "/k8s-1.11.yaml"
				id := "k8s-1.11"

				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Selector: map[string]string{"k8s-app": "metrics-server"},
					Manifest: new(location),
					Id:       id,
					NeedsPKI: !fi.ValueOf(b.Cluster.Spec.MetricsServer.Insecure),
				})
			}
		}
	}

	if b.Cluster.Spec.CertManager != nil && fi.ValueOf(b.Cluster.Spec.CertManager.Enabled) && (b.Cluster.Spec.CertManager.Managed == nil || fi.ValueOf(b.Cluster.Spec.CertManager.Managed)) {
		{
			key := "certmanager.io"

			{
				location := key + "/k8s-1.16.yaml"
				id := "k8s-1.16"

				addon := addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Manifest: new(location),
					Id:       id,
				})
				addon.BuildPrune = true
			}
		}

		if len(b.Cluster.Spec.CertManager.HostedZoneIDs) > 0 {
			serviceAccountRoles = append(serviceAccountRoles, &certmanager.ServiceAccount{})
		}
	}

	if b.Cluster.Spec.CloudProvider.AWS != nil {
		nth := b.Cluster.Spec.CloudProvider.AWS.NodeTerminationHandler

		if nth != nil && fi.ValueOf(nth.Enabled) {

			key := "node-termination-handler.aws"

			{
				location := key + "/k8s-1.11.yaml"
				id := "k8s-1.11"

				addon := addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
					Id:       id,
				})
				addon.BuildPrune = true
			}

			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &nodeterminationhandler.ServiceAccount{})
			}
		}
	}

	npd := b.Cluster.Spec.NodeProblemDetector

	if npd != nil && fi.ValueOf(npd.Enabled) {

		key := "node-problem-detector.addons.k8s.io"

		{
			location := key + "/k8s-1.17.yaml"
			id := "k8s-1.17"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	nvidia := b.Cluster.Spec.Containerd.NvidiaGPU
	igNvidia := false
	for _, ig := range b.KopsModelContext.InstanceGroups {
		if ig.Spec.Containerd != nil && ig.Spec.Containerd.NvidiaGPU != nil && fi.ValueOf(ig.Spec.Containerd.NvidiaGPU.Enabled) {
			igNvidia = true
			break
		}
	}

	if nvidia != nil && fi.ValueOf(nvidia.Enabled) || igNvidia {

		key := "nvidia.addons.k8s.io"

		{
			location := key + "/k8s-1.16.yaml"
			id := "k8s-1.16"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
	}

	igGVisor := false
	for _, ig := range b.KopsModelContext.InstanceGroups {
		if ig.HasGVisor() {
			igGVisor = true
			break
		}
	}

	if igGVisor {
		key := "gvisor.addons.k8s.io"

		{
			location := key + "/k8s-1.20.yaml"
			id := "k8s-1.20"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.CloudProvider.AWS != nil {
		if b.Cluster.Spec.CloudProvider.AWS.LoadBalancerController != nil && fi.ValueOf(b.Cluster.Spec.CloudProvider.AWS.LoadBalancerController.Enabled) {

			key := "aws-load-balancer-controller.addons.k8s.io"

			// The IRSA variant drops hostNetwork / control-plane tolerations /
			// control-plane nodeAffinity / RollingUpdate maxSurge=0 so the pod
			// can schedule on any node using its projected SA token for AWS
			// credentials. Without IRSA the pod must land on a control-plane
			// node with hostNetwork to reach the node IAM role.
			id := "k8s-1.19"
			if b.UseServiceAccountExternalPermissions() {
				id = "k8s-1.19-irsa"
			}
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
				NeedsPKI: true,
			})
			addon.BuildPrune = true

			// Generate aws-load-balancer-controller ServiceAccount IAM permissions
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &awsloadbalancercontroller.ServiceAccount{})
			}
		}

		if b.Cluster.Spec.CloudProvider.AWS.PodIdentityWebhook != nil && fi.ValueOf(&b.Cluster.Spec.CloudProvider.AWS.PodIdentityWebhook.Enabled) {

			key := "eks-pod-identity-webhook.addons.k8s.io"

			{
				id := "k8s-1.16"
				location := key + "/" + id + ".yaml"

				addons.Add(&channelsapi.AddonSpec{
					Name: new(key),
					Selector: map[string]string{
						"k8s-addon": key,
					},
					Manifest: new(location),
					Id:       id,
					NeedsPKI: true,
				})
			}
		}
	}

	if fi.ValueOf(b.Cluster.Spec.CloudConfig.ManageStorageClasses) {
		if b.Cluster.GetCloudProvider() == kops.CloudProviderAWS {
			key := "storage-aws.addons.k8s.io"

			{
				id := "v1.15.0"
				location := key + "/" + id + ".yaml"

				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
					Id:       id,
				})
			}
		}

		if b.Cluster.GetCloudProvider() == kops.CloudProviderAzure {
			key := "storage-azure.addons.k8s.io"

			{
				id := "k8s-1.31"
				location := key + "/" + id + ".yaml"

				addon := addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
					Id:       id,
				})
				addon.BuildPrune = true
			}
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderDO {
		key := "digitalocean-cloud-controller.addons.k8s.io"

		{
			id := "k8s-1.8"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}

		key = "digitalocean-csi-driver.addons.k8s.io"

		{
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderHetzner {
		{
			key := "hcloud-config.addons.k8s.io"
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
		{
			key := "hcloud-cloud-controller.addons.k8s.io"
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
		{
			key := "hcloud-csi-driver.addons.k8s.io"
			id := "k8s-1.22"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.IsKubernetesGTE("1.31") && b.Cluster.GetCloudProvider() == kops.CloudProviderAzure {
		{
			key := "azure-cloud-config.addons.k8s.io"
			id := "k8s-1.31"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
		{
			key := "azure-cloud-controller.addons.k8s.io"
			id := "k8s-1.31"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
		{
			key := "azuredisk-csi-driver.addons.k8s.io"
			id := "k8s-1.31"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderGCE {
		if fi.ValueOf(b.Cluster.Spec.CloudConfig.ManageStorageClasses) {
			key := "storage-gce.addons.k8s.io"

			{
				id := "v1.7.0"
				location := key + "/" + id + ".yaml"

				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: new(location),
					Id:       id,
				})
			}
		}
		if b.Cluster.Spec.CloudProvider.GCE != nil && b.Cluster.Spec.CloudProvider.GCE.PDCSIDriver != nil && fi.ValueOf(b.Cluster.Spec.CloudProvider.GCE.PDCSIDriver.Enabled) {
			key := "gcp-pd-csi-driver.addons.k8s.io"
			{
				id := "k8s-1.23"
				location := key + "/" + id + ".yaml"
				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Manifest: new(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderScaleway {
		{
			key := "scaleway-cloud-controller.addons.k8s.io"
			id := "k8s-1.24"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
		{
			key := "scaleway-csi-driver.addons.k8s.io"
			id := "k8s-1.24"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
		}
	}

	if featureflag.Spotinst.Enabled() && featureflag.SpotinstController.Enabled() {
		key := "spotinst-kubernetes-cluster-controller.addons.k8s.io"

		{
			id := "v1.14.0"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: new(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderGCE {
		{
			key := "gcp-cloud-controller.addons.k8s.io"
			useBuiltin := !b.hasExternalAddon(key)

			if !useBuiltin {
				klog.Infof("Found cloud-controller-manager in addons; won't use builtin")

				// Until we make the manifest extensible, we still need to inject our arguments.
				// TODO(justinsb): we don't really want to do this, it limits the ability for users to override things.
				// However, this is behind a feature flag at the moment, and this way we can work towards something better.
				gkDaemonset := schema.GroupKind{Group: "apps", Kind: "DaemonSet"}
				for _, addon := range b.ClusterAddons {
					if addon.GroupVersionKind().GroupKind() == gkDaemonset &&
						addon.GetName() == "cloud-controller-manager" &&
						addon.GetNamespace() == "kube-system" {

						klog.Infof("replacing arguments in externally provided cloud-controller-manager")

						args, err := b.addonRenderer.CloudControllerConfigArgv()
						if err != nil {
							return nil, nil, fmt.Errorf("in TemplateFunction CloudControllerConfigArgv: %w", err)
						}

						if err := addon.VisitContainers(func(container map[string]interface{}) error {
							// TODO: Check name?
							container["args"] = args
							return nil
						}); err != nil {
							return nil, nil, fmt.Errorf("error visiting containers: %w", err)
						}
					}
				}
			}

			if useBuiltin {
				id := "k8s-1.23"
				location := key + "/" + id + ".yaml"
				addon := addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Manifest: new(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
				addon.BuildPrune = true
			}
		}
	}

	if b.Cluster.Spec.Networking.Kopeio != nil && !featureflag.UseAddonOperators.Enabled() {
		key := "networking.kope.io"
		useBuiltin := !b.hasExternalAddon(key)

		if !useBuiltin {
			klog.Infof("Found kopeio-networking-agent in addons; won't use builtin")
		}

		if useBuiltin {
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: networkingSelector(),
				Manifest: new(location),
				Id:       id,
			})

			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.Flannel != nil {
		key := "networking.flannel"

		{
			id := "k8s-1.25"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: networkingSelector(),
				Manifest: new(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.Calico != nil {
		key := "networking.projectcalico.org"

		{
			id := "k8s-1.25"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: networkingSelector(),
				Manifest: new(location),
				Id:       id,
			})
			addon.BuildPrune = true
		}
	}

	if b.Cluster.Spec.Networking.KubeRouter != nil {
		key := "networking.kuberouter"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Selector: networkingSelector(),
				Manifest: new(location),
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
				Name:               new(key),
				Selector:           networkingSelector(),
				Manifest:           new(location),
				Id:                 id,
				NeedsRollingUpdate: channelsapi.NeedsRollingUpdateAll,
			})
		}
	}

	if b.Cluster.Spec.Networking.Kindnet != nil {
		key := "networking.kindnet"

		{
			id := "k8s-1.32"
			location := key + "/" + id + ".yaml"

			addons.Add(&channelsapi.AddonSpec{
				Name:               new(key),
				Selector:           networkingSelector(),
				Manifest:           new(location),
				Id:                 id,
				NeedsRollingUpdate: channelsapi.NeedsRollingUpdateAll,
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
					Name:     new(key),
					Selector: authenticationSelector,
					Manifest: new(location),
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
					Name:     new(key),
					Selector: authenticationSelector,
					Manifest: new(location),
					Id:       id,
				})
			}
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderOpenstack {
		{
			key := "storage-openstack.addons.k8s.io"

			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Manifest: new(location),
				Selector: map[string]string{"k8s-addon": key},
				Id:       id,
			})
			addon.BuildPrune = true
		}

		// cloudprovider specific out-of-tree controller
		{
			key := "openstack.addons.k8s.io"

			location := key + "/k8s-1.13.yaml"
			id := "k8s-1.13-ccm"

			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Manifest: new(location),
				Selector: map[string]string{"k8s-addon": key},
				Id:       id,
			})
		}
	}

	if b.Cluster.GetCloudProvider() == kops.CloudProviderAWS {

		{
			key := "aws-cloud-controller.addons.k8s.io"

			{
				id := "k8s-1.18"
				location := key + "/" + id + ".yaml"
				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Manifest: new(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &awscloudcontrollermanager.ServiceAccount{})
			}
		}
		if b.Cluster.Spec.CloudProvider.AWS != nil &&
			(b.Cluster.Spec.CloudProvider.AWS.EBSCSIDriver.Managed == nil || fi.ValueOf(b.Cluster.Spec.CloudProvider.AWS.EBSCSIDriver.Managed)) {
			key := "aws-ebs-csi-driver.addons.k8s.io"

			{
				id := "k8s-1.17"
				location := key + "/" + id + ".yaml"
				addons.Add(&channelsapi.AddonSpec{
					Name:     new(key),
					Manifest: new(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &awsebscsidriver.ServiceAccount{})
			}
		}
	}

	if b.Cluster.Spec.SnapshotController != nil && fi.ValueOf(b.Cluster.Spec.SnapshotController.Enabled) {
		key := "snapshot-controller.addons.k8s.io"

		{
			id := "k8s-1.20"
			location := key + "/" + id + ".yaml"
			addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Manifest: new(location),
				Selector: map[string]string{"k8s-addon": key},
				NeedsPKI: true,
				Id:       id,
			})
		}
	}
	if b.Cluster.Spec.Karpenter != nil && b.Cluster.Spec.Karpenter.Enabled {
		key := "karpenter.sh"

		{
			id := "k8s-1.19"
			location := key + "/" + id + ".yaml"
			addon := addons.Add(&channelsapi.AddonSpec{
				Name:     new(key),
				Manifest: new(location),
				Selector: map[string]string{"k8s-addon": key},
				Id:       id,
			})
			if b.UseServiceAccountExternalPermissions() {
				serviceAccountRoles = append(serviceAccountRoles, &karpenter.ServiceAccount{})
			}
			addon.BuildPrune = true
		}
	}

	serviceAccounts := make(map[types.NamespacedName]iam.Subject)

	if b.Cluster.GetCloudProvider() == kops.CloudProviderAWS && b.Cluster.Spec.KubeAPIServer.ServiceAccountIssuer != nil {
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
			serviceAccounts[sa] = serviceAccountRole
		}
	}
	return addons, serviceAccounts, nil
}

// hasExternalAddon checks if the user has overridden a built-in manifest via additional objects.
// We identify this by looking for objects with the matching label.
func (b *BootstrapChannelBuilder) hasExternalAddon(key string) bool {
	for _, o := range b.ClusterAddons {
		labels := o.ToUnstructured().GetLabels()
		if labels[addonmanifests.KopsAddonLabelKey] == key {
			return true
		}
	}
	return false
}

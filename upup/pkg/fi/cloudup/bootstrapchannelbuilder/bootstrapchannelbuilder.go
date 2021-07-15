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

	"k8s.io/klog/v2"
	channelsapi "k8s.io/kops/channels/pkg/api"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/awsmodel"
	"k8s.io/kops/pkg/model/components/addonmanifests"
	"k8s.io/kops/pkg/model/components/addonmanifests/awsebscsidriver"
	"k8s.io/kops/pkg/model/components/addonmanifests/awsloadbalancercontroller"
	"k8s.io/kops/pkg/model/components/addonmanifests/clusterautoscaler"
	"k8s.io/kops/pkg/model/components/addonmanifests/dnscontroller"
	"k8s.io/kops/pkg/model/components/addonmanifests/externaldns"
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
	addons, err := b.buildAddons(c)
	if err != nil {
		return err
	}

	if err := addons.Verify(); err != nil {
		return err
	}

	for _, a := range addons.Spec.Addons {
		key := *a.Name
		if a.Id != "" {
			key = key + "-" + a.Id
		}
		name := b.Cluster.ObjectMeta.Name + "-addons-" + key
		manifestPath := "addons/" + *a.Manifest
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
		remapped, err := addonmanifests.RemapAddonManifest(a, b.KopsModelContext, b.assetBuilder, manifestBytes)
		if err != nil {
			klog.Infof("invalid manifest: %s", string(manifestBytes))
			return fmt.Errorf("error remapping manifest %s: %v", manifestPath, err)
		}
		manifestBytes = remapped

		// Trim whitespace
		manifestBytes = []byte(strings.TrimSpace(string(manifestBytes)))

		rawManifest := string(manifestBytes)
		klog.V(4).Infof("Manifest %v", rawManifest)

		manifestHash, err := utils.HashString(rawManifest)
		klog.V(4).Infof("hash %s", manifestHash)
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
	}

	if featureflag.UseAddonOperators.Enabled() {
		ob := &wellknownoperators.Builder{
			Cluster: b.Cluster,
		}

		wellKnownAddons, crds, err := ob.Build()
		if err != nil {
			return fmt.Errorf("error building well-known operators: %v", err)
		}

		for _, a := range wellKnownAddons {
			key := *a.Spec.Name
			if a.Spec.Id != "" {
				key = key + "-" + a.Spec.Id
			}
			name := b.Cluster.ObjectMeta.Name + "-addons-" + key
			manifestPath := "addons/" + *a.Spec.Manifest

			// Go through any transforms that are best expressed as code
			manifestBytes, err := addonmanifests.RemapAddonManifest(&a.Spec, b.KopsModelContext, b.assetBuilder, a.Manifest)
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

			addons.Spec.Addons = append(addons.Spec.Addons, &a.Spec)
		}

		b.ClusterAddons = append(b.ClusterAddons, crds...)
	}

	if b.ClusterAddons != nil {
		key := "cluster-addons.kops.k8s.io"
		location := key + "/default.yaml"

		a := &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		}

		name := b.Cluster.ObjectMeta.Name + "-addons-" + key
		manifestPath := "addons/" + *a.Manifest

		manifestBytes, err := b.ClusterAddons.ToYAML()
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

		addons.Spec.Addons = append(addons.Spec.Addons, a)
	}

	addonsYAML, err := utils.YamlMarshal(addons)
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

func (b *BootstrapChannelBuilder) buildAddons(c *fi.ModelBuilderContext) (*channelsapi.Addons, error) {
	serviceAccountRoles := []iam.Subject{}

	addons := &channelsapi.Addons{}
	addons.Kind = "Addons"
	addons.ObjectMeta.Name = "bootstrap"

	{
		key := "kops-controller.addons.k8s.io"

		{
			location := key + "/k8s-1.16.yaml"
			id := "k8s-1.16"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:               fi.String(key),
				Selector:           map[string]string{"k8s-addon": key},
				Manifest:           fi.String(location),
				NeedsRollingUpdate: "control-plane",
				Id:                 id,
			})
		}
	}

	{
		key := "core.addons.k8s.io"
		version := "1.4.0"
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
	}

	// @check if podsecuritypolicies are enabled and if so, push the default kube-system policy
	if b.Cluster.Spec.KubeAPIServer != nil && b.Cluster.Spec.KubeAPIServer.HasAdmissionController("PodSecurityPolicy") {
		key := "podsecuritypolicy.addons.k8s.io"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	kubeDNS := b.Cluster.Spec.KubeDNS

	// This checks if the Kubernetes version is greater than or equal to 1.20
	// and makes the default DNS server as CoreDNS if the DNS provider is not specified
	// and the Kubernetes version is >=1.19
	if kubeDNS.Provider == "" {
		kubeDNS.Provider = "KubeDNS"
		if b.Cluster.IsKubernetesGTE("1.20") {
			kubeDNS.Provider = "CoreDNS"
		}
	}

	if kubeDNS.Provider == "KubeDNS" {

		{
			key := "kube-dns.addons.k8s.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
	}

	if b.Cluster.Spec.ExternalDNS == nil || b.Cluster.Spec.ExternalDNS.Provider == kops.ExternalDNSProviderDNSController {
		// @check the dns-controller has not been disabled
		externalDNS := b.Cluster.Spec.ExternalDNS
		if externalDNS == nil || !externalDNS.Disable {
			{
				key := "dns-controller.addons.k8s.io"

				{
					location := key + "/k8s-1.12.yaml"
					id := "k8s-1.12"

					addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
						Name:     fi.String(key),
						Selector: map[string]string{"k8s-addon": key},
						Manifest: fi.String(location),
						Id:       id,
					})
				}
			}

			// Generate dns-controller ServiceAccount IAM permissions
			if b.UseServiceAccountIAM() {
				serviceAccountRoles = append(serviceAccountRoles, &dnscontroller.ServiceAccount{})
			}
		}
	} else {
		{
			key := "external-dns.addons.k8s.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}

			if b.UseServiceAccountIAM() {
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

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}

		if b.UseServiceAccountIAM() {
			serviceAccountRoles = append(serviceAccountRoles, &clusterautoscaler.ServiceAccount{})
		}

	}

	if b.Cluster.Spec.MetricsServer != nil && fi.BoolValue(b.Cluster.Spec.MetricsServer.Enabled) {
		{
			key := "metrics-server.addons.k8s.io"

			{
				location := key + "/k8s-1.11.yaml"
				id := "k8s-1.11"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	nth := b.Cluster.Spec.NodeTerminationHandler

	if nth != nil && fi.BoolValue(nth.Enabled) {

		key := "node-termination-handler.aws"

		{
			location := key + "/k8s-1.11.yaml"
			id := "k8s-1.11"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	npd := b.Cluster.Spec.NodeProblemDetector

	if npd != nil && fi.BoolValue(npd.Enabled) {

		key := "node-problem-detector.addons.k8s.io"

		{
			location := key + "/k8s-1.17.yaml"
			id := "k8s-1.17"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.AWSLoadBalancerController != nil && fi.BoolValue(b.Cluster.Spec.AWSLoadBalancerController.Enabled) {

		key := "aws-load-balancer-controller.addons.k8s.io"

		{
			location := key + "/k8s-1.9.yaml"
			id := "k8s-1.9"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
				NeedsPKI: true,
			})
		}

		// Generate aws-load-balancer-controller ServiceAccount IAM permissions
		if b.UseServiceAccountIAM() {
			serviceAccountRoles = append(serviceAccountRoles, &awsloadbalancercontroller.ServiceAccount{})
		}
	}

	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderAWS {
		key := "storage-aws.addons.k8s.io"

		{
			id := "v1.15.0"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderDO {
		key := "digitalocean-cloud-controller.addons.k8s.io"

		{
			id := "k8s-1.8"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderGCE {
		key := "storage-gce.addons.k8s.io"

		{
			id := "v1.7.0"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if featureflag.Spotinst.Enabled() && featureflag.SpotinstController.Enabled() {
		key := "spotinst-kubernetes-cluster-controller.addons.k8s.io"

		{
			id := "v1.14.0"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	// The metadata-proxy daemonset conceals node metadata endpoints in GCE.
	// It will land on nodes labeled cloud.google.com/metadata-proxy-ready=true
	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderGCE {
		key := "metadata-proxy.addons.k8s.io"

		{
			id := "v0.1.12"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: map[string]string{"k8s-addon": key},
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Kopeio != nil {
		key := "networking.kope.io"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Weave != nil {
		key := "networking.weave"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Flannel != nil {
		key := "networking.flannel"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Calico != nil {
		key := "networking.projectcalico.org"

		{
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Canal != nil {
		key := "networking.projectcalico.org.canal"

		{
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.Kuberouter != nil {
		key := "networking.kuberouter"

		{
			location := key + "/k8s-1.12.yaml"
			id := "k8s-1.12"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Selector: networkingSelector(),
				Manifest: fi.String(location),
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.Networking.AmazonVPC != nil {
		key := "networking.amazon-vpc-routed-eni"

		{
			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
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
		return nil, fmt.Errorf("failed to add cilium addon: %w", err)
	}

	authenticationSelector := map[string]string{"role.kubernetes.io/authentication": "1"}

	if b.Cluster.Spec.Authentication != nil {
		if b.Cluster.Spec.Authentication.Kopeio != nil {
			key := "authentication.kope.io"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: authenticationSelector,
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
		if b.Cluster.Spec.Authentication.Aws != nil {
			key := "authentication.aws"

			{
				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: authenticationSelector,
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderOpenstack {
		{
			key := "storage-openstack.addons.k8s.io"

			id := "k8s-1.16"
			location := key + "/" + id + ".yaml"

			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Manifest: fi.String(location),
				Selector: map[string]string{"k8s-addon": key},
				Id:       id,
			})
		}

		if b.Cluster.Spec.ExternalCloudControllerManager != nil {
			// cloudprovider specific out-of-tree controller
			{
				key := "openstack.addons.k8s.io"

				location := key + "/k8s-1.13.yaml"
				id := "k8s-1.13-ccm"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
		} else {
			{
				key := "core.addons.k8s.io"

				location := key + "/k8s-1.12.yaml"
				id := "k8s-1.12-ccm"

				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Selector: map[string]string{"k8s-addon": key},
					Manifest: fi.String(location),
					Id:       id,
				})
			}
		}
	}

	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderAWS {

		if b.IsKubernetesGTE("1.18") && b.Cluster.Spec.ExternalCloudControllerManager != nil {
			key := "aws-cloud-controller.addons.k8s.io"

			{
				id := "k8s-1.18"
				location := key + "/" + id + ".yaml"
				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}
		}
		if b.Cluster.Spec.CloudConfig != nil && b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver != nil && fi.BoolValue(b.Cluster.Spec.CloudConfig.AWSEBSCSIDriver.Enabled) {
			key := "aws-ebs-csi-driver.addons.k8s.io"

			{
				id := "k8s-1.17"
				location := key + "/" + id + ".yaml"
				addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
					Name:     fi.String(key),
					Manifest: fi.String(location),
					Selector: map[string]string{"k8s-addon": key},
					Id:       id,
				})
			}

			// Generate aws-load-balancer-controller ServiceAccount IAM permissions
			if b.UseServiceAccountIAM() {
				serviceAccountRoles = append(serviceAccountRoles, &awsebscsidriver.ServiceAccount{})
			}
		}
	}

	if b.IsKubernetesGTE("1.20") && b.Cluster.Spec.SnapshotController != nil && fi.BoolValue(b.Cluster.Spec.SnapshotController.Enabled) {
		key := "snapshot-controller.addons.k8s.io"

		{
			id := "k8s-1.20"
			location := key + "/" + id + ".yaml"
			addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
				Name:     fi.String(key),
				Manifest: fi.String(location),
				Selector: map[string]string{"k8s-addon": key},
				NeedsPKI: true,
				Id:       id,
			})
		}
	}

	if b.Cluster.Spec.KubeScheduler.UsePolicyConfigMap != nil {
		key := "scheduler.addons.k8s.io"
		version := "1.7.0"
		location := key + "/v" + version + ".yaml"

		addons.Spec.Addons = append(addons.Spec.Addons, &channelsapi.AddonSpec{
			Name:     fi.String(key),
			Selector: map[string]string{"k8s-addon": key},
			Manifest: fi.String(location),
		})
	}

	if kops.CloudProviderID(b.Cluster.Spec.CloudProvider) == kops.CloudProviderAWS && b.Cluster.Spec.KubeAPIServer.ServiceAccountIssuer != nil {
		awsModelContext := &awsmodel.AWSModelContext{
			KopsModelContext: b.KopsModelContext,
		}

		for _, serviceAccountRole := range serviceAccountRoles {
			iamModelBuilder := &awsmodel.IAMModelBuilder{AWSModelContext: awsModelContext, Lifecycle: b.Lifecycle, Cluster: b.Cluster}

			_, err := iamModelBuilder.BuildServiceAccountRoleTasks(serviceAccountRole, c)
			if err != nil {
				return nil, err
			}
		}
	}
	return addons, nil
}

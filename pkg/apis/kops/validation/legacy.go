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

package validation

import (
	"fmt"
	"net"
	"strings"

	"github.com/blang/semver/v4"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/util/subnet"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/vfs"
)

// legacy contains validation functions that don't match the apimachinery style

// ValidateCluster is responsible for checking the validity of the Cluster spec
func ValidateCluster(c *kops.Cluster, strict bool, vfsContext *vfs.VFSContext) field.ErrorList {
	fieldSpec := field.NewPath("spec")
	allErrs := field.ErrorList{}

	// KubernetesVersion
	// This is one case we return the error because a large part of the rest of the validation logic depends on a valid kubernetes version.

	var k8sVersion *semver.Version
	var err error
	if c.Spec.KubernetesVersion == "" {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubernetesVersion"), ""))
		return allErrs
	} else {
		k8sVersion, err = util.ParseKubernetesVersion(c.Spec.KubernetesVersion)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubernetesVersion"), c.Spec.KubernetesVersion, "unable to determine kubernetes version"))
			return allErrs
		}
	}

	if strict && c.Spec.Kubelet == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubelet"), "kubelet not configured"))
	}
	if strict && c.Spec.ControlPlaneKubelet == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("controlPlaneKubelet"), "controlPlaneKubelet not configured"))
	}
	if strict && c.Spec.KubeControllerManager == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeControllerManager"), "kubeControllerManager not configured"))
	}
	if strict && c.Spec.KubeDNS == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeDNS"), "kubeDNS not configured"))
	}
	if strict && c.Spec.KubeScheduler == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeScheduler"), "kubeScheduler not configured"))
	}
	if strict && c.Spec.KubeAPIServer == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeAPIServer"), "kubeAPIServer not configured"))
	}
	if strict && c.Spec.KubeProxy == nil {
		allErrs = append(allErrs, field.Required(fieldSpec.Child("kubeProxy"), "kubeProxy not configured"))
	}

	var nonMasqueradeCIDR *net.IPNet
	var serviceClusterIPRange *net.IPNet

	if c.Spec.Networking.NonMasqueradeCIDR != "" {
		_, nonMasqueradeCIDR, _ = net.ParseCIDR(c.Spec.Networking.NonMasqueradeCIDR)
	}
	if c.Spec.Networking.ServiceClusterIPRange != "" {
		_, serviceClusterIPRange, _ = net.ParseCIDR(c.Spec.Networking.ServiceClusterIPRange)
	}

	// Check ClusterCIDR
	if c.Spec.KubeControllerManager != nil {
		var clusterCIDR *net.IPNet
		clusterCIDRString := c.Spec.KubeControllerManager.ClusterCIDR
		if clusterCIDRString != "" {
			_, clusterCIDR, err = net.ParseCIDR(clusterCIDRString)
			if err != nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeControllerManager", "clusterCIDR"), clusterCIDRString, "cluster had an invalid kubeControllerManager.clusterCIDR"))
			} else if nonMasqueradeCIDR != nil && !subnet.BelongsTo(nonMasqueradeCIDR, clusterCIDR) {
				allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeControllerManager", "clusterCIDR"), fmt.Sprintf("kubeControllerManager.clusterCIDR %q must be a subnet of nonMasqueradeCIDR %q", clusterCIDRString, c.Spec.Networking.NonMasqueradeCIDR)))
			}
		}
	}

	// @check the custom kubedns options are valid
	if c.Spec.KubeDNS != nil {
		if c.Spec.KubeDNS.ServerIP != "" {
			address := c.Spec.KubeDNS.ServerIP
			ip := net.ParseIP(address)
			if ip == nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "serverIP"), address, "Cluster had an invalid kubeDNS.serverIP"))
			} else {
				if serviceClusterIPRange != nil && !serviceClusterIPRange.Contains(ip) {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeDNS", "serverIP"), fmt.Sprintf("ServiceClusterIPRange %q must contain the DNS Server IP %q", c.Spec.Networking.ServiceClusterIPRange, address)))
				}
				if !featureflag.ExperimentalClusterDNS.Enabled() {
					if isExperimentalClusterDNS(c.Spec.Kubelet, c.Spec.KubeDNS) {
						allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubelet", "clusterDNS"), "Kubelet ClusterDNS did not match cluster kubeDNS.serverIP or nodeLocalDNS.localIP"))
					}
					if isExperimentalClusterDNS(c.Spec.ControlPlaneKubelet, c.Spec.KubeDNS) {
						allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("controlPlaneKubelet", "clusterDNS"), "ControlPlaneKubelet ClusterDNS did not match cluster kubeDNS.serverIP or nodeLocalDNS.localIP"))
					}
				}
			}

			// @ check that NodeLocalDNS addon is configured correctly
			if c.Spec.KubeDNS.NodeLocalDNS != nil && fi.ValueOf(c.Spec.KubeDNS.NodeLocalDNS.Enabled) {
				if c.Spec.KubeDNS.Provider != "CoreDNS" && c.Spec.KubeDNS.Provider != "" {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeDNS", "provider"), "KubeDNS provider must be set to CoreDNS if NodeLocalDNS addon is enabled"))
				}

				allErrs = append(allErrs, validateNodeLocalDNS(&c.Spec, fieldSpec.Child("spec"))...)
			}
		}

		// @check the nameservers are valid
		for i, x := range c.Spec.KubeDNS.UpstreamNameservers {
			if ip := net.ParseIP(x); ip == nil {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "upstreamNameservers").Index(i), x, "Invalid nameserver given, should be a valid ip address"))
			}
		}

		// @check the stubdomain if any
		for domain, nameservers := range c.Spec.KubeDNS.StubDomains {
			if len(nameservers) <= 0 {
				allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "stubDomains").Key(domain), domain, "No nameservers specified for the stub domain"))
			}
			for i, x := range nameservers {
				if ip := net.ParseIP(x); ip == nil {
					allErrs = append(allErrs, field.Invalid(fieldSpec.Child("kubeDNS", "stubDomains").Key(domain).Index(i), x, "Invalid nameserver given, should be a valid ip address"))
				}
			}
		}
	}

	// Check CloudProvider
	{

		var k8sCloudProvider string
		switch c.GetCloudProvider() {
		case kops.CloudProviderAWS:
			k8sCloudProvider = "aws"
		case kops.CloudProviderGCE:
			k8sCloudProvider = "gce"
		case kops.CloudProviderDO:
			k8sCloudProvider = "external"
		case kops.CloudProviderHetzner:
			k8sCloudProvider = "external"
		case kops.CloudProviderOpenstack:
			k8sCloudProvider = "openstack"
		case kops.CloudProviderAzure:
			k8sCloudProvider = "azure"
		case kops.CloudProviderScaleway:
			k8sCloudProvider = "external"
		default:
			// We already added an error above
			k8sCloudProvider = "ignore"
		}

		if k8sCloudProvider != "ignore" {
			if c.Spec.Kubelet != nil && (strict || c.Spec.Kubelet.CloudProvider != "") {
				if c.Spec.Kubelet.CloudProvider != "external" && k8sCloudProvider != c.Spec.Kubelet.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubelet", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
			if c.Spec.ControlPlaneKubelet != nil && (strict || c.Spec.ControlPlaneKubelet.CloudProvider != "") {
				if c.Spec.ControlPlaneKubelet.CloudProvider != "external" && k8sCloudProvider != c.Spec.ControlPlaneKubelet.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("controlPlaneKubelet", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
			if c.Spec.KubeAPIServer != nil && (strict || c.Spec.KubeAPIServer.CloudProvider != "") {
				if k8sVersion != nil && k8sVersion.LT(semver.MustParse("1.33.0-alpha.3")) {
					if c.Spec.KubeAPIServer.CloudProvider != "external" && k8sCloudProvider != c.Spec.KubeAPIServer.CloudProvider {
						allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeAPIServer", "cloudProvider"), "Did not match cluster cloudProvider"))
					}
				}
			}
			if c.Spec.KubeControllerManager != nil && (strict || c.Spec.KubeControllerManager.CloudProvider != "") {
				if c.Spec.KubeControllerManager.CloudProvider != "external" && k8sCloudProvider != c.Spec.KubeControllerManager.CloudProvider {
					allErrs = append(allErrs, field.Forbidden(fieldSpec.Child("kubeControllerManager", "cloudProvider"), "Did not match cluster cloudProvider"))
				}
			}
		}
	}

	allErrs = append(allErrs, newValidateCluster(c, strict)...)

	said := c.Spec.ServiceAccountIssuerDiscovery
	allErrs = append(allErrs, validateServiceAccountIssuerDiscovery(c, said, fieldSpec.Child("serviceAccountIssuerDiscovery"), vfsContext)...)

	return allErrs
}

func validateServiceAccountIssuerDiscovery(c *kops.Cluster, said *kops.ServiceAccountIssuerDiscoveryConfig, fieldSpec *field.Path, vfsContext *vfs.VFSContext) field.ErrorList {
	if said == nil {
		return nil
	}
	allErrs := field.ErrorList{}
	saidStore := said.DiscoveryStore
	if saidStore != "" {
		saidStoreField := fieldSpec.Child("serviceAccountIssuerDiscovery", "discoveryStore")
		base, err := vfsContext.BuildVfsPath(saidStore)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(saidStoreField, saidStore, "not a valid VFS path"))
		} else {
			switch base := base.(type) {
			case *vfs.S3Path:
				// S3 bucket should not contain dots because of the wildcard certificate
				if strings.Contains(base.Bucket(), ".") {
					allErrs = append(allErrs, field.Invalid(saidStoreField, saidStore, "Bucket name cannot contain dots"))
				}
			case *vfs.GSPath:
				// No known restrictions currently. Added here to avoid falling into the default catch all below.
			case *vfs.MemFSPath:
				// memfs is ok for tests; not OK otherwise
				if !base.IsClusterReadable() {
					// (If this _is_ a test, we should call MarkClusterReadable)
					allErrs = append(allErrs, field.Invalid(saidStoreField, saidStore, "S3 is the only supported VFS for discoveryStore"))
				}
			default:
				allErrs = append(allErrs, field.Invalid(saidStoreField, saidStore, "S3 is the only supported VFS for discoveryStore"))
			}
		}
	}
	if said.EnableAWSOIDCProvider {
		enableOIDCField := fieldSpec.Child("serviceAccountIssuerDiscovery", "enableAWSOIDCProvider")
		if saidStore == "" {
			allErrs = append(allErrs, field.Forbidden(enableOIDCField, "AWS OIDC Provider requires a discovery store"))
		}
	}

	return allErrs
}

// DeepValidate is responsible for validating the instancegroups within the cluster spec
func DeepValidate(c *kops.Cluster, groups []*kops.InstanceGroup, strict bool, vfsContext *vfs.VFSContext, cloud fi.Cloud) error {
	if errs := ValidateCluster(c, strict, vfsContext); len(errs) != 0 {
		return errs.ToAggregate()
	}

	if len(groups) == 0 {
		return fmt.Errorf("must configure at least one InstanceGroup")
	}

	controlPlaneGroupCount := 0
	nodeGroupCount := 0
	for _, g := range groups {
		if g.IsControlPlane() {
			controlPlaneGroupCount++
		} else {
			nodeGroupCount++
		}
	}

	if controlPlaneGroupCount == 0 {
		return fmt.Errorf("must configure at least one ControlPlane InstanceGroup")
	}

	for _, g := range groups {
		errs := CrossValidateInstanceGroup(g, c, cloud, strict)

		// Additional cloud-specific validation rules
		if c.GetCloudProvider() != kops.CloudProviderAWS && len(g.Spec.Volumes) > 0 {
			errs = append(errs, field.Forbidden(field.NewPath("spec", "volumes"), "instancegroup volumes are only available with aws at present"))
		}

		if len(errs) != 0 {
			return errs.ToAggregate()
		}
	}

	return nil
}

func isExperimentalClusterDNS(k *kops.KubeletConfigSpec, dns *kops.KubeDNSConfig) bool {
	return k != nil && k.ClusterDNS != dns.ServerIP && dns.NodeLocalDNS != nil && k.ClusterDNS != dns.NodeLocalDNS.LocalIP
}

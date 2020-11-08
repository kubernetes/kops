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

/******************************************************************************
Template Functions are what map functions in the models, to internal logic in
kops. This is the point where we connect static YAML configuration to dynamic
runtime values in memory.

When defining a new function:
	- Build the new function here
	- Define the new function in AddTo()
		dest["MyNewFunction"] = MyNewFunction // <-- Function Pointer
******************************************************************************/

package cloudup

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	kopscontrollerconfig "k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/pkg/apis/kops"
	apiModel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	"k8s.io/kops/util/pkg/env"
)

// TemplateFunctions provides a collection of methods used throughout the templates
type TemplateFunctions struct {
	model.KopsModelContext
}

// AddTo defines the available functions we can use in our YAML models.
// If we are trying to get a new function implemented it MUST
// be defined here.
func (tf *TemplateFunctions) AddTo(dest template.FuncMap, secretStore fi.SecretStore) (err error) {
	cluster := tf.Cluster

	dest["EtcdScheme"] = tf.EtcdScheme
	dest["SharedVPC"] = tf.SharedVPC
	dest["ToJSON"] = tf.ToJSON
	dest["UseBootstrapTokens"] = tf.UseBootstrapTokens
	dest["UseEtcdTLS"] = tf.UseEtcdTLS
	// Remember that we may be on a different arch from the target.  Hard-code for now.
	dest["replace"] = func(s, find, replace string) string {
		return strings.Replace(s, find, replace, -1)
	}
	dest["join"] = func(a []string, sep string) string {
		return strings.Join(a, sep)
	}

	sprigTxtFuncMap := sprig.TxtFuncMap()
	dest["indent"] = sprigTxtFuncMap["indent"]
	dest["contains"] = sprigTxtFuncMap["contains"]

	dest["ClusterName"] = tf.ClusterName
	dest["WithDefaultBool"] = func(v *bool, defaultValue bool) bool {
		if v != nil {
			return *v
		}
		return defaultValue
	}

	dest["GetInstanceGroup"] = tf.GetInstanceGroup
	dest["GetNodeInstanceGroups"] = tf.GetNodeInstanceGroups
	dest["CloudTags"] = tf.CloudTagsForInstanceGroup
	dest["KubeDNS"] = func() *kops.KubeDNSConfig {
		return cluster.Spec.KubeDNS
	}

	dest["NodeLocalDNSClusterIP"] = func() string {
		if cluster.Spec.KubeProxy.ProxyMode == "ipvs" {
			return cluster.Spec.KubeDNS.ServerIP
		}
		return "__PILLAR__CLUSTER__DNS__"
	}
	dest["NodeLocalDNSHealthCheck"] = func() string {
		return fmt.Sprintf("%d", wellknownports.NodeLocalDNSHealthCheck)
	}

	dest["KopsControllerArgv"] = tf.KopsControllerArgv
	dest["KopsControllerConfig"] = tf.KopsControllerConfig
	dest["DnsControllerArgv"] = tf.DNSControllerArgv
	dest["ExternalDnsArgv"] = tf.ExternalDNSArgv
	dest["CloudControllerConfigArgv"] = tf.CloudControllerConfigArgv
	// TODO: Only for GCE?
	dest["EncodeGCELabel"] = gce.EncodeGCELabel
	dest["Region"] = func() string {
		return tf.Region
	}

	// will return openstack external ccm image location for current kubernetes version
	dest["OpenStackCCMTag"] = tf.OpenStackCCMTag
	dest["ProxyEnv"] = tf.ProxyEnv

	dest["KopsSystemEnv"] = tf.KopsSystemEnv
	dest["UseKopsControllerForNodeBootstrap"] = func() bool {
		return tf.UseKopsControllerForNodeBootstrap()
	}

	dest["DO_TOKEN"] = func() string {
		return os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}

	if featureflag.Spotinst.Enabled() {
		if creds, err := spotinst.LoadCredentials(); err == nil {
			dest["SpotinstToken"] = func() string { return creds.Token }
			dest["SpotinstAccount"] = func() string { return creds.Account }
			dest["SpotinstTokenBase64"] = func() string { return base64.StdEncoding.EncodeToString([]byte(creds.Token)) }
			dest["SpotinstAccountBase64"] = func() string { return base64.StdEncoding.EncodeToString([]byte(creds.Account)) }
		}
	}

	if cluster.Spec.Networking != nil && cluster.Spec.Networking.Flannel != nil {
		flannelBackendType := cluster.Spec.Networking.Flannel.Backend
		if flannelBackendType == "" {
			klog.Warningf("Defaulting flannel backend to udp (not a recommended configuration)")
			flannelBackendType = "udp"
		}
		dest["FlannelBackendType"] = func() string { return flannelBackendType }
	}

	if cluster.Spec.Networking != nil && cluster.Spec.Networking.Weave != nil {
		weavesecretString := ""
		weavesecret, _ := secretStore.Secret("weavepassword")
		if weavesecret != nil {
			weavesecretString, err = weavesecret.AsString()
			if err != nil {
				return err
			}
			klog.V(4).Info("Weave secret function successfully registered")
		}

		dest["WeaveSecret"] = func() string { return weavesecretString }
	}

	if cluster.Spec.Networking != nil && cluster.Spec.Networking.Cilium != nil {
		ciliumsecretString := ""
		ciliumsecret, _ := secretStore.Secret("ciliumpassword")
		if ciliumsecret != nil {
			ciliumsecretString, err = ciliumsecret.AsString()
			if err != nil {
				return err
			}
			klog.V(4).Info("Cilium secret function successfully registered")
		}

		dest["CiliumSecret"] = func() string { return ciliumsecretString }
	}

	return nil
}

// ToJSON returns a json representation of the struct or on error an empty string
func (tf *TemplateFunctions) ToJSON(data interface{}) string {
	encoded, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	return string(encoded)
}

// EtcdScheme parses and grabs the protocol to the etcd cluster
func (tf *TemplateFunctions) EtcdScheme() string {
	if tf.UseEtcdTLS() {
		return "https"
	}

	return "http"
}

// SharedVPC is a simple helper function which makes the templates for a shared VPC clearer
func (tf *TemplateFunctions) SharedVPC() bool {
	return tf.Cluster.SharedVPC()
}

// GetInstanceGroup returns the instance group with the specified name
func (tf *TemplateFunctions) GetInstanceGroup(name string) (*kops.InstanceGroup, error) {
	ig := tf.KopsModelContext.FindInstanceGroup(name)
	if ig == nil {
		return nil, fmt.Errorf("InstanceGroup %q not found", name)
	}
	return ig, nil
}

// CloudControllerConfigArgv returns the args to external cloud controller
func (tf *TemplateFunctions) CloudControllerConfigArgv() ([]string, error) {
	cluster := tf.Cluster

	if cluster.Spec.ExternalCloudControllerManager == nil {
		return nil, fmt.Errorf("ExternalCloudControllerManager is nil")
	}
	var argv []string

	if cluster.Spec.ExternalCloudControllerManager.Master != "" {
		argv = append(argv, fmt.Sprintf("--master=%s", cluster.Spec.ExternalCloudControllerManager.Master))
	}
	if cluster.Spec.ExternalCloudControllerManager.LogLevel != 0 {
		argv = append(argv, fmt.Sprintf("--v=%d", cluster.Spec.ExternalCloudControllerManager.LogLevel))
	} else {
		argv = append(argv, "--v=2")
	}
	if cluster.Spec.ExternalCloudControllerManager.CloudProvider != "" {
		argv = append(argv, fmt.Sprintf("--cloud-provider=%s", cluster.Spec.ExternalCloudControllerManager.CloudProvider))
	} else if cluster.Spec.CloudProvider != "" {
		argv = append(argv, fmt.Sprintf("--cloud-provider=%s", cluster.Spec.CloudProvider))
	} else {
		return nil, fmt.Errorf("Cloud Provider is not set")
	}
	if cluster.Spec.ExternalCloudControllerManager.ClusterName != "" {
		argv = append(argv, fmt.Sprintf("--cluster-name=%s", cluster.Spec.ExternalCloudControllerManager.ClusterName))
	}
	if cluster.Spec.ExternalCloudControllerManager.ClusterCIDR != "" {
		argv = append(argv, fmt.Sprintf("--cluster-cidr=%s", cluster.Spec.ExternalCloudControllerManager.ClusterCIDR))
	}
	if cluster.Spec.ExternalCloudControllerManager.AllocateNodeCIDRs != nil {
		argv = append(argv, fmt.Sprintf("--allocate-node-cidrs=%t", *cluster.Spec.ExternalCloudControllerManager.AllocateNodeCIDRs))
	}
	if cluster.Spec.ExternalCloudControllerManager.ConfigureCloudRoutes != nil {
		argv = append(argv, fmt.Sprintf("--configure-cloud-routes=%t", *cluster.Spec.ExternalCloudControllerManager.ConfigureCloudRoutes))
	}
	if cluster.Spec.ExternalCloudControllerManager.CIDRAllocatorType != nil && *cluster.Spec.ExternalCloudControllerManager.CIDRAllocatorType != "" {
		argv = append(argv, fmt.Sprintf("--cidr-allocator-type=%s", *cluster.Spec.ExternalCloudControllerManager.CIDRAllocatorType))
	}
	if cluster.Spec.ExternalCloudControllerManager.UseServiceAccountCredentials != nil {
		argv = append(argv, fmt.Sprintf("--use-service-account-credentials=%t", *cluster.Spec.ExternalCloudControllerManager.UseServiceAccountCredentials))
	} else {
		argv = append(argv, fmt.Sprintf("--use-service-account-credentials=%t", true))
	}

	return argv, nil
}

// DNSControllerArgv returns the args to the DNS controller
func (tf *TemplateFunctions) DNSControllerArgv() ([]string, error) {
	cluster := tf.Cluster

	var argv []string

	argv = append(argv, "/dns-controller")

	// @check if the dns controller has custom configuration
	if cluster.Spec.ExternalDNS == nil {
		argv = append(argv, []string{"--watch-ingress=false"}...)

		klog.V(4).Infof("watch-ingress=false set on dns-controller")
	} else {
		// @check if the watch ingress is set
		var watchIngress bool
		if cluster.Spec.ExternalDNS.WatchIngress != nil {
			watchIngress = fi.BoolValue(cluster.Spec.ExternalDNS.WatchIngress)
		}

		if watchIngress {
			klog.Warningln("--watch-ingress=true set on dns-controller")
			klog.Warningln("this may cause problems with previously defined services: https://github.com/kubernetes/kops/issues/2496")
		}
		argv = append(argv, fmt.Sprintf("--watch-ingress=%t", watchIngress))
		if cluster.Spec.ExternalDNS.WatchNamespace != "" {
			argv = append(argv, fmt.Sprintf("--watch-namespace=%s", cluster.Spec.ExternalDNS.WatchNamespace))
		}
	}

	if dns.IsGossipHostname(cluster.Spec.MasterInternalName) {
		argv = append(argv, "--dns=gossip")

		// Configuration specifically for the DNS controller gossip
		if cluster.Spec.DNSControllerGossipConfig != nil {
			if cluster.Spec.DNSControllerGossipConfig.Protocol != nil {
				argv = append(argv, "--gossip-protocol="+*cluster.Spec.DNSControllerGossipConfig.Protocol)
			}
			if cluster.Spec.DNSControllerGossipConfig.Listen != nil {
				argv = append(argv, "--gossip-listen="+*cluster.Spec.DNSControllerGossipConfig.Listen)
			}
			if cluster.Spec.DNSControllerGossipConfig.Secret != nil {
				argv = append(argv, "--gossip-secret="+*cluster.Spec.DNSControllerGossipConfig.Secret)
			}

			if cluster.Spec.DNSControllerGossipConfig.Seed != nil {
				argv = append(argv, "--gossip-seed="+*cluster.Spec.DNSControllerGossipConfig.Seed)
			} else {
				argv = append(argv, fmt.Sprintf("--gossip-seed=127.0.0.1:%d", wellknownports.ProtokubeGossipWeaveMesh))
			}

			if cluster.Spec.DNSControllerGossipConfig.Secondary != nil {
				if cluster.Spec.DNSControllerGossipConfig.Secondary.Protocol != nil {
					argv = append(argv, "--gossip-protocol-secondary="+*cluster.Spec.DNSControllerGossipConfig.Secondary.Protocol)
				}
				if cluster.Spec.DNSControllerGossipConfig.Secondary.Listen != nil {
					argv = append(argv, "--gossip-listen-secondary="+*cluster.Spec.DNSControllerGossipConfig.Secondary.Listen)
				}
				if cluster.Spec.DNSControllerGossipConfig.Secondary.Secret != nil {
					argv = append(argv, "--gossip-secret-secondary="+*cluster.Spec.DNSControllerGossipConfig.Secondary.Secret)
				}

				if cluster.Spec.DNSControllerGossipConfig.Secondary.Seed != nil {
					argv = append(argv, "--gossip-seed-secondary="+*cluster.Spec.DNSControllerGossipConfig.Secondary.Seed)
				} else {
					argv = append(argv, fmt.Sprintf("--gossip-seed-secondary=127.0.0.1:%d", wellknownports.ProtokubeGossipMemberlist))
				}
			}
		} else {
			// Default to primary mesh and secondary memberlist
			argv = append(argv, fmt.Sprintf("--gossip-seed=127.0.0.1:%d", wellknownports.ProtokubeGossipWeaveMesh))

			argv = append(argv, "--gossip-protocol-secondary=memberlist")
			argv = append(argv, fmt.Sprintf("--gossip-listen-secondary=0.0.0.0:%d", wellknownports.DNSControllerGossipMemberlist))
			argv = append(argv, fmt.Sprintf("--gossip-seed-secondary=127.0.0.1:%d", wellknownports.ProtokubeGossipMemberlist))
		}
	} else {
		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			if strings.HasPrefix(os.Getenv("AWS_REGION"), "cn-") {
				argv = append(argv, "--dns=gossip")
			} else {
				argv = append(argv, "--dns=aws-route53")
			}
		case kops.CloudProviderGCE:
			argv = append(argv, "--dns=google-clouddns")
		case kops.CloudProviderDO:
			argv = append(argv, "--dns=digitalocean")

		default:
			return nil, fmt.Errorf("unhandled cloudprovider %q", cluster.Spec.CloudProvider)
		}
	}

	zone := cluster.Spec.DNSZone
	if zone != "" {
		if strings.Contains(zone, ".") {
			// match by name
			argv = append(argv, "--zone="+zone)
		} else {
			// match by id
			argv = append(argv, "--zone=*/"+zone)
		}
	}
	// permit wildcard updates
	argv = append(argv, "--zone=*/*")
	// Verbose, but not crazy logging
	argv = append(argv, "-v=2")

	return argv, nil
}

// KopsControllerConfig returns the yaml configuration for kops-controller
func (tf *TemplateFunctions) KopsControllerConfig() (string, error) {
	cluster := tf.Cluster

	config := &kopscontrollerconfig.Options{
		Cloud:      cluster.Spec.CloudProvider,
		ConfigBase: cluster.Spec.ConfigBase,
	}

	if featureflag.CacheNodeidentityInfo.Enabled() {
		config.CacheNodeidentityInfo = true
	}

	if tf.UseKopsControllerForNodeBootstrap() {
		certNames := []string{"kubelet", "kubelet-server"}
		signingCAs := []string{fi.CertificateIDCA}
		if apiModel.UseCiliumEtcd(cluster) {
			certNames = append(certNames, "etcd-client-cilium")
			signingCAs = append(signingCAs, "etcd-clients-ca-cilium")
		}
		if cluster.Spec.KubeProxy.Enabled == nil || *cluster.Spec.KubeProxy.Enabled {
			certNames = append(certNames, "kube-proxy")
		}
		if cluster.Spec.Networking.Kuberouter != nil {
			certNames = append(certNames, "kube-router")
		}

		pkiDir := "/etc/kubernetes/kops-controller/pki"
		config.Server = &kopscontrollerconfig.ServerOptions{
			Listen:                fmt.Sprintf(":%d", wellknownports.KopsControllerPort),
			ServerCertificatePath: path.Join(pkiDir, "kops-controller.crt"),
			ServerKeyPath:         path.Join(pkiDir, "kops-controller.key"),
			CABasePath:            pkiDir,
			SigningCAs:            signingCAs,
			CertNames:             certNames,
		}

		switch kops.CloudProviderID(cluster.Spec.CloudProvider) {
		case kops.CloudProviderAWS:
			nodesRoles := sets.String{}
			for _, ig := range tf.InstanceGroups {
				if ig.Spec.Role == kops.InstanceGroupRoleNode {
					profile, err := tf.LinkToIAMInstanceProfile(ig)
					if err != nil {
						return "", fmt.Errorf("getting role for ig %s: %v", ig.Name, err)
					}
					nodesRoles.Insert(*profile.Name)
				}
			}
			config.Server.Provider.AWS = &awsup.AWSVerifierOptions{
				NodesRoles: nodesRoles.List(),
				Region:     tf.Region,
			}
		default:
			return "", fmt.Errorf("unsupported cloud provider %s", cluster.Spec.CloudProvider)
		}
	}

	// To avoid indentation problems, we marshal as json.  json is a subset of yaml
	b, err := json.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to serialize kops-controller config: %v", err)
	}

	return string(b), nil
}

// KopsControllerArgv returns the args to kops-controller
func (tf *TemplateFunctions) KopsControllerArgv() ([]string, error) {
	var argv []string

	argv = append(argv, "/kops-controller")

	// Verbose, but not excessive logging
	argv = append(argv, "--v=2")

	argv = append(argv, "--conf=/etc/kubernetes/kops-controller/config/config.yaml")

	return argv, nil
}

func (tf *TemplateFunctions) ExternalDNSArgv() ([]string, error) {
	cluster := tf.Cluster

	var argv []string

	cloudProvider := cluster.Spec.CloudProvider

	switch kops.CloudProviderID(cloudProvider) {
	case kops.CloudProviderAWS:
		argv = append(argv, "--provider=aws")
	case kops.CloudProviderGCE:
		project := cluster.Spec.Project
		argv = append(argv, "--provider=google")
		argv = append(argv, "--google-project="+project)
	default:
		return nil, fmt.Errorf("unhandled cloudprovider %q", cluster.Spec.CloudProvider)
	}

	argv = append(argv, "--source=ingress")

	return argv, nil
}

func (tf *TemplateFunctions) ProxyEnv() map[string]string {
	cluster := tf.Cluster

	envs := map[string]string{}
	proxies := cluster.Spec.EgressProxy
	if proxies == nil {
		return envs
	}
	httpProxy := proxies.HTTPProxy
	if httpProxy.Host != "" {
		var portSuffix string
		if httpProxy.Port != 0 {
			portSuffix = ":" + strconv.Itoa(httpProxy.Port)
		} else {
			portSuffix = ""
		}
		url := "http://" + httpProxy.Host + portSuffix
		envs["http_proxy"] = url
		envs["https_proxy"] = url
	}
	if proxies.ProxyExcludes != "" {
		envs["no_proxy"] = proxies.ProxyExcludes
		envs["NO_PROXY"] = proxies.ProxyExcludes
	}
	return envs
}

// KopsSystemEnv builds the env vars for a system component
func (tf *TemplateFunctions) KopsSystemEnv() []corev1.EnvVar {
	envMap := env.BuildSystemComponentEnvVars(&tf.Cluster.Spec)

	return envMap.ToEnvVars()
}

// OpenStackCCM returns OpenStack external cloud controller manager current image
// with tag specified to k8s version
func (tf *TemplateFunctions) OpenStackCCMTag() string {
	var tag string
	parsed, err := util.ParseKubernetesVersion(tf.Cluster.Spec.KubernetesVersion)
	if err != nil {
		tag = "latest"
	} else {
		if parsed.Minor == 13 {
			// The bugfix release
			tag = "1.13.1"
		} else {
			// otherwise we use always .0 ccm image, if needed that can be overrided using clusterspec
			tag = fmt.Sprintf("v%d.%d.0", parsed.Major, parsed.Minor)
		}
	}
	return tag
}

// GetNodeInstanceGroups returns a map containing the defined instance groups of role "Node".
func (tf *TemplateFunctions) GetNodeInstanceGroups() map[string]kops.InstanceGroupSpec {
	nodegroups := make(map[string]kops.InstanceGroupSpec)
	for _, ig := range tf.KopsModelContext.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleNode {
			nodegroups[ig.ObjectMeta.Name] = ig.Spec
		}
	}
	return nodegroups
}

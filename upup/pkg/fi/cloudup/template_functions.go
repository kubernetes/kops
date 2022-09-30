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
	"net"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/aws/aws-sdk-go/service/ec2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/flagbuilder"
	"sigs.k8s.io/yaml"

	kopscontrollerconfig "k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/pkg/apis/kops"
	apiModel "k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/dns"
	"k8s.io/kops/pkg/featureflag"
	"k8s.io/kops/pkg/kubemanifest"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/components/kopscontroller"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/pkg/resources/spotinst"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce"
	gcetpm "k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm"
	"k8s.io/kops/util/pkg/env"
)

// TemplateFunctions provides a collection of methods used throughout the templates
type TemplateFunctions struct {
	model.KopsModelContext

	cloud fi.Cloud
}

// AddTo defines the available functions we can use in our YAML models.
// If we are trying to get a new function implemented it MUST
// be defined here.
func (tf *TemplateFunctions) AddTo(dest template.FuncMap, secretStore fi.SecretStore) (err error) {
	cluster := tf.Cluster

	dest["ToJSON"] = tf.ToJSON
	dest["ToYAML"] = tf.ToYAML
	dest["KubeObjectToApplyYAML"] = kubemanifest.KubeObjectToApplyYAML

	dest["SharedVPC"] = tf.SharedVPC
	dest["UseBootstrapTokens"] = tf.UseBootstrapTokens
	// Remember that we may be on a different arch from the target.  Hard-code for now.
	dest["replace"] = func(s, find, replace string) string {
		return strings.Replace(s, find, replace, -1)
	}
	dest["join"] = func(a []string, sep string) string {
		return strings.Join(a, sep)
	}
	dest["joinHostPort"] = net.JoinHostPort

	sprigTxtFuncMap := sprig.TxtFuncMap()
	dest["nindent"] = sprigTxtFuncMap["nindent"]
	dest["indent"] = sprigTxtFuncMap["indent"]
	dest["contains"] = sprigTxtFuncMap["contains"]
	dest["trimPrefix"] = sprigTxtFuncMap["trimPrefix"]
	dest["semverCompare"] = sprigTxtFuncMap["semverCompare"]
	dest["ternary"] = sprigTxtFuncMap["ternary"]

	dest["ClusterName"] = tf.ClusterName
	dest["WithDefaultBool"] = func(v *bool, defaultValue bool) bool {
		if v != nil {
			return *v
		}
		return defaultValue
	}

	dest["GetCloudProvider"] = cluster.Spec.GetCloudProvider
	dest["GetInstanceGroup"] = tf.GetInstanceGroup
	dest["GetNodeInstanceGroups"] = tf.GetNodeInstanceGroups
	dest["HasHighlyAvailableControlPlane"] = tf.HasHighlyAvailableControlPlane
	dest["ControlPlaneControllerReplicas"] = tf.ControlPlaneControllerReplicas
	dest["APIServerNodeRole"] = tf.APIServerNodeRole

	dest["CloudTags"] = tf.CloudTagsForInstanceGroup
	dest["KubeDNS"] = func() *kops.KubeDNSConfig {
		return cluster.Spec.KubeDNS
	}

	dest["GossipDomains"] = func() []string {
		var names []string

		if dns.IsGossipHostname(cluster.Spec.MasterInternalName) {
			names = append(names, "k8s.local")
		}

		return names
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
	kopscontroller.AddTemplateFunctions(cluster, dest)
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
	dest["OpenStackCSITag"] = tf.OpenStackCSITag
	dest["ProxyEnv"] = tf.ProxyEnv

	dest["KopsSystemEnv"] = tf.KopsSystemEnv
	dest["UseKopsControllerForNodeBootstrap"] = func() bool {
		return tf.UseKopsControllerForNodeBootstrap()
	}

	dest["DO_TOKEN"] = func() string {
		return os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
	}

	dest["HCLOUD_TOKEN"] = func() string {
		return os.Getenv("HCLOUD_TOKEN")
	}
	dest["HCLOUD_NETWORK"] = func() string {
		if cluster.Spec.NetworkID != "" {
			return cluster.Spec.NetworkID
		}
		return cluster.Name
	}

	dest["YANDEX_CLOUD_CREDENTIAL_FILE"] = func() string {
		return os.Getenv("YANDEX_CLOUD_CREDENTIAL_FILE")
	}

	if featureflag.Spotinst.Enabled() {
		if creds, err := spotinst.LoadCredentials(); err == nil {
			dest["SpotinstToken"] = func() string { return creds.Token }
			dest["SpotinstAccount"] = func() string { return creds.Account }
			dest["SpotinstTokenBase64"] = func() string { return base64.StdEncoding.EncodeToString([]byte(creds.Token)) }
			dest["SpotinstAccountBase64"] = func() string { return base64.StdEncoding.EncodeToString([]byte(creds.Account)) }
		}
	}

	if cluster.Spec.Networking != nil && cluster.Spec.Networking.AmazonVPC != nil {
		c := cluster.Spec.Networking.AmazonVPC
		dest["AmazonVpcEnvVars"] = func() map[string]string {
			envVars := map[string]string{
				// Use defaults from the official AWS VPC CNI Helm chart:
				// https://github.com/aws/amazon-vpc-cni-k8s/blob/master/charts/aws-vpc-cni/values.yaml
				"AWS_VPC_CNI_NODE_PORT_SUPPORT":         "true",
				"AWS_VPC_ENI_MTU":                       "9001",
				"AWS_VPC_K8S_CNI_CONFIGURE_RPFILTER":    "false",
				"AWS_VPC_K8S_CNI_CUSTOM_NETWORK_CFG":    "false",
				"AWS_VPC_K8S_CNI_EXTERNALSNAT":          "false",
				"AWS_VPC_K8S_CNI_LOG_FILE":              "/host/var/log/aws-routed-eni/ipamd.log",
				"AWS_VPC_K8S_CNI_LOGLEVEL":              "DEBUG",
				"AWS_VPC_K8S_CNI_RANDOMIZESNAT":         "prng",
				"AWS_VPC_K8S_CNI_VETHPREFIX":            "eni",
				"AWS_VPC_K8S_PLUGIN_LOG_FILE":           "/var/log/aws-routed-eni/plugin.log",
				"AWS_VPC_K8S_PLUGIN_LOG_LEVEL":          "DEBUG",
				"DISABLE_INTROSPECTION":                 "false",
				"DISABLE_METRICS":                       "false",
				"ENABLE_POD_ENI":                        "false",
				"WARM_ENI_TARGET":                       "1",
				"DISABLE_NETWORK_RESOURCE_PROVISIONING": "false",
			}
			for _, e := range c.Env {
				envVars[e.Name] = e.Value
			}
			envVars["ENABLE_IPv4"] = strconv.FormatBool(!cluster.Spec.IsIPv6Only())
			envVars["ENABLE_IPv6"] = strconv.FormatBool(cluster.Spec.IsIPv6Only())
			if cluster.Spec.IsIPv6Only() {
				envVars["ENABLE_PREFIX_DELEGATION"] = "true"
				envVars["WARM_PREFIX_TARGET"] = "1"
			}
			envVars["ADDITIONAL_ENI_TAGS"] = fmt.Sprintf(
				"{\\\"KubernetesCluster\\\":\\\"%s\\\",\\\"kubernetes.io/cluster/%s\\\":\\\"owned\\\"}",
				tf.ClusterName(),
				tf.ClusterName(),
			)

			return envVars
		}
	}

	if cluster.Spec.Networking != nil && cluster.Spec.Networking.Calico != nil {
		c := cluster.Spec.Networking.Calico
		dest["CalicoIPv4PoolIPIPMode"] = func() string {
			if c.EncapsulationMode != "ipip" {
				return "Never"
			}
			if c.IPIPMode != "" {
				return c.IPIPMode
			}
			if cluster.Spec.GetCloudProvider() == kops.CloudProviderOpenstack {
				return "Always"
			}
			return "CrossSubnet"
		}
		dest["CalicoIPv4PoolVXLANMode"] = func() string {
			if c.EncapsulationMode != "vxlan" {
				return "Never"
			}
			if c.VXLANMode != "" {
				return c.VXLANMode
			}
			return "CrossSubnet"
		}
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

	dest["CloudLabels"] = func() string {
		labels := []string{
			fmt.Sprintf("KubernetesCluster=%s", cluster.ObjectMeta.Name),
		}
		for n, v := range cluster.Spec.CloudLabels {
			labels = append(labels, fmt.Sprintf("%s=%s", n, v))
		}
		// ensure stable sorting of tags
		sort.Strings(labels)
		return strings.Join(labels, ",")
	}

	dest["IsIPv6Only"] = tf.IsIPv6Only
	dest["UseServiceAccountExternalPermissions"] = tf.UseServiceAccountExternalPermissions

	if cluster.Spec.NodeTerminationHandler != nil {
		dest["DefaultQueueName"] = func() string {
			s := strings.Replace(tf.ClusterName(), ".", "-", -1)
			domain := ".amazonaws.com/"
			if strings.Contains(tf.Region, "cn-") {
				domain = ".amazonaws.com.cn/"
			}
			url := "https://sqs." + tf.Region + domain + tf.AWSAccountID + "/" + s + "-nth"
			return url
		}

		dest["EnableSQSTerminationDraining"] = func() bool { return *cluster.Spec.NodeTerminationHandler.EnableSQSTerminationDraining }
	}

	dest["ArchitectureOfAMI"] = tf.architectureOfAMI

	dest["ParseTaint"] = util.ParseTaint

	dest["UsesInstanceIDForNodeName"] = func() bool {
		return nodeup.UsesInstanceIDForNodeName(tf.Cluster)
	}

	dest["KarpenterInstanceTypes"] = func(ig kops.InstanceGroupSpec) ([]string, error) {
		return karpenterInstanceTypes(tf.cloud.(awsup.AWSCloud), ig)
	}

	dest["PodIdentityWebhookConfigMapData"] = tf.podIdentityWebhookConfigMapData

	dest["HasSnapshotController"] = func() bool {
		sc := cluster.Spec.SnapshotController
		return sc != nil && fi.BoolValue(sc.Enabled)
	}

	dest["IsKubernetesGTE"] = tf.IsKubernetesGTE
	dest["IsKubernetesLT"] = tf.IsKubernetesLT

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

// ToYAML returns a yaml representation of the struct or on error an empty string
func (tf *TemplateFunctions) ToYAML(data interface{}) string {
	encoded, err := yaml.Marshal(data)
	if err != nil {
		return ""
	}

	return string(encoded)
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

// ControlPlaneControllerReplicas returns the amount of replicas for a controllers that should run in the cluster
// If the cluster has a highly available control plane, this function will return 2, if it has 1 control plane node, it will return 1
// deployOnWorkersIfExternalPermissons should be true if a controller runs on worker nodes when external IAM permissions is enabled for the cluster.
// In this case it is assumed that it can run 2 replicas.
func (tf *TemplateFunctions) ControlPlaneControllerReplicas(deployOnWorkersIfExternalPermissions bool) int {
	if deployOnWorkersIfExternalPermissions && tf.Cluster.Spec.IAM != nil && fi.BoolValue(tf.Cluster.Spec.IAM.UseServiceAccountExternalPermissions) {
		return 2
	}
	if tf.HasHighlyAvailableControlPlane() {
		return 2
	}
	return 1
}

func (tf *TemplateFunctions) APIServerNodeRole() string {
	if featureflag.APIServerNodes.Enabled() {
		return "node-role.kubernetes.io/api-server"
	} else if tf.Cluster.IsKubernetesGTE("1.24") {
		return "node-role.kubernetes.io/control-plane"
	}
	return "node-role.kubernetes.io/master"
}

// HasHighlyAvailableControlPlane returns true of the cluster has more than one control plane node. False otherwise.
func (tf *TemplateFunctions) HasHighlyAvailableControlPlane() bool {
	cp := 0
	for _, ig := range tf.InstanceGroups {
		if ig.Spec.Role == kops.InstanceGroupRoleMaster {
			cp++
			if cp > 1 {
				return true
			}
		}
	}
	return false
}

// CloudControllerConfigArgv returns the args to external cloud controller
func (tf *TemplateFunctions) CloudControllerConfigArgv() ([]string, error) {
	cluster := tf.Cluster

	if cluster.Spec.ExternalCloudControllerManager == nil {
		return nil, fmt.Errorf("ExternalCloudControllerManager is nil")
	}

	argv, err := flagbuilder.BuildFlagsList(cluster.Spec.ExternalCloudControllerManager)
	if err != nil {
		return nil, err
	}

	// default verbosity to 2
	if cluster.Spec.ExternalCloudControllerManager.LogLevel == 0 {
		argv = append(argv, "--v=2")
	}

	// take the cloud provider value from clusterSpec if unset
	if cluster.Spec.ExternalCloudControllerManager.CloudProvider == "" {
		if cluster.Spec.GetCloudProvider() != "" {
			argv = append(argv, fmt.Sprintf("--cloud-provider=%s", cluster.Spec.GetCloudProvider()))
		} else {
			return nil, fmt.Errorf("Cloud Provider is not set")
		}
	}

	// default use-service-account-credentials to true
	if cluster.Spec.ExternalCloudControllerManager.UseServiceAccountCredentials == nil {
		argv = append(argv, fmt.Sprintf("--use-service-account-credentials=%t", true))
	}

	if cluster.Spec.GetCloudProvider() != kops.CloudProviderHetzner {
		argv = append(argv, "--cloud-config=/etc/kubernetes/cloud.config")
	}

	return argv, nil
}

// DNSControllerArgv returns the args to the DNS controller
func (tf *TemplateFunctions) DNSControllerArgv() ([]string, error) {
	cluster := tf.Cluster

	var argv []string

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
		switch cluster.Spec.GetCloudProvider() {
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
			return nil, fmt.Errorf("unhandled cloudprovider %q", cluster.Spec.GetCloudProvider())
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

	if cluster.Spec.IsIPv6Only() {
		argv = append(argv, "--internal-ipv6")
	} else {
		argv = append(argv, "--internal-ipv4")
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
		Cloud:      string(cluster.Spec.GetCloudProvider()),
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

		switch cluster.Spec.GetCloudProvider() {
		case kops.CloudProviderAWS:
			nodesRoles := sets.String{}
			for _, ig := range tf.InstanceGroups {
				if ig.Spec.Role == kops.InstanceGroupRoleNode || ig.Spec.Role == kops.InstanceGroupRoleAPIServer {
					profile, err := tf.LinkToIAMInstanceProfile(ig)
					if err != nil {
						return "", fmt.Errorf("getting profile for ig %s: %v", ig.Name, err)
					}
					// The IAM Instance Profile has not been created at this point if it is not specified.
					// Because the IAM Instance Profile and the IAM Role are created in IAMModelBuilder tasks.
					// Therefore, the IAM Role associated with IAM Instance Profile is acquired only when it is not specified.
					if ig.Spec.IAM != nil && ig.Spec.IAM.Profile != nil {
						c := tf.cloud.(awsup.AWSCloud)
						roles, err := awsup.GetRolesInInstanceProfile(c, *profile.Name)
						if err != nil {
							return "", fmt.Errorf("getting role from profile %s: %v", *profile.Name, err)
						}
						nodesRoles.Insert(roles...)
					} else {
						// When the IAM Instance Profile is not specified, IAM Instance Profile is created by kOps.
						// In this case, the IAM Instance Profile name and IAM Role name are same.
						// So there is no problem even if IAM Instance Profile name is inserted as role name in nodesRoles.
						nodesRoles.Insert(*profile.Name)
					}
				}
			}
			config.Server.Provider.AWS = &awsup.AWSVerifierOptions{
				NodesRoles: nodesRoles.List(),
				Region:     tf.Region,
			}

			if cluster.Spec.ExternalCloudControllerManager != nil && cluster.IsKubernetesGTE("1.22") {
				config.Server.UseInstanceIDForNodeName = true
			}

		case kops.CloudProviderGCE:
			c := tf.cloud.(gce.GCECloud)

			config.Server.Provider.GCE = &gcetpm.TPMVerifierOptions{
				ProjectID:   c.Project(),
				ClusterName: tf.ClusterName(),
				Region:      tf.Region,
				MaxTimeSkew: 300,
			}
		default:
			return "", fmt.Errorf("unsupported cloud provider %s", cluster.Spec.GetCloudProvider())
		}
	}

	if cluster.Spec.IsKopsControllerIPAM() {
		config.EnableCloudIPAM = true
	}

	if dns.IsGossipHostname(cluster.Spec.MasterInternalName) {
		config.Discovery = &kopscontrollerconfig.DiscoveryOptions{
			Enabled: true,
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

	// Verbose, but not excessive logging
	argv = append(argv, "--v=2")

	argv = append(argv, "--conf=/etc/kubernetes/kops-controller/config/config.yaml")

	return argv, nil
}

func (tf *TemplateFunctions) ExternalDNSArgv() ([]string, error) {
	cluster := tf.Cluster
	externalDNS := tf.Cluster.Spec.ExternalDNS

	var argv []string

	cloudProvider := cluster.Spec.GetCloudProvider()

	switch cloudProvider {
	case kops.CloudProviderAWS:
		argv = append(argv, "--provider=aws")
	case kops.CloudProviderGCE:
		project := cluster.Spec.Project
		argv = append(argv, "--provider=google")
		argv = append(argv, "--google-project="+project)
	default:
		return nil, fmt.Errorf("unhandled cloudprovider %q", cluster.Spec.GetCloudProvider())
	}

	argv = append(argv, "--events")
	if externalDNS.WatchIngress == nil || *externalDNS.WatchIngress {
		argv = append(argv, "--source=ingress")
	}
	argv = append(argv, "--source=pod")
	argv = append(argv, "--source=service")
	argv = append(argv, "--compatibility=kops-dns-controller")
	argv = append(argv, "--registry=txt")
	argv = append(argv, "--txt-owner-id=kops-"+tf.ClusterName())
	argv = append(argv, "--zone-id-filter="+tf.Cluster.Spec.DNSZone)
	if externalDNS.WatchNamespace != "" {
		argv = append(argv, "--namespace="+externalDNS.WatchNamespace)
	}

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
		} else if parsed.Minor == 23 {
			// The bugfix release, see https://github.com/kubernetes/cloud-provider-openstack/releases
			tag = "v1.23.1"
		} else if parsed.Minor == 24 {
			tag = "v1.24.1"
		} else {
			// otherwise we use always .0 ccm image, if needed that can be overrided using clusterspec
			tag = fmt.Sprintf("v%d.%d.0", parsed.Major, parsed.Minor)
		}
	}
	return tag
}

// OpenStackCSI returns OpenStack csi current image
// with tag specified to k8s version
func (tf *TemplateFunctions) OpenStackCSITag() string {
	var tag string
	parsed, err := util.ParseKubernetesVersion(tf.Cluster.Spec.KubernetesVersion)
	if err != nil {
		tag = "latest"
	} else {
		if parsed.Minor == 24 {
			tag = "v1.24.1"
		} else {
			// otherwise we use always .0 csi image, if needed that can be overrided using cloud config spec
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

func (tf *TemplateFunctions) architectureOfAMI(amiID string) string {
	image, _ := tf.cloud.(awsup.AWSCloud).ResolveImage(amiID)
	switch *image.Architecture {
	case "x86_64":
		return "amd64"
	}
	return "arm64"
}

type podIdentityWebhookMapping struct {
	RoleARN         string
	Audience        string
	UseRegionalSTS  bool
	TokenExpiration int64
}

func (tf *TemplateFunctions) podIdentityWebhookConfigMapData() (string, error) {
	sas := tf.Cluster.Spec.IAM.ServiceAccountExternalPermissions
	mappings := make(map[string]podIdentityWebhookMapping)
	for _, sa := range sas {
		if sa.AWS == nil {
			continue
		}
		key := sa.Namespace + "/" + sa.Name
		mappings[key] = podIdentityWebhookMapping{
			RoleARN:        fmt.Sprintf("arn:%s:iam::%s:role/%s", tf.AWSPartition, tf.AWSAccountID, iam.IAMNameForServiceAccountRole(sa.Name, sa.Namespace, tf.ClusterName())),
			Audience:       "amazonaws.com",
			UseRegionalSTS: true,
		}
	}
	jsonBytes, err := json.Marshal(mappings)
	return fmt.Sprintf("%q", jsonBytes), err
}

func karpenterInstanceTypes(cloud awsup.AWSCloud, ig kops.InstanceGroupSpec) ([]string, error) {
	var mixedInstancesPolicy *kops.MixedInstancesPolicySpec

	if ig.MachineType == "" && ig.MixedInstancesPolicy == nil {
		// Karpenter thinks all clusters run VPC CNI and schedules thinking Node Capacity is constrainted by number of ENIs.

		// cpuMin is the reasonable lower limit for a Kubernetes Node
		// Generally, it also avoids instances Karpenter thinks it can only schedule 4 Pods on.
		cpuMin := resource.MustParse("2")
		memoryMin := resource.MustParse(("2G"))

		mixedInstancesPolicy = &kops.MixedInstancesPolicySpec{
			InstanceRequirements: &kops.InstanceRequirementsSpec{
				CPU: &kops.MinMaxSpec{
					Min: &cpuMin,
				},
				Memory: &kops.MinMaxSpec{
					Min: &memoryMin,
				},
			},
		}
	}
	if ig.MixedInstancesPolicy != nil {
		mixedInstancesPolicy = ig.MixedInstancesPolicy
	}

	if mixedInstancesPolicy != nil {
		if len(mixedInstancesPolicy.Instances) > 0 {
			return mixedInstancesPolicy.Instances, nil
		}
		if mixedInstancesPolicy.InstanceRequirements != nil {
			instanceRequirements := mixedInstancesPolicy.InstanceRequirements
			ami, err := cloud.ResolveImage(ig.Image)
			if err != nil {
				return nil, err
			}
			arch := ami.Architecture
			hv := ami.VirtualizationType

			ir := &ec2.InstanceRequirementsRequest{
				VCpuCount:            &ec2.VCpuCountRangeRequest{},
				MemoryMiB:            &ec2.MemoryMiBRequest{},
				BurstablePerformance: fi.String("included"),
				InstanceGenerations:  []*string{fi.String("current")},
			}
			cpu := instanceRequirements.CPU
			if cpu != nil {
				if cpu.Max != nil {
					cpuMax, _ := instanceRequirements.CPU.Max.AsInt64()
					ir.VCpuCount.Max = &cpuMax
				}
				cpu := instanceRequirements.CPU
				if cpu != nil {
					if cpu.Max != nil {
						cpuMax, _ := instanceRequirements.CPU.Max.AsInt64()
						ir.VCpuCount.Max = &cpuMax
					}
					if cpu.Min != nil {
						cpuMin, _ := instanceRequirements.CPU.Min.AsInt64()
						ir.VCpuCount.Min = &cpuMin
					}
				} else {
					ir.VCpuCount.Min = fi.Int64(0)
				}

				memory := instanceRequirements.Memory
				if memory != nil {
					if memory.Max != nil {
						memoryMax := instanceRequirements.Memory.Max.ScaledValue(resource.Mega)
						ir.MemoryMiB.Max = &memoryMax
					}
					if memory.Min != nil {
						memoryMin := instanceRequirements.Memory.Min.ScaledValue(resource.Mega)
						ir.MemoryMiB.Min = &memoryMin
					}
				} else {
					ir.MemoryMiB.Min = fi.Int64(0)
				}

				ir.AcceleratorCount = &ec2.AcceleratorCountRequest{
					Min: fi.Int64(0),
					Max: fi.Int64(0),
				}

				response, err := cloud.EC2().GetInstanceTypesFromInstanceRequirements(
					&ec2.GetInstanceTypesFromInstanceRequirementsInput{
						ArchitectureTypes:    []*string{arch},
						VirtualizationTypes:  []*string{hv},
						InstanceRequirements: ir,
					},
				)
				if err != nil {
					return nil, err
				}
				types := []string{}
				for _, it := range response.InstanceTypes {
					types = append(types, *it.InstanceType)
				}
				if len(types) == 0 {
					return nil, fmt.Errorf("no instances matched requirements")
				}
				return types, nil
			}
		}
	}

	return []string{ig.MachineType}, nil
}

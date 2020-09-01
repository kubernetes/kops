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
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	toml "github.com/pelletier/go-toml"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops/model"
	"k8s.io/kops/upup/pkg/fi/utils"
	"sigs.k8s.io/yaml"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
	"k8s.io/kops/util/pkg/reflectutils"
)

type NodeUpConfigBuilder interface {
	BuildConfig(ig *kops.InstanceGroup, apiserverAdditionalIPs []string, caTasks map[string]*fitasks.Keypair) (*nodeup.Config, *nodeup.BootConfig, error)
}

// BootstrapScriptBuilder creates the bootstrap script
type BootstrapScriptBuilder struct {
	*KopsModelContext
	Lifecycle           fi.Lifecycle
	NodeUpAssets        map[architectures.Architecture]*mirrors.MirroredAsset
	NodeUpConfigBuilder NodeUpConfigBuilder
	Cluster             *kops.Cluster
}

type BootstrapScript struct {
	Name      string
	Lifecycle fi.Lifecycle
	ig        *kops.InstanceGroup
	builder   *BootstrapScriptBuilder
	resource  fi.TaskDependentResource
	// alternateNameTasks are tasks that contribute api-server IP addresses.
	alternateNameTasks []fi.HasAddress

	// caTasks hold the CA tasks, for dependency analysis.
	caTasks map[string]*fitasks.Keypair

	// nodeupConfig contains the nodeup config.
	nodeupConfig fi.TaskDependentResource
}

var _ fi.Task = &BootstrapScript{}
var _ fi.HasName = &BootstrapScript{}
var _ fi.HasDependencies = &BootstrapScript{}

// This template is a TOML file defined here:
// https://github.com/bottlerocket-os/bottlerocket#kubernetes-settings

type bottleRocketUserdataSettingsHostContainersAdmin struct {
	Enabled bool `toml:"enabled"`
}

type bottleRocketUserdataSettingsHostContainers struct {
	Admin bottleRocketUserdataSettingsHostContainersAdmin `toml:"admin"`
}

type bottleRocketUserdataSettingsKubernetes struct {
	APIServer          string `toml:"api-server"`
	ClusterCertificate string `toml:"cluster-certificate"`
	ClusterName        string `toml:"cluster-name"`

	NodeTaints            map[string]string `toml:"node-taints,omitempty"`
	AllowedUnsafeSysctls  []string          `toml:"allowed-unsafe-sysctls,omitempty"`
	SystemReserved        map[string]string `toml:"system-reserved,omitempty"`
	RegistryQPS           *int32            `toml:"registry-qps"`
	RegistryBurst         *int32            `toml:"registry-burst"`
	EventQPS              *int32            `toml:"event-qps"`
	EventBurst            *int32            `toml:"event-burst"`
	ContainerLogMaxSize   *string           `toml:"container-log-max-size"`
	ContainerLogMaxFiles  *int32            `toml:"container-log-max-files"`
	CPUManagerPolicy      *string           `toml:"cpu-manager-policy"`
	TopologyManagerPolicy *string           `toml:"topology-manager-policy"`

	PodInfraContainerImage *string           `toml:"pod-infra-container-image"`
	KubeReserved           map[string]string `toml:"kube-reserved,omitempty"`
}

type bottleRocketUserdataSettings struct {
	HostContainers bottleRocketUserdataSettingsHostContainers `toml:"host-containers"`
	Kubernetes     bottleRocketUserdataSettingsKubernetes     `toml:"kubernetes"`
}
type bottleRocketUserdata struct {
	Settings bottleRocketUserdataSettings `toml:"settings"`
}

// kubeEnv returns the boot config for the instance group
func (b *BootstrapScript) kubeEnv(ig *kops.InstanceGroup, c *fi.Context) (string, error) {
	var alternateNames []string

	for _, hasAddress := range b.alternateNameTasks {
		address, err := hasAddress.FindIPAddress(c)
		if err != nil {
			return "", fmt.Errorf("error finding address for %v: %v", hasAddress, err)
		}
		if address == nil {
			klog.Warningf("Task did not have an address: %v", hasAddress)
			continue
		}
		klog.V(8).Infof("Resolved alternateName %q for %q", *address, hasAddress)
		alternateNames = append(alternateNames, *address)
	}

	sort.Strings(alternateNames)
	config, bootConfig, err := b.builder.NodeUpConfigBuilder.BuildConfig(ig, alternateNames, b.caTasks)
	if err != nil {
		return "", err
	}

	configData, err := utils.YamlMarshal(config)
	if err != nil {
		return "", fmt.Errorf("error converting nodeup config to yaml: %v", err)
	}
	sum256 := sha256.Sum256(configData)
	bootConfig.NodeupConfigHash = base64.StdEncoding.EncodeToString(sum256[:])
	b.nodeupConfig.Resource = fi.NewBytesResource(configData)

	bootConfigData, err := utils.YamlMarshal(bootConfig)
	if err != nil {
		return "", fmt.Errorf("error converting boot config to yaml: %v", err)
	}

	return string(bootConfigData), nil
}

func (b *BootstrapScript) buildEnvironmentVariables(cluster *kops.Cluster) (map[string]string, error) {
	env := make(map[string]string)

	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		env["GOSSIP_DNS_CONN_LIMIT"] = os.Getenv("GOSSIP_DNS_CONN_LIMIT")
	}

	if os.Getenv("S3_ENDPOINT") != "" {
		env["S3_ENDPOINT"] = os.Getenv("S3_ENDPOINT")
		env["S3_REGION"] = os.Getenv("S3_REGION")
		env["S3_ACCESS_KEY_ID"] = os.Getenv("S3_ACCESS_KEY_ID")
		env["S3_SECRET_ACCESS_KEY"] = os.Getenv("S3_SECRET_ACCESS_KEY")
	}

	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderOpenstack {

		osEnvs := []string{
			"OS_TENANT_ID", "OS_TENANT_NAME", "OS_PROJECT_ID", "OS_PROJECT_NAME",
			"OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID",
			"OS_DOMAIN_NAME", "OS_DOMAIN_ID",
			"OS_AUTH_URL",
			"OS_REGION_NAME",
		}

		hasCCM := cluster.Spec.ExternalCloudControllerManager != nil
		appCreds := os.Getenv("OS_APPLICATION_CREDENTIAL_ID") != "" && os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET") != ""
		if !hasCCM && appCreds {
			klog.Warning("application credentials only supported when using external cloud controller manager. Continuing with passwords.")
		}

		if hasCCM && appCreds {
			osEnvs = append(osEnvs,
				"OS_APPLICATION_CREDENTIAL_ID",
				"OS_APPLICATION_CREDENTIAL_SECRET",
			)
		} else {
			klog.Warning("exporting username and password. Consider using application credentials instead.")
			osEnvs = append(osEnvs,
				"OS_USERNAME",
				"OS_PASSWORD",
			)
		}

		// Pass in required credentials when using user-defined swift endpoint
		if os.Getenv("OS_AUTH_URL") != "" {
			for _, envVar := range osEnvs {
				env[envVar] = fmt.Sprintf("'%s'", os.Getenv(envVar))
			}
		}
	}

	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderDO {
		doToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
		if doToken != "" {
			env["DIGITALOCEAN_ACCESS_TOKEN"] = doToken
		}
	}

	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderAWS {
		region, err := awsup.FindRegion(cluster)
		if err != nil {
			return nil, err
		}
		if region == "" {
			klog.Warningf("unable to determine cluster region")
		} else {
			env["AWS_REGION"] = region
		}
	}

	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderALI {
		region := os.Getenv("OSS_REGION")
		if region != "" {
			env["OSS_REGION"] = os.Getenv("OSS_REGION")
		}

		aliID := os.Getenv("ALIYUN_ACCESS_KEY_ID")
		if aliID != "" {
			env["ALIYUN_ACCESS_KEY_ID"] = os.Getenv("ALIYUN_ACCESS_KEY_ID")
		}

		aliSecret := os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
		if aliSecret != "" {
			env["ALIYUN_ACCESS_KEY_SECRET"] = os.Getenv("ALIYUN_ACCESS_KEY_SECRET")
		}
	}

	if kops.CloudProviderID(cluster.Spec.CloudProvider) == kops.CloudProviderAzure {
		env["AZURE_STORAGE_ACCOUNT"] = os.Getenv("AZURE_STORAGE_ACCOUNT")
		azureEnv := os.Getenv("AZURE_ENVIRONMENT")
		if azureEnv != "" {
			env["AZURE_ENVIRONMENT"] = os.Getenv("AZURE_ENVIRONMENT")
		}
	}

	return env, nil
}

// ResourceNodeUp generates and returns a nodeup (bootstrap) script from a
// template file, substituting in specific env vars & cluster spec configuration
func (b *BootstrapScriptBuilder) ResourceNodeUp(c *fi.ModelBuilderContext, ig *kops.InstanceGroup) (fi.Resource, error) {
	keypairs := []string{"kubernetes-ca", "etcd-clients-ca"}
	for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
		k := etcdCluster.Name
		keypairs = append(keypairs, "etcd-manager-ca-"+k, "etcd-peers-ca-"+k)
		if k != "events" && k != "main" {
			keypairs = append(keypairs, "etcd-clients-ca-"+k)
		}
	}

	if model.UseCiliumEtcd(b.Cluster) && !model.UseKopsControllerForNodeBootstrap(b.Cluster) {
		keypairs = append(keypairs, "etcd-client-cilium")
	}
	if ig.HasAPIServer() {
		keypairs = append(keypairs, "apiserver-aggregator-ca", "service-account", "etcd-clients-ca")
	} else if !model.UseKopsControllerForNodeBootstrap(b.Cluster) {
		keypairs = append(keypairs, "kubelet", "kube-proxy")
		if b.Cluster.Spec.Networking.Kuberouter != nil {
			keypairs = append(keypairs, "kube-router")
		}
	}

	if ig.IsBastion() {
		keypairs = nil

		// Bastions can have AdditionalUserData, but if there isn't any skip this part
		if len(ig.Spec.AdditionalUserData) == 0 {
			templateResource, err := NewTemplateResource("nodeup", "", nil, nil)
			if err != nil {
				return nil, err
			}
			return templateResource, nil
		}
	}

	if ig.Spec.ImageFamily != nil && ig.Spec.ImageFamily.Bottlerocket != nil {
		keypairs = []string{"kubernetes-ca"}
	}

	caTasks := map[string]*fitasks.Keypair{}
	for _, keypair := range keypairs {
		caTaskObject, found := c.Tasks["Keypair/"+keypair]
		if !found {
			return nil, fmt.Errorf("keypair/%s task not found", keypair)
		}
		caTasks[keypair] = caTaskObject.(*fitasks.Keypair)
	}

	task := &BootstrapScript{
		Name:      ig.Name,
		Lifecycle: b.Lifecycle,
		ig:        ig,
		builder:   b,
		caTasks:   caTasks,
	}
	task.resource.Task = task
	if ig.Spec.ImageFamily == nil || ig.Spec.ImageFamily.Bottlerocket == nil {
		task.nodeupConfig.Task = task

		c.AddTask(&fitasks.ManagedFile{
			Name:      fi.String("nodeupconfig-" + ig.Name),
			Lifecycle: b.Lifecycle,
			Location:  fi.String("igconfig/" + strings.ToLower(string(ig.Spec.Role)) + "/" + ig.Name + "/nodeupconfig.yaml"),
			Contents:  &task.nodeupConfig,
		})
	}
	c.AddTask(task)
	return &task.resource, nil
}

func (b *BootstrapScript) GetName() *string {
	return &b.Name
}

func (b *BootstrapScript) GetDependencies(tasks map[string]fi.Task) []fi.Task {
	var deps []fi.Task

	for _, task := range tasks {
		if hasAddress, ok := task.(fi.HasAddress); ok && hasAddress.IsForAPIServer() {
			deps = append(deps, task)
			b.alternateNameTasks = append(b.alternateNameTasks, hasAddress)
		}
		if b.ig.Spec.ImageFamily != nil && b.ig.Spec.ImageFamily.Bottlerocket != nil {
			if t, ok := task.(*fitasks.Keypair); ok && fi.StringValue(t.Name) == fi.CertificateIDCA {
				deps = append(deps, task)
			}
		}
	}

	for _, task := range b.caTasks {
		deps = append(deps, task)
	}

	return deps
}

func (b *BootstrapScript) Run(c *fi.Context) error {
	if b.Lifecycle == fi.LifecycleIgnore {
		return nil
	}
	if b.ig.Spec.ImageFamily != nil && b.ig.Spec.ImageFamily.Bottlerocket != nil {
		return b.runBottleRocket(c)
	}
	config, err := b.kubeEnv(b.ig, c)
	if err != nil {
		return err
	}

	functions := template.FuncMap{
		"NodeUpSourceAmd64": func() string {
			if b.builder.NodeUpAssets[architectures.ArchitectureAmd64] != nil {
				return strings.Join(b.builder.NodeUpAssets[architectures.ArchitectureAmd64].Locations, ",")
			}
			return ""
		},
		"NodeUpSourceHashAmd64": func() string {
			if b.builder.NodeUpAssets[architectures.ArchitectureAmd64] != nil {
				return b.builder.NodeUpAssets[architectures.ArchitectureAmd64].Hash.Hex()
			}
			return ""
		},
		"NodeUpSourceArm64": func() string {
			if b.builder.NodeUpAssets[architectures.ArchitectureArm64] != nil {
				return strings.Join(b.builder.NodeUpAssets[architectures.ArchitectureArm64].Locations, ",")
			}
			return ""
		},
		"NodeUpSourceHashArm64": func() string {
			if b.builder.NodeUpAssets[architectures.ArchitectureArm64] != nil {
				return b.builder.NodeUpAssets[architectures.ArchitectureArm64].Hash.Hex()
			}
			return ""
		},
		"KubeEnv": func() string {
			return config
		},

		"EnvironmentVariables": func() (string, error) {
			env, err := b.buildEnvironmentVariables(c.Cluster)
			if err != nil {
				return "", err
			}

			// Sort keys to have a stable sequence of "export xx=xxx"" statements
			var keys []string
			for k := range env {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var b bytes.Buffer
			for _, k := range keys {
				b.WriteString(fmt.Sprintf("export %s=%s\n", k, env[k]))
			}
			return b.String(), nil
		},

		"ProxyEnv": func() string {
			return b.createProxyEnv(c.Cluster.Spec.EgressProxy)
		},

		"ClusterSpec": func() (string, error) {
			cs := c.Cluster.Spec

			spec := make(map[string]interface{})
			spec["cloudConfig"] = cs.CloudConfig
			spec["containerRuntime"] = cs.ContainerRuntime
			spec["containerd"] = cs.Containerd
			spec["docker"] = cs.Docker
			spec["kubeProxy"] = cs.KubeProxy
			spec["kubelet"] = cs.Kubelet

			if cs.KubeAPIServer != nil && cs.KubeAPIServer.EnableBootstrapAuthToken != nil {
				spec["kubeAPIServer"] = map[string]interface{}{
					"enableBootstrapAuthToken": cs.KubeAPIServer.EnableBootstrapAuthToken,
				}
			}

			if b.ig.IsMaster() {
				spec["encryptionConfig"] = cs.EncryptionConfig
				spec["etcdClusters"] = make(map[string]kops.EtcdClusterSpec)
				spec["kubeAPIServer"] = cs.KubeAPIServer
				spec["kubeControllerManager"] = cs.KubeControllerManager
				spec["kubeScheduler"] = cs.KubeScheduler
				spec["masterKubelet"] = cs.MasterKubelet

				for _, etcdCluster := range cs.EtcdClusters {
					c := kops.EtcdClusterSpec{
						Image:         etcdCluster.Image,
						Version:       etcdCluster.Version,
						Manager:       etcdCluster.Manager,
						CPURequest:    etcdCluster.CPURequest,
						MemoryRequest: etcdCluster.MemoryRequest,
					}
					for _, etcdMember := range etcdCluster.Members {
						if fi.StringValue(etcdMember.InstanceGroup) == b.ig.Name && etcdMember.VolumeSize != nil {
							m := kops.EtcdMemberSpec{
								Name:       etcdMember.Name,
								VolumeSize: etcdMember.VolumeSize,
							}
							c.Members = append(c.Members, m)
						}
					}
					spec["etcdClusters"].(map[string]kops.EtcdClusterSpec)[etcdCluster.Name] = c
				}
			}

			content, err := yaml.Marshal(spec)
			if err != nil {
				return "", fmt.Errorf("error converting cluster spec to yaml for inclusion within bootstrap script: %v", err)
			}
			return string(content), nil
		},

		"CompressUserData": func() *bool {
			return b.ig.Spec.CompressUserData
		},

		"GzipBase64": func(data string) (string, error) {
			return gzipBase64(data)
		},

		"SetSysctls": func() string {
			// By setting some sysctls early, we avoid broken configurations that prevent nodeup download.
			// See https://github.com/kubernetes/kops/issues/10206 for details.
			return setSysctls()
		},
	}

	nodeUpTemplate, err := resources.NodeUpTemplate(b.ig)
	if err != nil {
		return err
	}

	templateResource, err := NewTemplateResource("nodeup", nodeUpTemplate, functions, nil)
	if err != nil {
		return err
	}

	b.resource.Resource = templateResource
	return nil
}

func (b *BootstrapScript) runBottleRocket(c *fi.Context) error {
	var caTask *fitasks.Keypair
	for _, task := range c.AllTasks() {
		if t, ok := task.(*fitasks.Keypair); ok && fi.StringValue(t.Name) == fi.CertificateIDCA {
			caTask = task.(*fitasks.Keypair)
		}
	}
	if caTask == nil {
		return fmt.Errorf("CA task not found")
	}
	caResource := caTask.Certificates()
	ca, err := fi.ResourceAsBytes(caResource)
	if err != nil {
		return err
	}

	// Kubelet's --register-with-taints just takes a list of strings
	// but bottlerocket's TOML requires they be in map form
	nodeTaints := make(map[string]string)
	for _, t := range b.ig.Spec.Taints {
		taint := strings.Split(t, "=")
		if len(taint) != 2 {
			return fmt.Errorf("unrecognized node taint format %v", t)
		}
		nodeTaints[taint[0]] = taint[1]
	}

	kubelet := &kops.KubeletConfigSpec{}
	if b.ig.Spec.Role == kops.InstanceGroupRoleMaster {
		reflectutils.JSONMergeStruct(kubelet, c.Cluster.Spec.MasterKubelet)
	} else {
		reflectutils.JSONMergeStruct(kubelet, c.Cluster.Spec.Kubelet)
	}

	if b.ig.Spec.Kubelet != nil {
		reflectutils.JSONMergeStruct(kubelet, b.ig.Spec.Kubelet)
	}

	userdata := bottleRocketUserdata{
		Settings: bottleRocketUserdataSettings{
			HostContainers: bottleRocketUserdataSettingsHostContainers{
				Admin: bottleRocketUserdataSettingsHostContainersAdmin{
					Enabled: true,
				},
			},
			Kubernetes: bottleRocketUserdataSettingsKubernetes{
				APIServer:          fmt.Sprintf("https://%v", c.Cluster.Spec.MasterInternalName),
				ClusterName:        c.Cluster.Name,
				ClusterCertificate: base64.StdEncoding.EncodeToString(ca),
				NodeTaints:         nodeTaints,

				AllowedUnsafeSysctls: kubelet.AllowedUnsafeSysctls,
				SystemReserved:       kubelet.SystemReserved,
				RegistryQPS:          kubelet.RegistryPullQPS,
				RegistryBurst:        kubelet.RegistryBurst,
				EventQPS:             kubelet.EventQPS,
				EventBurst:           kubelet.EventBurst,
				ContainerLogMaxFiles: kubelet.ContainerLogMaxFiles,
				KubeReserved:         kubelet.KubeReserved,
			},
		},
	}
	if kubelet.CpuManagerPolicy != "" {
		userdata.Settings.Kubernetes.CPUManagerPolicy = &kubelet.CpuManagerPolicy
	}
	if kubelet.TopologyManagerPolicy != "" {
		userdata.Settings.Kubernetes.TopologyManagerPolicy = &kubelet.TopologyManagerPolicy
	}
	if kubelet.ContainerLogMaxSize != "" {
		userdata.Settings.Kubernetes.ContainerLogMaxSize = &kubelet.ContainerLogMaxSize
	}
	if kubelet.PodInfraContainerImage != "" {
		userdata.Settings.Kubernetes.PodInfraContainerImage = &kubelet.PodInfraContainerImage
	}
	var userdataBytes bytes.Buffer
	err = toml.NewEncoder(&userdataBytes).QuoteMapKeys(true).Encode(&userdata)
	if err != nil {
		return err
	}
	b.resource.Resource = fi.NewBytesResource(userdataBytes.Bytes())
	return nil
}

func (b *BootstrapScript) createProxyEnv(ps *kops.EgressProxySpec) string {
	var buffer bytes.Buffer

	if ps != nil && ps.HTTPProxy.Host != "" {
		var httpProxyURL string

		// TODO double check that all the code does this
		// TODO move this into a validate so we can enforce the string syntax
		if !strings.HasPrefix(ps.HTTPProxy.Host, "http://") {
			httpProxyURL = "http://"
		}

		if ps.HTTPProxy.Port != 0 {
			httpProxyURL += ps.HTTPProxy.Host + ":" + strconv.Itoa(ps.HTTPProxy.Port)
		} else {
			httpProxyURL += ps.HTTPProxy.Host
		}

		// Set env variables for base environment
		buffer.WriteString(`echo "http_proxy=` + httpProxyURL + `" >> /etc/environment` + "\n")
		buffer.WriteString(`echo "https_proxy=` + httpProxyURL + `" >> /etc/environment` + "\n")
		buffer.WriteString(`echo "no_proxy=` + ps.ProxyExcludes + `" >> /etc/environment` + "\n")
		buffer.WriteString(`echo "NO_PROXY=` + ps.ProxyExcludes + `" >> /etc/environment` + "\n")

		// Load the proxy environment variables
		buffer.WriteString("while read in; do export $in; done < /etc/environment\n")

		// Set env variables for package manager depending on OS Distribution (N/A for Flatcar)
		// Note: Nodeup will source the `/etc/environment` file within docker config in the correct location
		buffer.WriteString("case `cat /proc/version` in\n")
		buffer.WriteString("*[Dd]ebian*)\n")
		buffer.WriteString(`  echo "Acquire::http::Proxy \"${http_proxy}\";" > /etc/apt/apt.conf.d/30proxy ;;` + "\n")
		buffer.WriteString("*[Uu]buntu*)\n")
		buffer.WriteString(`  echo "Acquire::http::Proxy \"${http_proxy}\";" > /etc/apt/apt.conf.d/30proxy ;;` + "\n")
		buffer.WriteString("*[Rr]ed[Hh]at*)\n")
		buffer.WriteString(`  echo "proxy=${http_proxy}" >> /etc/yum.conf ;;` + "\n")
		buffer.WriteString("esac\n")

		// Set env variables for systemd
		buffer.WriteString(`echo "DefaultEnvironment=\"http_proxy=${http_proxy}\" \"https_proxy=${http_proxy}\"`)
		buffer.WriteString(` \"NO_PROXY=${no_proxy}\" \"no_proxy=${no_proxy}\""`)
		buffer.WriteString(" >> /etc/systemd/system.conf\n")

		// Restart stuff
		buffer.WriteString("systemctl daemon-reload\n")
		buffer.WriteString("systemctl daemon-reexec\n")
	}
	return buffer.String()
}

func gzipBase64(data string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err := gz.Write([]byte(data))
	if err != nil {
		return "", err
	}

	if err = gz.Flush(); err != nil {
		return "", err
	}

	if err = gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

func setSysctls() string {
	var b bytes.Buffer

	// Based on https://github.com/kubernetes/kops/issues/10206#issuecomment-766852332
	b.WriteString("sysctl -w net.core.rmem_max=16777216 || true\n")
	b.WriteString("sysctl -w net.core.wmem_max=16777216 || true\n")
	b.WriteString("sysctl -w net.ipv4.tcp_rmem='4096 87380 16777216' || true\n")
	b.WriteString("sysctl -w net.ipv4.tcp_wmem='4096 87380 16777216' || true\n")

	return b.String()
}

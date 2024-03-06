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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/model/resources"
	"k8s.io/kops/pkg/wellknownservices"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/fitasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
)

type NodeUpConfigBuilder interface {
	BuildConfig(ig *kops.InstanceGroup, wellKnownAddresses WellKnownAddresses, keysets map[string]*fi.Keyset) (*nodeup.Config, *nodeup.BootConfig, error)
}

// WellKnownAddresses holds known addresses for well-known services
type WellKnownAddresses map[wellknownservices.WellKnownService][]string

// BootstrapScriptBuilder creates the bootstrap script
type BootstrapScriptBuilder struct {
	*KopsModelContext
	Lifecycle           fi.Lifecycle
	NodeUpAssets        map[architectures.Architecture]*mirrors.MirroredAsset
	NodeUpConfigBuilder NodeUpConfigBuilder
}

type BootstrapScript struct {
	Name      string
	Lifecycle fi.Lifecycle
	cluster   *kops.Cluster
	ig        *kops.InstanceGroup
	builder   *BootstrapScriptBuilder
	resource  fi.CloudupTaskDependentResource

	// hasAddressTasks holds fi.HasAddress tasks, that contribute well-known services.
	hasAddressTasks []fi.HasAddress

	// caTasks hold the CA tasks, for dependency analysis.
	caTasks map[string]*fitasks.Keypair

	// nodeupConfig contains the nodeup config.
	nodeupConfig fi.CloudupTaskDependentResource
}

var (
	_ fi.CloudupTask            = &BootstrapScript{}
	_ fi.HasName                = &BootstrapScript{}
	_ fi.CloudupHasDependencies = &BootstrapScript{}
)

// kubeEnv returns the boot config for the instance group
func (b *BootstrapScript) kubeEnv(ig *kops.InstanceGroup, c *fi.CloudupContext) (*nodeup.BootConfig, error) {
	wellKnownAddresses := make(WellKnownAddresses)

	for _, hasAddress := range b.hasAddressTasks {
		addresses, err := hasAddress.FindAddresses(c)
		if err != nil {
			return nil, fmt.Errorf("error finding address for %v: %v", hasAddress, err)
		}
		if len(addresses) == 0 {
			// Such tasks won't have an address in dry-run mode, until the resource is created
			klog.V(2).Infof("Task did not have an address: %v", hasAddress)
			continue
		}

		klog.V(8).Infof("Resolved alternateNames %q for %q", addresses, hasAddress)

		for _, wellKnownService := range hasAddress.GetWellKnownServices() {
			wellKnownAddresses[wellKnownService] = append(wellKnownAddresses[wellKnownService], addresses...)
		}
	}

	for k := range wellKnownAddresses {
		sort.Strings(wellKnownAddresses[k])
	}

	keysets := make(map[string]*fi.Keyset)
	for _, caTask := range b.caTasks {
		name := *caTask.Name
		keyset := caTask.Keyset()
		if keyset == nil {
			return nil, fmt.Errorf("failed to get keyset from %q", name)
		}
		keysets[name] = keyset
	}
	config, bootConfig, err := b.builder.NodeUpConfigBuilder.BuildConfig(ig, wellKnownAddresses, keysets)
	if err != nil {
		return nil, err
	}

	configData, err := utils.YamlMarshal(config)
	if err != nil {
		return nil, fmt.Errorf("error converting nodeup config to yaml: %v", err)
	}
	sum256 := sha256.Sum256(configData)
	bootConfig.NodeupConfigHash = base64.StdEncoding.EncodeToString(sum256[:])
	b.nodeupConfig.Resource = fi.NewBytesResource(configData)

	return bootConfig, nil
}

func (b *BootstrapScript) buildEnvironmentVariables() (map[string]string, error) {
	cluster := b.cluster

	env := make(map[string]string)

	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		env["GOSSIP_DNS_CONN_LIMIT"] = os.Getenv("GOSSIP_DNS_CONN_LIMIT")
	}

	if os.Getenv("S3_ENDPOINT") != "" {
		if b.ig.IsControlPlane() {
			env["S3_ENDPOINT"] = os.Getenv("S3_ENDPOINT")
			env["S3_REGION"] = os.Getenv("S3_REGION")
			env["S3_ACCESS_KEY_ID"] = os.Getenv("S3_ACCESS_KEY_ID")
			env["S3_SECRET_ACCESS_KEY"] = os.Getenv("S3_SECRET_ACCESS_KEY")
		}
	}

	if cluster.Spec.GetCloudProvider() == kops.CloudProviderOpenstack {

		osEnvs := []string{
			"OS_TENANT_ID", "OS_TENANT_NAME", "OS_PROJECT_ID", "OS_PROJECT_NAME",
			"OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID",
			"OS_DOMAIN_NAME", "OS_DOMAIN_ID",
			"OS_AUTH_URL",
			"OS_REGION_NAME",
		}

		appCreds := os.Getenv("OS_APPLICATION_CREDENTIAL_ID") != "" && os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET") != ""
		if appCreds {
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

		// credentials needed always in control-plane and when using gossip also in nodes
		passEnvs := false
		if b.ig.IsControlPlane() || cluster.UsesLegacyGossip() {
			passEnvs = true
		}
		// Pass in required credentials when using user-defined swift endpoint
		if os.Getenv("OS_AUTH_URL") != "" && passEnvs {
			for _, envVar := range osEnvs {
				env[envVar] = fmt.Sprintf("'%s'", os.Getenv(envVar))
			}
		}
	}

	if cluster.Spec.GetCloudProvider() == kops.CloudProviderDO {
		if b.ig.IsControlPlane() {
			doToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
			if doToken != "" {
				env["DIGITALOCEAN_ACCESS_TOKEN"] = doToken
			}
		}
	}

	if cluster.Spec.GetCloudProvider() == kops.CloudProviderHetzner && (b.ig.IsControlPlane() || cluster.UsesLegacyGossip()) {
		hcloudToken := os.Getenv("HCLOUD_TOKEN")
		if hcloudToken != "" {
			env["HCLOUD_TOKEN"] = hcloudToken
		}
	}

	if cluster.Spec.GetCloudProvider() == kops.CloudProviderAWS {
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

	if cluster.Spec.GetCloudProvider() == kops.CloudProviderAzure {
		env["AZURE_STORAGE_ACCOUNT"] = os.Getenv("AZURE_STORAGE_ACCOUNT")
		azureEnv := os.Getenv("AZURE_ENVIRONMENT")
		if azureEnv != "" {
			env["AZURE_ENVIRONMENT"] = os.Getenv("AZURE_ENVIRONMENT")
		}
	}

	if cluster.Spec.GetCloudProvider() == kops.CloudProviderScaleway && (b.ig.IsControlPlane() || cluster.UsesLegacyGossip()) {
		profile, err := scaleway.CreateValidScalewayProfile()
		if err != nil {
			return nil, err
		}
		env["SCW_ACCESS_KEY"] = fi.ValueOf(profile.AccessKey)
		env["SCW_SECRET_KEY"] = fi.ValueOf(profile.SecretKey)
		env["SCW_DEFAULT_PROJECT_ID"] = fi.ValueOf(profile.DefaultProjectID)
	}

	return env, nil
}

// ResourceNodeUp generates and returns a nodeup (bootstrap) script from a
// template file, substituting in specific env vars & cluster spec configuration
func (b *BootstrapScriptBuilder) ResourceNodeUp(c *fi.CloudupModelBuilderContext, ig *kops.InstanceGroup) (fi.Resource, error) {
	keypairs := []string{"kubernetes-ca", "etcd-clients-ca"}
	for _, etcdCluster := range b.Cluster.Spec.EtcdClusters {
		k := etcdCluster.Name
		keypairs = append(keypairs, "etcd-manager-ca-"+k, "etcd-peers-ca-"+k)
		if k != "events" && k != "main" {
			keypairs = append(keypairs, "etcd-clients-ca-"+k)
		}
	}

	if ig.HasAPIServer() {
		keypairs = append(keypairs, "apiserver-aggregator-ca", "service-account", "etcd-clients-ca")
	}

	if ig.IsBastion() {
		keypairs = nil

		// Bastions can have AdditionalUserData, but if there isn't any skip this part
		if len(ig.Spec.AdditionalUserData) == 0 {
			return nil, nil
		}
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
		cluster:   b.Cluster,
		ig:        ig,
		builder:   b,
		caTasks:   caTasks,
	}
	task.resource.Task = task
	task.nodeupConfig.Task = task
	c.AddTask(task)

	c.AddTask(&fitasks.ManagedFile{
		Name:      fi.PtrTo("nodeupconfig-" + ig.Name),
		Lifecycle: b.Lifecycle,
		Location:  fi.PtrTo("igconfig/" + ig.Spec.Role.ToLowerString() + "/" + ig.Name + "/nodeupconfig.yaml"),
		Contents:  &task.nodeupConfig,
	})
	return &task.resource, nil
}

func (b *BootstrapScript) GetName() *string {
	return &b.Name
}

func (b *BootstrapScript) GetDependencies(tasks map[string]fi.CloudupTask) []fi.CloudupTask {
	var deps []fi.CloudupTask

	for _, task := range tasks {
		if hasAddress, ok := task.(fi.HasAddress); ok && len(hasAddress.GetWellKnownServices()) > 0 {
			deps = append(deps, task)
			b.hasAddressTasks = append(b.hasAddressTasks, hasAddress)
		}
	}

	for _, task := range b.caTasks {
		deps = append(deps, task)
	}

	return deps
}

func (b *BootstrapScript) Run(c *fi.CloudupContext) error {
	if b.Lifecycle == fi.LifecycleIgnore {
		return nil
	}

	bootConfig, err := b.kubeEnv(b.ig, c)
	if err != nil {
		return err
	}

	var nodeupScript resources.NodeUpScript
	nodeupScript.NodeUpAssets = b.builder.NodeUpAssets
	nodeupScript.BootConfig = bootConfig

	{
		nodeupScript.EnvironmentVariables = func() (string, error) {
			env, err := b.buildEnvironmentVariables()
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
		}

		nodeupScript.ProxyEnv = func() (string, error) {
			return b.createProxyEnv(c.T.Cluster.Spec.Networking.EgressProxy)
		}
	}

	nodeupScript.CompressUserData = fi.ValueOf(b.ig.Spec.CompressUserData)

	// By setting some sysctls early, we avoid broken configurations that prevent nodeup download.
	// See https://github.com/kubernetes/kops/issues/10206 for details.
	nodeupScript.SetSysctls = setSysctls()

	nodeupScript.CloudProvider = string(c.T.Cluster.Spec.GetCloudProvider())

	nodeupScriptResource, err := nodeupScript.Build()
	if err != nil {
		return err
	}

	b.resource.Resource = fi.FunctionToResource(func() ([]byte, error) {
		nodeupScript, err := fi.ResourceAsString(nodeupScriptResource)
		if err != nil {
			return nil, err
		}

		awsUserData, err := resources.AWSMultipartMIME(nodeupScript, b.ig)
		if err != nil {
			return nil, err
		}

		return []byte(awsUserData), nil
	})
	return nil
}

func (b *BootstrapScript) createProxyEnv(ps *kops.EgressProxySpec) (string, error) {
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
		buffer.WriteString(`{` + "\n")
		buffer.WriteString(`  echo "http_proxy=` + httpProxyURL + `"` + "\n")
		buffer.WriteString(`  echo "https_proxy=` + httpProxyURL + `"` + "\n")
		buffer.WriteString(`  echo "no_proxy=` + ps.ProxyExcludes + `"` + "\n")
		buffer.WriteString(`  echo "NO_PROXY=` + ps.ProxyExcludes + `"` + "\n")
		buffer.WriteString(`} >> /etc/environment` + "\n")

		// Load the proxy environment variables
		buffer.WriteString("while read -r in; do export \"${in?}\"; done < /etc/environment\n")

		// Set env variables for package manager depending on OS Distribution (N/A for Flatcar)
		// Note: Nodeup will source the `/etc/environment` file within docker config in the correct location
		buffer.WriteString("case $(cat /proc/version) in\n")
		buffer.WriteString("*[Dd]ebian* | *[Uu]buntu*)\n")
		buffer.WriteString(`  echo "Acquire::http::Proxy \"` + httpProxyURL + `\";" > /etc/apt/apt.conf.d/30proxy ;;` + "\n")
		buffer.WriteString("*[Rr]ed[Hh]at*)\n")
		buffer.WriteString(`  echo "proxy=` + httpProxyURL + `" >> /etc/yum.conf ;;` + "\n")
		buffer.WriteString("esac\n")

		// Set env variables for systemd
		buffer.WriteString(`echo "DefaultEnvironment=\"http_proxy=` + httpProxyURL + `\" \"https_proxy=` + httpProxyURL + `\"`)
		buffer.WriteString(` \"NO_PROXY=` + ps.ProxyExcludes + `\" \"no_proxy=` + ps.ProxyExcludes + `\""`)
		buffer.WriteString(" >> /etc/systemd/system.conf\n")

		// Restart stuff
		buffer.WriteString("systemctl daemon-reload\n")
		buffer.WriteString("systemctl daemon-reexec\n")
	}
	return buffer.String(), nil
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

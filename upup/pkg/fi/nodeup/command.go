/*
Copyright 2016 The Kubernetes Authors.

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
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kops/nodeup/pkg/distros"
	"k8s.io/kops/nodeup/pkg/model"
	api "k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/kops/registry"
	"k8s.io/kops/pkg/apis/kops/util"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/vfs"
)

// We should probably retry for a long time - there is not really any great fallback
const MaxTaskDuration = 365 * 24 * time.Hour

type NodeUpCommand struct {
	config         *nodeup.Config
	cluster        *api.Cluster
	instanceGroup  *api.InstanceGroup
	ConfigLocation string
	ModelDir       vfs.Path
	CacheDir       string
	Target         string
	FSRoot         string
}

func (c *NodeUpCommand) Run(out io.Writer) error {
	if c.FSRoot == "" {
		return fmt.Errorf("FSRoot is required")
	}

	if c.ConfigLocation != "" {
		config, err := vfs.Context.ReadFile(c.ConfigLocation)
		if err != nil {
			return fmt.Errorf("error loading configuration %q: %v", c.ConfigLocation, err)
		}

		err = utils.YamlUnmarshal(config, &c.config)
		if err != nil {
			return fmt.Errorf("error parsing configuration %q: %v", c.ConfigLocation, err)
		}
	} else {
		return fmt.Errorf("ConfigLocation is required")
	}

	if c.CacheDir == "" {
		return fmt.Errorf("CacheDir is required")
	}
	assetStore := fi.NewAssetStore(c.CacheDir)
	for _, asset := range c.config.Assets {
		err := assetStore.Add(asset)
		if err != nil {
			return fmt.Errorf("error adding asset %q: %v", asset, err)
		}
	}

	var configBase vfs.Path
	if fi.StringValue(c.config.ConfigBase) != "" {
		var err error
		configBase, err = vfs.Context.BuildVfsPath(*c.config.ConfigBase)
		if err != nil {
			return fmt.Errorf("cannot parse ConfigBase %q: %v", *c.config.ConfigBase, err)
		}
	} else if fi.StringValue(c.config.ClusterLocation) != "" {
		basePath := *c.config.ClusterLocation
		lastSlash := strings.LastIndex(basePath, "/")
		if lastSlash != -1 {
			basePath = basePath[0:lastSlash]
		}

		var err error
		configBase, err = vfs.Context.BuildVfsPath(basePath)
		if err != nil {
			return fmt.Errorf("cannot parse inferred ConfigBase %q: %v", basePath, err)
		}
	} else {
		return fmt.Errorf("ConfigBase is required")
	}

	c.cluster = &api.Cluster{}
	{
		clusterLocation := fi.StringValue(c.config.ClusterLocation)

		var p vfs.Path
		if clusterLocation != "" {
			var err error
			p, err = vfs.Context.BuildVfsPath(clusterLocation)
			if err != nil {
				return fmt.Errorf("error parsing ClusterLocation %q: %v", clusterLocation, err)
			}
		} else {
			p = configBase.Join(registry.PathClusterCompleted)
		}

		b, err := p.ReadFile()
		if err != nil {
			return fmt.Errorf("error loading Cluster %q: %v", p, err)
		}

		err = utils.YamlUnmarshal(b, c.cluster)
		if err != nil {
			return fmt.Errorf("error parsing Cluster %q: %v", p, err)
		}
	}

	if c.config.InstanceGroupName != "" {
		instanceGroupLocation := configBase.Join("instancegroup", c.config.InstanceGroupName)

		c.instanceGroup = &api.InstanceGroup{}
		b, err := instanceGroupLocation.ReadFile()
		if err != nil {
			return fmt.Errorf("error loading InstanceGroup %q: %v", instanceGroupLocation, err)
		}

		err = utils.YamlUnmarshal(b, c.instanceGroup)
		if err != nil {
			return fmt.Errorf("error parsing InstanceGroup %q: %v", instanceGroupLocation, err)
		}
	} else {
		glog.Warningf("No instance group defined in nodeup config")
	}

	err := evaluateSpec(c.cluster)
	if err != nil {
		return err
	}

	//if c.Config.ConfigurationStore != "" {
	//	// TODO: If we ever delete local files, we need to filter so we only copy
	//	// certain directories (i.e. not secrets / keys), because dest is a parent dir!
	//	p, err := c.buildPath(c.Config.ConfigurationStore)
	//	if err != nil {
	//		return fmt.Errorf("error building config store: %v", err)
	//	}
	//
	//	dest := vfs.NewFSPath("/etc/kubernetes")
	//	scanner := vfs.NewVFSScan(p)
	//	err = vfs.SyncDir(scanner, dest)
	//	if err != nil {
	//		return fmt.Errorf("error copying config store: %v", err)
	//	}
	//
	//	c.Config.Tags = append(c.Config.Tags, "_config_store")
	//} else {
	//	c.Config.Tags = append(c.Config.Tags, "_not_config_store")
	//}

	distribution, err := distros.FindDistribution(c.FSRoot)
	if err != nil {
		return fmt.Errorf("error determining OS distribution: %v", err)
	}

	osTags := distribution.BuildTags()

	nodeTags := sets.NewString()
	nodeTags.Insert(osTags...)
	nodeTags.Insert(c.config.Tags...)

	glog.Infof("Config tags: %v", c.config.Tags)
	glog.Infof("OS tags: %v", osTags)

	tf, err := newTemplateFunctions(c.config, c.cluster, c.instanceGroup, nodeTags)
	if err != nil {
		return fmt.Errorf("error initializing: %v", err)
	}

	k8sVersion, err := util.ParseKubernetesVersion(c.cluster.Spec.KubernetesVersion)
	if err != nil || k8sVersion == nil {
		return fmt.Errorf("unable to parse KubernetesVersion %q", c.cluster.Spec.KubernetesVersion)
	}

	modelContext := &model.NodeupModelContext{
		NodeupConfig:  c.config,
		Cluster:       c.cluster,
		Distribution:  distribution,
		Architecture:  model.ArchitectureAmd64,
		InstanceGroup: c.instanceGroup,
		IsMaster:      nodeTags.Has(TagMaster),
		Assets:        assetStore,
		KeyStore:      tf.keyStore,
		SecretStore:   tf.secretStore,

		KubernetesVersion: *k8sVersion,
	}

	loader := NewLoader(c.config, c.cluster, assetStore, nodeTags)
	loader.Builders = append(loader.Builders, &model.DirectoryBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.DockerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.ProtokubeBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.CloudConfigBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeletBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubectlBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.EtcdBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.LogrotateBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.PackagesBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.SecretBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.FirewallBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.NetworkBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.SysctlBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeAPIServerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeControllerManagerBuilder{NodeupModelContext: modelContext})
	loader.Builders = append(loader.Builders, &model.KubeSchedulerBuilder{NodeupModelContext: modelContext})
	if c.cluster.Spec.Networking.Kuberouter == nil {
		loader.Builders = append(loader.Builders, &model.KubeProxyBuilder{NodeupModelContext: modelContext})
	} else {
		loader.Builders = append(loader.Builders, &model.KubeRouterBuilder{NodeupModelContext: modelContext})
	}
	loader.Builders = append(loader.Builders, &model.HookBuilder{NodeupModelContext: modelContext})

	tf.populate(loader.TemplateFunctions)

	taskMap, err := loader.Build(c.ModelDir)
	if err != nil {
		return fmt.Errorf("error building loader: %v", err)
	}

	for i, image := range c.config.Images {
		taskMap["LoadImage."+strconv.Itoa(i)] = &nodetasks.LoadImageTask{
			Source: image.Source,
			Hash:   image.Hash,
		}
	}
	if c.config.ProtokubeImage != nil {
		taskMap["LoadImage.protokube"] = &nodetasks.LoadImageTask{
			Source: c.config.ProtokubeImage.Source,
			Hash:   c.config.ProtokubeImage.Hash,
		}
	}

	var cloud fi.Cloud
	var keyStore fi.Keystore
	var secretStore fi.SecretStore
	var target fi.Target
	checkExisting := true

	switch c.Target {
	case "direct":
		target = &local.LocalTarget{
			CacheDir: c.CacheDir,
			Tags:     nodeTags,
		}
	case "dryrun":
		assetBuilder := assets.NewAssetBuilder()
		target = fi.NewDryRunTarget(assetBuilder, out)
	case "cloudinit":
		checkExisting = false
		target = cloudinit.NewCloudInitTarget(out, nodeTags)
	default:
		return fmt.Errorf("unsupported target type %q", c.Target)
	}

	context, err := fi.NewContext(target, cloud, keyStore, secretStore, configBase, checkExisting, taskMap)
	if err != nil {
		glog.Exitf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(MaxTaskDuration)
	if err != nil {
		glog.Exitf("error running tasks: %v", err)
	}

	err = target.Finish(taskMap)
	if err != nil {
		glog.Exitf("error closing target: %v", err)
	}

	return nil
}

func evaluateSpec(c *api.Cluster) error {
	var err error

	c.Spec.Kubelet.HostnameOverride, err = evaluateHostnameOverride(c.Spec.Kubelet.HostnameOverride)
	if err != nil {
		return err
	}

	c.Spec.MasterKubelet.HostnameOverride, err = evaluateHostnameOverride(c.Spec.MasterKubelet.HostnameOverride)
	if err != nil {
		return err
	}

	if c.Spec.KubeProxy != nil {
		c.Spec.KubeProxy.HostnameOverride, err = evaluateHostnameOverride(c.Spec.KubeProxy.HostnameOverride)
		if err != nil {
			return err
		}
	}

	if c.Spec.Docker != nil {
		err = evaluateDockerSpecStorage(c.Spec.Docker)
		if err != nil {
			return err
		}
	}

	return nil
}

func evaluateHostnameOverride(hostnameOverride string) (string, error) {
	if hostnameOverride == "" {
		return "", nil
	}
	k := strings.TrimSpace(hostnameOverride)
	k = strings.ToLower(k)

	if k != "@aws" {
		return hostnameOverride, nil
	}

	// We recognize @aws as meaning "the local-hostname from the aws metadata service"
	vBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/local-hostname")
	if err != nil {
		return "", fmt.Errorf("error reading local hostname from AWS metadata: %v", err)
	}

	// The local-hostname gets it's hostname from the AWS DHCP Option Set, which
	// may provide multiple hostnames separated by spaces. For now just choose
	// the first one as the hostname.
	domains := strings.Fields(string(vBytes))
	if len(domains) == 0 {
		glog.Warningf("Local hostname from AWS metadata service was empty")
		return "", nil
	} else {
		domain := domains[0]
		glog.Infof("Using hostname from AWS metadata service: %s", domain)

		return domain, nil
	}
}

// evaluateDockerSpec selects the first supported storage mode, if it is a list
func evaluateDockerSpecStorage(spec *api.DockerConfig) error {
	storage := fi.StringValue(spec.Storage)
	if strings.Contains(fi.StringValue(spec.Storage), ",") {
		precedence := strings.Split(storage, ",")
		for _, opt := range precedence {
			fs := opt
			if fs == "overlay2" {
				fs = "overlay"
			}
			supported, err := kernelHasFilesystem(fs)
			if err != nil {
				glog.Warningf("error checking if %q filesystem is supported: %v", fs, err)
				continue
			}

			if !supported {
				// overlay -> overlay
				// aufs -> aufs
				module := fs
				err := modprobe(fs)
				if err != nil {
					glog.Warningf("error running `modprobe %q`: %v", module, err)
				}
			}

			supported, err = kernelHasFilesystem(fs)
			if err != nil {
				glog.Warningf("error checking if %q filesystem is supported: %v", fs, err)
				continue
			}

			if supported {
				glog.Infof("Using supported docker storage %q", opt)
				spec.Storage = fi.String(opt)
				return nil
			}

			glog.Warningf("%q docker storage was specified, but filesystem is not supported", opt)
		}

		// Just in case we don't recognize the driver?
		// TODO: Is this the best behaviour
		glog.Warningf("No storage module was supported from %q, will default to %q", storage, precedence[0])
		spec.Storage = fi.String(precedence[0])
		return nil
	}

	return nil
}

// kernelHasFilesystem checks if /proc/filesystems contains the specified filesystem
func kernelHasFilesystem(fs string) (bool, error) {
	contents, err := ioutil.ReadFile("/proc/filesystems")
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
	glog.Infof("Doing modprobe for module %v", module)
	out, err := exec.Command("/sbin/modprobe", module).CombinedOutput()
	outString := string(out)
	if err != nil {
		return fmt.Errorf("modprobe for module %q failed (%v): %s", module, err, outString)
	}
	if outString != "" {
		glog.Infof("Output from modprobe %s:\n%s", module, outString)
	}
	return nil
}

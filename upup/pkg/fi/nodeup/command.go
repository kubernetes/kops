package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kops/upup/pkg/api"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kops/upup/pkg/fi/nodeup/local"
	"k8s.io/kops/upup/pkg/fi/nodeup/nodetasks"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/upup/pkg/fi/vfs"
	"strconv"
	"strings"
)

// We should probably retry for a long time - there is not really any great fallback
const MaxAttemptsWithNoProgress = 100

type NodeUpCommand struct {
	config         *NodeUpConfig
	cluster        *api.Cluster
	ConfigLocation string
	ModelDir       string
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
	assets := fi.NewAssetStore(c.CacheDir)
	for _, asset := range c.config.Assets {
		err := assets.Add(asset)
		if err != nil {
			return fmt.Errorf("error adding asset %q: %v", asset, err)
		}
	}

	//c.nodeset = &cloudup.NodeSetConfig{}
	//if c.config.NodeSetLocation != "" {
	//	b, err := vfs.Context.ReadFile(c.config.NodeSetLocation)
	//	if err != nil {
	//		return fmt.Errorf("error loading NodeSet %q: %v", c.config.NodeSetLocation, err)
	//	}
	//
	//	err = utils.YamlUnmarshal(b, c.nodeset)
	//	if err != nil {
	//		return fmt.Errorf("error parsing NodeSet %q: %v", c.config.NodeSetLocation, err)
	//	}
	//} else {
	//	return fmt.Errorf("NodeSetLocation is required")
	//}

	c.cluster = &api.Cluster{}
	if c.config.ClusterLocation != "" {
		b, err := vfs.Context.ReadFile(c.config.ClusterLocation)
		if err != nil {
			return fmt.Errorf("error loading Cluster %q: %v", c.config.ClusterLocation, err)
		}

		err = utils.YamlUnmarshal(b, c.cluster)
		if err != nil {
			return fmt.Errorf("error parsing Cluster %q: %v", c.config.ClusterLocation, err)
		}
	} else {
		// TODO Infer this from NodeSetLocation?
		return fmt.Errorf("ClusterLocation is required")
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

	osTags, err := FindOSTags(c.FSRoot)
	if err != nil {
		return fmt.Errorf("error determining OS tags: %v", err)
	}

	tags := make(map[string]struct{})
	for _, tag := range osTags {
		tags[tag] = struct{}{}
	}
	for _, tag := range c.config.Tags {
		tags[tag] = struct{}{}
	}

	loader := NewLoader(c.config, c.cluster, assets, tags)

	tf, err := newTemplateFunctions(c.config, c.cluster, tags)
	if err != nil {
		return fmt.Errorf("error initializing: %v", err)
	}
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

	var cloud fi.Cloud
	var caStore fi.CAStore
	var secretStore fi.SecretStore
	var target fi.Target
	checkExisting := true

	switch c.Target {
	case "direct":
		target = &local.LocalTarget{
			CacheDir: c.CacheDir,
		}
	case "dryrun":
		target = fi.NewDryRunTarget(out)
	case "cloudinit":
		checkExisting = false
		target = cloudinit.NewCloudInitTarget(out)
	default:
		return fmt.Errorf("unsupported target type %q", c.Target)
	}

	context, err := fi.NewContext(target, cloud, caStore, secretStore, checkExisting)
	if err != nil {
		glog.Exitf("error building context: %v", err)
	}
	defer context.Close()

	err = context.RunTasks(taskMap, MaxAttemptsWithNoProgress)
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

	return nil
}

func evaluateHostnameOverride(hostnameOverride string) (string, error) {
	k := strings.TrimSpace(hostnameOverride)
	k = strings.ToLower(k)

	if hostnameOverride != "@aws" {
		return hostnameOverride, nil
	}

	// We recognize @aws as meaning "the local-hostname from the aws metadata service"
	vBytes, err := vfs.Context.ReadFile("metadata://aws/meta-data/local-hostname")
	if err != nil {
		return "", fmt.Errorf("error reading local hostname from AWS metadata: %v", err)
	}
	v := strings.TrimSpace(string(vBytes))
	if v == "" {
		glog.Warningf("Local hostname from AWS metadata service was empty")
	} else {
		glog.Infof("Using hostname from AWS metadata service: %s", v)
	}
	return v, nil
}

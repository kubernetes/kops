package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/local"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
	"k8s.io/kube-deploy/upup/pkg/fi/vfs"
)

type NodeUpCommand struct {
	Config         *NodeConfig
	ConfigLocation string
	ModelDir       string
	AssetDir       string
	Target         string
}

func (c *NodeUpCommand) Run(out io.Writer) error {
	if c.ConfigLocation != "" {
		config, err := vfs.Context.ReadFile(c.ConfigLocation)
		if err != nil {
			return fmt.Errorf("error loading configuration %q: %v", c.ConfigLocation, err)
		}

		err = utils.YamlUnmarshal(config, c.Config)
		if err != nil {
			return fmt.Errorf("error parsing configuration %q: %v", c.ConfigLocation, err)
		}
	}

	if c.AssetDir == "" {
		return fmt.Errorf("AssetDir is required")
	}
	assets := fi.NewAssetStore(c.AssetDir)
	for _, asset := range c.Config.Assets {
		err := assets.Add(asset)
		if err != nil {
			return fmt.Errorf("error adding asset %q: %v", asset, err)
		}
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

	loader := NewLoader(c.Config, assets)

	err := buildTemplateFunctions(c.Config, loader.TemplateFunctions)
	if err != nil {
		return fmt.Errorf("error initializing: %v", err)
	}

	taskMap, err := loader.Build(c.ModelDir)
	if err != nil {
		return fmt.Errorf("error building loader: %v", err)
	}

	var cloud fi.Cloud
	var caStore fi.CAStore
	var secretStore fi.SecretStore
	var target fi.Target
	checkExisting := true

	switch c.Target {
	case "direct":
		target = &local.LocalTarget{}
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

	err = context.RunTasks(taskMap)
	if err != nil {
		glog.Exitf("error running tasks: %v", err)
	}

	err = target.Finish(taskMap)
	if err != nil {
		glog.Exitf("error closing target: %v", err)
	}

	return nil
}

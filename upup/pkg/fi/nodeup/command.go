package nodeup

import (
	"fmt"
	"github.com/golang/glog"
	"io"
	"k8s.io/kube-deploy/upup/pkg/fi"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/cloudinit"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup/local"
	"k8s.io/kube-deploy/upup/pkg/fi/utils"
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
		config, err := utils.ReadLocation(c.ConfigLocation)
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

	loader := NewLoader(c.Config, assets)

	taskMap, err := loader.Build(c.ModelDir)
	if err != nil {
		glog.Exitf("error building: %v", err)
	}

	var cloud fi.Cloud
	var caStore fi.CAStore
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

	context, err := fi.NewContext(target, cloud, caStore, checkExisting)
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

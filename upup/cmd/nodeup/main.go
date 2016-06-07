package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/upup/pkg/fi/nodeup"
	"os"
)

func main() {
	flagModel := "model"
	flag.StringVar(&flagModel, "model", flagModel, "directory to use as model for desired configuration")
	var flagConf string
	flag.StringVar(&flagConf, "conf", "node.yaml", "configuration location")
	var flagAssetDir string
	flag.StringVar(&flagAssetDir, "assets", "/var/cache/nodeup", "the location for the local asset cache")

	dryrun := false
	flag.BoolVar(&dryrun, "dryrun", false, "Don't create cloud resources; just show what would be done")
	target := "direct"
	flag.StringVar(&target, "target", target, "Target - direct, cloudinit")

	if dryrun {
		target = "dryrun"
	}

	flag.Set("logtostderr", "true")
	flag.Parse()

	if flagConf == "" {
		glog.Exitf("--conf is required")
	}

	config := &nodeup.NodeConfig{}
	cmd := &nodeup.NodeUpCommand{
		Config:         config,
		ConfigLocation: flagConf,
		ModelDir:       flagModel,
		Target:         target,
		AssetDir:       flagAssetDir,
	}
	err := cmd.Run(os.Stdout)
	if err != nil {
		glog.Exitf("error running nodeup: %v", err)
		os.Exit(1)
	}
	fmt.Printf("success")
}

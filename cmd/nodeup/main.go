package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops/upup/pkg/fi/nodeup"
	"os"
)

func main() {
	flagModel := "model"
	flag.StringVar(&flagModel, "model", flagModel, "directory to use as model for desired configuration")
	var flagConf string
	flag.StringVar(&flagConf, "conf", "node.yaml", "configuration location")
	var flagAssetDir string
	flag.StringVar(&flagAssetDir, "assets", "/var/cache/nodeup", "the location for the local asset cache")
	var flagRootFS string
	flag.StringVar(&flagRootFS, "rootfs", "/", "the location of the machine root (for running in a container)")

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

	cmd := &nodeup.NodeUpCommand{
		ConfigLocation: flagConf,
		ModelDir:       flagModel,
		Target:         target,
		AssetDir:       flagAssetDir,
		FSRoot:         flagRootFS,
	}
	err := cmd.Run(os.Stdout)
	if err != nil {
		glog.Exitf("error running nodeup: %v", err)
		os.Exit(1)
	}
	fmt.Printf("success")
}

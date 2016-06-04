package main

import (
	"flag"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/protokube/pkg/protokube"
	"os"
)

func main() {
	//flagModel := "model"
	//flag.StringVar(&flagModel, "model", flagModel, "directory to use as model for desired configuration")
	//var flagConf string
	//flag.StringVar(&flagConf, "conf", "node.yaml", "configuration location")
	//var flagAssetDir string
	//flag.StringVar(&flagAssetDir, "assets", "/var/cache/nodeup", "the location for the local asset cache")
	//
	//dryrun := false
	//flag.BoolVar(&dryrun, "dryrun", false, "Don't create cloud resources; just show what would be done")
	//target := "direct"
	//flag.StringVar(&target, "target", target, "Target - direct, cloudinit")

	//if dryrun {
	//	target = "dryrun"
	//}

	flag.Set("logtostderr", "true")
	flag.Parse()

	volumes, err := protokube.NewAWSVolumes()
	if err != nil {
		glog.Errorf("Error initializing AWS: %q", err)
		os.Exit(1)
	}

	//if flagConf == "" {
	//	glog.Exitf("--conf is required")
	//}

	kubeboot := protokube.NewKubeBoot(volumes)
	err = kubeboot.Bootstrap()
	if err != nil {
		glog.Errorf("Error during bootstrap: %q", err)
		os.Exit(1)
	}

	glog.Infof("Bootstrap complete; starting kubelet")

	err = kubeboot.RunBootstrapTasks()
	if err != nil {
		glog.Errorf("Error during bootstrap: %q", err)
		os.Exit(1)
	}

	glog.Infof("Unexpected exited from kubelet run")
	os.Exit(1)
}

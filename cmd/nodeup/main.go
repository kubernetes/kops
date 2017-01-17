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

package main // import "k8s.io/kops/cmd/nodeup"

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kops"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi/nodeup"
	"os"
	"time"
)

const retryInterval = 30 * time.Second

func main() {
	gitVersion := ""
	if kops.GitVersion != "" {
		gitVersion = " (git-" + kops.GitVersion + ")"
	}
	fmt.Printf("nodeup version %s%s\n", kops.Version, gitVersion)

	var flagConf string
	flag.StringVar(&flagConf, "conf", "node.yaml", "configuration location")
	var flagCacheDir string
	flag.StringVar(&flagCacheDir, "cache", "/var/cache/nodeup", "the location for the local asset cache")
	var flagRootFS string
	flag.StringVar(&flagRootFS, "rootfs", "/", "the location of the machine root (for running in a container)")
	var flagRetries int
	flag.IntVar(&flagRetries, "retries", -1, "maximum number of retries on failure: -1 means retry forever")

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

	retries := flagRetries

	for {
		cmd := &nodeup.NodeUpCommand{
			ConfigLocation: flagConf,
			Target:         target,
			CacheDir:       flagCacheDir,
			FSRoot:         flagRootFS,
			ModelDir:       models.NewAssetPath("nodeup"),
		}
		err := cmd.Run(os.Stdout)
		if err == nil {
			fmt.Printf("success")
			os.Exit(0)
		}

		if retries == 0 {
			glog.Exitf("error running nodeup: %v", err)
			os.Exit(1)
		}

		if retries > 0 {
			retries--
		}

		glog.Warningf("got error running nodeup (will retry in %s): %v", retryInterval, err)
		time.Sleep(retryInterval)
	}
}

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

package main // import "k8s.io/kops/cmd/nodeup"

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"k8s.io/klog"
	"k8s.io/kops"
	"k8s.io/kops/nodeup/pkg/bootstrap"
	"k8s.io/kops/upup/models"
	"k8s.io/kops/upup/pkg/fi/nodeup"
)

const (
	retryInterval = 30 * time.Second
	procSelfExe   = "/proc/self/exe"
)

func main() {
	klog.InitFlags(nil)

	var flagConf, flagCacheDir, flagRootFS, gitVersion string
	var flagRetries int
	var dryrun, installSystemdUnit bool
	target := "direct"

	if kops.GitVersion != "" {
		gitVersion = fmt.Sprintf(" (git-%s)", kops.GitVersion)
	}
	fmt.Printf("nodeup version %s%s\n", kops.Version, gitVersion)
	flag.StringVar(&flagConf, "conf", "node.yaml", "configuration location")
	flag.StringVar(&flagCacheDir, "cache", "/var/cache/nodeup", "the location for the local asset cache")
	flag.StringVar(&flagRootFS, "rootfs", "/", "the location of the machine root (for running in a container)")
	flag.IntVar(&flagRetries, "retries", -1, "maximum number of retries on failure: -1 means retry forever")
	flag.BoolVar(&dryrun, "dryrun", false, "Don't create cloud resources; just show what would be done")
	flag.StringVar(&target, "target", target, "Target - direct, cloudinit")
	flag.BoolVar(&installSystemdUnit, "install-systemd-unit", installSystemdUnit, "If true, will install a systemd unit instead of running directly")

	if dryrun {
		target = "dryrun"
	}

	flag.Set("logtostderr", "true")
	flag.Parse()

	if flagConf == "" {
		klog.Exitf("--conf is required")
	}

	retries := flagRetries

	for {
		var err error
		if installSystemdUnit {
			// create a systemd unit to bootstrap kops
			// using the same args as we were called with
			var command []string
			for i := 0; i < len(os.Args); i++ {
				s := os.Args[i]
				if s == "-install-systemd-unit" || s == "--install-systemd-unit" {
					continue
				}
				if i == 0 {
					// We could also try to evaluate based on cwd
					if _, err := os.Stat(procSelfExe); os.IsNotExist(err) {
						klog.Fatalf("file %v does not exist", procSelfExe)
					}

					fi, err := os.Lstat(procSelfExe)
					if err != nil {
						klog.Fatalf("error doing lstat on %q: %v", procSelfExe, err)
					}
					if fi.Mode()&os.ModeSymlink != os.ModeSymlink {
						klog.Fatalf("file %v is not a symlink", procSelfExe)
					}

					s, err = os.Readlink(procSelfExe)
					if err != nil {
						klog.Fatalf("error reading %v link: %v", procSelfExe, err)
					}
				}
				command = append(command, s)
			}
			i := bootstrap.Installation{
				CacheDir: flagCacheDir,
				Command:  command,
				FSRoot:   flagRootFS,
			}
			i.RunTasksOptions.InitDefaults()
			i.RunTasksOptions.MaxTaskDuration = 5 * time.Minute
			err = i.Run()
			if err == nil {
				fmt.Printf("service installed")
				os.Exit(0)
			}
		} else {
			ctx := context.Background()
			cmd := &nodeup.NodeUpCommand{
				ConfigLocation: flagConf,
				Target:         target,
				CacheDir:       flagCacheDir,
				FSRoot:         flagRootFS,
				ModelDir:       models.NewAssetPath("nodeup"),
			}
			err = cmd.Run(ctx, os.Stdout)
			if err == nil {
				fmt.Printf("success")
				os.Exit(0)
			}
		}

		if retries == 0 {
			klog.Exitf("error running nodeup: %v", err)
			os.Exit(1)
		}

		if retries > 0 {
			retries--
		}

		klog.Warningf("got error running nodeup (will retry in %s): %v", retryInterval, err)
		time.Sleep(retryInterval)
	}
}

/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"fmt"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/kube-deploy/cluster-api-gcp/cloud/google"
)

type options struct {
	version      string
	role         string
	dockerImages []string
}

var opts options

var generateCmd = &cobra.Command{
	Use:   "generate_image",
	Short: "Outputs a script to generate a preloaded image",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runGenerate(opts); err != nil {
			glog.Exit(err)
		}
	},
}

func init() {
	generateCmd.Flags().StringVar(&opts.version, "version", "1.7.3", "The version of kubernetes to install")
	generateCmd.Flags().StringVar(&opts.role, "role", "master", "The role of the machine (master or node)")
	generateCmd.Flags().StringArrayVar(&opts.dockerImages, "extra-docker-images", []string{}, "extra docker images to preload")
}

func runGenerate(o options) error {
	var script string
	var err error
	switch o.role {
	case "master":
		script, err = google.PreloadMasterScript(o.version, o.dockerImages)
	case "node":
		script, err = google.PreloadMasterScript(o.version, o.dockerImages)
	default:
		return fmt.Errorf("unrecognized role: %q", o.role)
	}

	if err != nil {
		return err
	}

	// just print the script for now
	// TODO actually start a VM, let it run the script, stop the VM, then create the image
	fmt.Println(script)
	return nil
}

func main() {
	if err := generateCmd.Execute(); err != nil {
		glog.Exit(err)
	}
}

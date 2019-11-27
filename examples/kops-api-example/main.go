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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"k8s.io/kops/util/pkg/vfs"
)

// registryBase is the base path where state files are kept (the state store)
var registryBase vfs.Path

// clusterName is the name of the cluster to create
var clusterName string

// nodeZones is the set of zones in which we will run nodes
var nodeZones []string

// masterZones is the set of zones in which we will run masters
var masterZones []string

var sshPublicKey = "~/.ssh/id_rsa.pub"

var flagRegistryBase = flag.String("registry", os.Getenv("KOPS_STATE_STORE"), "VFS path where files are kept")
var flagClusterName = flag.String("name", "", "Name of cluster to create")
var flagZones = flag.String("zones", "", "Comma separated list of zones to create")

func main() {
	flag.Parse()

	err := parseFlags()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	err = up()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error from up: %v\n", err)
		os.Exit(1)
	}

	err = apply()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error from apply: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() error {
	var err error

	registryBase, err = vfs.Context.BuildVfsPath(*flagRegistryBase)
	if err != nil {
		return fmt.Errorf("error parsing registry path %q: %v", *flagRegistryBase, err)
	}

	clusterName = *flagClusterName
	if clusterName == "" {
		return fmt.Errorf("Must pass -name with cluster name")
	}

	if *flagZones == "" {
		return fmt.Errorf("Must pass -zones with comma-separated list of zones")
	}
	nodeZones = strings.Split(*flagZones, ",")
	masterZones = []string{nodeZones[0]}

	return nil
}

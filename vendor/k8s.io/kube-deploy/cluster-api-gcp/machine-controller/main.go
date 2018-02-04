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
	goflag "flag"
	"fmt"
	"os"

	"github.com/spf13/pflag"

	"k8s.io/apiserver/pkg/util/flag"
	"k8s.io/apiserver/pkg/util/logs"

	"k8s.io/kube-deploy/cluster-api-gcp/machine-controller/controller"
)

func main() {
	c := controller.NewConfiguration()
	c.AddFlags(pflag.CommandLine)

	flag.InitFlags()
	pflag.Parse()
	// Suppress warning
	// https://github.com/kubernetes/kubernetes/issues/17162
	goflag.CommandLine.Parse([]string{})
	logs.InitLogs()
	defer logs.FlushLogs()

	mc := controller.NewMachineController(c)
	if err := mc.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

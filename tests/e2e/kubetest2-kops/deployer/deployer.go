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

// Package deployer implements the kubetest2 Kops deployer
package deployer

import (
	"flag"
	"sync"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"
	"k8s.io/kops/tests/e2e/kubetest2-kops/builder"

	"sigs.k8s.io/kubetest2/pkg/types"
)

// Name is the name of the deployer
const Name = "kops"

type deployer struct {
	// generic parts
	commonOptions types.Options
	// doInit helps to make sure the initialization is performed only once
	doInit sync.Once

	KopsRoot      string `flag:"kops-root" desc:"Path to root of the kops repo. Used with --build."`
	StageLocation string `flag:"stage-location" desc:"Storage location for kops artifacts. Only gs:// paths are supported."`
	BuildOptions  *builder.BuildOptions
}

// assert that New implements types.NewDeployer
var _ types.NewDeployer = New

// assert that deployer implements types.Deployer
var _ types.Deployer = &deployer{}

func (d *deployer) Provider() string {
	return Name
}

func (d *deployer) Up() error {
	klog.Warning("Up is not implemented")
	return nil
}

func (d *deployer) IsUp() (bool, error) {
	klog.Warning("IsUp is not implemented")
	return true, nil
}

func (d *deployer) Down() error {
	klog.Warning("Down is not implemented")
	return nil
}

func (d *deployer) DumpClusterLogs() error {
	klog.Warning("DumpClusterLogs is not implemented")
	return nil
}

// New implements deployer.New for kops
func New(opts types.Options) (types.Deployer, *pflag.FlagSet) {
	// create a deployer object and set fields that are not flag controlled
	d := &deployer{
		commonOptions: opts,
		BuildOptions:  &builder.BuildOptions{},
	}

	// register flags
	fs := bindFlags(d)

	// register flags for klog
	klog.InitFlags(nil)
	fs.AddGoFlagSet(flag.CommandLine)
	return d, fs
}

func bindFlags(d *deployer) *pflag.FlagSet {
	flags, err := gpflag.Parse(d)
	if err != nil {
		klog.Fatalf("unable to generate flags from deployer")
		return nil
	}
	return flags
}

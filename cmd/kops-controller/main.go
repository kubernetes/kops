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
	"io/ioutil"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"k8s.io/kops/cmd/kops-controller/controllers"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/pkg/nodeidentity"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	nodeidentitydo "k8s.io/kops/pkg/nodeidentity/do"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	nodeidentityos "k8s.io/kops/pkg/nodeidentity/openstack"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/yaml"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {

	// +kubebuilder:scaffold:scheme
}

func main() {
	klog.InitFlags(nil)

	// Disable metrics by default (avoid port conflicts, also risky because we are host network)
	metricsAddress := ":0"
	//flag.StringVar(&metricsAddr, "metrics-addr", metricsAddress, "The address the metric endpoint binds to.")

	configPath := "/etc/kubernetes/kops-controller/config.yaml"
	flag.StringVar(&configPath, "conf", configPath, "Location of yaml configuration file")

	flag.Parse()

	if configPath == "" {
		klog.Fatalf("must specify --conf")
	}

	var opt config.Options
	opt.PopulateDefaults()

	{
		b, err := ioutil.ReadFile(configPath)
		if err != nil {
			klog.Fatalf("failed to read configuration file %q: %v", configPath, err)
		}

		if err := yaml.Unmarshal(b, &opt); err != nil {
			klog.Fatalf("failed to parse configuration file %q: %v", configPath, err)
		}
	}

	ctrl.SetLogger(klogr.New())

	if err := buildScheme(); err != nil {
		setupLog.Error(err, "error building scheme")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddress,
		LeaderElection:     true,
		LeaderElectionID:   "kops-controller-leader",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err := addNodeController(mgr, &opt); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeController")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func buildScheme() error {
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering corev1: %v", err)
	}
	return nil
}

func addNodeController(mgr manager.Manager, opt *config.Options) error {
	var identifier nodeidentity.Identifier
	var err error
	switch opt.Cloud {
	case "aws":
		identifier, err = nodeidentityaws.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}
	case "gce":
		identifier, err = nodeidentitygce.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "openstack":
		identifier, err = nodeidentityos.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "digitalocean":
		identifier, err = nodeidentitydo.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "":
		return fmt.Errorf("must specify cloud")

	default:
		return fmt.Errorf("identifier for cloud %q not implemented", opt.Cloud)
	}

	if opt.ConfigBase == "" {
		return fmt.Errorf("must specify configBase")
	}

	nodeController, err := controllers.NewNodeReconciler(mgr, opt.ConfigBase, identifier)
	if err != nil {
		return err
	}
	if err := nodeController.SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}

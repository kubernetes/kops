/*

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

	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
	"k8s.io/klog/klogr"
	"k8s.io/kops/cmd/kops-controller/pkg/apis"
	"k8s.io/kops/cmd/kops-controller/pkg/controller"
	"k8s.io/kops/cmd/kops-controller/pkg/webhook"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
)

func main() {
	klog.InitFlags(nil)

	flag.Parse()
	logf.SetLogger(klogr.New())

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func run() error {
	// Get a config to talk to the apiserver
	klog.Infof("setting up client for manager")
	cfg, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("unable to set up client config: %v", err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	klog.Infof("setting up manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		return fmt.Errorf("unable to set up manager: %v", err)
	}

	klog.Infof("Registering Components.")

	// Setup Scheme for all resources
	klog.Infof("setting up scheme")
	scheme := mgr.GetScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return fmt.Errorf("unable to set up scheme: %v", err)
	}

	if err := v1alpha2.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error setting up scheme: %v", err)
	}

	// Setup all Controllers
	klog.Infof("Setting up controller")
	if err := controller.AddToManager(mgr); err != nil {
		return fmt.Errorf("unable to set up controllers: %v", err)
	}

	klog.Infof("setting up webhooks")
	if err := webhook.AddToManager(mgr); err != nil {
		return fmt.Errorf("unable to set up webhooks: %v", err)
	}

	// Start the Cmd
	klog.Infof("Starting the Cmd.")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error running the manager: %v", err)
	}

	return nil
}

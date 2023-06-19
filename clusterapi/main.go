/*
Copyright 2023 The Kubernetes Authors.

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
	"context"
	"flag"
	"fmt"
	"os"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"k8s.io/kops/clusterapi/bootstrap/controllers"
	bootstrapapi "k8s.io/kops/clusterapi/bootstrap/kops/api/v1beta1"
	controlplaneapi "k8s.io/kops/clusterapi/controlplane/kops/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/server"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	// +kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	klog.InitFlags(nil)

	// Disable metrics by default (avoid port conflicts, also risky because we are host network)
	metricsAddress := ":0"

	flag.Parse()

	ctrl.SetLogger(klogr.New())

	if err := buildScheme(); err != nil {
		return fmt.Errorf("error building scheme: %w", err)
	}

	kubeConfig := ctrl.GetConfigOrDie()
	options := ctrl.Options{
		Scheme: scheme,
		// MetricsBindAddress: metricsAddress,
		// LeaderElection:      true,
		// LeaderElectionID:    "kops-clusterapi-leader",
	}
	options.Metrics = server.Options{
		BindAddress: metricsAddress,
	}
	mgr, err := ctrl.NewManager(kubeConfig, options)

	if err != nil {
		return fmt.Errorf("error starting manager: %w", err)
	}

	if err := controllers.NewKopsConfigReconciler(mgr); err != nil {
		return fmt.Errorf("error creating controller: %w", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return fmt.Errorf("error running manager: %w", err)
	}
	return nil
}

func buildScheme() error {
	if err := corev1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering corev1: %v", err)
	}

	if err := bootstrapapi.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering api: %w", err)
	}

	if err := controlplaneapi.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering api: %w", err)
	}

	// Needed so that the leader-election system can post events
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		return fmt.Errorf("error registering coordinationv1: %v", err)
	}
	return nil
}

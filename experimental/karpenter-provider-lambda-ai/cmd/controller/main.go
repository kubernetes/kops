/*
Copyright The Kubernetes Authors.

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
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"sigs.k8s.io/karpenter/pkg/controllers"
	"sigs.k8s.io/karpenter/pkg/controllers/state"
	"sigs.k8s.io/karpenter/pkg/operator"

	"k8s.io/kops/karpenter-provider-lambda-ai/pkg/apis/v1alpha1"
	"k8s.io/kops/karpenter-provider-lambda-ai/pkg/cloudprovider"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	ctx, op := operator.NewOperator()
	if op == nil {
		setupLog.Error(nil, "failed to create operator")
		os.Exit(1)
	}

	// Register our custom types to the operator's scheme
	if err := v1alpha1.AddToScheme(op.GetScheme()); err != nil {
		setupLog.Error(err, "unable to add v1alpha1 scheme")
		os.Exit(1)
	}

	// Initialize the Cloud Provider
	cp := cloudprovider.New(op.GetClient())

	// Initialize Cluster State
	cluster := state.NewCluster(op.Clock, op.GetClient(), cp)

	// Initialize Controllers
	coreControllers := controllers.NewControllers(
		ctx,
		op.Manager,
		op.Clock,
		op.GetClient(),
		op.EventRecorder,
		cp,
		cp,
		cluster,
		op.InstanceTypeStore,
	)

	// Register controllers
	op.WithControllers(ctx, coreControllers...)

	if err := op.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := op.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	op.Start(ctx)
}

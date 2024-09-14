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
	"context"
	"flag"
	"fmt"
	"os"

	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog/v2"
	"k8s.io/klog/v2/klogr"
	"k8s.io/kops/cmd/kops-controller/controllers"
	"k8s.io/kops/cmd/kops-controller/pkg/config"
	"k8s.io/kops/cmd/kops-controller/pkg/server"
	"k8s.io/kops/pkg/apis/kops/v1alpha2"
	"k8s.io/kops/pkg/bootstrap"
	"k8s.io/kops/pkg/bootstrap/pkibootstrap"
	"k8s.io/kops/pkg/nodeidentity"
	nodeidentityaws "k8s.io/kops/pkg/nodeidentity/aws"
	nodeidentityazure "k8s.io/kops/pkg/nodeidentity/azure"
	nodeidentitydo "k8s.io/kops/pkg/nodeidentity/do"
	nodeidentitygce "k8s.io/kops/pkg/nodeidentity/gce"
	nodeidentityhetzner "k8s.io/kops/pkg/nodeidentity/hetzner"
	nodeidentitymetal "k8s.io/kops/pkg/nodeidentity/metal"
	nodeidentityos "k8s.io/kops/pkg/nodeidentity/openstack"
	nodeidentityscw "k8s.io/kops/pkg/nodeidentity/scaleway"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/azure"
	"k8s.io/kops/upup/pkg/fi/cloudup/do"
	"k8s.io/kops/upup/pkg/fi/cloudup/gce/tpm/gcetpmverifier"
	"k8s.io/kops/upup/pkg/fi/cloudup/hetzner"
	"k8s.io/kops/upup/pkg/fi/cloudup/openstack"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/util/pkg/vfs"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/yaml"
	// +kubebuilder:scaffold:imports
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	// +kubebuilder:scaffold:scheme
}

func main() {
	ctx := context.Background()

	klog.InitFlags(nil)

	// Disable metrics by default (avoid port conflicts, also risky because we are host network)
	metricsAddress := ":0"
	// flag.StringVar(&metricsAddr, "metrics-addr", metricsAddress, "The address the metric endpoint binds to.")

	configPath := "/etc/kubernetes/kops-controller/config.yaml"
	flag.StringVar(&configPath, "conf", configPath, "Location of yaml configuration file")

	flag.Parse()

	if configPath == "" {
		klog.Fatalf("must specify --conf")
	}

	var opt config.Options
	opt.PopulateDefaults()

	{
		b, err := os.ReadFile(configPath)
		if err != nil {
			klog.Fatalf("failed to read configuration file %q: %v", configPath, err)
		}

		if err := yaml.Unmarshal(b, &opt); err != nil {
			klog.Fatalf("failed to parse configuration file %q: %v", configPath, err)
		}
	}

	ctrl.SetLogger(klogr.New())

	scheme, err := buildScheme()
	if err != nil {
		setupLog.Error(err, "error building scheme")
		os.Exit(1)
	}

	kubeConfig := ctrl.GetConfigOrDie()
	kubeConfig.Burst = 200
	kubeConfig.QPS = 100

	mgr, err := ctrl.NewManager(kubeConfig, ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddress,
		},
		LeaderElection:   true,
		LeaderElectionID: "kops-controller-leader",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	vfsContext := vfs.NewVFSContext()

	if opt.Server != nil {
		var verifiers []bootstrap.Verifier
		var err error
		if opt.Server.Provider.AWS != nil {
			verifier, err := awsup.NewAWSVerifier(ctx, opt.Server.Provider.AWS)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}
		if opt.Server.Provider.GCE != nil {
			verifier, err := gcetpmverifier.NewTPMVerifier(opt.Server.Provider.GCE)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}
		if opt.Server.Provider.Hetzner != nil {
			verifier, err := hetzner.NewHetznerVerifier(opt.Server.Provider.Hetzner)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}
		if opt.Server.Provider.OpenStack != nil {
			verifier, err := openstack.NewOpenstackVerifier(opt.Server.Provider.OpenStack)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}
		if opt.Server.Provider.DigitalOcean != nil {
			verifier, err := do.NewVerifier(ctx, opt.Server.Provider.DigitalOcean)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}
		if opt.Server.Provider.Scaleway != nil {
			verifier, err := scaleway.NewScalewayVerifier(ctx, opt.Server.Provider.Scaleway)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}
		if opt.Server.Provider.Azure != nil {
			verifier, err := azure.NewAzureVerifier(ctx, opt.Server.Provider.Azure)
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}

		if opt.Server.PKI != nil {
			verifier, err := pkibootstrap.NewVerifier(opt.Server.PKI, mgr.GetClient())
			if err != nil {
				setupLog.Error(err, "unable to create verifier")
				os.Exit(1)
			}
			verifiers = append(verifiers, verifier)
		}

		if len(verifiers) == 0 {
			klog.Fatalf("server verifiers not provided")
		}

		uncachedClient, err := client.New(mgr.GetConfig(), client.Options{
			Scheme: mgr.GetScheme(),
			Mapper: mgr.GetRESTMapper(),
		})
		if err != nil {
			setupLog.Error(err, "error creating uncached client")
			os.Exit(1)
		}

		verifier := bootstrap.NewChainVerifier(verifiers...)

		srv, err := server.NewServer(vfsContext, &opt, verifier, uncachedClient)
		if err != nil {
			setupLog.Error(err, "unable to create server")
			os.Exit(1)
		}
		mgr.Add(srv)
	}

	if opt.EnableCloudIPAM {
		if err := setupCloudIPAM(ctx, mgr, &opt); err != nil {
			setupLog.Error(err, "unable to setup cloud IPAM")
			os.Exit(1)

		}
	}

	if err := addNodeController(ctx, mgr, vfsContext, &opt); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NodeController")
		os.Exit(1)
	}

	if err := addGossipController(mgr, &opt); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GossipController")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func buildScheme() (*runtime.Scheme, error) {
	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error registering corev1: %v", err)
	}
	if err := v1alpha2.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error registering kops/v1alpha2 API: %v", err)
	}
	// Needed so that the leader-election system can post events
	if err := coordinationv1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("error registering coordinationv1: %v", err)
	}
	return scheme, nil
}

func addNodeController(ctx context.Context, mgr manager.Manager, vfsContext *vfs.VFSContext, opt *config.Options) error {
	var legacyIdentifier nodeidentity.LegacyIdentifier
	var identifier nodeidentity.Identifier
	var err error
	switch opt.Cloud {
	case "aws":
		identifier, err = nodeidentityaws.New(ctx, opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "gce":
		legacyIdentifier, err = nodeidentitygce.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "openstack":
		identifier, err = nodeidentityos.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "digitalocean":
		legacyIdentifier, err = nodeidentitydo.New()
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "hetzner":
		identifier, err = nodeidentityhetzner.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %w", err)
		}

	case "azure":
		identifier, err = nodeidentityazure.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %v", err)
		}

	case "scaleway":
		identifier, err = nodeidentityscw.New(opt.CacheNodeidentityInfo)
		if err != nil {
			return fmt.Errorf("error building identifier: %w", err)
		}

	case "metal":
		identifier, err = nodeidentitymetal.New()
		if err != nil {
			return fmt.Errorf("error building metal node identifier: %w", err)
		}

	case "":
		return fmt.Errorf("must specify cloud")

	default:
		return fmt.Errorf("identifier for cloud %q not implemented", opt.Cloud)
	}

	if identifier != nil {
		nodeController, err := controllers.NewNodeReconciler(mgr, identifier)
		if err != nil {
			return err
		}
		if err := nodeController.SetupWithManager(mgr); err != nil {
			return err
		}
	} else {
		if opt.ConfigBase == "" {
			return fmt.Errorf("must specify configBase")
		}
		if opt.SecretStore == "" {
			return fmt.Errorf("must specify secretStore")
		}

		nodeController, err := controllers.NewLegacyNodeReconciler(mgr, vfsContext, opt.ConfigBase, legacyIdentifier)
		if err != nil {
			return err
		}
		if err := nodeController.SetupWithManager(mgr); err != nil {
			return err
		}
	}

	return nil
}

func addGossipController(mgr manager.Manager, opt *config.Options) error {
	if opt.Discovery == nil || !opt.Discovery.Enabled {
		return nil
	}

	configMapID := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "coredns",
	}

	controller, err := controllers.NewHostsReconciler(mgr, opt, configMapID)
	if err != nil {
		return err
	}

	if err := controller.SetupWithManager(mgr); err != nil {
		return err
	}

	return nil
}

// Reconciler is the interface for a standard Reconciler.
type Reconciler interface {
	SetupWithManager(mgr manager.Manager) error
}

func setupCloudIPAM(ctx context.Context, mgr manager.Manager, opt *config.Options) error {
	setupLog.Info("enabling IPAM controller")
	var controller Reconciler
	switch opt.Cloud {
	case "aws":
		ipamController, err := controllers.NewAWSIPAMReconciler(ctx, mgr)
		if err != nil {
			return fmt.Errorf("creating aws IPAM controller: %w", err)
		}
		controller = ipamController
	case "gce":
		ipamController, err := controllers.NewGCEIPAMReconciler(mgr)
		if err != nil {
			return fmt.Errorf("creating gce IPAM controller: %w", err)
		}
		controller = ipamController
	default:
		return fmt.Errorf("kOps IPAM controller is not supported on cloud %q", opt.Cloud)
	}

	if err := controller.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("registering IPAM controller: %w", err)
	}

	return nil
}

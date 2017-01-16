/*
Copyright 2016 The Kubernetes Authors.

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
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/watchers"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/core/v1"
	client_extensions "k8s.io/kubernetes/pkg/client/clientset_generated/clientset/typed/extensions/v1beta1"
	kubectl_util "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"

	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)
)

func main() {
	dnsProviderId := "aws-route53"
	flags.StringVar(&dnsProviderId, "dns", dnsProviderId, "DNS provider we should use (aws-route53, google-clouddns)")

	var zones []string
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")

	watchIngress := true
	flags.BoolVar(&watchIngress, "watch-ingress", watchIngress, "Configure hostnames found in ingress resources")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)
	clientConfig := kubectl_util.DefaultClientConfig(flags)

	flags.Parse(os.Args)

	zoneRules, err := dns.ParseZoneRules(zones)
	if err != nil {
		glog.Errorf("unexpected zone flags: %q", err)
		os.Exit(1)
	}

	config, err := clientConfig.ClientConfig()
	if err != nil {
		glog.Errorf("error building client configuration: %v", err)
		os.Exit(1)
	}

	kubeClient, err := client.NewForConfig(config)
	if err != nil {
		glog.Fatalf("error building REST client: %v", err)
	}

	kubeExtensionsClient, err := client_extensions.NewForConfig(config)
	if err != nil {
		glog.Fatalf("error building extensions REST client: %v", err)
	}

	dnsProvider, err := dnsprovider.GetDnsProvider(dnsProviderId, nil)
	if err != nil {
		glog.Errorf("Error initializing DNS provider %q: %v", dnsProviderId, err)
		os.Exit(1)
	}
	if dnsProvider == nil {
		glog.Errorf("DNS provider was nil %q: %v", dnsProviderId, err)
		os.Exit(1)
	}

	dnsCache, err := dns.NewDNSCache(dnsProvider)
	if err != nil {
		glog.Errorf("Error initializing DNS cache: %v", err)
		os.Exit(1)
	}

	dnsController, err := dns.NewDNSController(dnsCache, zoneRules)
	if err != nil {
		glog.Errorf("Error building DNS controller: %v", err)
		os.Exit(1)
	}

	nodeController, err := watchers.NewNodeController(kubeClient, dnsController)
	if err != nil {
		glog.Errorf("Error building node controller: %v", err)
		os.Exit(1)
	}

	podController, err := watchers.NewPodController(kubeClient, dnsController)
	if err != nil {
		glog.Errorf("Error building pod controller: %v", err)
		os.Exit(1)
	}

	serviceController, err := watchers.NewServiceController(kubeClient, dnsController)
	if err != nil {
		glog.Errorf("Error building service controller: %v", err)
		os.Exit(1)
	}

	var ingressController *watchers.IngressController
	if watchIngress {
		ingressController, err = watchers.NewIngressController(kubeExtensionsClient, dnsController)
		if err != nil {
			glog.Errorf("Error building ingress controller: %v", err)
			os.Exit(1)
		}
	} else {
		glog.Infof("Ingress controller disabled")
	}

	go nodeController.Run()
	go podController.Run()
	go serviceController.Run()

	if ingressController != nil {
		go ingressController.Run()
	}

	dnsController.Run()
}

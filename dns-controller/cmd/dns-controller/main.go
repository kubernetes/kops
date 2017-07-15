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
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/watchers"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipdns "k8s.io/kops/protokube/pkg/gossip/dns"
	gossipdnsprovider "k8s.io/kops/protokube/pkg/gossip/dns/provider"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/kops/protokube/pkg/gossip/mesh"
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	k8scoredns "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/coredns"
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)

	// value overwritten during build. This can be used to resolve issues.
	BuildVersion = "0.1"
)

func main() {
	fmt.Printf("dns-controller version %s\n", BuildVersion)

	// Be sure to get the glog flags
	glog.Flush()

	dnsProviderId := "aws-route53"
	flags.StringVar(&dnsProviderId, "dns", dnsProviderId, "DNS provider we should use (aws-route53, google-clouddns, coredns, gossip)")

	gossipListen := "0.0.0.0:3998"
	flags.StringVar(&gossipListen, "gossip-listen", gossipListen, "The address on which to listen if gossip is enabled")

	var gossipSeeds []string
	flags.StringSliceVar(&gossipSeeds, "gossip-seed", gossipSeeds, "If set, will enable gossip zones and seed using the provided addresses")

	var gossipSecret string
	flags.StringVar(&gossipSecret, "gossip-secret", gossipSecret, "Secret to use to secure gossip")

	var zones []string
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")

	watchIngress := true
	flags.BoolVar(&watchIngress, "watch-ingress", watchIngress, "Configure hostnames found in ingress resources")

	dnsServer := ""
	flag.StringVar(&dnsServer, "dns-server", dnsServer, "DNS Server")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)

	flags.Parse(os.Args)

	zoneRules, err := dns.ParseZoneRules(zones)
	if err != nil {
		glog.Errorf("unexpected zone flags: %q", err)
		os.Exit(1)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("error building client configuration: %v", err)
		os.Exit(1)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("error building REST client: %v", err)
	}

	//kubeExtensionsClient, err := client_extensions.NewForConfig(config)
	//if err != nil {
	//	glog.Fatalf("error building extensions REST client: %v", err)
	//}

	var dnsProviders []dnsprovider.Interface
	if dnsProviderId != "gossip" {
		var file io.Reader
		if dnsProviderId == k8scoredns.ProviderName {
			var lines []string
			lines = append(lines, "etcd-endpoints = "+dnsServer)
			lines = append(lines, "zones = "+zones[0])
			config := "[global]\n" + strings.Join(lines, "\n") + "\n"
			file = bytes.NewReader([]byte(config))
		}
		dnsProvider, err := dnsprovider.GetDnsProvider(dnsProviderId, file)
		if err != nil {
			glog.Errorf("Error initializing DNS provider %q: %v", dnsProviderId, err)
			os.Exit(1)
		}
		if dnsProvider == nil {
			glog.Errorf("DNS provider was nil %q: %v", dnsProviderId, err)
			os.Exit(1)
		}
		dnsProviders = append(dnsProviders, dnsProvider)
	}

	if len(gossipSeeds) != 0 {
		gossipSeeds := gossip.NewStaticSeedProvider(gossipSeeds)

		id := os.Getenv("HOSTNAME")
		if id == "" {
			glog.Fatalf("Unable to fetch HOSTNAME for use as node identifier")
		}
		gossipName := "dns-controller." + id

		channelName := "dns"
		gossipState, err := mesh.NewMeshGossiper(gossipListen, channelName, gossipName, []byte(gossipSecret), gossipSeeds)
		if err != nil {
			glog.Errorf("Error initializing gossip: %v", err)
			os.Exit(1)
		}

		go func() {
			err := gossipState.Start()
			if err != nil {
				glog.Fatalf("gossip exited unexpectedly: %v", err)
			} else {
				glog.Fatalf("gossip exited unexpectedly, but without error")
			}
		}()

		dnsView := gossipdns.NewDNSView(gossipState)
		dnsProvider, err := gossipdnsprovider.New(dnsView)
		if err != nil {
			glog.Errorf("Error initializing gossip DNS provider: %v", err)
			os.Exit(1)
		}
		if dnsProvider == nil {
			glog.Errorf("Gossip DNS provider was nil: %v", err)
			os.Exit(1)
		}
		dnsProviders = append(dnsProviders, dnsProvider)
	}

	dnsController, err := dns.NewDNSController(dnsProviders, zoneRules)
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
		ingressController, err = watchers.NewIngressController(kubeClient, dnsController)
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

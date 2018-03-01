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

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/watchers"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/aws/route53"
	k8scoredns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/coredns"
	_ "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/google/clouddns"
	_ "k8s.io/kops/pkg/resources/digitalocean/dns"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipdns "k8s.io/kops/protokube/pkg/gossip/dns"
	gossipdnsprovider "k8s.io/kops/protokube/pkg/gossip/dns/provider"
	"k8s.io/kops/protokube/pkg/gossip/mesh"
)

var (
	flags        = pflag.NewFlagSet("", pflag.ExitOnError)
	BuildVersion = "0.1"
)

func main() {
	fmt.Printf("dns-controller version %s\n", BuildVersion)
	var dnsServer, dnsProviderID, gossipListen, gossipSecret, watchNamespace string
	var gossipSeeds, zones []string
	var watchIngress bool

	// Be sure to get the glog flags
	glog.Flush()

	flag.StringVar(&dnsServer, "dns-server", "", "DNS Server")
	flags.BoolVar(&watchIngress, "watch-ingress", true, "Configure hostnames found in ingress resources")
	flags.StringSliceVar(&gossipSeeds, "gossip-seed", gossipSeeds, "If set, will enable gossip zones and seed using the provided addresses")
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")
	flags.StringVar(&dnsProviderID, "dns", "aws-route53", "DNS provider we should use (aws-route53, google-clouddns, digitalocean, coredns, gossip)")
	flags.StringVar(&gossipListen, "gossip-listen", "0.0.0.0:3998", "The address on which to listen if gossip is enabled")
	flags.StringVar(&gossipSecret, "gossip-secret", gossipSecret, "Secret to use to secure gossip")
	flags.StringVar(&watchNamespace, "watch-namespace", "", "Limits the functionality for pods, services and ingress to specific namespace, by default all")
	flag.IntVar(&route53.MaxBatchSize, "route53-batch-size", route53.MaxBatchSize, "Maximum number of operations performed per changeset batch")

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

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Fatalf("error building REST client: %v", err)
	}

	var dnsProviders []dnsprovider.Interface
	if dnsProviderID != "gossip" {
		var file io.Reader
		if dnsProviderID == k8scoredns.ProviderName {
			var lines []string
			lines = append(lines, "etcd-endpoints = "+dnsServer)
			lines = append(lines, "zones = "+zones[0])
			config := "[global]\n" + strings.Join(lines, "\n") + "\n"
			file = bytes.NewReader([]byte(config))
		}
		dnsProvider, err := dnsprovider.GetDnsProvider(dnsProviderID, file)
		if err != nil {
			glog.Errorf("Error initializing DNS provider %q: %v", dnsProviderID, err)
			os.Exit(1)
		}
		if dnsProvider == nil {
			glog.Errorf("DNS provider was nil %q: %v", dnsProviderID, err)
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

	// @step: initialize the watchers
	if err := initializeWatchers(client, dnsController, watchNamespace, watchIngress); err != nil {
		glog.Errorf("%s", err)
		os.Exit(1)
	}

	// start and wait on the dns controller
	dnsController.Run()
}

// initializeWatchers is responsible for creating the watchers
func initializeWatchers(client kubernetes.Interface, dnsctl *dns.DNSController, namespace string, watchIngress bool) error {
	glog.V(1).Info("initializing the watch controllers, namespace: %q", namespace)

	nodeController, err := watchers.NewNodeController(client, dnsctl)
	if err != nil {
		return fmt.Errorf("failed to initialize the node controller, error: %v", err)
	}

	podController, err := watchers.NewPodController(client, dnsctl, namespace)
	if err != nil {
		return fmt.Errorf("failed to initialize the pod controller, error: %v", err)
	}

	serviceController, err := watchers.NewServiceController(client, dnsctl, namespace)
	if err != nil {
		return fmt.Errorf("failed to initialize the service controller, error: %v", err)
	}

	var ingressController *watchers.IngressController
	if watchIngress {
		ingressController, err = watchers.NewIngressController(client, dnsctl, namespace)
		if err != nil {
			return fmt.Errorf("failed to initialize the ingress controller, error: %v", err)
		}
	} else {
		glog.Infof("Ingress controller disabled")
	}

	go nodeController.Run()
	go podController.Run()
	go serviceController.Run()

	if watchIngress {
		go ingressController.Run()
	}

	return nil
}

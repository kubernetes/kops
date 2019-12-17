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
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"k8s.io/klog"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client metric registration

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/watchers"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider"
	"k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/aws/route53"
	k8scoredns "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/coredns"
	_ "k8s.io/kops/dnsprovider/pkg/dnsprovider/providers/google/clouddns"
	_ "k8s.io/kops/pkg/resources/digitalocean/dns"
	"k8s.io/kops/pkg/wellknownports"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipdns "k8s.io/kops/protokube/pkg/gossip/dns"
	gossipdnsprovider "k8s.io/kops/protokube/pkg/gossip/dns/provider"
	_ "k8s.io/kops/protokube/pkg/gossip/memberlist"
	_ "k8s.io/kops/protokube/pkg/gossip/mesh"
)

var (
	flags        = pflag.NewFlagSet("", pflag.ExitOnError)
	BuildVersion = "0.1"
)

func main() {
	fmt.Printf("dns-controller version %s\n", BuildVersion)
	var dnsServer, dnsProviderID, gossipListen, gossipSecret, watchNamespace, metricsListen, gossipProtocol, gossipSecretSecondary, gossipListenSecondary, gossipProtocolSecondary string
	var gossipSeeds, gossipSeedsSecondary, zones []string
	var watchIngress bool
	var updateInterval int

	// Be sure to get the glog flags
	klog.InitFlags(nil)
	klog.Flush()

	flag.StringVar(&dnsServer, "dns-server", "", "DNS Server")
	flags.BoolVar(&watchIngress, "watch-ingress", true, "Configure hostnames found in ingress resources")
	flags.StringSliceVar(&gossipSeeds, "gossip-seed", gossipSeeds, "If set, will enable gossip zones and seed using the provided addresses")
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")
	flags.StringVar(&dnsProviderID, "dns", "aws-route53", "DNS provider we should use (aws-route53, google-clouddns, digitalocean, coredns, gossip)")
	flag.StringVar(&gossipProtocol, "gossip-protocol", "mesh", "mesh/memberlist")
	flags.StringVar(&gossipListen, "gossip-listen", fmt.Sprintf("0.0.0.0:%d", wellknownports.DNSControllerGossipWeaveMesh), "The address on which to listen if gossip is enabled")
	flags.StringVar(&gossipSecret, "gossip-secret", gossipSecret, "Secret to use to secure gossip")
	flag.StringVar(&gossipProtocolSecondary, "gossip-protocol-secondary", "", "mesh/memberlist")
	flag.StringVar(&gossipListenSecondary, "gossip-listen-secondary", fmt.Sprintf("0.0.0.0:%d", wellknownports.DNSControllerGossipMemberlist), "address:port on which to bind for gossip")
	flags.StringVar(&gossipSecretSecondary, "gossip-secret-secondary", gossipSecret, "Secret to use to secure gossip")
	flags.StringSliceVar(&gossipSeedsSecondary, "gossip-seed-secondary", gossipSeedsSecondary, "If set, will enable gossip zones and seed using the provided addresses")
	flags.StringVar(&watchNamespace, "watch-namespace", "", "Limits the functionality for pods, services and ingress to specific namespace, by default all")
	flag.IntVar(&route53.MaxBatchSize, "route53-batch-size", route53.MaxBatchSize, "Maximum number of operations performed per changeset batch")
	flag.StringVar(&metricsListen, "metrics-listen", "", "The address on which to listen for Prometheus metrics.")
	flags.IntVar(&updateInterval, "update-interval", 5, "Configure interval at which to update DNS records.")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	if metricsListen != "" {
		go func() {
			http.Handle("/metrics", promhttp.Handler())
			log.Fatal(http.ListenAndServe(metricsListen, nil))
		}()
	}

	zoneRules, err := dns.ParseZoneRules(zones)
	if err != nil {
		klog.Errorf("unexpected zone flags: %q", err)
		os.Exit(1)
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Errorf("error building client configuration: %v", err)
		os.Exit(1)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("error building REST client: %v", err)
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
			klog.Errorf("Error initializing DNS provider %q: %v", dnsProviderID, err)
			os.Exit(1)
		}
		if dnsProvider == nil {
			klog.Errorf("DNS provider was nil %q: %v", dnsProviderID, err)
			os.Exit(1)
		}
		dnsProviders = append(dnsProviders, dnsProvider)
	}

	if len(gossipSeeds) != 0 {
		gossipSeeds := gossip.NewStaticSeedProvider(gossipSeeds)

		id := os.Getenv("HOSTNAME")
		if id == "" {
			klog.Fatalf("Unable to fetch HOSTNAME for use as node identifier")
		}
		gossipName := "dns-controller." + id

		channelName := "dns"
		var gossipState gossip.GossipState

		gossipState, err = gossip.GetGossipState(gossipProtocol, gossipListen, channelName, gossipName, []byte(gossipSecret), gossipSeeds)
		if err != nil {
			klog.Errorf("Error initializing gossip: %v", err)
			os.Exit(1)
		}

		if gossipProtocolSecondary != "" {

			secondaryGossipState, err := gossip.GetGossipState(gossipProtocolSecondary, gossipListenSecondary, channelName, gossipName, []byte(gossipSecretSecondary), gossip.NewStaticSeedProvider(gossipSeedsSecondary))
			if err != nil {
				klog.Errorf("Error initializing secondary gossip: %v", err)
				os.Exit(1)
			}

			gossipState = &gossip.MultiGossipState{
				Primary:   gossipState,
				Secondary: secondaryGossipState,
			}
		}
		go func() {
			err := gossipState.Start()
			if err != nil {
				klog.Fatalf("gossip exited unexpectedly: %v", err)
			} else {
				klog.Fatalf("gossip exited unexpectedly, but without error")
			}
		}()

		dnsView := gossipdns.NewDNSView(gossipState)
		dnsProvider, err := gossipdnsprovider.New(dnsView)
		if err != nil {
			klog.Errorf("Error initializing gossip DNS provider: %v", err)
			os.Exit(1)
		}
		if dnsProvider == nil {
			klog.Errorf("Gossip DNS provider was nil: %v", err)
			os.Exit(1)
		}
		dnsProviders = append(dnsProviders, dnsProvider)
	}

	dnsController, err := dns.NewDNSController(dnsProviders, zoneRules, updateInterval)
	if err != nil {
		klog.Errorf("Error building DNS controller: %v", err)
		os.Exit(1)
	}

	// @step: initialize the watchers
	if err := initializeWatchers(client, dnsController, watchNamespace, watchIngress); err != nil {
		klog.Errorf("%s", err)
		os.Exit(1)
	}

	// start and wait on the dns controller
	dnsController.Run()
}

// initializeWatchers is responsible for creating the watchers
func initializeWatchers(client kubernetes.Interface, dnsctl *dns.DNSController, namespace string, watchIngress bool) error {
	klog.V(1).Infof("initializing the watch controllers, namespace: %q", namespace)

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
		klog.Infof("Ingress controller disabled")
	}

	go nodeController.Run()
	go podController.Run()
	go serviceController.Run()

	if watchIngress {
		go ingressController.Run()
	}

	return nil
}

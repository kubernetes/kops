package main

import (
	"flag"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/dns-controller/pkg/watchers"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	client "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/core/v1"
	client_extensions "k8s.io/kubernetes/pkg/client/clientset_generated/release_1_3/typed/extensions/v1beta1"
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

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)
	clientConfig := kubectl_util.DefaultClientConfig(flags)

	flags.Parse(os.Args)

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

	dnsController, err := dns.NewDNSController(dnsProvider)
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

	ingressController, err := watchers.NewIngressController(kubeExtensionsClient, dnsController)
	if err != nil {
		glog.Errorf("Error building ingress controller: %v", err)
		os.Exit(1)
	}

	go nodeController.Run()
	go podController.Run()
	go serviceController.Run()
	go ingressController.Run()

	dnsController.Run()
}

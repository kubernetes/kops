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
	"fmt"
	"github.com/golang/glog"
	"github.com/spf13/pflag"
	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/protokube/pkg/protokube"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	"net"
	"os"
	"strings"

	// Load DNS plugins
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)

	// value overwritten during build. This can be used to resolve issues.
	BuildVersion = "0.1"
)

func main() {
	fmt.Printf("protokube version %s\n", BuildVersion)

	err := run()
	if err != nil {
		glog.Errorf("Error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	dnsProviderId := "aws-route53"
	flags.StringVar(&dnsProviderId, "dns", dnsProviderId, "DNS provider we should use (aws-route53, google-clouddns)")

	var zones []string
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")

	master := false
	flag.BoolVar(&master, "master", master, "Act as master")

	containerized := false
	flag.BoolVar(&containerized, "containerized", containerized, "Set if we are running containerized.")

	dnsInternalSuffix := ""
	flag.StringVar(&dnsInternalSuffix, "dns-internal-suffix", dnsInternalSuffix, "DNS suffix for internal domain names")

	clusterID := ""
	flag.StringVar(&clusterID, "cluster-id", clusterID, "Cluster ID")

	flagChannels := ""
	flag.StringVar(&flagChannels, "channels", flagChannels, "channels to install")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")

	flags.AddGoFlagSet(flag.CommandLine)

	flags.Parse(os.Args)

	volumes, err := protokube.NewAWSVolumes()
	if err != nil {
		return fmt.Errorf("Error initializing AWS: %q", err)
	}

	if clusterID == "" {
		clusterID = volumes.ClusterID()
		if clusterID == "" {
			return fmt.Errorf("cluster-id is required (cannot be determined from cloud)")
		} else {
			glog.Infof("Setting cluster-id from cloud: %s", clusterID)
		}
	}

	if dnsInternalSuffix == "" {
		// TODO: Maybe only master needs DNS?
		dnsInternalSuffix = ".internal." + clusterID
		glog.Infof("Setting dns-internal-suffix to %q", dnsInternalSuffix)
	}

	// Make sure it's actually a suffix (starts with .)
	if !strings.HasPrefix(dnsInternalSuffix, ".") {
		dnsInternalSuffix = "." + dnsInternalSuffix
	}

	// Get internal IP from cloud, to avoid problems if we're in a container
	// TODO: Just run with --net=host ??
	//internalIP, err := findInternalIP()
	//if err != nil {
	//	glog.Errorf("Error finding internal IP: %q", err)
	//	os.Exit(1)
	//}
	internalIP := volumes.InternalIP()

	var dnsScope dns.Scope
	var dnsController *dns.DNSController
	{
		dnsProvider, err := dnsprovider.GetDnsProvider(dnsProviderId, nil)
		if err != nil {
			return fmt.Errorf("Error initializing DNS provider %q: %v", dnsProviderId, err)
		}
		if dnsProvider == nil {
			return fmt.Errorf("DNS provider %q could not be initialized", dnsProviderId)
		}

		zoneRules, err := dns.ParseZoneRules(zones)
		if err != nil {
			return fmt.Errorf("unexpected zone flags: %q", err)
		}

		dnsCache, err := dns.NewDNSCache(dnsProvider)
		if err != nil {
			return err
		}

		dnsController, err = dns.NewDNSController(dnsCache, zoneRules)
		if err != nil {
			return err
		}

		dnsScope, err = dnsController.CreateScope("protokube")
		if err != nil {
			return err
		}

		// We don't really use readiness - our records are simple
		dnsScope.MarkReady()
	}

	rootfs := "/"
	if containerized {
		rootfs = "/rootfs/"
	}
	protokube.RootFS = rootfs
	protokube.Containerized = containerized

	modelDir := "model/etcd"

	var channels []string
	if flagChannels != "" {
		channels = strings.Split(flagChannels, ",")
	}

	k := &protokube.KubeBoot{
		Master:            master,
		InternalDNSSuffix: dnsInternalSuffix,
		InternalIP:        internalIP,
		//MasterID          : fromVolume
		//EtcdClusters   : fromVolume

		ModelDir: modelDir,
		DNSScope: dnsScope,

		Channels: channels,

		Kubernetes: protokube.NewKubernetesContext(),
	}
	k.Init(volumes)

	go dnsController.Run()

	k.RunSyncLoop()

	return fmt.Errorf("Unexpected exit")
}

// TODO: run with --net=host ??
func findInternalIP() (net.IP, error) {
	var ips []net.IP

	networkInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("error querying interfaces to determine internal ip: %v", err)
	}

	for i := range networkInterfaces {
		networkInterface := &networkInterfaces[i]
		flags := networkInterface.Flags
		name := networkInterface.Name

		if (flags & net.FlagLoopback) != 0 {
			glog.V(2).Infof("Ignoring interface %s - loopback", name)
			continue
		}

		// Not a lot else to go on...
		if !strings.HasPrefix(name, "eth") {
			glog.V(2).Infof("Ignoring interface %s - name does not look like ethernet device", name)
			continue
		}

		addrs, err := networkInterface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("error querying network interface %s for IP adddresses: %v", name, err)
		}

		for _, addr := range addrs {
			ip, _, err := net.ParseCIDR(addr.String())
			if err != nil {
				return nil, fmt.Errorf("error parsing address %s on network interface %s: %v", addr.String(), name, err)
			}

			if ip.IsLoopback() {
				glog.V(2).Infof("Ignoring address %s (loopback)", ip)
				continue
			}

			if ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
				glog.V(2).Infof("Ignoring address %s (link-local)", ip)
				continue
			}

			ips = append(ips, ip)
		}
	}

	if len(ips) == 0 {
		return nil, fmt.Errorf("unable to determine internal ip (no adddresses found)")
	}

	if len(ips) != 1 {
		glog.Warningf("Found multiple internal IPs; making arbitrary choice")
		for _, ip := range ips {
			glog.Warningf("\tip: %s", ip.String())
		}
	}
	return ips[0], nil
}

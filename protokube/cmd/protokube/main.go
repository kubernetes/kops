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
	"net"
	"os"
	"path"
	"strings"

	"k8s.io/kops/dns-controller/pkg/dns"
	"k8s.io/kops/protokube/pkg/gossip"
	gossipdns "k8s.io/kops/protokube/pkg/gossip/dns"
	"k8s.io/kops/protokube/pkg/gossip/mesh"
	"k8s.io/kops/protokube/pkg/protokube"
	"k8s.io/kubernetes/federation/pkg/dnsprovider"
	// Load DNS plugins
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/aws/route53"
	k8scoredns "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/coredns"
	_ "k8s.io/kubernetes/federation/pkg/dnsprovider/providers/google/clouddns"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)
	// BuildVersion is overwritten during build. This can be used to resolve issues.
	BuildVersion = "0.1"
)

func main() {
	fmt.Printf("protokube version %s\n", BuildVersion)

	if err := run(); err != nil {
		glog.Errorf("Error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

// run is responsible for running the protokube service controller
func run() error {
	var zones []string
	var applyTaints, initializeRBAC, containerized, master bool
	var cloud, clusterID, dnsServer, dnsProviderID, dnsInternalSuffix, gossipSecret, gossipListen string
	var flagChannels, tlsCert, tlsKey, tlsCA, peerCert, peerKey, peerCA, etcdImageSource string

	flag.BoolVar(&applyTaints, "apply-taints", applyTaints, "Apply taints to nodes based on the role")
	flag.BoolVar(&containerized, "containerized", containerized, "Set if we are running containerized.")
	flag.BoolVar(&initializeRBAC, "initialize-rbac", initializeRBAC, "Set if we should initialize RBAC")
	flag.BoolVar(&master, "master", master, "Whether or not this node is a master")
	flag.StringVar(&cloud, "cloud", "aws", "CloudProvider we are using (aws,gce)")
	flag.StringVar(&clusterID, "cluster-id", clusterID, "Cluster ID")
	flag.StringVar(&dnsInternalSuffix, "dns-internal-suffix", dnsInternalSuffix, "DNS suffix for internal domain names")
	flag.StringVar(&dnsServer, "dns-server", dnsServer, "DNS Server")
	flag.StringVar(&flagChannels, "channels", flagChannels, "channels to install")
	flag.StringVar(&gossipListen, "gossip-listen", "0.0.0.0:3999", "address:port on which to bind for gossip")
	flag.StringVar(&peerCA, "peer-ca", peerCA, "Path to a file containing the peer ca in PEM format")
	flag.StringVar(&peerCert, "peer-cert", peerCert, "Path to a file containing the peer certificate")
	flag.StringVar(&peerKey, "peer-key", peerKey, "Path to a file containing the private key for the peers")
	flag.StringVar(&tlsCA, "tls-ca", tlsCA, "Path to a file containing the ca for client certificates")
	flag.StringVar(&tlsCert, "tls-cert", tlsCert, "Path to a file containing the certificate for etcd server")
	flag.StringVar(&tlsKey, "tls-key", tlsKey, "Path to a file containing the private key for etcd server")
	flags.StringSliceVarP(&zones, "zone", "z", []string{}, "Configure permitted zones and their mappings")
	flags.StringVar(&dnsProviderID, "dns", "aws-route53", "DNS provider we should use (aws-route53, google-clouddns, coredns)")
	flags.StringVar(&etcdImageSource, "etcd-image-source", etcdImageSource, "Etcd Source Container Registry")
	flags.StringVar(&gossipSecret, "gossip-secret", gossipSecret, "Secret to use to secure gossip")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	var volumes protokube.Volumes
	var internalIP net.IP

	if cloud == "aws" {
		awsVolumes, err := protokube.NewAWSVolumes()
		if err != nil {
			glog.Errorf("Error initializing AWS: %q", err)
			os.Exit(1)
		}
		volumes = awsVolumes

		if clusterID == "" {
			clusterID = awsVolumes.ClusterID()
		}
		if internalIP == nil {
			internalIP = awsVolumes.InternalIP()
		}
	} else if cloud == "gce" {
		gceVolumes, err := protokube.NewGCEVolumes()
		if err != nil {
			glog.Errorf("Error initializing GCE: %q", err)
			os.Exit(1)
		}

		volumes = gceVolumes

		if clusterID == "" {
			clusterID = gceVolumes.ClusterID()
		}

		if internalIP == nil {
			internalIP = gceVolumes.InternalIP()
		}
	} else if cloud == "vsphere" {
		glog.Info("Initializing vSphere volumes")
		vsphereVolumes, err := protokube.NewVSphereVolumes()
		if err != nil {
			glog.Errorf("Error initializing vSphere: %q", err)
			os.Exit(1)
		}
		volumes = vsphereVolumes
		if internalIP == nil {
			internalIP = vsphereVolumes.InternalIp()
		}

	} else {
		glog.Errorf("Unknown cloud %q", cloud)
		os.Exit(1)
	}

	if clusterID == "" {
		if clusterID == "" {
			return fmt.Errorf("cluster-id is required (cannot be determined from cloud)")
		}
		glog.Infof("Setting cluster-id from cloud: %s", clusterID)
	}

	if internalIP == nil {
		glog.Errorf("Cannot determine internal IP")
		os.Exit(1)
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

	rootfs := "/"
	if containerized {
		rootfs = "/rootfs/"
	}

	protokube.RootFS = rootfs
	protokube.Containerized = containerized

	var dnsProvider protokube.DNSProvider

	if dnsProviderID == "gossip" {
		dnsTarget := &gossipdns.HostsFile{
			Path: path.Join(rootfs, "etc/hosts"),
		}

		var gossipSeeds gossip.SeedProvider
		var err error
		var gossipName string
		if cloud == "aws" {
			gossipSeeds, err = volumes.(*protokube.AWSVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.AWSVolumes).InstanceID()
		} else if cloud == "gce" {
			gossipSeeds, err = volumes.(*protokube.GCEVolumes).GossipSeeds()
			if err != nil {
				return err
			}
			gossipName = volumes.(*protokube.GCEVolumes).InstanceName()
		} else {
			glog.Fatalf("seed provider for %q not yet implemented", cloud)
		}

		id := os.Getenv("HOSTNAME")
		if id == "" {
			glog.Warningf("Unable to fetch HOSTNAME for use as node identifier")
		}

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
		go func() {
			gossipdns.RunDNSUpdates(dnsTarget, dnsView)
			glog.Fatalf("RunDNSUpdates exited unexpectedly")
		}()

		zoneInfo := gossipdns.DNSZoneInfo{
			Name: gossipdns.DefaultZoneName,
		}
		dnsProvider = &protokube.GossipDnsProvider{DNSView: dnsView, Zone: zoneInfo}
	} else {
		var dnsScope dns.Scope
		var dnsController *dns.DNSController
		{
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
				return fmt.Errorf("Error initializing DNS provider %q: %v", dnsProviderID, err)
			}
			if dnsProvider == nil {
				return fmt.Errorf("DNS provider %q could not be initialized", dnsProviderID)
			}

			zoneRules, err := dns.ParseZoneRules(zones)
			if err != nil {
				return fmt.Errorf("unexpected zone flags: %q", err)
			}

			dnsController, err = dns.NewDNSController([]dnsprovider.Interface{dnsProvider}, zoneRules)
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

		dnsProvider = &protokube.KopsDnsProvider{
			DNSScope:      dnsScope,
			DNSController: dnsController,
		}
	}
	modelDir := "model/etcd"

	var channels []string
	if flagChannels != "" {
		channels = strings.Split(flagChannels, ",")
	}

	k := &protokube.KubeBoot{
		ApplyTaints:       applyTaints,
		Channels:          channels,
		DNS:               dnsProvider,
		EtcdImageSource:   etcdImageSource,
		InitializeRBAC:    initializeRBAC,
		InternalDNSSuffix: dnsInternalSuffix,
		InternalIP:        internalIP,
		Kubernetes:        protokube.NewKubernetesContext(),
		Master:            master,
		ModelDir:          modelDir,
		PeerCA:            peerCA,
		PeerCert:          peerCert,
		PeerKey:           peerKey,
		TLSCA:             tlsCA,
		TLSCert:           tlsCert,
		TLSKey:            tlsKey,
	}

	k.Init(volumes)

	if dnsProvider != nil {
		go dnsProvider.Run()
	}

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

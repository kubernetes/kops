/*
Copyright 2017 The Kubernetes Authors.

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
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/spf13/pflag"
	"k8s.io/klog"
	"k8s.io/kops/protokube/pkg/gossip/dns/hosts"
)

var (
	flags = pflag.NewFlagSet("", pflag.ExitOnError)
	// BuildVersion is overwritten during build. This can be used to resolve issues.
	BuildVersion = "0.1"
)

func main() {
	fmt.Printf("kube-discovery version %s\n", BuildVersion)

	if err := run(); err != nil {
		klog.Errorf("unexpected error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

type Options struct {
	Containerized       bool
	ClusterID           string
	DnsDiscoveryTimeout time.Duration
	Interval            time.Duration
	Prefixes            []string
}

func (o *Options) InitDefaults() {
	o.DnsDiscoveryTimeout = 5 * time.Second
	o.Interval = 60 * time.Second
	o.Prefixes = []string{"api.internal."}
}

// run is responsible for running the protokube service controller
func run() error {
	var o Options
	o.InitDefaults()

	flags.BoolVar(&o.Containerized, "containerized", o.Containerized, "Set if we are running containerized.")
	flags.StringVar(&o.ClusterID, "cluster-id", o.ClusterID, "Cluster ID")

	// Trick to avoid 'logging before flag.Parse' warning
	flag.CommandLine.Parse([]string{})

	flag.Set("logtostderr", "true")
	flags.AddGoFlagSet(flag.CommandLine)
	flags.Parse(os.Args)

	if o.ClusterID == "" {
		klog.Infof("updating records for all discovered clusters")
	} else {
		klog.Infof("updating records for cluster %q", o.ClusterID)
	}

	c := &DiscoveryController{
		Options: o,
	}
	return c.Run()
}

type DiscoveryController struct {
	Options Options
}

func (c *DiscoveryController) Run() error {
	for {
		err := c.runOnce()
		if err != nil {
			klog.Warningf("error updating records: %v", err)
		}
		time.Sleep(c.Options.Interval)
	}
}

func (c *DiscoveryController) runOnce() error {
	o := &c.Options

	rootfs := "/"
	if o.Containerized {
		rootfs = "/rootfs/"
	}

	clusters, err := discoverKubernetesClusters(o.DnsDiscoveryTimeout)
	if err != nil {
		return fmt.Errorf("error from dns resolve: %v", err)
	}

	klog.Infof("clusters: %v", clusters)

	hostsPath := filepath.Join(rootfs, "etc/hosts")

	addrToHosts := make(map[string][]string)
	for k, addrs := range clusters {
		if o.ClusterID != "" {
			if k != o.ClusterID {
				klog.V(2).Infof("skipping discovered cluster %q as does not match configured %q", k, o.ClusterID)
				continue
			}
		}
		for _, addr := range addrs {
			addrString := addr.String()
			for _, prefix := range o.Prefixes {
				addrToHosts[addrString] = append(addrToHosts[addrString], prefix+k)
			}
		}
	}

	if len(addrToHosts) == 0 {
		// We don't update if there are no records remaining, just in case it is a transient blip
		klog.Warningf("no records found; skipping update")
		return nil
	}

	// TODO: Combined with previously discovered records (with a sliding window?)
	// TODO: Support an iptables / ipvs backend?
	// TODO: Verify resolved records against certificates?
	if err := hosts.UpdateHostsFileWithRecords(hostsPath, addrToHosts); err != nil {
		return fmt.Errorf("error updating hosts file: %v", err)
	}
	klog.Infof("updated %s", hostsPath)

	return nil
}

func discoverKubernetesClusters(timeout time.Duration) (map[string][]net.IP, error) {
	addr := &net.UDPAddr{IP: net.IPv4(224, 0, 0, 251), Port: 5353}
	connection, err := net.ListenMulticastUDP("udp4", nil, addr)
	if err != nil {
		return nil, fmt.Errorf("error listening for multicast: %v", err)
	}

	defer func() {
		err := connection.Close()
		if err != nil {
			klog.Warningf("error closing multicast connection: %v", err)
		}
	}()

	serviceName := "_kubernetes._tcp.local."

	{
		m := new(dns.Msg)
		m.SetQuestion(serviceName, dns.TypePTR)
		m.RecursionDesired = false
		buf, err := m.Pack()
		if err != nil {
			return nil, fmt.Errorf("error building DNS query: %v", err)
		}
		if _, err := connection.WriteToUDP(buf, addr); err != nil {
			return nil, fmt.Errorf("error sending DNS query: %v", err)
		}
	}

	stopAt := time.Now().Add(timeout)

	if err := connection.SetReadDeadline(stopAt); err != nil {
		return nil, fmt.Errorf("error setting socket read deadline: %v", err)
	}

	ptrs := make(map[string][]*dns.PTR)
	srvs := make(map[string][]*dns.SRV)
	txts := make(map[string][]*dns.TXT)
	aaaas := make(map[string][]*dns.AAAA)
	as := make(map[string][]*dns.A)

	buf := make([]byte, 65536)
	for {
		n, err := connection.Read(buf)
		if err != nil {
			if err, ok := err.(net.Error); ok && err.Timeout() {
				break
			}
			return nil, fmt.Errorf("error reading UDP response: %v", err)
		}
		msg := new(dns.Msg)
		if err := msg.Unpack(buf[:n]); err != nil {
			klog.Warningf("got unparsable DNS packet: %v", err)
			continue
		}
		klog.V(4).Infof("got response: %v", msg)

		for _, rr := range msg.Answer {
			switch rr := rr.(type) {
			case *dns.PTR:
				klog.V(4).Infof("PTR %v", rr)
				ptrs[rr.Hdr.Name] = append(ptrs[rr.Hdr.Name], rr)
			case *dns.TXT:
				klog.V(4).Infof("TXT %v", rr)
				txts[rr.Hdr.Name] = append(txts[rr.Hdr.Name], rr)
			case *dns.SRV:
				klog.V(4).Infof("SRV %v", rr)
				srvs[rr.Hdr.Name] = append(srvs[rr.Hdr.Name], rr)
			case *dns.AAAA:
				klog.V(4).Infof("AAAA %v", rr)
				aaaas[rr.Hdr.Name] = append(aaaas[rr.Hdr.Name], rr)
			case *dns.A:
				klog.V(4).Infof("A %v", rr)
				as[rr.Hdr.Name] = append(as[rr.Hdr.Name], rr)
			default:
				klog.V(2).Infof("ignoring answer of unknown type %T: %v", rr, rr)
			}
		}
	}

	addrs := make(map[string][]net.IP)

	for _, ptr := range ptrs[serviceName] {
		instance := strings.TrimSuffix(ptr.Ptr, serviceName)
		instance = strings.TrimSuffix(instance, ".")

		// Dots in the instance name are escaped
		instance = strings.Replace(instance, "\\.", ".", -1)

		for _, srv := range srvs[ptr.Ptr] {
			// TODO: Ignore if port is not 443?
			for _, a := range as[srv.Target] {
				ensureInMap(addrs, instance, a.A)
			}
			for _, aaaa := range aaaas[srv.Target] {
				ensureInMap(addrs, instance, aaaa.AAAA)
			}
		}
	}

	return addrs, nil
}

func ensureInMap(addrs map[string][]net.IP, name string, ip net.IP) {
	// Avoid duplicates
	for _, existing := range addrs[name] {
		if ip.Equal(existing) {
			return
		}
	}
	addrs[name] = append(addrs[name], ip)
}

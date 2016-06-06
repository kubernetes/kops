package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"k8s.io/kube-deploy/protokube/pkg/protokube"
	"net"
	"os"
	"strings"
)

func main() {
	master := false
	flag.BoolVar(&master, "master", master, "Act as master")

	containerized := false
	flag.BoolVar(&containerized, "containerized", containerized, "Set if we are running containerized.")

	dnsZoneName := ""
	flag.StringVar(&dnsZoneName, "dns-zone-name", dnsZoneName, "Name of zone to use for DNS")

	dnsInternalSuffix := ""
	flag.StringVar(&dnsInternalSuffix, "dns-internal-suffix", dnsInternalSuffix, "DNS suffix for internal domain names")

	clusterID := ""
	flag.StringVar(&clusterID, "cluster-id", clusterID, "Cluster ID")

	flag.Set("logtostderr", "true")
	flag.Parse()

	volumes, err := protokube.NewAWSVolumes()
	if err != nil {
		glog.Errorf("Error initializing AWS: %q", err)
		os.Exit(1)
	}

	if clusterID == "" {
		clusterID = volumes.ClusterID()
		if clusterID == "" {
			glog.Errorf("cluster-id is required (cannot be determined from cloud)")
			os.Exit(1)
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

	if dnsZoneName == "" {
		tokens := strings.Split(dnsInternalSuffix, ".")
		dnsZoneName = strings.Join(tokens[len(tokens)-2:], ".")
	}

	// Get internal IP from cloud, to avoid problems if we're in a container
	// TODO: Just run with --net=host ??
	//internalIP, err := findInternalIP()
	//if err != nil {
	//	glog.Errorf("Error finding internal IP: %q", err)
	//	os.Exit(1)
	//}
	internalIP := volumes.InternalIP()

	dns, err := protokube.NewRoute53DNSProvider(dnsZoneName)
	if err != nil {
		glog.Errorf("Error initializing DNS: %q", err)
		os.Exit(1)
	}

	rootfs := "/"
	if containerized {
		rootfs = "/rootfs/"
	}
	k := &protokube.KubeBoot{
		Containerized: containerized,
		RootFS:        rootfs,

		Master:            master,
		InternalDNSSuffix: dnsInternalSuffix,
		InternalIP:        internalIP,
		//MasterID          : fromVolume
		//EtcdClusters   : fromVolume

		Volumes: volumes,
		DNS:     dns,
	}

	err = k.Bootstrap()
	if err != nil {
		glog.Errorf("Error during bootstrap: %q", err)
		os.Exit(1)
	}

	glog.Infof("Bootstrap complete; applying configuration")
	err = k.ApplyModel()
	if err != nil {
		glog.Errorf("Error during configuration: %q", err)
		os.Exit(1)
	}

	glog.Infof("Bootstrap complete; starting kubelet")
	err = k.RunBootstrapTasks()
	if err != nil {
		glog.Errorf("Error during bootstrap: %q", err)
		os.Exit(1)
	}

	glog.Infof("Unexpected exited from kubelet run")
	os.Exit(1)
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

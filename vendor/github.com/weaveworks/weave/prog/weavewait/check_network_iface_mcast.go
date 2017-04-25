// +build iface,mcast

package main

import (
	weavenet "github.com/weaveworks/weave/net"
)

func checkNetwork() error {
	_, err := weavenet.EnsureInterfaceAndMcastRoute(weavenet.VethName)
	return err
}

package main

import (
	weavenet "github.com/weaveworks/weave/net"
)

func exposeNAT(args []string) error {
	if len(args) < 1 {
		cmdUsage("expose-nat", "<cidr>...")
	}

	cidrs, err := parseCIDRs(args)
	if err != nil {
		return err
	}

	for _, cidr := range cidrs {
		if err := weavenet.ExposeNAT(*cidr); err != nil {
			return err
		}
	}
	return nil
}

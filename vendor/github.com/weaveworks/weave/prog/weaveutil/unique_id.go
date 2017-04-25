package main

import (
	"fmt"

	weavenet "github.com/weaveworks/weave/net"
)

func uniqueID(args []string) error {
	if len(args) != 2 {
		cmdUsage("unique-id", "<db-prefix> <host-root>")
	}
	dbPrefix := args[0]
	hostRoot := args[1]
	uid, err := weavenet.GetSystemPeerName(dbPrefix, hostRoot)
	if err != nil {
		return err
	}
	fmt.Print(uid)
	return nil
}

/* weaveutil: collection of operations required by weave script */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	weavenet "github.com/weaveworks/weave/net"
)

var commands map[string]func([]string) error

func init() {
	commands = map[string]func([]string) error{
		"help":                     help,
		"netcheck":                 netcheck,
		"docker-tls-args":          dockerTLSArgs,
		"create-datapath":          createDatapath,
		"delete-datapath":          deleteDatapath,
		"check-datapath":           checkDatapath,
		"add-datapath-interface":   addDatapathInterface,
		"create-plugin-network":    createPluginNetwork,
		"remove-plugin-network":    removePluginNetwork,
		"container-addrs":          containerAddrs,
		"process-addrs":            processAddrs,
		"attach-container":         attach,
		"detach-container":         detach,
		"configure-arp":            configureARP,
		"check-iface":              checkIface,
		"del-iface":                delIface,
		"setup-iface":              setupIface,
		"list-netdevs":             listNetDevs,
		"cni-net":                  cniNet,
		"cni-ipam":                 cniIPAM,
		"expose-nat":               exposeNAT,
		"bridge-ip":                bridgeIP,
		"unique-id":                uniqueID,
		"swarm-manager-peers":      swarmManagerPeers,
		"is-swarm-manager":         isSwarmManager,
		"is-docker-plugin-enabled": isDockerPluginEnabled,
	}
}

func main() {
	// force re-exec of this binary
	selfPath, err := filepath.EvalSymlinks("/proc/self/exe")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	weavenet.WeaveUtilCmd = selfPath

	// If no args passed, act as CNI plugin based on executable name
	switch {
	case len(os.Args) == 1 && strings.HasSuffix(os.Args[0], "weave-ipam"):
		cniIPAM(os.Args)
		os.Exit(0)
	case len(os.Args) == 1 && strings.HasSuffix(os.Args[0], "weave-net"):
		cniNet(os.Args)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd, found := commands[os.Args[1]]
	if !found {
		fmt.Fprintf(os.Stderr, "%q cmd is not found\n", os.Args[1])
		usage()
		os.Exit(1)
	}
	if err := cmd(os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func help(args []string) error {
	if len(args) > 0 {
		cmdUsage("help", "")
	}
	usage()
	return nil
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: weaveutil <command> <arg>...")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "where <command> is one of:")
	fmt.Fprintln(os.Stderr)
	for cmd := range commands {
		fmt.Fprintln(os.Stderr, cmd)
	}
}

func cmdUsage(cmd string, usage string) {
	fmt.Fprintf(os.Stderr, "usage: weaveutil %s %s\n", cmd, usage)
	os.Exit(1)
}

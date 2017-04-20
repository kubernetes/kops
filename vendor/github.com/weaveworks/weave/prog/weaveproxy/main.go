package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/common/mflag"
	"github.com/weaveworks/common/mflagext"
	"github.com/weaveworks/common/signals"
	"github.com/weaveworks/weave/common"
	"github.com/weaveworks/weave/proxy"
)

var version = "unreleased"

var Log = common.Log

func getenv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func main() {
	var (
		justVersion bool
		logLevel    = "info"
		c           proxy.Config
		withDNS     bool
	)

	c.Version = version

	mflag.BoolVar(&justVersion, []string{"#version", "-version"}, false, "print version and exit")
	mflag.StringVar(&logLevel, []string{"-log-level"}, "info", "logging level (debug, info, warning, error)")
	mflagext.ListVar(&c.ListenAddrs, []string{"H"}, nil, "addresses on which to listen")
	mflag.StringVar(&c.HostnameFromLabel, []string{"-hostname-from-label"}, "", "Key of container label from which to obtain the container's hostname")
	mflag.StringVar(&c.HostnameMatch, []string{"-hostname-match"}, "(.*)", "Regexp pattern to apply on container names (e.g. '^aws-[0-9]+-(.*)$')")
	mflag.StringVar(&c.HostnameReplacement, []string{"-hostname-replacement"}, "$1", "Expression to generate hostnames based on matches from --hostname-match (e.g. 'my-app-$1')")
	mflag.BoolVar(&c.RewriteInspect, []string{"-rewrite-inspect"}, false, "Rewrite 'inspect' calls to return the weave network settings (if attached)")
	mflag.BoolVar(&c.NoDefaultIPAM, []string{"#-no-default-ipam", "-no-default-ipalloc"}, false, "do not automatically allocate addresses for containers without a WEAVE_CIDR")
	mflag.BoolVar(&c.NoRewriteHosts, []string{"-no-rewrite-hosts"}, false, "do not automatically rewrite /etc/hosts. Use if you need the docker IP to remain in /etc/hosts")
	mflag.StringVar(&c.TLSConfig.CACert, []string{"#tlscacert", "-tlscacert"}, "", "Trust certs signed only by this CA")
	mflag.StringVar(&c.TLSConfig.Cert, []string{"#tlscert", "-tlscert"}, "", "Path to TLS certificate file")
	mflag.BoolVar(&c.TLSConfig.Enabled, []string{"#tls", "-tls"}, false, "Use TLS; implied by --tlsverify")
	mflag.StringVar(&c.TLSConfig.Key, []string{"#tlskey", "-tlskey"}, "", "Path to TLS key file")
	mflag.BoolVar(&c.TLSConfig.Verify, []string{"#tlsverify", "-tlsverify"}, false, "Use TLS and verify the remote")
	mflag.BoolVar(&withDNS, []string{"#-with-dns", "#w"}, false, "option removed")
	mflag.BoolVar(&c.WithoutDNS, []string{"-without-dns"}, false, "instruct created containers to never use weaveDNS as their nameserver")
	mflag.BoolVar(&c.NoMulticastRoute, []string{"-no-multicast-route"}, false, "do not add a multicast route via the weave interface when attaching containers")
	mflag.Parse()

	if justVersion {
		fmt.Printf("weave proxy  %s\n", version)
		os.Exit(0)
	}

	common.SetLogLevel(logLevel)

	Log.Infoln("weave proxy", version)
	Log.Infoln("Command line arguments:", strings.Join(os.Args[1:], " "))

	if withDNS {
		Log.Warning("--with-dns option has been removed; DNS is on by default")
	}

	c.Image = getenv("EXEC_IMAGE", "weaveworks/weaveexec")
	c.DockerBridge = getenv("DOCKER_BRIDGE", "docker0")
	c.DockerHost = getenv("DOCKER_HOST", "unix:///var/run/docker.sock")

	p, err := proxy.NewProxy(c)
	if err != nil {
		Log.Fatalf("Could not start proxy: %s", err)
	}
	defer p.Stop()

	listeners := p.Listen()
	p.AttachExistingContainers()
	go p.Serve(listeners)
	go p.ListenAndServeStatus("/home/weave/status.sock")
	signals.SignalHandlerLoop(common.Log)
}
